package chat

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model     string    `json:"model"`
	System    string    `json:"system,omitempty"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
	Stream    bool      `json:"stream"`
}

type OMLXResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

func stripThinking(text string) string {
	// oMLX wraps responses in a thinking process. The thinking starts with
	// "Here's a thinking process" and contains numbered sections.
	// The actual output (tool calls or final answer) comes at the end.
	for _, prefix := range []string{"Here's a thinking process", "Let me think through"} {
		idx := strings.Index(text, prefix)
		if idx < 0 {
			continue
		}
		// Find the last [TOOL call in the text — that's the real tool call,
		// not the mentions in the thinking process
		lastTool := strings.LastIndex(text[idx:], "[TOOL")
		if lastTool >= 0 {
			return strings.TrimSpace(text[idx+lastTool:])
		}
		// No tool call — find the output section after thinking
		for _, marker := range []string{"\n## ", "\n##\n", "Output Generation:"} {
			fullIdx := strings.Index(text[idx:], marker)
			if fullIdx >= 0 {
				return strings.TrimSpace(text[idx+fullIdx+len(marker):])
			}
		}
		// Fallback: return everything after the thinking header
		return strings.TrimSpace(text[idx:])
	}
	// No thinking prefix found — return as-is
	return strings.TrimSpace(text)
}

func getCommonConfig() (string, string, string) {
	baseURL := os.Getenv("LLM_BASE_URL")
	model := os.Getenv("LLM_MODEL")
	apiKey := os.Getenv("LLM_API_KEY")

	if baseURL == "" {
		baseURL = "http://127.0.0.1:8000"
	}
	if model == "" {
		model = "claude-sonnet-4-6"
	}
	if apiKey == "" {
		apiKey = "1111"
	}
	return baseURL, model, apiKey
}

func ChatStream(messages []Message, systemPrompt string, w http.ResponseWriter, flush func(), sendDone bool) error {
	baseURL, model, apiKey := getCommonConfig()

	reqBody := ChatRequest{
		Model:     model,
		System:    systemPrompt,
		Messages:  messages,
		MaxTokens: 4000,
		Stream:    true,
	}

	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", baseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 600 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("LLM API error %d: %s", resp.StatusCode, string(b))
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")

		var event map[string]interface{}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		if event["type"] == "content_block_delta" {
			if block, ok := event["delta"].(map[string]interface{}); ok {
				deltaType := block["type"].(string)
				// oMLX sends "text_delta" with "text" field and "thinking_delta" with "thinking" field
				if deltaType == "text_delta" {
					if text, ok := block["text"].(string); ok {
						fmt.Fprintf(w, "data: %s\n\n", text)
						flush()
					}
				}
			}
		}

		if event["type"] == "message_stop" {
			if sendDone {
				fmt.Fprintln(w, "data: [DONE]")
				flush()
			}
			break
		}
	}

	return nil
}

func ChatOnce(messages []Message, systemPrompt string) (string, error) {
	baseURL, model, apiKey := getCommonConfig()

	reqBody := ChatRequest{
		Model:     model,
		System:    systemPrompt,
		Messages:  messages,
		MaxTokens: 800,
		Stream:    false,
	}

	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", baseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 600 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	rawBody, readErr := io.ReadAll(resp.Body)
	fmt.Printf("[ChatOnce] status=%d, bodyLen=%d, readErr=%v, first200=%q\n", resp.StatusCode, len(rawBody), readErr, rawBody[:min(len(rawBody), 200)])

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("LLM API error %d: %s", resp.StatusCode, string(rawBody))
	}

	var result OMLXResponse
	if err := json.Unmarshal(rawBody, &result); err != nil {
		return "", fmt.Errorf("parse error: %w (body: %s)", err, string(rawBody))
	}

	// Find the text content (skip thinking blocks)
	for _, c := range result.Content {
		if c.Type == "text" && c.Text != "" {
			text := stripThinking(c.Text)
			return text, nil
		}
	}
	return "", fmt.Errorf("empty text response")
}
