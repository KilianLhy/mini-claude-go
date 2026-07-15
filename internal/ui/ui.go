package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gitlab.com/marseille-bb/mini-claude/internal/apiclient"
	"gitlab.com/marseille-bb/mini-claude/internal/chat"
	"gitlab.com/marseille-bb/mini-claude/internal/client"
	"gitlab.com/marseille-bb/mini-claude/internal/config"
	"gitlab.com/marseille-bb/mini-claude/internal/shared"
	"gitlab.com/marseille-bb/mini-claude/internal/store"
)

type mode int

const (
	modeChat mode = iota
	modeModelLoading
	modeModelPicker
	modeSettings
	modeAuth
	modeHelp
)

type Model struct {
	cfg     config.Config
	client  *client.Client
	history *chat.History
	ctx     context.Context

	viewport viewport.Model
	textarea textarea.Model
	spinner  spinner.Model

	api           *apiclient.Client
	authEmail     textinput.Model
	authPass      textinput.Model
	authFocus     int  // 0 = email, 1 = password
	authRegister  bool // register vs login mode on the auth screen
	loggedInEmail string

	streaming bool
	current   string
	tokens    <-chan string
	errs      <-chan error
	lastErr   error
	notice    string

	mode        mode
	models      []string
	modelCursor int

	themeCursor   int
	themeOriginal string // theme to restore if the settings screen is cancelled

	width  int
	height int
	ready  bool
}

type tokenMsg struct{ content string }
type streamDoneMsg struct{ err error }
type modelsMsg struct {
	models []string
	err    error
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

	theme := themeByName(cfg.Theme)
	applyTheme(theme)

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(theme.Primary)

	// Restore any persisted conversation. A load error is non-fatal: we start
	// with an empty history and tell the user.
	var notice string
	st, err := store.LoadState()
	if err != nil {
		notice = "could not load saved history (starting fresh)"
	}

	// Auth inputs for the sign-in screen.
	emailInput := textinput.New()
	emailInput.Placeholder = "you@example.com"
	emailInput.CharLimit = 254
	passInput := textinput.New()
	passInput.Placeholder = "password"
	passInput.EchoMode = textinput.EchoPassword
	passInput.CharLimit = 128

	// Restore a saved login session, if any, so the user stays logged in.
	var loggedInEmail, token string
	if creds, found, _ := store.LoadCredentials(); found {
		loggedInEmail = creds.Email
		token = creds.Token
	}

	return Model{
		cfg:           cfg,
		client:        cli,
		history:       chat.Load(cfg.SystemPrompt, st.Messages),
		ctx:           ctx,
		textarea:      ta,
		spinner:       sp,
		notice:        notice,
		api:           apiclient.New(cfg.ServerURL, token),
		authEmail:     emailInput,
		authPass:      passInput,
		loggedInEmail: loggedInEmail,
	}
}

// persistState saves the current conversation to disk. Failures are surfaced
// as a notice but never interrupt the chat (offline/local-first).
func (m *Model) persistState() {
	if err := store.SaveState(shared.State{Messages: m.history.Conversation()}); err != nil {
		m.notice = "could not save history"
	}
}

// persistConfig saves the current settings. Non-fatal on failure.
func (m *Model) persistConfig() {
	if err := config.Save(m.cfg); err != nil {
		m.notice = "could not save settings"
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.spinner.Tick)
}

// UI styles live in theme.go and are (re)built by applyTheme.

