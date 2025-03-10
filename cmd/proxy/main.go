// In cmd/proxy/main.go

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/devashar13/ssh-proxy/internal/config"
	"github.com/devashar13/ssh-proxy/internal/proxy"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadYAML(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Print configuration summary
	fmt.Println("Configuration loaded")
	fmt.Println("Server Settings:")
	fmt.Printf("  - Port: %d\n", cfg.Server.Port)
	fmt.Printf("  - Host Key Path: %s\n", cfg.Server.HostKeyPath)
	
	fmt.Println("\nUpstream Server:")
	fmt.Printf("  - Host: %s\n", cfg.Upstream.Host)
	fmt.Printf("  - Port: %d\n", cfg.Upstream.Port)
	fmt.Printf("  - Username: %s\n", cfg.Upstream.Username)
	fmt.Printf("  - Auth Type: %s\n", cfg.Upstream.Auth.Type)
	
	fmt.Println("\nConfigured Users:")
	for i, user := range cfg.Users {
		fmt.Printf("  User #%d: %s (Auth: %s)\n", i+1, user.Username, user.Auth.Type)
	}
	
	fmt.Println("\nLogging:")
	fmt.Printf("  - Directory: %s\n", cfg.Logging.Directory)

	// Create and start SSH proxy server
	server, err := proxy.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Handle graceful shutdown
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		<-signalCh
		fmt.Println("\nShutting down SSH proxy server...")
		server.Shutdown()
	}()

	// Start the server (this will block until server is shut down)
	fmt.Println("\nStarting SSH proxy server...")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
