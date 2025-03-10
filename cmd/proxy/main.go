package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/devashar13/ssh-proxy/internal/config"
)


func main(){ 
	configPath := flag.String("config", "configs/config.yaml", "Path to configuration file")
	flag.Parse()
	cfg, err := config.LoadYAML(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
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

}
