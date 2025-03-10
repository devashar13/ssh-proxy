// internal/llm/summarizer.go
package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/devashar13/ssh-proxy/internal/config"
)

type Summarizer struct {
	config *config.Config
}

// NewSummarizer creates a new summarizer instance
func NewSummarizer(cfg *config.Config) *Summarizer {
	return &Summarizer{
		config: cfg,
	}
}
func (s *Summarizer) SummarizeSessionAsync(logFilePath string) {
	if !s.config.LLM.Enabled || s.config.LLM.APIKey == "" {
		log.Printf("LLM summarization is disabled or API key is missing")
		return
	}

	go func() {
		log.Printf("Starting asynchronous security analysis of session: %s", filepath.Base(logFilePath))
		if err := s.SummarizeSession(logFilePath); err != nil {
			log.Printf("Error summarizing session: %v", err)
		} else {
			log.Printf("Security analysis completed for %s", filepath.Base(logFilePath))
		}
	}()
}
func (s *Summarizer) SummarizeSession(logFilePath string) error {
	logContent, err := os.ReadFile(logFilePath)
	if err != nil {
		return fmt.Errorf("failed to read log file: %w", err)
	}

	var summary string
	switch s.config.LLM.Provider {
	case "openai":
		summary, err = s.callOpenAI(string(logContent))
	default:
		return fmt.Errorf("unsupported LLM provider: %s", s.config.LLM.Provider)
	}

	if err != nil {
		return fmt.Errorf("failed to generate summary: %w", err)
	}

	summaryFilePath := logFilePath + ".summary"
	err = os.WriteFile(summaryFilePath, []byte(summary), 0644)
	if err != nil {
		return fmt.Errorf("failed to write summary file: %w", err)
	}

	return nil
}


func (s *Summarizer) callOpenAI(logContent string) (string, error) {
	apiURL := "https://api.openai.com/v1/chat/completions"

	// Create the prompt for security analysis
	prompt := fmt.Sprintf(`
Analyze the following SSH session log and provide a security assessment:

1. Identify all commands executed during the session
2. Flag any potentially suspicious or dangerous commands
3. Evaluate the overall security risk (low, medium, high)
4. Provide recommendations if any security concerns are identified

SSH Session Log:
%s
`, logContent)

	requestBody, err := json.Marshal(map[string]interface{}{
		"model": s.config.LLM.Model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a security analyst specializing in SSH session analysis.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.3, 
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.LLM.APIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("invalid response format")
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid choice format")
	}

	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid message format")
	}

	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("invalid content format")
	}

	return content, nil
}