const (
	logoMini = `███╗   ███╗██╗███╗   ██╗██╗
████╗ ████║██║████╗  ██║██║
██╔████╔██║██║██╔██╗ ██║██║
██║╚██╔╝██║██║██║╚██╗██║██║
██║ ╚═╝ ██║██║██║ ╚████║██║
╚═╝     ╚═╝╚═╝╚═╝  ╚═══╝╚═╝`

	logoClaude = ` ██████╗██╗      █████╗ ██╗   ██╗██████╗ ███████╗
██╔════╝██║     ██╔══██╗██║   ██║██╔══██╗██╔════╝
██║     ██║     ███████║██║   ██║██║  ██║█████╗
██║     ██║     ██╔══██║██║   ██║██║  ██║██╔══╝
╚██████╗███████╗██║  ██║╚██████╔╝██████╔╝███████╗
 ╚═════╝╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ╚══════╝`

	logoMinWidth = 49
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resize()
		m.refresh()
		m.ready = true

	case tea.KeyMsg:
		if m.mode == modeAuth {
			return m.updateAuth(msg)
		}
		if m.mode == modeSettings {
			return m.updateSettings(msg)
		}
		if m.mode == modeModelPicker {
			return m.updatePicker(msg)
		}
		if m.mode == modeHelp {
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			m.mode = modeChat
			m.refresh()
			return m, nil
		}
		if m.mode == modeModelLoading {
			if s := msg.String(); s == "esc" || s == "ctrl+c" {
				m.mode = modeChat
				m.refresh()
				return m, nil
			}
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.streaming {
				return m, nil
			}
			return m, tea.Quit
		case "enter":
			if m.streaming {
				return m, nil
			}
			input := strings.TrimSpace(m.textarea.Value())
			if input == "" {
				return m, nil
			}
			m.textarea.Reset()
			if strings.HasPrefix(input, "/") {
				return m.handleCommand(input)
			}
			m.history.Add(chat.RoleUser, input)
			m.current = ""
			m.streaming = true
			m.lastErr = nil
			m.notice = ""
			m.tokens, m.errs = m.client.Stream(m.ctx, m.history.Messages())
			m.refresh()
			return m, tea.Batch(nextEvent(m.tokens, m.errs), m.spinner.Tick)
		case "ctrl+j":
			m.textarea.InsertString("\n")
			return m, nil
		}

	case tokenMsg:
		m.current += msg.content
		m.refresh()
		return m, nextEvent(m.tokens, m.errs)

	case streamDoneMsg:
		if msg.err != nil {
			m.lastErr = msg.err
		} else if m.current != "" {
			m.history.Add(chat.RoleAssistant, m.current)
			m.persistState()
		}
		m.current = ""
		m.streaming = false
		m.tokens = nil
		m.errs = nil
		m.refresh()
		return m, nil

	case modelsMsg:
		if msg.err != nil {
			m.lastErr = msg.err
			m.mode = modeChat
			m.refresh()
			return m, nil
		}
		m.models = msg.models
		m.modelCursor = 0
		for i, name := range m.models {
			if name == m.cfg.Model {
				m.modelCursor = i
				break
			}
		}
		m.mode = modeModelPicker
		m.refresh()
		return m, nil

	case authDoneMsg:
		if msg.err != nil {
			// Stay on the auth screen so the user can retry.
			m.notice = ""
			m.lastErr = msg.err
			m.refresh()
			return m, nil
		}
		m.api.SetToken(msg.creds.Token)
		m.loggedInEmail = msg.creds.Email
		m.lastErr = nil
		if err := store.SaveCredentials(msg.creds); err != nil {
			m.notice = "signed in (could not save session)"
		} else {
			m.notice = "signed in as " + msg.creds.Email
		}
		m.mode = modeChat
		m.refresh()
		return m, nil

	case syncDoneMsg:
		if msg.err != nil {
			m.lastErr = msg.err
			m.notice = ""
			m.refresh()
			return m, nil
		}
		m.lastErr = nil
		switch msg.verb {
		case "export":
			m.notice = "exported to server (backup #" + msg.backupID + ")"
		case "import":
			if msg.payload.UpdatedAt.IsZero() {
				m.notice = "nothing to import (server is empty)"
			} else {
				m.applyImported(msg.payload)
				m.notice = "imported from server"
			}
		}
		m.refresh()
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		if m.streaming {
			cmds = append(cmds, cmd)
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

	account := "not signed in"
	if m.loggedInEmail != "" {
		account = m.loggedInEmail
	}
	header := headerStyle.Render("mini-claude") +
		subtleStyle.Render(fmt.Sprintf("  model: %s  ·  %s  ·  %s", m.cfg.Model, m.cfg.BaseURL, account))

	status := m.statusLine()

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		viewportStyle.Render(m.viewport.View()),
		m.textarea.View(),
		status,
	)
}

