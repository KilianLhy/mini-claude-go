package chat

import "github.com/KilianLhy/mini-claude-go/internal/shared"

const (
	RoleSystem    = shared.RoleSystem
	RoleUser      = shared.RoleUser
	RoleAssistant = shared.RoleAssistant
)

type Message = shared.Message

type History struct {
	messages []Message
}

func New(systemPrompt string) *History {
	h := &History{}
	if systemPrompt != "" {
		h.messages = append(h.messages, Message{Role: RoleSystem, Content: systemPrompt})
	}
	return h
}

func Load(systemPrompt string, msgs []Message) *History {
	h := New(systemPrompt)
	h.messages = append(h.messages, msgs...)
	return h
}

func (h *History) Add(role, content string) {
	h.messages = append(h.messages, Message{Role: role, Content: content})
}

func (h *History) Conversation() []Message {
	out := make([]Message, 0, len(h.messages))
	for _, m := range h.messages {
		if m.Role == RoleSystem {
			continue
		}
		out = append(out, m)
	}
	return out
}

func (h *History) Messages() []Message {
	out := make([]Message, len(h.messages))
	copy(out, h.messages)
	return out
}

func (h *History) Len() int {
	return len(h.messages)
}
