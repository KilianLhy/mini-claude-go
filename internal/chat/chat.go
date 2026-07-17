package chat

import "github.com/KilianLhy/mini-claude-go/internal/shared"

// Roles are re-exported from the shared contract so callers of this package
// need not import two packages for the constants and the type.
const (
	RoleSystem    = shared.RoleSystem
	RoleUser      = shared.RoleUser
	RoleAssistant = shared.RoleAssistant
)

// Message is the shared wire type. Aliasing (rather than redefining) keeps a
// single source of truth: the CLI, the server, and this package all use the
// exact same struct.
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

// Load rebuilds a history from a system prompt plus previously persisted
// (non-system) messages.
func Load(systemPrompt string, msgs []Message) *History {
	h := New(systemPrompt)
	h.messages = append(h.messages, msgs...)
	return h
}

func (h *History) Add(role, content string) {
	h.messages = append(h.messages, Message{Role: role, Content: content})
}

// Conversation returns the messages excluding the system prompt, i.e. the part
// worth persisting as application state.
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
