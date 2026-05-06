package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	defaultBaseURL = "http://localhost:1234/v1"
	defaultModel   = "local-model"
	defaultAPIKey  = "lm-studio"
)

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type completionRequest struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
}

type completionResponse struct {
	Choices []struct {
		Message message `json:"message"`
	} `json:"choices"`
	// Error accepts both shapes LM Studio returns:
	//   {"error":{"message":"..."}}        - OpenAI-compatible
	//   {"error":"Context length exceeded"} - LM Studio's terse form on overflow
	Error json.RawMessage `json:"error,omitempty"`
}

// errorMessage extracts a human-readable error string from either shape.
// Falls back to the raw JSON when neither shape yields a usable message —
// surfacing unknown-but-non-empty error payloads beats hiding them as "".
func errorMessage(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	var obj struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(raw, &obj); err == nil && obj.Message != "" {
		return obj.Message
	}
	return string(raw)
}

type modelsResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
	Error json.RawMessage `json:"error,omitempty"`
}

func listModels() ([]string, error) {
	baseURL := envOr("LMS_BASE_URL", defaultBaseURL)
	apiKey := envOr("LMS_API_KEY", defaultAPIKey)

	req, err := http.NewRequest(http.MethodGet, baseURL+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET /models: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var mr modelsResponse
	if err := json.Unmarshal(raw, &mr); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	if msg := errorMessage(mr.Error); msg != "" {
		return nil, fmt.Errorf("API error: %s", msg)
	}

	ids := make([]string, len(mr.Data))
	for i, m := range mr.Data {
		ids[i] = m.ID
	}
	return ids, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func complete(msgs []message, maxTokens int, temperature float64) (string, error) {
	baseURL := envOr("LMS_BASE_URL", defaultBaseURL)
	model := envOr("LMS_MODEL", defaultModel)
	apiKey := envOr("LMS_API_KEY", defaultAPIKey)

	body, err := json.Marshal(completionRequest{
		Model:       model,
		Messages:    msgs,
		MaxTokens:   maxTokens,
		Temperature: temperature,
	})
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("POST %s: %w", baseURL, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	var cr completionResponse
	if err := json.Unmarshal(raw, &cr); err != nil {
		return "", fmt.Errorf("parse response: %w\nbody: %s", err, raw)
	}
	if msg := errorMessage(cr.Error); msg != "" {
		return "", fmt.Errorf("API error: %s", msg)
	}
	if len(cr.Choices) == 0 {
		return "", fmt.Errorf("empty response from model")
	}
	return cr.Choices[0].Message.Content, nil
}
