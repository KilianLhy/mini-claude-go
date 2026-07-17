package ui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name      string
	Label     string
	Primary   lipgloss.Color
	Accent    lipgloss.Color
	Assistant lipgloss.Color
	Error     lipgloss.Color
	Text      lipgloss.Color
	Muted     lipgloss.Color
	Border    lipgloss.Color
}

var themeOrder = []string{"claude", "midnight", "mono"}

var themes = map[string]Theme{
	"claude": {
		Name:      "claude",
		Label:     "Claude — orange & rose",
		Primary:   lipgloss.Color("213"),
		Accent:    lipgloss.Color("209"),
		Assistant: lipgloss.Color("82"),
		Error:     lipgloss.Color("196"),
		Text:      lipgloss.Color("231"),
		Muted:     lipgloss.Color("245"),
		Border:    lipgloss.Color("240"),
	},
	"midnight": {
		Name:      "midnight",
		Label:     "Sombre — bleu nuit",
		Primary:   lipgloss.Color("75"),
		Accent:    lipgloss.Color("39"),
		Assistant: lipgloss.Color("79"),
		Error:     lipgloss.Color("203"),
		Text:      lipgloss.Color("231"),
		Muted:     lipgloss.Color("244"),
		Border:    lipgloss.Color("238"),
	},
	"mono": {
		Name:      "mono",
		Label:     "Mono — sobre",
		Primary:   lipgloss.Color("252"),
		Accent:    lipgloss.Color("250"),
		Assistant: lipgloss.Color("254"),
		Error:     lipgloss.Color("203"),
		Text:      lipgloss.Color("255"),
		Muted:     lipgloss.Color("241"),
		Border:    lipgloss.Color("238"),
	},
}

func themeByName(name string) Theme {
	if t, ok := themes[name]; ok {
		return t
	}
	return themes["claude"]
}

var (
	headerStyle         lipgloss.Style
	subtleStyle         lipgloss.Style
	userStyle           lipgloss.Style
	assistantStyle      lipgloss.Style
	errorStyle          lipgloss.Style
	viewportStyle       lipgloss.Style
	welcomeChipStyle    lipgloss.Style
	welcomeStarStyle    lipgloss.Style
	welcomeTitleStyle   lipgloss.Style
	welcomeLogoStyle    lipgloss.Style
	welcomeLabelStyle   lipgloss.Style
	welcomeValueStyle   lipgloss.Style
	welcomeTipStyle     lipgloss.Style
	welcomeSectionStyle lipgloss.Style
	welcomeAccentStyle  lipgloss.Style
	pickerArrowStyle    lipgloss.Style
	pickerSelectedStyle lipgloss.Style
	pickerItemStyle     lipgloss.Style
	noticeStyle         lipgloss.Style
)

func applyTheme(t Theme) {
	headerStyle = lipgloss.NewStyle().Foreground(t.Primary).Bold(true)
	subtleStyle = lipgloss.NewStyle().Foreground(t.Muted)
	userStyle = lipgloss.NewStyle().Foreground(t.Primary).Bold(true)
	assistantStyle = lipgloss.NewStyle().Foreground(t.Assistant).Bold(true)
	errorStyle = lipgloss.NewStyle().Foreground(t.Error)
	viewportStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Border).
		Padding(0, 1)
	welcomeChipStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Accent).
		Padding(0, 2)
	welcomeStarStyle = lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	welcomeTitleStyle = lipgloss.NewStyle().Foreground(t.Text).Bold(true)
	welcomeLogoStyle = lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	welcomeLabelStyle = lipgloss.NewStyle().Foreground(t.Muted)
	welcomeValueStyle = lipgloss.NewStyle().Foreground(t.Text)
	welcomeTipStyle = lipgloss.NewStyle().Foreground(t.Muted).Italic(true)
	welcomeSectionStyle = lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	welcomeAccentStyle = lipgloss.NewStyle().Foreground(t.Primary)
	pickerArrowStyle = lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	pickerSelectedStyle = lipgloss.NewStyle().Foreground(t.Text).Bold(true)
	pickerItemStyle = lipgloss.NewStyle().Foreground(t.Muted)
	noticeStyle = lipgloss.NewStyle().Foreground(t.Assistant)
}
