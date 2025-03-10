package proxy

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"golang.org/x/crypto/ssh"

	"github.com/devashar13/ssh-proxy/internal/config"
)

type Server struct {
	config     *config.Config
	sshConfig  *ssh.ServerConfig
	listener   net.Listener
	shutdownWg sync.WaitGroup
	running    bool
	mu         sync.Mutex
}

func NewServer(cfg *config.Config) (*Server, error) {
	server := &Server{
		config: cfg,
	}


	sshConfig := &ssh.ServerConfig{
	
		PasswordCallback:  server.handlePasswordAuth,
		PublicKeyCallback: server.handlePublicKeyAuth,
	}


	hostKey, err := server.loadOrGenerateHostKey(cfg.Server.HostKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load/generate host key: %w", err)
	}
	sshConfig.AddHostKey(hostKey)

	server.sshConfig = sshConfig
	return server, nil
}

func (s *Server) ListenAndServe() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("server already running")
	}
	s.running = true
	s.mu.Unlock()

	addr := fmt.Sprintf(":%d", s.config.Server.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = listener

	log.Printf("SSH proxy server listening on %s", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
		
			s.mu.Lock()
			if !s.running {
				s.mu.Unlock()
				return nil
			}
			s.mu.Unlock()
			
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

	
		s.shutdownWg.Add(1)
		go func() {
			defer s.shutdownWg.Done()
			if err := s.handleConnection(conn); err != nil {
				log.Printf("Error handling connection: %v", err)
			}
		}()
	}
}

func (s *Server) Shutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.running = false
	if s.listener != nil {
		s.listener.Close()
	}


	s.shutdownWg.Wait()
	log.Println("SSH proxy server shutdown complete")
}

func (s *Server) handleConnection(conn net.Conn) error {
    defer conn.Close()
    
    log.Printf("New connection from %s", conn.RemoteAddr())


    sshConn, chans, reqs, err := ssh.NewServerConn(conn, s.sshConfig)
    if err != nil {
        return fmt.Errorf("failed to handshake: %w", err)
    }
    defer sshConn.Close()

    log.Printf("User %s authenticated from %s", sshConn.User(), conn.RemoteAddr())


    go ssh.DiscardRequests(reqs)


    for newChannel := range chans {
    
        if newChannel.ChannelType() != "session" {
            newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
            continue
        }

    
        channel, requests, err := newChannel.Accept()
        if err != nil {
            return fmt.Errorf("failed to accept channel: %w", err)
        }

    
        session, err := NewSession(s.config, sshConn.User(), channel, requests)
        if err != nil {
            log.Printf("Failed to create session: %v", err)
        
            fmt.Fprintf(channel, "Error: %v\r\n", err)
            channel.Close()
            continue
        }

    
        go func() {
            if err := session.Start(); err != nil {
                log.Printf("Session error: %v", err)
            
                fmt.Fprintf(channel, "Error: %v\r\n", err)
            }
        }()
    }
    
    return nil
}
func (s *Server) handlePasswordAuth(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	username := conn.User()
	

	for _, user := range s.config.Users {
		if user.Username == username && user.Auth.Type == "password" {
			if user.Auth.Password == string(password) {
				return &ssh.Permissions{
				
					Extensions: map[string]string{
						"username": username,
					},
				}, nil
			}
		}
	}

	log.Printf("Failed password auth attempt for user %s from %s", username, conn.RemoteAddr())
	return nil, fmt.Errorf("authentication failed")
}

func (s *Server) handlePublicKeyAuth(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	username := conn.User()
	

	for _, user := range s.config.Users {
		if user.Username == username && user.Auth.Type == "publickey" {
		
			authorizedKeysBytes, err := os.ReadFile(user.Auth.KeyPath)
			if err != nil {
				log.Printf("Failed to read authorized keys for %s: %v", username, err)
				return nil, err
			}

		
			parsedKey, _, _, _, err := ssh.ParseAuthorizedKey(authorizedKeysBytes)
			if err != nil {
				log.Printf("Failed to parse authorized key for %s: %v", username, err)
				return nil, err
			}

		
			if KeysEqual(parsedKey, key) {
				return &ssh.Permissions{
					Extensions: map[string]string{
						"username": username,
					},
				}, nil
			}
		}
	}

	log.Printf("Failed public key auth attempt for user %s from %s", username, conn.RemoteAddr())
	return nil, fmt.Errorf("authentication failed")
}

func (s *Server) loadOrGenerateHostKey(path string) (ssh.Signer, error) {

	if _, err := os.Stat(path); err == nil {
		return loadHostKey(path)
	}


	log.Printf("Generating new ED25519 host key at %s", path)
	
	key, err := generateED25519Key()
	if err != nil {
		return nil, fmt.Errorf("failed to generate host key: %w", err)
	}


	if err := saveHostKey(key, path); err != nil {
		return nil, fmt.Errorf("failed to save host key: %w", err)
	}

	return key, nil
}