func (m Model) statusLine() string {
	switch m.mode {
	case modeModelLoading:
		return m.spinner.View() + " " + subtleStyle.Render("loading models…  esc to cancel")
	case modeModelPicker:
		return subtleStyle.Render("↑/↓ navigate  ·  enter select  ·  esc cancel")
	case modeSettings:
		return subtleStyle.Render("↑/↓ preview theme  ·  enter keep  ·  esc cancel")
	case modeAuth:
		return subtleStyle.Render("tab switch  ·  enter submit  ·  ctrl+r login/register  ·  esc cancel")
	case modeHelp:
		return subtleStyle.Render("press any key to return")
	}
	hint := subtleStyle.Render("enter send · ctrl+j newline · /model · /clear · esc/ctrl+c quit")
	if m.lastErr != nil {
		return errorStyle.Render("error: "+m.lastErr.Error()) + "  " + hint
	}
	if m.notice != "" {
		return noticeStyle.Render(m.notice) + "  " + hint
	}
	if m.streaming {
		return m.spinner.View() + " " + subtleStyle.Render("generating…") + "  " + hint
	}
	return hint
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

func (m *Model) refresh() {
	switch m.mode {
	case modeModelLoading:
		m.viewport.SetContent(subtleStyle.Render("fetching models from ") + welcomeValueStyle.Render(m.cfg.BaseURL) + subtleStyle.Render("…"))
	case modeModelPicker:
		m.viewport.SetContent(m.renderPicker())
	case modeSettings:
		m.viewport.SetContent(m.renderSettings())
	case modeAuth:
		m.viewport.SetContent(m.renderAuth())
	case modeHelp:
		m.viewport.SetContent(m.renderHelp())
	default:
		m.viewport.SetContent(m.renderHistory())
	}
	m.viewport.GotoBottom()
}

func (m Model) renderPicker() string {
	var sb strings.Builder
	sb.WriteString(welcomeSectionStyle.Render("Choose a model") + "\n\n")
	if len(m.models) == 0 {
		sb.WriteString(subtleStyle.Render("No models found. Pull one with:") + "\n")
		sb.WriteString(welcomeValueStyle.Render("  ollama pull llama3.2:3b") + "\n")
		return sb.String()
	}
	for i, name := range m.models {
		prefix := "   "
		line := pickerItemStyle.Render(name)
		if i == m.modelCursor {
			prefix = " " + pickerArrowStyle.Render("›") + " "
			line = pickerSelectedStyle.Render(name)
		}
		marker := ""
		if name == m.cfg.Model {
			marker = subtleStyle.Render("  (current)")
		}
		sb.WriteString(prefix + line + marker + "\n")
	}
	return sb.String()
}

func (m Model) renderHelp() string {
	cmdW := lipgloss.NewStyle().Width(12)
	line := func(name, desc string) string {
		return cmdW.Render(welcomeAccentStyle.Bold(true).Render(name)) + welcomeLabelStyle.Render(desc)
	}
	section := func(title string) string {
		return welcomeSectionStyle.Render(title)
	}

	var sb strings.Builder
	sb.WriteString(section("Chat") + "\n")
	sb.WriteString(line("enter", "send your message") + "\n")
	sb.WriteString(line("ctrl+j", "insert a newline") + "\n\n")

	sb.WriteString(section("Commands") + "\n")
	sb.WriteString(line("/help", "show this help") + "\n")
	sb.WriteString(line("/model", "pick or switch the model") + "\n")
	sb.WriteString(line("/settings", "change the color theme") + "\n")
	sb.WriteString(line("/clear", "start a fresh conversation") + "\n\n")

	sb.WriteString(section("Account & sync") + "\n")
	sb.WriteString(line("/login", "sign in to your account") + "\n")
	sb.WriteString(line("/register", "create a new account") + "\n")
	sb.WriteString(line("/logout", "sign out of this device") + "\n")
	sb.WriteString(line("/export", "push config + history to the server") + "\n")
	sb.WriteString(line("/import", "pull config + history from the server") + "\n\n")

	sb.WriteString(section("Quit") + "\n")
	sb.WriteString(line("/quit", "exit mini-claude") + "\n")
	sb.WriteString(line("esc / ctrl+c", "exit mini-claude") + "\n")

	return sb.String()
}

func (m Model) handleCommand(input string) (tea.Model, tea.Cmd) {
	parts := strings.Fields(input)
	cmd := parts[0]
	switch cmd {
	case "/help", "/?":
		m.mode = modeHelp
		m.lastErr = nil
		m.notice = ""
		m.refresh()
		return m, nil
	case "/model":
		if len(parts) > 1 {
			name := parts[1]
			m.client.SetModel(name)
			m.cfg.Model = name
			m.notice = "switched to " + name
			m.lastErr = nil
			m.persistConfig()
			m.refresh()
			return m, nil
		}
		m.mode = modeModelLoading
		m.lastErr = nil
		m.notice = ""
		m.refresh()
		return m, tea.Batch(fetchModelsCmd(m.client, m.ctx), m.spinner.Tick)
	case "/login", "/signin":
		m.openAuth(false)
		return m, nil
	case "/register", "/signup":
		m.openAuth(true)
		return m, nil
	case "/logout":
		m.api.SetToken("")
		m.loggedInEmail = ""
		if err := store.ClearCredentials(); err != nil {
			m.lastErr = err
		} else {
			m.notice = "signed out"
		}
		m.refresh()
		return m, nil
	case "/export":
		if m.api.Token() == "" {
			m.lastErr = fmt.Errorf("sign in first with /login")
			m.refresh()
			return m, nil
		}
		m.notice = "exporting…"
		m.lastErr = nil
		m.refresh()
		return m, exportCmd(m.api, m.ctx, m.currentPayload())
	case "/import":
		if m.api.Token() == "" {
			m.lastErr = fmt.Errorf("sign in first with /login")
			m.refresh()
			return m, nil
		}
		m.notice = "importing…"
		m.lastErr = nil
		m.refresh()
		return m, importCmd(m.api, m.ctx)
	case "/settings", "/theme":
		m.themeOriginal = m.cfg.Theme
		m.themeCursor = 0
		for i, name := range themeOrder {
			if name == m.cfg.Theme {
				m.themeCursor = i
				break
			}
		}
		m.mode = modeSettings
		m.lastErr = nil
		m.notice = ""
		m.refresh()
		return m, nil
	case "/clear":
		m.history = chat.New(m.cfg.SystemPrompt)
		m.lastErr = nil
		m.notice = "conversation cleared"
		m.persistState()
		m.refresh()
		return m, nil
	case "/quit", "/exit":
		return m, tea.Quit
	}
	m.lastErr = fmt.Errorf("unknown command: %s (try /model, /settings, /login, /export, /import, /clear, /quit)", cmd)
	return m, nil
}

// previewTheme applies a theme immediately so the settings screen shows a live
// preview. It updates the spinner too, then re-renders.
func (m *Model) previewTheme(name string) {
	t := themeByName(name)
	applyTheme(t)
	m.spinner.Style = lipgloss.NewStyle().Foreground(t.Primary)
	m.refresh()
}

func (m Model) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.themeCursor > 0 {
			m.themeCursor--
			m.previewTheme(themeOrder[m.themeCursor])
		}
	case "down", "j":
		if m.themeCursor < len(themeOrder)-1 {
			m.themeCursor++
			m.previewTheme(themeOrder[m.themeCursor])
		}
	case "enter":
		name := themeOrder[m.themeCursor]
		m.cfg.Theme = name
		m.previewTheme(name)
		m.persistConfig()
		m.notice = "theme: " + name
		m.mode = modeChat
		m.refresh()
	case "esc":
		m.previewTheme(m.themeOriginal) // revert the live preview
		m.mode = modeChat
		m.refresh()
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) renderSettings() string {
	var sb strings.Builder
	sb.WriteString(welcomeSectionStyle.Render("Settings — theme") + "\n\n")
	for i, name := range themeOrder {
		t := themeByName(name)
		prefix := "   "
		label := pickerItemStyle.Render(t.Label)
		if i == m.themeCursor {
			prefix = " " + pickerArrowStyle.Render("›") + " "
			label = pickerSelectedStyle.Render(t.Label)
		}
		swatch := lipgloss.NewStyle().Foreground(t.Primary).Render("●") +
			lipgloss.NewStyle().Foreground(t.Accent).Render("●") +
			lipgloss.NewStyle().Foreground(t.Assistant).Render("●")
		marker := ""
		if name == m.cfg.Theme {
			marker = subtleStyle.Render("  (current)")
		}
		sb.WriteString(prefix + label + "  " + swatch + marker + "\n")
	}
	sb.WriteString("\n" + subtleStyle.Render("preview is live  ·  enter keep  ·  esc cancel"))
	return sb.String()
}

