package ui

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gemini/game-save-backup-manager-reimagined/internal/backup"
	"github.com/gemini/game-save-backup-manager-reimagined/internal/config"
	"github.com/gemini/game-save-backup-manager-reimagined/internal/tui"
	"github.com/gemini/game-save-backup-manager-reimagined/internal/validation"
)

// viewState represents the current view of the application.
type viewState int

const (
	initializingView viewState = iota
	mainMenuView
	backupListView
	createBackupView
	restoreConfirmationView
	deleteConfirmationView
	deletingView
	settingsView
	changeSavePathView
	changeBackupDirView
	firstRunView
)

// A message to indicate that the database is now initialized.
type dbInitializedMsg struct{}

// A message to clear the notification after a delay
type clearNotificationMsg struct{}

// model holds the state of the entire application.
type model struct {
	styles    *tui.Styles
	state     viewState
	list      list.Model
	textInput textinput.Model
	config    *config.Config
	db        *backup.DB
	width     int
	height    int
	err       error
	message   string
	selected  map[int]struct{}
}

// NewModel initializes the application's state.
func NewModel(cfg *config.Config, isFirstRun bool) *model {
	styles := tui.DefaultStyles()
	selected := make(map[int]struct{})

	// Create the list and text input components right away.
	list := list.New(nil, newNormalItemDelegate(), 0, 0)
	textInput := textinput.New()

	m := &model{
		styles:    styles,
		config:    cfg,
		list:      list,
		textInput: textInput,
		selected:  selected,
	}

	if isFirstRun {
		m.state = firstRunView
		m.textInput.Placeholder = "Enter your game's save file path"
		m.textInput.Focus()
		m.textInput.Width = 50 // Will be updated on first WindowSizeMsg
	} else {
		m.state = initializingView
		// DB will be initialized via a command.
	}
	return m
}

// Init is the first command that is run when the application starts.
func (m *model) Init() tea.Cmd {
	// If it's not the first run, we need to initialize the database.
	if m.state == initializingView {
		return m.initDB
	}
	return nil
}

// Update handles all messages and updates the model accordingly.
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Do not process any messages if we are in an error state.
	if m.err != nil {
		if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Update the list size.
		m.list.SetSize(m.width, m.height-6) // Adjust for title/help text with more spacing
		// Update text input width to scale with window
		m.textInput.Width = m.width - 10 // Leave some margin
		if m.textInput.Width < 20 {
			m.textInput.Width = 20 // Minimum width
		}
		return m, nil

	case tea.KeyMsg:
		// Global key bindings.
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		// Per-view key bindings.
		if m.state != firstRunView && m.state != mainMenuView && msg.String() == "q" {
			// Allow quitting from most views back to the main menu.
			m.state = mainMenuView
			m.message = "" // Clear any pending messages when returning to main menu
			return m, nil
		}

	case clearNotificationMsg:
		// Clear the notification message
		m.message = ""
		return m, nil

	case dbInitializedMsg:
		// This is a command that signals the DB is ready.
		m.state = mainMenuView
		return m, nil

	case error:
		m.err = msg
		return m, nil
	}

	// Process updates based on the current view.
	switch m.state {
	case initializingView:
		// No updates while initializing
		return m, nil
	case mainMenuView:
		return m.updateMainMenu(msg)
	case backupListView, restoreConfirmationView, deleteConfirmationView:
		return m.updateBackupViews(msg)
	case deletingView:
		return m.updateDeletingView(msg)
	case createBackupView:
		return m.updateCreateBackup(msg)
	case settingsView:
		return m.updateSettings(msg)
	case changeSavePathView:
		return m.updateChangeSavePath(msg)
	case changeBackupDirView:
		return m.updateChangeBackupDir(msg)
	case firstRunView:
		return m.updateFirstRun(msg)
	}

	return m, cmd
}

// View renders the UI based on the current model state.
func (m *model) View() string {
	// Always render errors first.
	if m.err != nil {
		errorStyle := m.styles.Error.Width(m.width)
		return errorStyle.Render(fmt.Sprintf("A critical error occurred:\n\n%v\n\nPress Ctrl+C to exit.", m.err))
	}

	// Render messages if they exist.
	if m.message != "" {
		// Show the message for industry standard time (don't clear immediately)
		return m.styles.Success.Render(m.message) + "\n" + m.currentView()
	}

	return m.currentView()
}

