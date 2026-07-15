package ui

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"gitlab.com/marseille-bb/mini-claude/internal/apiclient"
	"gitlab.com/marseille-bb/mini-claude/internal/chat"
	"gitlab.com/marseille-bb/mini-claude/internal/shared"
	"gitlab.com/marseille-bb/mini-claude/internal/store"
)

// authDoneMsg is the result of a login/register network call.
type authDoneMsg struct {
	creds store.Credentials
	err   error
}

// syncDoneMsg is the result of an export/import network call.
type syncDoneMsg struct {
	verb     string // "export" or "import"
	backupID string
	payload  shared.DataPayload
	err      error
}

// currentPayload snapshots the local config + conversation for syncing.
func (m Model) currentPayload() shared.DataPayload {
	return shared.DataPayload{
		Config: m.cfg,
		State:  shared.State{Messages: m.history.Conversation()},
	}
}

func loginCmd(api *apiclient.Client, ctx context.Context, email, password string, register bool) tea.Cmd {
	return func() tea.Msg {
		var (
			resp shared.AuthResponse
			err  error
		)
		if register {
			resp, err = api.Register(ctx, email, password)
		} else {
			resp, err = api.Login(ctx, email, password)
		}
		if err != nil {
			return authDoneMsg{err: err}
		}
		return authDoneMsg{creds: store.Credentials{Email: email, Token: resp.Token, ExpiresAt: resp.ExpiresAt}}
	}
}

func exportCmd(api *apiclient.Client, ctx context.Context, payload shared.DataPayload) tea.Cmd {
	return func() tea.Msg {
		backup, err := api.Export(ctx, payload)
		return syncDoneMsg{verb: "export", backupID: backup.ID, err: err}
	}
}

func importCmd(api *apiclient.Client, ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		payload, err := api.Import(ctx)
		return syncDoneMsg{verb: "import", payload: payload, err: err}
	}
}

// openAuth switches to the authentication screen in login or register mode.
func (m *Model) openAuth(register bool) {
	m.mode = modeAuth
	m.authRegister = register
	m.authFocus = 0
	m.authEmail.Reset()
	m.authPass.Reset()
	m.lastErr = nil
	m.notice = ""
	m.syncAuthFocus()
	m.refresh()
}

// syncAuthFocus focuses the active input and blurs the other.
func (m *Model) syncAuthFocus() {
	if m.authFocus == 0 {
		m.authEmail.Focus()
		m.authPass.Blur()
	} else {
		m.authEmail.Blur()
		m.authPass.Focus()
	}
}

func (m Model) updateAuth(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeChat
		m.refresh()
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	case "tab", "shift+tab", "up", "down":
		m.authFocus = 1 - m.authFocus
		m.syncAuthFocus()
		m.refresh()
		return m, nil
	case "ctrl+r":
		m.authRegister = !m.authRegister
		m.refresh()
		return m, nil
	case "enter":
		email := strings.TrimSpace(m.authEmail.Value())
		password := m.authPass.Value()
		if email == "" || password == "" {
			m.notice = "email and password are required"
			m.refresh()
			return m, nil
		}
		m.notice = "connecting…"
		m.lastErr = nil
		m.refresh()
		return m, loginCmd(m.api, m.ctx, email, password, m.authRegister)
	}

	var cmd tea.Cmd
	if m.authFocus == 0 {
		m.authEmail, cmd = m.authEmail.Update(msg)
	} else {
		m.authPass, cmd = m.authPass.Update(msg)
	}
	m.refresh()
	return m, cmd
}

// applyImported replaces local config + conversation with data pulled from the
// server, then persists it locally.
func (m *Model) applyImported(payload shared.DataPayload) {
	m.cfg = payload.Config
	m.client.SetModel(m.cfg.Model)
	applyTheme(themeByName(m.cfg.Theme))
	m.history = chat.Load(m.cfg.SystemPrompt, payload.State.Messages)
	m.persistConfig()
	m.persistState()
}

func (m Model) renderAuth() string {
	var sb strings.Builder
	title := "Sign in"
	if m.authRegister {
		title = "Create an account"
	}
	sb.WriteString(welcomeSectionStyle.Render(title) + "\n\n")
	sb.WriteString(welcomeLabelStyle.Render("email") + "\n")
	sb.WriteString(m.authEmail.View() + "\n\n")
	sb.WriteString(welcomeLabelStyle.Render("password") + "\n")
	sb.WriteString(m.authPass.View() + "\n\n")
	sb.WriteString(subtleStyle.Render("tab switch field  ·  enter submit  ·  ctrl+r toggle login/register  ·  esc cancel"))

	// Surface the outcome right on the screen (the status line is hidden here).
	if m.lastErr != nil {
		sb.WriteString("\n\n" + errorStyle.Render("✗ "+m.lastErr.Error()))
	} else if m.notice != "" {
		sb.WriteString("\n\n" + noticeStyle.Render(m.notice))
	}
	return sb.String()
}
