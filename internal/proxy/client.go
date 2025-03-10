package proxy

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/ssh"

	"github.com/devashar13/ssh-proxy/internal/config"
)


type UpstreamClient struct {
	config *config.Config
	client *ssh.Client
}


// Create client instance
func NewUpstreamClient(cfg *config.Config) *UpstreamClient {
	return &UpstreamClient{
		config: cfg,
	}
}


func (c *UpstreamClient) Connect() error {
	clientConfig, err := c.createClientConfig()
	if err != nil {
		return fmt.Errorf("failed to create client config: %w", err)
	}
	addr := fmt.Sprintf("%s:%d", c.config.Upstream.Host, c.config.Upstream.Port)
	client, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to upstream server: %w", err)
	}
	c.client = client
	log.Printf("Connected to upstream server %s", addr)
	return nil
}

func (c *UpstreamClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}


// Create session
func (c *UpstreamClient) NewSession() (*ssh.Session, error) {
	if c.client == nil {
		return nil, fmt.Errorf("not connected to upstream server")
	}
	return c.client.NewSession()
}


func (c *UpstreamClient) GetClient() *ssh.Client {
	return c.client
}

func (c *UpstreamClient) createClientConfig() (*ssh.ClientConfig, error) {
	authMethod, err := c.getAuthMethod()
	if err != nil {
		return nil, err
	}

	return &ssh.ClientConfig{
		User: c.config.Upstream.Username,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
		// TODO:  use proper host key verification
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, nil
}

func (c *UpstreamClient) getAuthMethod() (ssh.AuthMethod, error) {
	switch c.config.Upstream.Auth.Type {
	case "password":
		return ssh.Password(c.config.Upstream.Auth.Password), nil
	case "publickey":
		keyPath := c.config.Upstream.Auth.KeyPath
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}

		return ssh.PublicKeys(signer), nil
	default:
		return nil, fmt.Errorf("unsupported authentication type: %s", c.config.Upstream.Auth.Type)
	}
}