// currentView returns the string for the current view.
func (m *model) currentView() string {
	body := new(strings.Builder)

	switch m.state {
	case initializingView:
		body.WriteString("Initializing...")
	case mainMenuView:
		body.WriteString(m.viewMainMenu())
	case backupListView, restoreConfirmationView, deleteConfirmationView:
		body.WriteString(m.viewBackup())
	case deletingView:
		body.WriteString(m.viewDeletingView())
	case createBackupView:
		body.WriteString(m.viewCreateBackup())
	case settingsView:
		body.WriteString(m.viewSettings())
	case changeSavePathView:
		body.WriteString(m.viewChangeSavePath())
	case changeBackupDirView:
		body.WriteString(m.viewChangeBackupDir())
	case firstRunView:
		body.WriteString(m.viewFirstRun())
	default:
		body.WriteString("Unknown view. This is a bug!")
	}

	// Add help text at the bottom.
	var help string
	switch m.state {
	case firstRunView:
		help = m.styles.Help.Render("Press 'enter' to confirm.")
	case mainMenuView:
		help = m.styles.Help.Render("Press 'ctrl+c' to quit.")
	case deletingView:
		help = m.styles.Help.Render("space: toggle, →: select all, ←: deselect all, enter: confirm, q: back")
	case initializingView:
		help = "" // No help text during initialization
	default:
		help = m.styles.Help.Render("Press 'q' to return to menu, 'ctrl+c' to quit.")
	}

	// Combine title, body, and help text.
	title := m.styles.Title.Render("Game Save Backup Manager")
	return fmt.Sprintf("%s\n\n%s\n\n%s", title, body.String(), help)
}

// initDB is a command to initialize the database.
func (m *model) initDB() tea.Msg {
	if m.config == nil || m.config.BackupDir == "" {
		return fmt.Errorf("configuration is missing or invalid")
	}

	db, err := backup.InitDB(m.config.BackupDir)
	if err != nil {
		return err // Return the error as a message.
	}
	m.db = db

	return dbInitializedMsg{} // Signal that the DB is ready.
}

// A helper to create the list of backups.
func (m *model) refreshBackupList(title string) tea.Cmd {
	m.list.Title = title
	items, err := m.getBackupItems()
	if err != nil {
		return func() tea.Msg { return err }
	}
	m.list.SetItems(items)
	m.list.SetSize(m.width, m.height-6)
	return nil
}

// listItem implements the list.Item interface.
type listItem backup.Backup

func (i listItem) Title() string       { return i.Name }
func (i listItem) Description() string { return i.CreatedAt.Format("2006-01-02 15:04:05") }
func (i listItem) FilterValue() string { return i.Name }

// getBackupItems fetches backups from the database and converts them to list.Item.
func (m *model) getBackupItems() ([]list.Item, error) {
	if m.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	backups, err := m.db.GetBackups()
	if err != nil {
		return nil, err
	}
	items := make([]list.Item, len(backups))
	for i, b := range backups {
		items[i] = listItem(b)
	}
	return items, nil
}

// --- Main Menu ---

func (m *model) updateMainMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1":
			m.state = createBackupView
			m.textInput.Placeholder = "My awesome backup"
			m.textInput.SetValue("") // Clear previous input
			m.textInput.Focus()
			m.textInput.CharLimit = 156
			m.textInput.Width = m.width - 10 // Scale with window width
			if m.textInput.Width < 20 {
				m.textInput.Width = 20
			}
			return m, nil
		case "2":
			m.state = backupListView
			// Switch back to normal delegate (no checkboxes)
			m.list.SetDelegate(newNormalItemDelegate())
			// Reset list selection to first item
			cmd := m.refreshBackupList("Select a backup to restore")
			m.list.Select(0)
			return m, cmd
		case "3":
			m.state = backupListView
			// Switch back to normal delegate (no checkboxes)
			m.list.SetDelegate(newNormalItemDelegate())
			// Reset list selection to first item
			cmd := m.refreshBackupList("Available Backups")
			m.list.Select(0)
			return m, cmd
		case "4":
			m.state = deletingView
			// Clear previous selections when entering delete view
			m.selected = make(map[int]struct{})
			// Switch to selection delegate for checkboxes
			m.list.SetDelegate(newSelectItemDelegate(m.selected))
			// Reset list selection to first item
			cmd := m.refreshBackupList("Select backups to delete")
			m.list.Select(0)
			return m, cmd
		case "5":
			m.state = settingsView
			return m, nil
		}
	}
	return m, nil
}