func (m Model) updatePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.modelCursor > 0 {
			m.modelCursor--
			m.refresh()
		}
	case "down", "j":
		if m.modelCursor < len(m.models)-1 {
			m.modelCursor++
			m.refresh()
		}
	case "enter":
		if len(m.models) > 0 {
			name := m.models[m.modelCursor]
			m.client.SetModel(name)
			m.cfg.Model = name
			m.notice = "switched to " + name
			m.persistConfig()
		}
		m.mode = modeChat
		m.refresh()
	case "esc":
		m.mode = modeChat
		m.refresh()
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func fetchModelsCmd(cli *client.Client, ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		models, err := cli.ListModels(ctx)
		return modelsMsg{models: models, err: err}
	}
}

func (m *Model) renderHistory() string {
	var sb strings.Builder
	msgs := m.history.Messages()
	first := true
	for _, msg := range msgs {
		if msg.Role == chat.RoleSystem {
			continue
		}
		if !first {
			sb.WriteString("\n")
		}
		first = false
		switch msg.Role {
		case chat.RoleUser:
			sb.WriteString(userStyle.Render("you") + "\n")
		case chat.RoleAssistant:
			sb.WriteString(assistantStyle.Render("mini-claude") + "\n")
		}
		sb.WriteString(msg.Content + "\n")
	}
	if m.streaming && m.current != "" {
		if !first {
			sb.WriteString("\n")
		}
		sb.WriteString(assistantStyle.Render("mini-claude") + "\n")
		sb.WriteString(m.current + "\n")
	}
	if sb.Len() == 0 {
		return m.welcomeView()
	}
	return sb.String()
}

