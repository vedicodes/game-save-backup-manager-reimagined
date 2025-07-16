package components

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/gemini/game-save-backup-manager-reimagined/internal/backup"
	"github.com/gemini/game-save-backup-manager-reimagined/internal/layout"
	"github.com/gemini/game-save-backup-manager-reimagined/internal/tui"
)

// ListItem wraps backup.Backup to implement list.Item interface
type ListItem backup.Backup

func (i ListItem) Title() string       { return i.Name }
func (i ListItem) Description() string { return i.CreatedAt.Format("2006-01-02 15:04:05") }
func (i ListItem) FilterValue() string { return i.Name }

// NormalItemDelegate handles rendering for normal list views (no checkboxes)
type NormalItemDelegate struct {
	list.DefaultDelegate
}

// NewNormalItemDelegate creates a delegate for normal list views
func NewNormalItemDelegate() *NormalItemDelegate {
	d := &NormalItemDelegate{}
	d.Styles = list.NewDefaultItemStyles()
	
	// Enhanced styling with rounded borders
	d.Styles.SelectedTitle = tui.DefaultStyles().Selected.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, layout.BorderPadding)
	d.Styles.SelectedDesc = tui.DefaultStyles().Selected.Copy().Faint(true).Padding(0, layout.BorderPadding)
	d.Styles.NormalTitle = tui.DefaultStyles().ListItem.Padding(0, layout.BorderPadding)
	d.Styles.NormalDesc = tui.DefaultStyles().ListItem.Copy().Faint(true).Padding(0, layout.BorderPadding)
	return d
}

// Render method for normal list items (no checkboxes)
func (d *NormalItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(ListItem)
	if !ok {
		return
	}

	title := i.Title()
	desc := i.Description()

	// Combine title and description for unified styling
	content := title + "\n" + desc

	if m.Index() == index {
		// Apply border to the entire content (title + description)
		styledContent := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, layout.BorderPadding).
			Render(content)
		fmt.Fprint(w, styledContent)
	} else {
		// Normal styling without border
		styledContent := lipgloss.NewStyle().
			Padding(0, layout.BorderPadding).
			Render(content)
		fmt.Fprint(w, styledContent)
	}
}

// SelectableItemDelegate handles rendering for delete view (with checkboxes)
type SelectableItemDelegate struct {
	list.DefaultDelegate
	selected map[int]struct{}
}

// NewSelectableItemDelegate creates a delegate for delete view with checkboxes
func NewSelectableItemDelegate(selected map[int]struct{}) *SelectableItemDelegate {
	d := &SelectableItemDelegate{
		selected: selected,
	}
	d.Styles = list.NewDefaultItemStyles()
	
	// Enhanced styling with rounded borders for selection view
	d.Styles.SelectedTitle = tui.DefaultStyles().Selected.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, layout.BorderPadding)
	d.Styles.SelectedDesc = tui.DefaultStyles().Selected.Copy().Faint(true).Padding(0, layout.BorderPadding)
	d.Styles.NormalTitle = tui.DefaultStyles().ListItem.Padding(0, layout.BorderPadding)
	d.Styles.NormalDesc = tui.DefaultStyles().ListItem.Copy().Faint(true).Padding(0, layout.BorderPadding)
	return d
}

// Render method for selectable list items (with checkboxes)
func (d *SelectableItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(ListItem)
	if !ok {
		return
	}

	var checkbox string
	if _, ok := d.selected[index]; ok {
		checkbox = "☑"
	} else {
		checkbox = "☐"
	}

	title := i.Title()
	desc := i.Description()

	// Combine checkbox, title and description with proper formatting
	content := fmt.Sprintf("%s%s%s\n%s%s", 
		checkbox, 
		generateSpaces(layout.CheckboxSpacing), 
		title,
		generateSpaces(layout.DescIndentation), 
		desc)

	if m.Index() == index {
		// Apply border to the entire content (checkbox + title + description)
		styledContent := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, layout.BorderPadding).
			Render(content)
		fmt.Fprint(w, styledContent)
	} else {
		// Normal styling without border
		styledContent := lipgloss.NewStyle().
			Padding(0, layout.BorderPadding).
			Render(content)
		fmt.Fprint(w, styledContent)
	}
}

// generateSpaces creates a string with the specified number of spaces
func generateSpaces(count int) string {
	spaces := ""
	for i := 0; i < count; i++ {
		spaces += " "
	}
	return spaces
}