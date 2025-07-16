package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gemini/game-save-backup-manager-reimagined/internal/app"
	"github.com/gemini/game-save-backup-manager-reimagined/internal/components"
	"github.com/gemini/game-save-backup-manager-reimagined/internal/config"
	"github.com/gemini/game-save-backup-manager-reimagined/internal/state"
	"github.com/gemini/game-save-backup-manager-reimagined/internal/views"
)

// Controller is the main UI controller that orchestrates the application
type Controller struct {
	app *app.Application
	
	// View handlers
	mainMenuHandler *views.MainMenuHandler
}

// NewController creates a new UI controller
func NewController(cfg *config.Config, isFirstRun bool) *Controller {
	application := app.NewApplication(cfg, isFirstRun)
	
	controller := &Controller{
		app: application,
	}
	
	// Initialize view handlers
	controller.mainMenuHandler = views.NewMainMenuHandler(application)
	
	return controller
}

// Init initializes the controller
func (c *Controller) Init() tea.Cmd {
	return c.app.Init()
}

// Update handles all messages and routes them to appropriate handlers
func (c *Controller) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle global messages first
	if cmd := c.handleGlobalMessages(msg); cmd != nil {
		return c, cmd
	}
	
	// Don't process any messages if we are in an error state
	if c.app.HasError() {
		if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "ctrl+c" {
			return c, tea.Quit
		}
		return c, nil
	}
	
	// Route to appropriate view handler based on current state
	switch c.app.GetCurrentState() {
	case state.MainMenuView:
		cmd := c.mainMenuHandler.Update(msg)
		return c, cmd
	case state.InitializingView:
		// No updates while initializing
		return c, nil
	default:
		// For now, handle other states in the controller
		// TODO: Create separate handlers for other views
		return c.handleOtherViews(msg)
	}
}

// View renders the current view
func (c *Controller) View() string {
	// Always render errors first
	if c.app.HasError() {
		width, _ := c.app.GetWindowDimensions()
		errorStyle := c.app.GetStyles().Error.Width(width)
		return errorStyle.Render(fmt.Sprintf("A critical error occurred:\n\n%v\n\nPress Ctrl+C to exit.", c.app.GetError()))
	}
	
	// Render notifications if they exist
	if c.app.HasNotification() {
		return c.app.GetStyles().Success.Render(c.app.GetNotificationMessage()) + "\n" + c.currentView()
	}
	
	return c.currentView()
}

// handleGlobalMessages handles messages that apply to all views
func (c *Controller) handleGlobalMessages(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.app.UpdateWindowSize(msg.Width, msg.Height)
		return nil
		
	case tea.KeyMsg:
		// Global key bindings
		if msg.String() == "ctrl+c" {
			return tea.Quit
		}
		// Per-view key bindings for returning to main menu
		if c.shouldAllowQuitToMainMenu() && msg.String() == "q" {
			c.app.TransitionToState(state.MainMenuView)
			c.app.ClearNotification() // Clear any pending messages
			return nil
		}
		
	case components.ClearNotificationMsg:
		c.app.ClearNotification()
		return nil
		
	case app.DatabaseInitializedMsg:
		c.app.TransitionToState(state.MainMenuView)
		return nil
		
	case error:
		c.app.SetError(msg)
		return nil
	}
	
	return nil
}

// shouldAllowQuitToMainMenu determines if 'q' should return to main menu
func (c *Controller) shouldAllowQuitToMainMenu() bool {
	currentState := c.app.GetCurrentState()
	return currentState != state.FirstRunView && currentState != state.MainMenuView
}

