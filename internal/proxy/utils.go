package proxy

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"bytes"

	"golang.org/x/crypto/ssh"
)

func loadHostKey(path string) (ssh.Signer, error) {
	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read host key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse host key: %w", err)
	}

	return signer, nil
}

func generateED25519Key() (ssh.Signer, error) {
	_, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ED25519 key: %w", err)
	}

	// Convert to SSH key
	signer, err := ssh.NewSignerFromKey(privKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer from key: %w", err)
	}

	return signer, nil
}


func saveHostKey(signer ssh.Signer, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory for host key: %w", err)
	}

	keyBytes := []byte(signer.PublicKey().Type() + " private key")
	
	
	pemBlock := &pem.Block{
		Type:  "OPENSSH PRIVATE KEY",
		Bytes: keyBytes,
	}
	
	
	keyData := pem.EncodeToMemory(pemBlock)
	if keyData == nil {
		return fmt.Errorf("failed to encode private key to PEM format")
	}
	
	
	if err := os.WriteFile(path, keyData, 0600); err != nil {
		return fmt.Errorf("failed to write host key to file: %w", err)
	}
	
	return nil
}

func KeysEqual(a, b ssh.PublicKey) bool {
	return bytes.Equal(a.Marshal(), b.Marshal())
}
