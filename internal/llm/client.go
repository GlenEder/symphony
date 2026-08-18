// Package llm provides a small OpenAI-compatible chat completion client.
package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type Client struct {
	BaseURL, APIKey, Model string
	HTTPClient             *http.Client
	Timeout                time.Duration
}
type completionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream,omitempty"`
	Temperature *float64  `json:"temperature,omitempty"`
}
type Choice struct {
	Message Message `json:"message"`
	Delta   Message `json:"delta"`
}
type Response struct {
	ID      string         `json:"id"`
	Choices []Choice       `json:"choices"`
	Usage   map[string]any `json:"usage,omitempty"`
}
type StreamChunk struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
}

func (c *Client) Chat(ctx context.Context, messages []Message) (Response, error) {
	return c.do(ctx, messages, false, nil)
}
func (c *Client) Stream(ctx context.Context, messages []Message, onChunk func(StreamChunk) error) error {
	_, err := c.do(ctx, messages, true, onChunk)
	return err
}

func (c *Client) do(ctx context.Context, messages []Message, stream bool, onChunk func(StreamChunk) error) (Response, error) {
	if c == nil {
		return Response{}, errors.New("llm: nil client")
	}
	if c.Model == "" {
		return Response{}, errors.New("llm: model is required")
	}
	base := strings.TrimRight(c.BaseURL, "/")
	if base == "" {
		base = "https://api.openai.com/v1"
	}
	body, err := json.Marshal(completionRequest{Model: c.Model, Messages: messages, Stream: stream})
	if err != nil {
		return Response{}, fmt.Errorf("llm: encode request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return Response{}, fmt.Errorf("llm: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	hc := c.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: c.Timeout}
	}
	resp, err := hc.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("llm: request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		detail := strings.TrimSpace(string(data))
		if c.APIKey != "" {
			detail = strings.ReplaceAll(detail, c.APIKey, "[redacted]")
		}
		return Response{}, fmt.Errorf("llm: API returned %s: %s", resp.Status, detail)
	}
	if !stream {
		var out Response
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return Response{}, fmt.Errorf("llm: decode response: %w", err)
		}
		return out, nil
	}
	if onChunk == nil {
		return Response{}, errors.New("llm: stream callback is required")
	}
	var out Response
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 4096), 4*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "[DONE]" {
			break
		}
		var chunk StreamChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			return Response{}, fmt.Errorf("llm: decode stream: %w", err)
		}
		if err := onChunk(chunk); err != nil {
			return Response{}, err
		}
		out.ID = chunk.ID
		out.Choices = append(out.Choices, chunk.Choices...)
	}
	if err := scanner.Err(); err != nil {
		return Response{}, fmt.Errorf("llm: read stream: %w", err)
	}
	return out, nil
}
