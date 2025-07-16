package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
	list := list.New(nil, newItemDelegate(selected), 0, 0)
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
		m.textInput.Width = 50
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
		m.list.SetSize(m.width, m.height-4) // Adjust for title/help text.
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
			return m, nil
		}

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
		// This is a transient message. We show it, then clear it.
		msg := m.message
		m.message = "" // Clear after rendering once.
		// We still render the underlying view, with the message on top.
		return m.styles.Success.Render(msg) + "\n" + m.currentView()
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
		help = m.styles.Help.Render("space: toggle, a: select all, n: deselect all, enter: confirm, q: back")
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
	m.list.SetSize(m.width, m.height-4)
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
			m.textInput.Focus()
			m.textInput.CharLimit = 156
			m.textInput.Width = 50
			return m, nil
		case "2":
			m.state = backupListView
			return m, m.refreshBackupList("Select a backup to restore")
		case "3":
			m.state = backupListView
			return m, m.refreshBackupList("Available Backups")
		case "4":
			m.state = deletingView
			return m, m.refreshBackupList("Select backups to delete")
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
					} else {
						m.message = "Backup restored successfully!"
					}
				}
				m.state = mainMenuView
				return m, nil
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
					} else {
						m.message = fmt.Sprintf("%d backup(s) deleted successfully!", len(backupsToDelete))
					}
				}
				m.state = mainMenuView
				return m, nil
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
		case "a":
			// Select all.
			for i := range m.list.Items() {
				m.selected[i] = struct{}{}
			}
			// Force a re-render of the list.
			return m, m.list.SetItems(m.list.Items())
		case "n":
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
			} else {
				m.message = "Backup created successfully!"
			}
			m.state = mainMenuView
			return m, nil
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
			m.textInput.Focus()
			m.textInput.Width = 50
			return m, nil
		case "2":
			m.state = changeBackupDirView
			m.textInput.Placeholder = "Enter new backup directory"
			m.textInput.Focus()
			m.textInput.Width = 50
			return m, nil
		case "3":
			m.config.AutoBackup = !m.config.AutoBackup
			if err := m.config.Save(); err != nil {
				m.err = err
			} else {
				m.message = "Auto-backup toggled successfully!"
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
				m.message = "Save path updated successfully!"
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
				m.message = "Backup directory updated successfully!"
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

func newItemDelegate(selected map[int]struct{}) *itemDelegate {
	d := &itemDelegate{
		selected: selected,
	}
	d.Styles = list.NewDefaultItemStyles()
	d.Styles.SelectedTitle = tui.DefaultStyles().Selected
	d.Styles.SelectedDesc = tui.DefaultStyles().Selected.Copy().Faint(true)
	d.Styles.NormalTitle = tui.DefaultStyles().ListItem
	d.Styles.NormalDesc = tui.DefaultStyles().ListItem.Copy().Faint(true)
	return d
}

func (d *itemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(listItem)
	if !ok {
		return
	}

	var checkbox string
	if _, ok := d.selected[index]; ok {
		checkbox = "[x] "
	} else {
		checkbox = "[ ] "
	}

	title := i.Title()
	desc := i.Description()

	if m.Index() == index {
		title = d.Styles.SelectedTitle.Render(checkbox + title)
		desc = d.Styles.SelectedDesc.Render("  " + desc)
	} else {
		title = d.Styles.NormalTitle.Render(checkbox + title)
		desc = d.Styles.NormalDesc.Render("  " + desc)
	}

	fmt.Fprint(w, title+"\n"+desc)
}

// Set a message to be displayed to the user.
func (m *model) setMessage(msg string) {
	m.message = msg
}