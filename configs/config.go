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

}
