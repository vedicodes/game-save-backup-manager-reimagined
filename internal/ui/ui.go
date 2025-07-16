package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gemini/game-save-backup-manager-reimagined/internal/backup"
	"github.com/gemini/game-save-backup-manager-reimagined/internal/config"
	"github.com/gemini/game-save-backup-manager-reimagined/internal/tui"
	"github.com/gemini/game-save-backup-manager-reimagined/internal/validation"
)

type viewState int

const (
	mainMenuView viewState = iota
	backupListView
	createBackupView
	deleteBackupView
	settingsView
	firstRunView
	restoreConfirmationView
	deleteConfirmationView
	changeSavePathView
	changeBackupDirView
)

type model struct {
	styles    *tui.Styles
	state     viewState
	list      list.Model
	textInput textinput.Model
	config    *config.Config
	db        *backup.DB
	err       error
	message   string
}

func NewModel(cfg *config.Config, db *backup.DB, isFirstRun bool) *model {
	styles := tui.DefaultStyles()
	m := &model{
		styles: styles,
		config: cfg,
		db:     db,
	}
	if isFirstRun {
		m.state = firstRunView
		m.textInput = textinput.New()
		m.textInput.Placeholder = "Enter your game's save file path"
		m.textInput.Focus()
		m.textInput.Width = 50
	} else {
		m.state = mainMenuView
	}
	return m
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	switch m.state {
	case mainMenuView:
		return m.updateMainMenu(msg)
	case backupListView:
		return m.updateBackupList(msg)
	case createBackupView:
		return m.updateCreateBackup(msg)
	case restoreConfirmationView:
		return m.updateRestoreConfirmation(msg)
	case deleteConfirmationView:
		return m.updateDeleteConfirmation(msg)
	case settingsView:
		return m.updateSettings(msg)
	case changeSavePathView:
		return m.updateChangeSavePath(msg)
	case changeBackupDirView:
		return m.updateChangeBackupDir(msg)
	case firstRunView:
		return m.updateFirstRun(msg)
	// Add other view updates here
	default:
		return m, nil
	}
}

func (m *model) View() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("Error: %v", m.err))
	}
	if m.message != "" {
		// A bit of a hack to show a message and then return to the main menu
		msg := m.message
		m.message = ""
		return msg
	}

	switch m.state {
	case mainMenuView:
		return m.viewMainMenu()
	case backupListView:
		return m.viewBackupList()
	case createBackupView:
		return m.viewCreateBackup()
	case restoreConfirmationView:
		return m.viewRestoreConfirmation()
	case deleteConfirmationView:
		return m.viewDeleteConfirmation()
	case settingsView:
		return m.viewSettings()
	case changeSavePathView:
		return m.viewChangeSavePath()
	case changeBackupDirView:
		return m.viewChangeBackupDir()
	case firstRunView:
		return m.viewFirstRun()
	// Add other views here
	default:
		return "Unknown view"
	}
}

func (m *model) updateMainMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1":
			m.state = createBackupView
			m.textInput = textinput.New()
			m.textInput.Placeholder = "My awesome backup"
			m.textInput.Focus()
			m.textInput.CharLimit = 156
			m.textInput.Width = 20
			return m, nil
		case "2":
			m.state = backupListView
			m.list.Title = "Select a backup to restore"
			items, err := m.getBackupItems()
			if err != nil {
				m.err = err
				return m, nil
			}
			m.list = list.New(items, list.NewDefaultDelegate(), 0, 0)
			return m, nil
		case "3":
			m.state = backupListView
			items, err := m.getBackupItems()
			if err != nil {
				m.err = err
				return m, nil
			}
			m.list = list.New(items, list.NewDefaultDelegate(), 0, 0)
			m.list.Title = "Available Backups"
			return m, nil
		case "4":
			m.state = backupListView
			m.list.Title = "Select backups to delete"
			items, err := m.getBackupItems()
			if err != nil {
				m.err = err
				return m, nil
			}
			m.list = list.New(items, list.NewDefaultDelegate(), 0, 0)
			return m, nil
		case "5":
			m.state = settingsView
			return m, nil
		}
	}
	return m, nil
}

func (m *model) updateBackupList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.list.Title == "Select a backup to restore" {
				m.state = restoreConfirmationView
				return m, nil
			}
			if m.list.Title == "Select backups to delete" {
				m.state = deleteConfirmationView
				return m, nil
			}
			m.state = mainMenuView
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *model) viewMainMenu() string {
	var b strings.Builder
	b.WriteString(m.styles.Title.Render("Game Save Backup Manager"))
	b.WriteString("\n")
	b.WriteString(m.styles.Subtitle.Render("What would you like to do?"))
	b.WriteString("\n\n")
	b.WriteString("1. Create Backup\n")
	b.WriteString("2. Restore Backup\n")
	b.WriteString("3. List Backups\n")
	b.WriteString("4. Delete Backups\n")
	b.WriteString("5. Settings\n")
	b.WriteString("\n")
	b.WriteString(m.styles.Help.Render("Press 'q' to quit."))
	return b.String()
}

