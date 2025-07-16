package tui

import "github.com/charmbracelet/lipgloss"

// Styles defines the styles for the UI.
type Styles struct {
	Title      lipgloss.Style
	Subtitle   lipgloss.Style
	Help       lipgloss.Style
	ListItem   lipgloss.Style
	ListHeader lipgloss.Style
	Selected   lipgloss.Style
	Error      lipgloss.Style
	Success    lipgloss.Style
	Warning    lipgloss.Style
	TextInput  lipgloss.Style
}

// DefaultStyles returns the default styles.
func DefaultStyles() *Styles {
	return &Styles{
		Title:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("62")).Margin(1, 0),
		Subtitle:   lipgloss.NewStyle().Foreground(lipgloss.Color("244")),
		Help:       lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		ListItem:   lipgloss.NewStyle().PaddingLeft(2),
		ListHeader: lipgloss.NewStyle().Bold(true).PaddingLeft(2),
		Selected:   lipgloss.NewStyle().Foreground(lipgloss.Color("62")).Bold(true),
		Error:      lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true),
		Success:    lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true),
		Warning:    lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true),
		TextInput:  lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1),
	}
}
