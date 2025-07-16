package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gemini/game-save-backup-manager-reimagined/internal/backup"
	"github.com/gemini/game-save-backup-manager-reimagined/internal/config"
	"github.com/gemini/game-save-backup-manager-reimagined/internal/ui"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("A critical error occurred:", r)
			fmt.Println("Press Enter to exit...")
			fmt.Scanln()
			os.Exit(1)
		}
	}()

	cfg, isFirstRun, err := config.Load()
	if err != nil {
		handleError(err)
	}

	if isFirstRun {
		model := ui.NewModel(cfg, nil, true) // No db yet, and it's the first run
		p := tea.NewProgram(model, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			handleError(err)
		}
		fmt.Println("Configuration complete. Please start the application again.")
		os.Exit(0)
	}

	db, err := backup.InitDB(cfg.BackupDir)
	if err != nil {
		handleError(err)
	}
	defer db.Close()

	model := ui.NewModel(cfg, db, false)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		handleError(err)
	}
}

func handleError(err error) {
	fmt.Printf("\nAn error occurred:\n%v\n\n", err)
	fmt.Println("Press Enter to exit...")
	fmt.Scanln()
	os.Exit(1)
}
