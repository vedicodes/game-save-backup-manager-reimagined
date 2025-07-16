package state

// ViewState represents the current view of the application
type ViewState int

const (
	InitializingView ViewState = iota
	MainMenuView
	BackupListView
	ViewBackupsView
	CreateBackupView
	RestoreConfirmationView
	DeleteConfirmationView
	DeletingView
	SettingsView
	ChangeSavePathView
	ChangeBackupDirView
	FirstRunView
	FirstRunBackupDirView
)

// StateManager handles view state transitions and validation
type StateManager struct {
	current ViewState
	previous ViewState
}

// NewStateManager creates a new state manager
func NewStateManager(initialState ViewState) *StateManager {
	return &StateManager{
		current: initialState,
		previous: initialState,
	}
}

// Current returns the current state
func (sm *StateManager) Current() ViewState {
	return sm.current
}

// Previous returns the previous state
func (sm *StateManager) Previous() ViewState {
	return sm.previous
}

// TransitionTo changes to a new state
func (sm *StateManager) TransitionTo(newState ViewState) {
	sm.previous = sm.current
	sm.current = newState
}

// CanTransitionTo validates if a state transition is allowed
func (sm *StateManager) CanTransitionTo(newState ViewState) bool {
	// Add validation logic here if needed
	// For now, all transitions are allowed
	return true
}

// IsInState checks if currently in the specified state
func (sm *StateManager) IsInState(state ViewState) bool {
	return sm.current == state
}

// IsInAnyState checks if currently in any of the specified states
func (sm *StateManager) IsInAnyState(states ...ViewState) bool {
	for _, state := range states {
		if sm.current == state {
			return true
		}
	}
	return false
}