func (m *model) viewMainMenu() string {
	return "What would you like to do?\n\n" +
		"1. Create Backup\n" +
		"2. Restore Backup\n" +
		"3. List Backups\n" +
		"4. Delete Backups\n" +
		"5. Settings"
}

// --- Backup Views ---

func (m *model) updateBackupViews(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			selectedItem := m.list.SelectedItem()
			if selectedItem == nil {
				return m, nil
			}
			switch m.list.Title {
			case "Select a backup to restore":
				m.state = restoreConfirmationView
			case "Select backups to delete":
				m.state = deleteConfirmationView
			}
			return m, nil

		case "y", "Y":
			if m.state == restoreConfirmationView {
				selectedItem := m.list.SelectedItem()
				if selectedItem != nil {
					b := selectedItem.(listItem)
					err := m.db.RestoreBackup(backup.Backup(b), m.config.SavePath)
					if err != nil {
						m.err = err
						return m, nil
					} else {
						m.state = mainMenuView
						return m, m.showNotification("Backup restored successfully!")
					}
				}
			}
			if m.state == deleteConfirmationView {
				var backupsToDelete []backup.Backup
				if len(m.selected) > 0 {
					for i, item := range m.list.Items() {
						if _, ok := m.selected[i]; ok {
							backupsToDelete = append(backupsToDelete, backup.Backup(item.(listItem)))
						}
					}
				} else {
					selectedItem := m.list.SelectedItem()
					if selectedItem != nil {
						backupsToDelete = append(backupsToDelete, backup.Backup(selectedItem.(listItem)))
					}
				}

				if len(backupsToDelete) > 0 {
					err := m.db.DeleteBackups(backupsToDelete)
					if err != nil {
						m.err = err
						return m, nil
					} else {
						m.state = mainMenuView
						return m, m.showNotification(fmt.Sprintf("%d backup(s) deleted successfully!", len(backupsToDelete)))
					}
				}
			}
		case "n", "N", "esc":
			if m.state == deleteConfirmationView {
				m.state = deletingView
			} else {
				m.state = backupListView
			}
			return m, nil
		}
	}
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *model) viewBackup() string {
	if m.state == restoreConfirmationView {
		selectedItem := m.list.SelectedItem()
		if selectedItem != nil {
			b := selectedItem.(listItem)
			return fmt.Sprintf("Are you sure you want to restore this backup?\n\n%s\n\n(y/n)", b.Name)
		}
	}
	if m.state == deleteConfirmationView {
		if len(m.selected) > 0 {
			return fmt.Sprintf("Are you sure you want to delete %d selected backups?\n\n(y/n)", len(m.selected))
		}
		selectedItem := m.list.SelectedItem()
		if selectedItem != nil {
			b := selectedItem.(listItem)
			return fmt.Sprintf("Are you sure you want to delete this backup?\n\n%s\n\n(y/n)", b.Name)
		}
	}
	return m.list.View()
}

// --- Deleting View ---
func (m *model) updateDeletingView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case " ":
			// Toggle selection.
			if i, ok := m.list.SelectedItem().(listItem); ok {
				index := m.list.Index()
				if _, ok := m.selected[index]; ok {
					delete(m.selected, index)
				} else {
					m.selected[index] = struct{}{}
				}
				m.list.SetItem(index, i)
			}
		case "right":
			// Select all.
			for i := range m.list.Items() {
				m.selected[i] = struct{}{}
			}
			// Force a re-render of the list.
			return m, m.list.SetItems(m.list.Items())
		case "left":
			// Deselect all.
			for k := range m.selected {
				delete(m.selected, k)
			}
			// Force a re-render of the list.
			return m, m.list.SetItems(m.list.Items())
		case "enter":
			if len(m.selected) > 0 {
				m.state = deleteConfirmationView
			}
		}
	}
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *model) viewDeletingView() string {
	return m.list.View()
}

