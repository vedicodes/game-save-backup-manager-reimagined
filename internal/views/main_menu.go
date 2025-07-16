package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/vedicodes/game-save-backup-manager-reimagined/internal/app"
	"github.com/vedicodes/game-save-backup-manager-reimagined/internal/components"
	"github.com/vedicodes/game-save-backup-manager-reimagined/internal/state"
)

// MainMenuHandler handles the main menu view
type MainMenuHandler struct {
	app *app.Application
}

// NewMainMenuHandler creates a new main menu handler
func NewMainMenuHandler(app *app.Application) *MainMenuHandler {
	return &MainMenuHandler{app: app}
}

// Update handles main menu input and returns commands
func (h *MainMenuHandler) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1":
			return h.handleCreateBackup()
		case "2":
			return h.handleRestoreBackup()
		case "3":
			return h.handleListBackups()
		case "4":
			return h.handleDeleteBackups()
		case "5":
			return h.handleSettings()
		}
	}
	return nil
}

// View renders the main menu
func (h *MainMenuHandler) View() string {
	return "What would you like to do?\n\n" +
		"1. Create Backup\n" +
		"2. Restore Backup\n" +
		"3. List Backups\n" +
		"4. Delete Backups\n" +
		"5. Settings"
}

// handleCreateBackup transitions to create backup view
func (h *MainMenuHandler) handleCreateBackup() tea.Cmd {
	h.app.TransitionToState(state.CreateBackupView)
	
	// Configure text input
	h.app.SetTextInputPlaceholder("My awesome backup")
	h.app.ClearTextInput()
	h.app.FocusTextInput()
	h.app.SetTextInputCharLimit(156)
	
	return nil
}

// handleRestoreBackup transitions to restore backup view
func (h *MainMenuHandler) handleRestoreBackup() tea.Cmd {
	h.app.TransitionToState(state.BackupListView)
	h.app.SetListDelegate(components.NewNormalItemDelegate())
	cmd := h.app.RefreshBackupList("Select a backup to restore")
	h.app.ResetListSelection()
	return cmd
}

// handleListBackups transitions to list backups view (view-only)
func (h *MainMenuHandler) handleListBackups() tea.Cmd {
	h.app.TransitionToState(state.ViewBackupsView)
	h.app.SetListDelegate(components.NewNormalItemDelegate())
	cmd := h.app.RefreshBackupList("Available Backups")
	h.app.ResetListSelection()
	return cmd
}

// handleDeleteBackups transitions to delete backups view
func (h *MainMenuHandler) handleDeleteBackups() tea.Cmd {
	h.app.TransitionToState(state.DeletingView)
	h.app.ClearSelections()
	h.app.SetListDelegate(components.NewSelectableItemDelegate(h.app.GetSelections()))
	cmd := h.app.RefreshBackupList("Select backups to delete")
	h.app.ResetListSelection()
	return cmd
}

// handleSettings transitions to settings view
func (h *MainMenuHandler) handleSettings() tea.Cmd {
	h.app.TransitionToState(state.SettingsView)
	return nil
}