package client

import (
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