// --- Create Backup ---

func (m *model) updateCreateBackup(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			backupName := m.textInput.Value()
			err := m.db.CreateBackup(m.config.SavePath, m.config.BackupDir, backupName)
			if err != nil {
				m.err = err
				return m, nil
			} else {
				m.state = mainMenuView
				return m, m.showNotification("Backup created successfully!")
			}
		case "esc":
			m.state = mainMenuView
			return m, nil
		}
	}

	return m, cmd
}

func (m *model) viewCreateBackup() string {
	return fmt.Sprintf(
		"Enter a name for your backup (optional):\n\n%s\n\n%s",
		m.textInput.View(),
		"(esc to cancel)",
	)
}

// --- Settings ---

func (m *model) updateSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1":
			m.state = changeSavePathView
			m.textInput.Placeholder = "Enter new save path"
			m.textInput.SetValue("") // Clear previous input
			m.textInput.Focus()
			m.textInput.Width = m.width - 10 // Scale with window width
			if m.textInput.Width < 20 {
				m.textInput.Width = 20
			}
			return m, nil
		case "2":
			m.state = changeBackupDirView
			m.textInput.Placeholder = "Enter new backup directory"
			m.textInput.SetValue("") // Clear previous input
			m.textInput.Focus()
			m.textInput.Width = m.width - 10 // Scale with window width
			if m.textInput.Width < 20 {
				m.textInput.Width = 20
			}
			return m, nil
		case "3":
			m.config.AutoBackup = !m.config.AutoBackup
			if err := m.config.Save(); err != nil {
				m.err = err
			} else {
				return m, m.showNotification("Auto-backup toggled successfully!")
			}
			return m, nil
		case "esc":
			m.state = mainMenuView
			return m, nil
		}
	}
	return m, nil
}

func (m *model) viewSettings() string {
	autoBackupStatus := "Disabled"
	if m.config.AutoBackup {
		autoBackupStatus = "Enabled"
	}
	return fmt.Sprintf(
		"Settings:\n\n1. Change Save File Path\n   (current: %s)\n\n2. Change Backup Directory\n   (current: %s)\n\n3. Toggle Auto-Backup on Restore\n   (current: %s)",
		m.config.SavePath, m.config.BackupDir, autoBackupStatus,
	)
}

// --- Change Settings ---

func (m *model) updateChangeSavePath(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			newPath := m.textInput.Value()
			m.config.SavePath = newPath
			if err := m.config.Save(); err != nil {
				m.err = err
			} else {
				return m, m.showNotification("Save path updated successfully!")
			}
			m.state = settingsView
			return m, nil
		case "esc":
			m.state = settingsView
			return m, nil
		}
	}

	return m, cmd
}

func (m *model) viewChangeSavePath() string {
	return fmt.Sprintf(
		"Enter the new save file path:\n\n%s\n\n%s",
		m.textInput.View(),
		"(esc to cancel)",
	)
}

func (m *model) updateChangeBackupDir(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			newDir := m.textInput.Value()
			m.config.BackupDir = newDir
			if err := m.config.Save(); err != nil {
				m.err = err
			} else {
				return m, m.showNotification("Backup directory updated successfully!")
			}
			m.state = settingsView
			return m, nil
		case "esc":
			m.state = settingsView
			return m, nil
		}
	}

	return m, cmd
}

func (m *model) viewChangeBackupDir() string {
	return fmt.Sprintf(
		"Enter the new backup directory:\n\n%s\n\n%s",
		m.textInput.View(),
		"(esc to cancel)",
	)
}

// --- First Run ---