func (m Model) welcomeView() string {
	w := m.viewport.Width
	if w <= 0 {
		w = m.width
	}

	chip := welcomeChipStyle.Render(
		welcomeStarStyle.Render("✻ ") +
			welcomeTitleStyle.Render("Welcome to ") +
			welcomeAccentStyle.Bold(true).Render("mini-claude"),
	)

	var logo string
	if w >= logoMinWidth+2 {
		logo = lipgloss.JoinVertical(lipgloss.Left,
			welcomeLogoStyle.Render(logoMini),
			welcomeLogoStyle.Render(logoClaude),
		)
	} else {
		logo = welcomeLogoStyle.Render("mini-claude")
	}

	tagline := welcomeTipStyle.Render("A fast, private TUI chat for self-hosted LLMs.\nNothing leaves your machine.")

	builtBy := lipgloss.JoinVertical(lipgloss.Left,
		welcomeSectionStyle.Render("Built by"),
		welcomeValueStyle.Render("Hugo Stawiarski  ·  Kilian Lahaye  ·  Moustapha Sow"),
	)

	labelW := lipgloss.NewStyle().Width(9)
	infoLine := func(label, value string) string {
		return labelW.Render(welcomeLabelStyle.Render(label)) + welcomeValueStyle.Render(value)
	}
	hyperlink := func(label, url, display string) string {
		linked := "\x1b]8;;" + url + "\x1b\\" + welcomeValueStyle.Render(display) + "\x1b]8;;\x1b\\"
		return labelW.Render(welcomeLabelStyle.Render(label)) + linked
	}
	info := lipgloss.JoinVertical(lipgloss.Left,
		infoLine("model", m.cfg.Model),
		infoLine("server", m.cfg.BaseURL),
		hyperlink("site", "https://miniclaude.fr", "miniclaude.fr"),
	)

	cmdW := lipgloss.NewStyle().Width(11)
	cmdLine := func(name, desc string) string {
		return cmdW.Render(welcomeAccentStyle.Bold(true).Render(name)) + welcomeLabelStyle.Render(desc)
	}
	commands := lipgloss.JoinVertical(lipgloss.Left,
		welcomeSectionStyle.Render("Commands"),
		cmdLine("/help", "show all commands"),
		cmdLine("/model", "pick or switch the model"),
		cmdLine("/settings", "change the color theme"),
		cmdLine("/login", "sign in to sync your data"),
		cmdLine("/export", "push config + history to the server"),
		cmdLine("/import", "pull config + history from the server"),
		cmdLine("/clear", "start a fresh conversation"),
		cmdLine("/quit", "exit mini-claude"),
	)

	keys := subtleStyle.Render("enter send  ·  ctrl+j newline  ·  esc/ctrl+c quit")

	return lipgloss.JoinVertical(lipgloss.Left,
		chip,
		"",
		logo,
		"",
		tagline,
		"",
		builtBy,
		"",
		info,
		"",
		commands,
		"",
		keys,
	)
}

func nextEvent(tokens <-chan string, errs <-chan error) tea.Cmd {
	return func() tea.Msg {
		tok, ok := <-tokens
		if !ok {
			err := <-errs
			return streamDoneMsg{err: err}
		}
		return tokenMsg{content: tok}
	}
}