// currentView returns the string for the current view
func (c *Controller) currentView() string {
	body := new(strings.Builder)
	
	switch c.app.GetCurrentState() {
	case state.InitializingView:
		body.WriteString("Initializing...")
	case state.FirstRunView:
		body.WriteString(c.renderFirstRunView())
	case state.FirstRunBackupDirView:
		body.WriteString(c.renderFirstRunBackupDirView())
	case state.MainMenuView:
		body.WriteString(c.mainMenuHandler.View())
	case state.CreateBackupView:
		body.WriteString(c.renderCreateBackupView())
	case state.BackupListView:
		body.WriteString(c.renderBackupListView())
	case state.ViewBackupsView:
		body.WriteString(c.renderViewBackupsView())
	case state.DeletingView:
		body.WriteString(c.renderDeletingView())
	case state.DeleteConfirmationView:
		body.WriteString(c.renderDeleteConfirmationView())
	case state.SettingsView:
		body.WriteString(c.renderSettingsView())
	case state.ChangeSavePathView:
		body.WriteString(c.renderChangeSavePathView())
	case state.ChangeBackupDirView:
		body.WriteString(c.renderChangeBackupDirView())
	default:
		// Fallback for any unhandled states
		body.WriteString("View not implemented yet")
	}
	
	// Get help text
	help := c.getHelpText()
	
	// Combine title and body
	title := c.app.GetStyles().Title.Render("Game Save Backup Manager")
	content := fmt.Sprintf("%s\n\n%s", title, body.String())
	
	// Calculate available space and position help at bottom
	_, windowHeight := c.app.GetWindowDimensions()
	contentLines := strings.Count(content, "\n") + 1
	helpLines := strings.Count(help, "\n") + 1
	
	// Add padding to push help to bottom
	availableLines := windowHeight - contentLines - helpLines - 2 // 2 for margins
	if availableLines > 0 {
		padding := strings.Repeat("\n", availableLines)
		return fmt.Sprintf("%s%s%s", content, padding, help)
	}
	
	// If no space available, just add help normally
	return fmt.Sprintf("%s\n\n%s", content, help)
}

// getHelpText returns appropriate help text for the current state
func (c *Controller) getHelpText() string {
	styles := c.app.GetStyles()
	
	switch c.app.GetCurrentState() {
	case state.FirstRunView:
		return styles.Help.Render("Press 'enter' to confirm.")
	case state.FirstRunBackupDirView:
		return styles.Help.Render("Press 'enter' to confirm.")
	case state.MainMenuView:
		return styles.Help.Render("Press 'ctrl+c' to quit.")
	case state.BackupListView:
		return styles.Help.Render("↑/↓: navigate, enter: restore backup, q: back")
	case state.ViewBackupsView:
		return styles.Help.Render("↑/↓: navigate, q: back")
	case state.DeletingView:
		return styles.Help.Render("space: toggle, →: select all, ←: deselect all, enter: confirm, q: back")
	case state.DeleteConfirmationView:
		return styles.Help.Render("y: confirm deletion, n/q: cancel")
	case state.SettingsView:
		return styles.Help.Render("1-3: select option, q: back")
	case state.CreateBackupView:
		return styles.Help.Render("enter: create backup (empty for auto-name), esc: cancel")
	case state.InitializingView:
		return "" // No help text during initialization
	default:
		return styles.Help.Render("Press 'q' to return to menu, 'ctrl+c' to quit.")
	}
}

// handleOtherViews handles views that don't have dedicated handlers yet
func (c *Controller) handleOtherViews(msg tea.Msg) (tea.Model, tea.Cmd) {
	currentState := c.app.GetCurrentState()
	
	// Handle text input views
	if c.isTextInputView(currentState) {
		return c.handleTextInputView(msg)
	}
	
	// Handle list views
	if c.isListView(currentState) {
		return c.handleListView(msg)
	}
	
	// Handle settings view
	if currentState == state.SettingsView {
		return c.handleSettingsView(msg)
	}
	
	// Handle delete confirmation view
	if currentState == state.DeleteConfirmationView {
		return c.handleDeleteConfirmationView(msg)
	}
	
	return c, nil
}

// renderFirstRunView renders the first run setup view
func (c *Controller) renderFirstRunView() string {
	width, _ := c.app.GetWindowDimensions()
	inputWidth := width - 8 // Leave some margin
	if inputWidth < 20 {
		inputWidth = 20 // Minimum width
	}
	
	inputStyle := c.app.GetStyles().TextInput.Width(inputWidth)
	
	return "Welcome to Game Save Backup Manager!\n\n" +
		"This appears to be your first time running the application.\n" +
		"Please enter the path to your game's save files:\n\n" +
		inputStyle.Render(c.app.GetTextInput().View())
}