func (m *model) updateFirstRun(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.err = nil // Clear error on any key press
	m.textInput, cmd = m.textInput.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.config.SavePath == "" {
				m.config.SavePath = m.textInput.Value()
				m.textInput.Placeholder = "Enter your backup directory"
				m.textInput.SetValue("")
			} else {
				m.config.BackupDir = m.textInput.Value()
				if err := validation.IsWritable(m.config.BackupDir); err != nil {
					m.err = err
					m.config.BackupDir = "" // Reset so we ask again
					m.textInput.SetValue("")
					m.textInput.Placeholder = "Invalid path. Try again."
					return m, nil
				}

				if err := m.config.Save(); err != nil {
					m.err = err
					return m, nil
				}

				// Configuration is done, transition to the main menu.
				m.state = initializingView
				m.message = "Configuration saved successfully!"
				return m, m.initDB // Initialize the database
			}
		}
	}

	return m, cmd
}

func (m *model) viewFirstRun() string {
	prompt := "Enter your game's save file path:"
	if m.config.SavePath != "" {
		prompt = "Enter your backup directory:"
	}
	return fmt.Sprintf(
		"Welcome to the Game Save Backup Manager!\n\nPlease configure your paths.\n\n%s\n\n%s",
		prompt,
		m.textInput.View(),
	)
}

// --- itemDelegate ---

type itemDelegate struct {
	list.DefaultDelegate
	selected map[int]struct{}
}

type normalItemDelegate struct {
	list.DefaultDelegate
}

// newNormalItemDelegate creates a delegate for normal list views (no checkboxes)
func newNormalItemDelegate() *normalItemDelegate {
	d := &normalItemDelegate{}
	d.Styles = list.NewDefaultItemStyles()
	
	// Enhanced styling with boxes
	d.Styles.SelectedTitle = tui.DefaultStyles().Selected.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1)
	d.Styles.SelectedDesc = tui.DefaultStyles().Selected.Copy().Faint(true).Padding(0, 1)
	d.Styles.NormalTitle = tui.DefaultStyles().ListItem.Padding(0, 1)
	d.Styles.NormalDesc = tui.DefaultStyles().ListItem.Copy().Faint(true).Padding(0, 1)
	return d
}

// newSelectItemDelegate creates a delegate for delete view (with checkboxes)
func newSelectItemDelegate(selected map[int]struct{}) *itemDelegate {
	d := &itemDelegate{
		selected: selected,
	}
	d.Styles = list.NewDefaultItemStyles()
	
	// Enhanced styling with boxes for selection view
	d.Styles.SelectedTitle = tui.DefaultStyles().Selected.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1)
	d.Styles.SelectedDesc = tui.DefaultStyles().Selected.Copy().Faint(true).Padding(0, 1)
	d.Styles.NormalTitle = tui.DefaultStyles().ListItem.Padding(0, 1)
	d.Styles.NormalDesc = tui.DefaultStyles().ListItem.Copy().Faint(true).Padding(0, 1)
	return d
}

// Render method for normalItemDelegate (no checkboxes)
func (d *normalItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(listItem)
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
			Padding(0, 1).
			Render(content)
		fmt.Fprint(w, styledContent)
	} else {
		// Normal styling without border
		styledContent := lipgloss.NewStyle().
			Padding(0, 1).
			Render(content)
		fmt.Fprint(w, styledContent)
	}
}

// Render method for itemDelegate (with checkboxes for delete view)
func (d *itemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(listItem)
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

	// Combine checkbox, title and description with better formatting
	content := fmt.Sprintf("%s  %s\n    %s", checkbox, title, desc)

	if m.Index() == index {
		// Apply border to the entire content (checkbox + title + description)
		styledContent := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1).
			Render(content)
		fmt.Fprint(w, styledContent)
	} else {
		// Normal styling without border
		styledContent := lipgloss.NewStyle().
			Padding(0, 1).
			Render(content)
		fmt.Fprint(w, styledContent)
	}
}

// Set a message to be displayed to the user with auto-clear timer.
func (m *model) setMessage(msg string) {
	m.message = msg
}

// showNotification sets a message and returns a command to clear it after 2 seconds
func (m *model) showNotification(msg string) tea.Cmd {
	m.message = msg
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return clearNotificationMsg{}
	})
}