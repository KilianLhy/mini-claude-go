package chat

const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

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

func (h *History) Add(role, content string) {
	h.messages = append(h.messages, Message{Role: role, Content: content})
}

func (h *History) Messages() []Message {
	out := make([]Message, len(h.messages))
	copy(out, h.messages)
	return out
}

func (h *History) Len() int {
	return len(h.messages)
}
