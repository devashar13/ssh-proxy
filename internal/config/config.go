package config

import (
	"fmt"
	"os"
	"gopkg.in/yaml.v3"
)


type Config struct {
	// Upstream 
	Upstream struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Auth     struct {
			Type     string `yaml:"type"` 
			Password string `yaml:"password,omitempty"`
			KeyPath  string `yaml:"key_path,omitempty"`
		} `yaml:"auth"`
	} `yaml:"upstream"`

	// Users allowed 
	Users []struct {
		Username string `yaml:"username"`
		Auth     struct {
			Type     string `yaml:"type,omitempty"`
			Password string `yaml:"password,omitempty"`
			KeyPath  string `yaml:"key_path,omitempty"`
		} `yaml:"auth"`
	} `yaml:"users"`

	// Logging configuration
	Logging struct {
		Directory string `yaml:"directory"`
	} `yaml:"logging"`

	// LLM
	LLM struct {
		Enabled  bool   `yaml:"enabled"`
		APIKey   string `yaml:"api_key"`
		Provider string `yaml:"provider"`
		Model    string `yaml:"model"`
	} `yaml:"llm"`

	// ssh server config
	Server struct {
		HostKeyPath string `yaml:"host_key_path"`
		Port        int    `yaml:"port"`
	} `yaml:"server"`
}

func LoadYAML(path string) (*Config,error){
	data,err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err) 
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	// TODO: add default port

	if err := validate(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}


func validate(cfg *Config) error{
	if cfg.Upstream.Host == "" {
		return fmt.Errorf("upstream host not specified")
	}
	if cfg.Upstream.Port <= 0 {
		return fmt.Errorf("invalid upstream port: %d", cfg.Upstream.Port)
	}
	if cfg.Upstream.Username == "" {
		return fmt.Errorf("upstream username not specified")
	}
	switch cfg.Upstream.Auth.Type {
	case "password":
		if cfg.Upstream.Auth.Password == "" {
			return fmt.Errorf("upstream password not specified")
		}
	case "publickey":
		if cfg.Upstream.Auth.KeyPath == "" {
			return fmt.Errorf("upstream key path not specified")
		}
	default:
		return fmt.Errorf("invalid upstream auth type: %s", cfg.Upstream.Auth.Type)
	}
	if len(cfg.Users) == 0 {
		return fmt.Errorf("no users specified")
	}
	if cfg.Logging.Directory == "" {
		return fmt.Errorf("log directory not specified")
	}
	return nil
}

