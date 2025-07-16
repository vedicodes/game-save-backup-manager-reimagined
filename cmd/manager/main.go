package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vedicodes/game-save-backup-manager-reimagined/internal/config"
	"github.com/vedicodes/game-save-backup-manager-reimagined/internal/ui"
)

func main() {
	// Set up a panic handler for graceful exit on critical errors.
	defer func() {
		if r := recover(); r != nil {
			// Stop Bubble Tea to release the terminal.
			// We can't pass the program here, so we send a signal.
			// This is a bit of a hack, but it's the best we can do.
			tea.NewProgram(nil).Quit() // This sends a quit message to the running program.
			fmt.Printf("\nA critical error occurred: %v\n", r)
			fmt.Println("Press Enter to exit...")
			fmt.Scanln()
			os.Exit(1)
		}
	}()

	cfg, isFirstRun, err := config.Load()
	if err != nil {
		handleError(err)
	}

	// Create the new controller-based UI
	controller := ui.NewController(cfg, isFirstRun)
	p := tea.NewProgram(controller, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		handleError(err)
	}
}

// handleError is a centralized function to display errors to the user.
func handleError(err error) {
	// Ensure the terminal is in a usable state.
	tea.NewProgram(nil).Quit()
	fmt.Printf("\nAn error occurred:\n%v\n\n", err)
	fmt.Println("Press Enter to exit...")
	fmt.Scanln()
	os.Exit(1)
}
