package app

import (
	"fmt"
	"time"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gemini/game-save-backup-manager-reimagined/internal/backup"
	"github.com/gemini/game-save-backup-manager-reimagined/internal/components"
	"github.com/gemini/game-save-backup-manager-reimagined/internal/config"
	"github.com/gemini/game-save-backup-manager-reimagined/internal/layout"
	"github.com/gemini/game-save-backup-manager-reimagined/internal/services"
	"github.com/gemini/game-save-backup-manager-reimagined/internal/state"
	"github.com/gemini/game-save-backup-manager-reimagined/internal/tui"
)

// Application represents the main application
type Application struct {
	// Core components
	stateManager        *state.StateManager
	backupService       *services.BackupService
	notificationManager *components.NotificationManager
	
	// UI components
	styles    *tui.Styles
	list      list.Model
	textInput textinput.Model
	
	// Configuration and state
	config   *config.Config
	selected map[int]struct{}
	
	// Window dimensions
	width  int
	height int
	
	// Error state
	err error
}

// NewApplication creates a new application instance
func NewApplication(cfg *config.Config, isFirstRun bool) *Application {
	// Initialize core components
	var initialState state.ViewState
	if isFirstRun {
		initialState = state.FirstRunView
	} else {
		initialState = state.InitializingView
	}
	
	stateManager := state.NewStateManager(initialState)
	backupService := services.NewBackupService(nil, cfg)
	notificationManager := components.NewNotificationManager()
	
	// Initialize UI components
	styles := tui.DefaultStyles()
	list := list.New(nil, components.NewNormalItemDelegate(), 0, 0)
	textInput := textinput.New()
	selected := make(map[int]struct{})
	
	// Configure text input for first run
	if isFirstRun {
		textInput.Placeholder = "Enter your game's save file path"
		textInput.Focus()
		textInput.Width = 50 // Will be updated on first WindowSizeMsg
	}
	
	return &Application{
		stateManager:        stateManager,
		backupService:       backupService,
		notificationManager: notificationManager,
		styles:              styles,
		list:                list,
		textInput:           textInput,
		config:              cfg,
		selected:            selected,
	}
}

// Init initializes the application
func (app *Application) Init() tea.Cmd {
	if app.stateManager.IsInState(state.InitializingView) {
		return app.initializeDatabase
	}
	return nil
}

// GetCurrentState returns the current application state
func (app *Application) GetCurrentState() state.ViewState {
	return app.stateManager.Current()
}

// TransitionToState changes the application state
func (app *Application) TransitionToState(newState state.ViewState) {
	app.stateManager.TransitionTo(newState)
}

// ShowNotification displays a notification message
func (app *Application) ShowNotification(message string) tea.Cmd {
	return app.notificationManager.Show(message)
}

// HasNotification returns true if there's a notification to display
func (app *Application) HasNotification() bool {
	return app.notificationManager.HasMessage()
}

// GetNotificationMessage returns the current notification message
func (app *Application) GetNotificationMessage() string {
	return app.notificationManager.GetMessage()
}

// ClearNotification clears the current notification
func (app *Application) ClearNotification() {
	app.notificationManager.Clear()
}

// SetError sets an error state
func (app *Application) SetError(err error) {
	app.err = err
}

// GetError returns the current error
func (app *Application) GetError() error {
	return app.err
}

// HasError returns true if there's an error
func (app *Application) HasError() bool {
	return app.err != nil
}

// UpdateWindowSize updates the window dimensions and adjusts UI components
func (app *Application) UpdateWindowSize(width, height int) {
	app.width = width
	app.height = height
	
	// Update list size using layout constants
	listHeight := layout.CalculateListHeight(height)
	app.list.SetSize(width, listHeight)
	
	// Update text input width using layout constants
	app.textInput.Width = layout.CalculateInputWidth(width)
}

// GetWindowDimensions returns current window dimensions
func (app *Application) GetWindowDimensions() (int, int) {
	return app.width, app.height
}

// RefreshBackupList refreshes the backup list with the given title
func (app *Application) RefreshBackupList(title string) tea.Cmd {
	app.list.Title = title
	items, err := app.backupService.GetBackupItems()
	if err != nil {
		return func() tea.Msg { return err }
	}
	app.list.SetItems(items)
	
	// Update list size using current window dimensions
	listHeight := layout.CalculateListHeight(app.height)
	app.list.SetSize(app.width, listHeight)
	return nil
}

// initializeDatabase initializes the backup database
func (app *Application) initializeDatabase() tea.Msg {
	err := app.backupService.InitializeDatabase()
	if err != nil {
		return err
	}
	
	// Return a message indicating database is ready
	return DatabaseInitializedMsg{}
}