// renderFirstRunBackupDirView renders the backup directory setup view
func (c *Controller) renderFirstRunBackupDirView() string {
	width, _ := c.app.GetWindowDimensions()
	inputWidth := width - 8 // Leave some margin
	if inputWidth < 20 {
		inputWidth = 20 // Minimum width
	}
	
	inputStyle := c.app.GetStyles().TextInput.Width(inputWidth)
	
	return "Setup Complete - Step 2 of 2\n\n" +
		"Now please enter the directory where you want to store your backups:\n" +
		"(This can be any folder on your computer)\n\n" +
		inputStyle.Render(c.app.GetTextInput().View())
}

// renderCreateBackupView renders the create backup view
func (c *Controller) renderCreateBackupView() string {
	width, _ := c.app.GetWindowDimensions()
	inputWidth := width - 8 // Leave some margin
	if inputWidth < 20 {
		inputWidth = 20 // Minimum width
	}
	
	inputStyle := c.app.GetStyles().TextInput.Width(inputWidth)
	
	return "Create a new backup\n\n" +
		"Enter a name for your backup:\n\n" +
		inputStyle.Render(c.app.GetTextInput().View())
}

// renderBackupListView renders the backup list view (for restoration)
func (c *Controller) renderBackupListView() string {
	return c.app.GetList().View()
}

// renderViewBackupsView renders the view-only backup list
func (c *Controller) renderViewBackupsView() string {
	return c.app.GetList().View()
}

// renderDeletingView renders the delete backups view
func (c *Controller) renderDeletingView() string {
	return c.app.GetList().View()
}

// renderDeleteConfirmationView renders the delete confirmation view
func (c *Controller) renderDeleteConfirmationView() string {
	selections := c.app.GetSelections()
	count := len(selections)
	
	if count == 0 {
		return "No backups selected for deletion.\n\nPress 'q' to go back."
	}
	
	return fmt.Sprintf("Delete Confirmation\n\n"+
		"Are you sure you want to delete %d backup(s)?\n"+
		"This action cannot be undone.\n\n"+
		"Press 'y' to confirm deletion\n"+
		"Press 'n' or 'q' to cancel", count)
}

// renderSettingsView renders the settings view
func (c *Controller) renderSettingsView() string {
	autoBackupStatus := "OFF"
	if c.app.GetConfig().AutoBackup {
		autoBackupStatus = "ON"
	}
	
	return "Settings\n\n" +
		"1. Change Save Path\n" +
		"2. Change Backup Directory\n" +
		"3. Auto-Backup Before Restore: " + autoBackupStatus
}

// renderChangeSavePathView renders the change save path view
func (c *Controller) renderChangeSavePathView() string {
	width, _ := c.app.GetWindowDimensions()
	inputWidth := width - 8 // Leave some margin
	if inputWidth < 20 {
		inputWidth = 20 // Minimum width
	}
	
	inputStyle := c.app.GetStyles().TextInput.Width(inputWidth)
	
	return "Change Save Path\n\n" +
		"Enter the new path to your game's save files:\n\n" +
		inputStyle.Render(c.app.GetTextInput().View())
}

// renderChangeBackupDirView renders the change backup directory view
func (c *Controller) renderChangeBackupDirView() string {
	width, _ := c.app.GetWindowDimensions()
	inputWidth := width - 8 // Leave some margin
	if inputWidth < 20 {
		inputWidth = 20 // Minimum width
	}
	
	inputStyle := c.app.GetStyles().TextInput.Width(inputWidth)
	
	return "Change Backup Directory\n\n" +
		"Enter the new backup directory path:\n\n" +
		inputStyle.Render(c.app.GetTextInput().View())
}

// isTextInputView checks if the current state uses text input
func (c *Controller) isTextInputView(currentState state.ViewState) bool {
	return currentState == state.FirstRunView ||
		currentState == state.FirstRunBackupDirView ||
		currentState == state.CreateBackupView ||
		currentState == state.ChangeSavePathView ||
		currentState == state.ChangeBackupDirView
}

// isListView checks if the current state uses list
func (c *Controller) isListView(currentState state.ViewState) bool {
	return currentState == state.BackupListView ||
		currentState == state.ViewBackupsView ||
		currentState == state.DeletingView
}

