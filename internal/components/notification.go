package components

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vedicodes/game-save-backup-manager-reimagined/internal/layout"
)

// NotificationManager handles notification display and timing
type NotificationManager struct {
	message string
}

// ClearNotificationMsg is sent when a notification should be cleared
type ClearNotificationMsg struct{}

// NewNotificationManager creates a new notification manager
func NewNotificationManager() *NotificationManager {
	return &NotificationManager{}
}

// Show sets a message and returns a command to clear it after the configured duration
func (nm *NotificationManager) Show(msg string) tea.Cmd {
	nm.message = msg
	return tea.Tick(layout.NotificationDuration, func(t time.Time) tea.Msg {
		return ClearNotificationMsg{}
	})
}

// Clear clears the current notification
func (nm *NotificationManager) Clear() {
	nm.message = ""
}

// HasMessage returns true if there's a message to display
func (nm *NotificationManager) HasMessage() bool {
	return nm.message != ""
}

// GetMessage returns the current message
func (nm *NotificationManager) GetMessage() string {
	return nm.message
}

// Update handles notification-related messages
func (nm *NotificationManager) Update(msg tea.Msg) {
	if _, ok := msg.(ClearNotificationMsg); ok {
		nm.Clear()
	}
}