// SetTextInputPlaceholder sets the text input placeholder
func (app *Application) SetTextInputPlaceholder(placeholder string) {
	app.textInput.Placeholder = placeholder
}

// ClearTextInput clears the text input value
func (app *Application) ClearTextInput() {
	app.textInput.SetValue("")
}

// FocusTextInput focuses the text input
func (app *Application) FocusTextInput() {
	app.textInput.Focus()
}

// SetTextInputCharLimit sets the character limit for text input
func (app *Application) SetTextInputCharLimit(limit int) {
	app.textInput.CharLimit = limit
}

// SetListDelegate sets the list delegate
func (app *Application) SetListDelegate(delegate list.ItemDelegate) {
	app.list.SetDelegate(delegate)
}

// ResetListSelection resets the list selection to the first item
func (app *Application) ResetListSelection() {
	app.list.Select(0)
}

// ClearSelections clears all selections
func (app *Application) ClearSelections() {
	app.selected = make(map[int]struct{})
}

// GetSelections returns the current selections
func (app *Application) GetSelections() map[int]struct{} {
	return app.selected
}

// GetStyles returns the application styles
func (app *Application) GetStyles() *tui.Styles {
	return app.styles
}

// DatabaseInitializedMsg indicates the database is ready
type DatabaseInitializedMsg struct{}

// GetTextInput returns the text input component
func (app *Application) GetTextInput() *textinput.Model {
	return &app.textInput
}

// GetList returns the list component
func (app *Application) GetList() *list.Model {
	return &app.list
}

// tempSavePath holds the save path temporarily during first-run setup
var tempSavePath string

// SetTempSavePath temporarily stores the save path during first-run setup
func (app *Application) SetTempSavePath(path string) {
	tempSavePath = path
}

// GetTempSavePath returns the temporarily stored save path
func (app *Application) GetTempSavePath() string {
	return tempSavePath
}

// ClearTempSavePath clears the temporarily stored save path
func (app *Application) ClearTempSavePath() {
	tempSavePath = ""
}

// GetConfig returns the application configuration
func (app *Application) GetConfig() *config.Config {
	return app.config
}

// CreateBackup creates a new backup with the given name
func (app *Application) CreateBackup(name string) error {
	return app.backupService.CreateBackup(name)
}

// RestoreSelectedBackup restores the currently selected backup
func (app *Application) RestoreSelectedBackup() error {
	selectedIndex := app.list.Index()
	items := app.list.Items()
	
	if selectedIndex >= len(items) {
		return fmt.Errorf("no backup selected")
	}
	
	// Convert the selected item to a backup
	if listItem, ok := items[selectedIndex].(components.ListItem); ok {
		backup := backup.Backup(listItem)
		return app.backupService.RestoreBackup(backup)
	}
	
	return fmt.Errorf("invalid backup selection")
}

// DeleteSelectedBackups deletes the currently selected backups
func (app *Application) DeleteSelectedBackups() error {
	items := app.list.Items()
	selectedBackups := app.backupService.GetSelectedBackups(items, app.selected)
	
	if len(selectedBackups) == 0 {
		return fmt.Errorf("no backups selected for deletion")
	}
	
	return app.backupService.DeleteBackups(selectedBackups)
}

// UpdateSavePath updates the save path in the configuration
func (app *Application) UpdateSavePath(newPath string) error {
	app.config.SavePath = newPath
	return app.config.Save()
}

// UpdateBackupDir updates the backup directory in the configuration
func (app *Application) UpdateBackupDir(newDir string) error {
	app.config.BackupDir = newDir
	return app.config.Save()
}

// ToggleAutoBackup toggles the auto-backup setting
func (app *Application) ToggleAutoBackup() error {
	app.config.AutoBackup = !app.config.AutoBackup
	return app.config.Save()
}

// RestoreSelectedBackupWithAutoBackup restores the selected backup with optional auto-backup
func (app *Application) RestoreSelectedBackupWithAutoBackup() error {
	selectedIndex := app.list.Index()
	items := app.list.Items()
	
	if selectedIndex >= len(items) {
		return fmt.Errorf("no backup selected")
	}
	
	// Convert the selected item to a backup
	if listItem, ok := items[selectedIndex].(components.ListItem); ok {
		backupToRestore := backup.Backup(listItem)
		
		// Create auto-backup before restoring if enabled
		if app.config.AutoBackup {
			autoBackupName := fmt.Sprintf("Backup_%s", time.Now().Format("2006-01-02_15-04-05"))
			if err := app.backupService.CreateBackup(autoBackupName); err != nil {
				return fmt.Errorf("failed to create auto-backup: %v", err)
			}
		}
		
		return app.backupService.RestoreBackup(backupToRestore)
	}
	
	return fmt.Errorf("invalid backup selection")
}