// handleTextInputView handles text input messages and updates
func (c *Controller) handleTextInputView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return c.handleTextInputSubmit()
		case "esc":
			// Cancel and go back to main menu
			c.app.TransitionToState(state.MainMenuView)
			return c, nil
		}
	}
	
	// Update the text input with the message
	textInput := c.app.GetTextInput()
	*textInput, cmd = textInput.Update(msg)
	
	return c, cmd
}

// handleTextInputSubmit handles when user presses enter in text input views
func (c *Controller) handleTextInputSubmit() (tea.Model, tea.Cmd) {
	currentState := c.app.GetCurrentState()
	inputValue := c.app.GetTextInput().Value()
	
	// Don't proceed if input is empty for required fields (not backup creation)
	if strings.TrimSpace(inputValue) == "" && currentState != state.CreateBackupView {
		return c, nil
	}
	
	switch currentState {
	case state.FirstRunView:
		// Save the save path temporarily and move to backup directory setup
		c.app.SetTempSavePath(inputValue)
		c.app.TransitionToState(state.FirstRunBackupDirView)
		c.app.SetTextInputPlaceholder("Enter backup directory path...")
		c.app.ClearTextInput()
		c.app.FocusTextInput()
		return c, nil
		
	case state.FirstRunBackupDirView:
		// Now we have both paths, save config and initialize
		return c.handleFirstRunComplete(inputValue)
		
	case state.CreateBackupView:
		// Create backup with the given name (or empty for auto-generated name)
		backupName := strings.TrimSpace(inputValue)
		if err := c.app.CreateBackup(backupName); err != nil {
			c.app.SetError(fmt.Errorf("failed to create backup: %v", err))
			return c, nil
		}
		var notificationCmd tea.Cmd
		if backupName == "" {
			notificationCmd = c.app.ShowNotification("Backup created successfully with auto-generated name")
		} else {
			notificationCmd = c.app.ShowNotification("Backup created successfully: " + backupName)
		}
		c.app.TransitionToState(state.MainMenuView)
		return c, notificationCmd
		
	case state.ChangeSavePathView:
		// Update save path
		if err := c.app.UpdateSavePath(inputValue); err != nil {
			c.app.SetError(fmt.Errorf("failed to update save path: %v", err))
			return c, nil
		}
		notificationCmd := c.app.ShowNotification("Save path updated: " + inputValue)
		c.app.TransitionToState(state.SettingsView)
		return c, notificationCmd
		
	case state.ChangeBackupDirView:
		// Update backup directory
		if err := c.app.UpdateBackupDir(inputValue); err != nil {
			c.app.SetError(fmt.Errorf("failed to update backup directory: %v", err))
			return c, nil
		}
		notificationCmd := c.app.ShowNotification("Backup directory updated: " + inputValue)
		c.app.TransitionToState(state.SettingsView)
		return c, notificationCmd
	}
	
	return c, nil
}

// handleListView handles list-based views
func (c *Controller) handleListView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return c.handleListSelection()
		case " ":
			if c.app.GetCurrentState() == state.DeletingView {
				return c.handleToggleSelection()
			}
		case "right", "→":
			if c.app.GetCurrentState() == state.DeletingView {
				return c.handleSelectAll()
			}
		case "left", "←":
			if c.app.GetCurrentState() == state.DeletingView {
				return c.handleDeselectAll()
			}
		}
	}
	
	// Update the list with the message
	list := c.app.GetList()
	*list, cmd = list.Update(msg)
	
	return c, cmd
}

// handleListSelection handles when user selects an item in list views
func (c *Controller) handleListSelection() (tea.Model, tea.Cmd) {
	currentState := c.app.GetCurrentState()
	
	switch currentState {
	case state.BackupListView:
		// Handle backup restoration with auto-backup if enabled
		if err := c.app.RestoreSelectedBackupWithAutoBackup(); err != nil {
			c.app.SetError(fmt.Errorf("failed to restore backup: %v", err))
			return c, nil
		}
		notificationCmd := c.app.ShowNotification("Backup restored successfully!")
		c.app.TransitionToState(state.MainMenuView)
		return c, notificationCmd
		
	case state.ViewBackupsView:
		// View-only mode - do nothing on Enter
		return c, nil
		
	case state.DeletingView:
		// Move to confirmation screen
		selections := c.app.GetSelections()
		if len(selections) == 0 {
			return c, nil // No items selected
		}
		c.app.TransitionToState(state.DeleteConfirmationView)
		return c, nil
	}
	
	return c, nil
}

