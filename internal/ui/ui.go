package ui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gitlab.com/marseille-bb/mini-claude/internal/chat"
	"gitlab.com/marseille-bb/mini-claude/internal/client"
	"gitlab.com/marseille-bb/mini-claude/internal/config"
)

type Model struct {
	cfg     config.Config
	client  *client.Client
	history *chat.History
	ctx     context.Context

	viewport viewport.Model
	textarea textarea.Model
	spinner  spinner.Model

	width  int
	height int
	ready  bool
}

func New(cfg config.Config, cli *client.Client, ctx context.Context) Model {
	ta := textarea.New()
	ta.Placeholder = "Type a message…"
	ta.Prompt = "│ "
	ta.CharLimit = 0
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.Focus()
	ta.KeyMap.InsertNewline.SetEnabled(false)

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("213"))

	return Model{
		cfg:      cfg,
		client:   cli,
		history:  chat.New(cfg.SystemPrompt),
		ctx:      ctx,
		textarea: ta,
		spinner:  sp,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.spinner.Tick)
}

var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("213")).
			Bold(true)
	subtleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
	viewportStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resize()
		m.ready = true

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "ctrl+j":
			m.textarea.InsertString("\n")
			return m, nil
		}
	}

	var tcmd, vcmd tea.Cmd
	m.textarea, tcmd = m.textarea.Update(msg)
	m.viewport, vcmd = m.viewport.Update(msg)
	cmds = append(cmds, tcmd, vcmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if !m.ready {
		return "loading…"
	}

	header := headerStyle.Render("mini-claude") +
		subtleStyle.Render(fmt.Sprintf("  model: %s  ·  %s", m.cfg.Model, m.cfg.BaseURL))
	status := subtleStyle.Render("ctrl+j newline · esc/ctrl+c quit")

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		viewportStyle.Render(m.viewport.View()),
		m.textarea.View(),
		status,
	)
}

func (m *Model) resize() {
	headerH := 1
	statusH := 1
	taH := m.textarea.Height()
	viewportH := m.height - headerH - statusH - taH - 2
	if viewportH < 3 {
		viewportH = 3
	}
	w := m.width
	if w < 20 {
		w = 20
	}
	innerW := w - 4
	if innerW < 10 {
		innerW = 10
	}

	if m.viewport.Width == 0 {
		m.viewport = viewport.New(innerW, viewportH)
	} else {
		m.viewport.Width = innerW
		m.viewport.Height = viewportH
	}
	m.textarea.SetWidth(w)
}
