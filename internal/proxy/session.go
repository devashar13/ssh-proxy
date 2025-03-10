package proxy

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/devashar13/ssh-proxy/internal/config"
	"github.com/devashar13/ssh-proxy/internal/llm"
)

type Session struct {
	config        *config.Config
	username      string
	clientChannel ssh.Channel
	clientReqs    <-chan *ssh.Request
	upstreamConn  *ssh.Client
	logFile       *os.File
	ptyWidth      uint32
	ptyHeight     uint32
	mu            sync.Mutex
}

func NewSession(cfg *config.Config, username string, clientChannel ssh.Channel, clientReqs <-chan *ssh.Request) (*Session, error) {
	logFile, err := createLogFile(cfg.Logging.Directory, username)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}
	upstreamClient := NewUpstreamClient(cfg)
	if err := upstreamClient.Connect(); err != nil {
		logFile.Close()
		return nil, fmt.Errorf("failed to connect to upstream server: %w", err)
	}
	return &Session{
		config:        cfg,
		username:      username,
		clientChannel: clientChannel,
		clientReqs:    clientReqs,
		upstreamConn:  upstreamClient.GetClient(),
		logFile:       logFile,
	}, nil
}

func (s *Session) Start() error {
	defer s.clientChannel.Close()
	defer s.upstreamConn.Close()
	
	// Get the logfile path before closing it
	logFilePath := s.logFile.Name()
	
	// Use defer with a function to ensure logFile is closed before summarization
	defer func() {
		s.logFile.Close()
		
		// If LLM is enabled, summarize the session
		if s.config.LLM.Enabled && s.config.LLM.APIKey != "" {
			log.Printf("Initiating security summarization for session: %s", filepath.Base(logFilePath))
			summarizer := llm.NewSummarizer(s.config)
			summarizer.SummarizeSessionAsync(logFilePath)
		}
	}()
	
	log.Printf("Starting session for user %s", s.username)
	upstreamChannel, upstreamReqs, err := s.upstreamConn.OpenChannel("session", nil)
	if err != nil {
		return fmt.Errorf("failed to open upstream channel: %w", err)
	}
	defer upstreamChannel.Close()
	go s.forwardRequests(upstreamChannel)
	stdinReader := io.TeeReader(s.clientChannel, s.logFile)
	errCh := make(chan error, 2)
	go func() {
		_, err := io.Copy(upstreamChannel, stdinReader)
		errCh <- err
	}()
	go func() {
		_, err := io.Copy(s.clientChannel, upstreamChannel)
		errCh <- err
	}()

	go ssh.DiscardRequests(upstreamReqs)
	err = <-errCh
	if err != nil && err != io.EOF {
		return fmt.Errorf("data forwarding error: %w", err)
	}

	log.Printf("Session ended for user %s", s.username)
	return nil
}

func (s *Session) forwardRequests(upstreamChannel ssh.Channel) {
	for req := range s.clientReqs {
		log.Printf("Forwarding request: %s", req.Type)
		if req.Type == "window-change" {
			s.handleWindowChange(req)
		}

		if req.Type == "pty-req" {
			s.handlePtyReq(req)
		}

		if req.Type == "exec" {
			s.logExecRequest(req)
		}

		ok, err := upstreamChannel.SendRequest(req.Type, req.WantReply, req.Payload)
		if err != nil {
			log.Printf("Failed to forward request: %v", err)
			if req.WantReply {
				req.Reply(false, nil)
			}
			continue
		}

		if req.WantReply {
			req.Reply(ok, nil)
		}
	}
}

func (s *Session) handleWindowChange(req *ssh.Request) {
	var params struct {
		Width  uint32
		Height uint32
		WidthPx  uint32
		HeightPx uint32
	}
	
	if err := ssh.Unmarshal(req.Payload, &params); err != nil {
		log.Printf("Failed to parse window-change payload: %v", err)
		return
	}

	s.mu.Lock()
	s.ptyWidth = params.Width
	s.ptyHeight = params.Height
	s.mu.Unlock()

	log.Printf("Window size changed to %dx%d", params.Width, params.Height)
}

func (s *Session) handlePtyReq(req *ssh.Request) {
	var params struct {
		Term     string
		Width    uint32
		Height   uint32
		WidthPx  uint32
		HeightPx uint32
		Modes    string
	}
	
	if err := ssh.Unmarshal(req.Payload, &params); err != nil {
		log.Printf("Failed to parse pty-req payload: %v", err)
		return
	}

	s.mu.Lock()
	s.ptyWidth = params.Width
	s.ptyHeight = params.Height
	s.mu.Unlock()

	log.Printf("PTY requested with term=%s, size=%dx%d", params.Term, params.Width, params.Height)
}

func (s *Session) logExecRequest(req *ssh.Request) {
	var params struct {
		Command string
	}
	
	if err := ssh.Unmarshal(req.Payload, &params); err != nil {
		log.Printf("Failed to parse exec payload: %v", err)
		return
	}

	log.Printf("Exec requested: %s", params.Command)
	fmt.Fprintf(s.logFile, "$ %s\n", params.Command)
}


func createLogFile(directory string, username string) (*os.File, error) {
	if err := os.MkdirAll(directory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s_%s.log", username, timestamp)
	path := filepath.Join(directory, filename)

	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	fmt.Fprintf(file, "--- SSH Session Log for %s ---\n", username)
	fmt.Fprintf(file, "Started: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(file, "------------------------------\n\n")

	log.Printf("Created log file: %s", path)
	return file, nil
}