// handleToggleSelection toggles selection of current item in delete view
func (c *Controller) handleToggleSelection() (tea.Model, tea.Cmd) {
	list := c.app.GetList()
	index := list.Index()
	selections := c.app.GetSelections()
	
	if _, exists := selections[index]; exists {
		delete(selections, index)
	} else {
		selections[index] = struct{}{}
	}
	
	// Update the list delegate with new selections
	c.app.SetListDelegate(components.NewSelectableItemDelegate(selections))
	
	return c, nil
}

// handleSelectAll selects all items in delete view
func (c *Controller) handleSelectAll() (tea.Model, tea.Cmd) {
	list := c.app.GetList()
	selections := c.app.GetSelections()
	
	// Select all items
	for i := 0; i < len(list.Items()); i++ {
		selections[i] = struct{}{}
	}
	
	// Update the list delegate with new selections
	c.app.SetListDelegate(components.NewSelectableItemDelegate(selections))
	
	return c, nil
}

// handleDeselectAll deselects all items in delete view
func (c *Controller) handleDeselectAll() (tea.Model, tea.Cmd) {
	// Clear all selections
	c.app.ClearSelections()
	
	// Update the list delegate with empty selections
	c.app.SetListDelegate(components.NewSelectableItemDelegate(c.app.GetSelections()))
	
	return c, nil
}

// handleSettingsView handles the settings menu
func (c *Controller) handleSettingsView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1":
			c.app.TransitionToState(state.ChangeSavePathView)
			c.app.SetTextInputPlaceholder("Enter new save path...")
			c.app.ClearTextInput()
			c.app.FocusTextInput()
			return c, nil
		case "2":
			c.app.TransitionToState(state.ChangeBackupDirView)
			c.app.SetTextInputPlaceholder("Enter new backup directory...")
			c.app.ClearTextInput()
			c.app.FocusTextInput()
			return c, nil
		case "3":
			// Toggle auto-backup setting
			if err := c.app.ToggleAutoBackup(); err != nil {
				c.app.SetError(fmt.Errorf("failed to update auto-backup setting: %v", err))
				return c, nil
			}
			status := "OFF"
			if c.app.GetConfig().AutoBackup {
				status = "ON"
			}
			notificationCmd := c.app.ShowNotification("Auto-backup setting: " + status)
			return c, notificationCmd
		}
	}
	
	return c, nil
}

// handleFirstRunComplete handles the completion of first-run setup
func (c *Controller) handleFirstRunComplete(backupDirPath string) (tea.Model, tea.Cmd) {
	// Get the temporarily saved save path
	savePath := c.app.GetTempSavePath()
	
	// Update the config with both paths
	config := c.app.GetConfig()
	config.SavePath = savePath
	config.BackupDir = backupDirPath
	
	// Save the config to disk
	if err := config.Save(); err != nil {
		c.app.SetError(fmt.Errorf("failed to save configuration: %v", err))
		return c, nil
	}
	
	// Clear the temporary save path
	c.app.ClearTempSavePath()
	
	// Transition to initializing and start database initialization
	c.app.TransitionToState(state.InitializingView)
	return c, c.app.Init()
}

// handleDeleteConfirmationView handles the delete confirmation view
func (c *Controller) handleDeleteConfirmationView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			// Confirm deletion
			selections := c.app.GetSelections()
			if err := c.app.DeleteSelectedBackups(); err != nil {
				c.app.SetError(fmt.Errorf("failed to delete backups: %v", err))
				return c, nil
			}
			notificationCmd := c.app.ShowNotification(fmt.Sprintf("Deleted %d backup(s)", len(selections)))
			c.app.TransitionToState(state.MainMenuView)
			return c, notificationCmd
		case "n", "N", "q":
			// Cancel deletion and go back to deleting view
			c.app.TransitionToState(state.DeletingView)
			return c, nil
		}
	}
	
	return c, nil
}