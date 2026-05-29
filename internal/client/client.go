package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"gitlab.com/marseille-bb/mini-claude/internal/chat"
)

type Client struct {
	baseURL string
	model   string
	temp    float64
	http    *http.Client
}

func New(baseURL, model string, temperature float64) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		temp:    temperature,
		http:    &http.Client{},
	}
}

func (c *Client) Model() string {
	return c.model
}

func (c *Client) SetModel(model string) {
	c.model = model
}

type modelsResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

func (c *Client) ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/models", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("upstream %d: %s", resp.StatusCode, strings.TrimSpace(string(msg)))
	}

	var r modelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(r.Data))
	for _, m := range r.Data {
		out = append(out, m.ID)
	}
	return out, nil
}

type chatRequest struct {
	Model       string         `json:"model"`
	Messages    []chat.Message `json:"messages"`
	Stream      bool           `json:"stream"`
	Temperature float64        `json:"temperature"`
}

type streamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason,omitempty"`
	} `json:"choices"`
}

func (c *Client) Stream(ctx context.Context, messages []chat.Message) (<-chan string, <-chan error) {
	tokens := make(chan string)
	errs := make(chan error, 1)

	go func() {
		defer close(tokens)
		defer close(errs)
		if err := c.stream(ctx, messages, tokens); err != nil {
			errs <- err
		}
	}()

	return tokens, errs
}

func (c *Client) stream(ctx context.Context, messages []chat.Message, tokens chan<- string) error {
	body, err := json.Marshal(chatRequest{
		Model:       c.model,
		Messages:    messages,
		Stream:      true,
		Temperature: c.temp,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("upstream %d: %s", resp.StatusCode, strings.TrimSpace(string(msg)))
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimPrefix(line, "data: ")
		if payload == "[DONE]" {
			return nil
		}
		var chunk streamChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			return fmt.Errorf("decode chunk: %w", err)
		}
		for _, choice := range chunk.Choices {
			if choice.Delta.Content == "" {
				continue
			}
			select {
			case tokens <- choice.Delta.Content:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return scanner.Err()
}
