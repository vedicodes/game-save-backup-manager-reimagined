# Game Save Backup Manager - Reimagined

A modern, interactive, and robust command-line tool for managing your game save backups, built with a clean, maintainable architecture using Bubble Tea and Lipgloss.

## Architecture

This application follows industry-standard software architecture principles:

- **Controller-Based UI**: Clean separation between UI logic and business logic
- **Service Layer**: Dedicated backup service for all backup operations
- **State Management**: Centralized state transitions and validation
- **Component Architecture**: Reusable UI components with consistent styling
- **Configuration Management**: Persistent settings with automatic first-run setup
- **No Magic Numbers**: All layout constants centralized for maintainability

## Features

- **Interactive TUI:** A completely interactive, terminal-based UI for easy navigation and use.
- **Database-backed:** Uses SQLite to store backup metadata, allowing for future expansion with features like notes and tags.
- **Create Backups:** Easily create a backup of your game save file.
- **Restore Backups:** Restore a previously created backup.
- **List Backups:** View a list of all your available backups.
- **Delete Backups:** Remove unwanted backups.
- **Auto-Backup:** Automatically creates a backup of the current save before restoring another.
- **Configuration:** Customize the save file path and backup directory.

## Getting Started

### Prerequisites

- [Go](https://golang.org/doc/install) (version 1.18 or higher)

### Installation

1.  **Clone the repository:**
    ```sh
    git clone https://github.com/vedicodes/game-save-backup-manager-reimagined.git
    cd game-save-backup-manager-reimagined
    ```

2.  **Build the application:**
    ```sh
    go build ./cmd/manager
    ```

3.  **Run the application:**
    -   On Windows: `manager.exe`
    -   On macOS/Linux: `./manager`

## Usage

When you first run the application, you will be guided through a first-time setup process to configure your game's save file path and the directory where you want to store your backups.

The main menu provides the following options:

1.  **Create Backup:** Prompts for a backup name and creates a copy of your save file.
2.  **Restore Backup:** Shows a list of backups and lets you choose one to restore.
3.  **List Backups:** Displays all the backups in your backup directory.
4.  **Delete Backups:** Allows you to select and delete one or more backups.
5.  **Settings:** Configure various application settings.

## Configuration

The application creates a `config.json` file in the same directory as the executable. You can edit this file to set your game's save file path and the directory where you want to store your backups.

The `config.json` file has the following structure:

```json
{
  "save_path": "path/to/your/game.sav",
  "backup_dir": "path/to/your/backups",
  "auto_backup": false
}
```

## Project Structure

```
internal/
├── app/           # Application orchestration layer
├── backup/        # Database and backup operations
├── components/    # Reusable UI components
├── config/        # Configuration management
├── layout/        # UI layout constants
├── services/      # Business logic services
├── state/         # State management
├── tui/           # Terminal UI styling
├── ui/            # UI controllers
├── validation/    # Input validation
└── views/         # View handlers
```

## Key Features

- **Two-Step First Run Setup**: Guides users through save path and backup directory configuration
- **Auto-Generated Backup Names**: Creates timestamped backups when no name is provided
- **Deletion Confirmation**: Standard confirmation dialog before deleting backups
- **Auto-Backup Before Restore**: Optional feature to create backup before restoration
- **Responsive UI**: Input fields adapt to terminal window size
- **Consistent Help Text**: Context-appropriate help positioned at window bottom