func (m *model) viewBackupList() string {
	return m.list.View()
}

type listItem backup.Backup

func (i listItem) Title() string       { return i.Name }
func (i listItem) Description() string { return i.CreatedAt.Format("2006-01-02 15:04:05") }
func (i listItem) FilterValue() string { return i.Name }

func (m *model) getBackupItems() ([]list.Item, error) {
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
				m.message = m.styles.Success.Render("Backup created successfully!")
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
		"Enter a name for your backup:\n\n%s\n\n%s",
		m.textInput.View(),
		"(esc to cancel)",
	)
}

func (m *model) updateRestoreConfirmation(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			selectedItem := m.list.SelectedItem()
			if selectedItem != nil {
				b := selectedItem.(listItem)
				err := m.db.RestoreBackup(backup.Backup(b), m.config.SavePath)
				if err != nil {
					m.err = err
				} else {
					m.message = m.styles.Success.Render("Backup restored successfully!")
				}
			}
			m.state = mainMenuView
			return m, nil
		case "n", "N", "esc":
			m.state = mainMenuView
			return m, nil
		}
	}
	return m, nil
}

func (m *model) viewRestoreConfirmation() string {
	selectedItem := m.list.SelectedItem()
	if selectedItem == nil {
		return "No backup selected."
	}
	b := selectedItem.(listItem)
	return fmt.Sprintf(
		"Are you sure you want to restore this backup?\n\n%s\n\n%s",
		b.Name,
		"(y/n)",
	)
}

func (m *model) updateDeleteConfirmation(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			selectedItem := m.list.SelectedItem()
			if selectedItem != nil {
				b := selectedItem.(listItem)
				err := m.db.DeleteBackup(backup.Backup(b))
				if err != nil {
					m.err = err
				} else {
					m.message = m.styles.Success.Render("Backup deleted successfully!")
				}
			}
			m.state = mainMenuView
			return m, nil
		case "n", "N", "esc":
			m.state = mainMenuView
			return m, nil
		}
	}
	return m, nil
}

func (m *model) viewDeleteConfirmation() string {
	selectedItem := m.list.SelectedItem()
	if selectedItem == nil {
		return "No backup selected."
	}
	b := selectedItem.(listItem)
	return fmt.Sprintf(
		"Are you sure you want to delete this backup?\n\n%s\n\n%s",
		b.Name,
		"(y/n)",
	)
}

func (m *model) updateSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1":
			m.state = changeSavePathView
			m.textInput = textinput.New()
			m.textInput.Placeholder = "Enter new save path"
			m.textInput.Focus()
			m.textInput.Width = 50
			return m, nil
		case "2":
			m.state = changeBackupDirView
			m.textInput = textinput.New()
			m.textInput.Placeholder = "Enter new backup directory"
			m.textInput.Focus()
			m.textInput.Width = 50
			return m, nil
		case "3":
			m.config.AutoBackup = !m.config.AutoBackup
			if err := m.config.Save(); err != nil {
				m.err = err
			} else {
				m.message = m.styles.Success.Render("Auto-backup toggled successfully!")
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
	var b strings.Builder
	b.WriteString(m.styles.Title.Render("Settings"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("1. Change Save File Path (current: %s)\n", m.config.SavePath))
	b.WriteString(fmt.Sprintf("2. Change Backup Directory (current: %s)\n", m.config.BackupDir))
	b.WriteString(fmt.Sprintf("3. Toggle Auto-Backup on Restore (current: %v)\n", m.config.AutoBackup))
	b.WriteString("\n")
	b.WriteString(m.styles.Help.Render("Press 'esc' to return to the main menu."))
	return b.String()
}

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
				m.message = m.styles.Success.Render("Save path updated successfully!")
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
				m.message = m.styles.Success.Render("Backup directory updated successfully!")
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
				} else {
					m.message = m.styles.Success.Render("Configuration saved successfully!")
				}
				return m, tea.Quit
			}
		}
	}

	return m, cmd
}

func (m *model) viewFirstRun() string {
	return fmt.Sprintf(
		"Welcome to the Game Save Backup Manager!\n\nPlease configure your paths.\n\n%s\n\n%s",
		m.textInput.View(),
		"(press enter to confirm)",
	)
}

