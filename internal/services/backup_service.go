package services

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/vedicodes/game-save-backup-manager-reimagined/internal/backup"
	"github.com/vedicodes/game-save-backup-manager-reimagined/internal/components"
	"github.com/vedicodes/game-save-backup-manager-reimagined/internal/config"
)

// BackupService handles all backup-related business logic
type BackupService struct {
	db     *backup.DB
	config *config.Config
}

// NewBackupService creates a new backup service
func NewBackupService(db *backup.DB, config *config.Config) *BackupService {
	return &BackupService{
		db:     db,
		config: config,
	}
}

// CreateBackup creates a new backup with the given name
func (bs *BackupService) CreateBackup(name string) error {
	return bs.db.CreateBackup(bs.config.SavePath, bs.config.BackupDir, name)
}

// RestoreBackup restores the specified backup
func (bs *BackupService) RestoreBackup(backup backup.Backup) error {
	return bs.db.RestoreBackup(backup, bs.config.SavePath)
}

// DeleteBackups deletes multiple backups
func (bs *BackupService) DeleteBackups(backups []backup.Backup) error {
	return bs.db.DeleteBackups(backups)
}

// GetBackupItems fetches all backups and converts them to list items
func (bs *BackupService) GetBackupItems() ([]list.Item, error) {
	if bs.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	
	backups, err := bs.db.GetBackups()
	if err != nil {
		return nil, err
	}
	
	items := make([]list.Item, len(backups))
	for i, b := range backups {
		items[i] = components.ListItem(b)
	}
	return items, nil
}

// GetSelectedBackups converts selected indices to backup objects
func (bs *BackupService) GetSelectedBackups(items []list.Item, selected map[int]struct{}) []backup.Backup {
	var backups []backup.Backup
	for i, item := range items {
		if _, ok := selected[i]; ok {
			if listItem, ok := item.(components.ListItem); ok {
				backups = append(backups, backup.Backup(listItem))
			}
		}
	}
	return backups
}

// InitializeDatabase initializes the backup database
func (bs *BackupService) InitializeDatabase() error {
	if bs.config == nil || bs.config.BackupDir == "" {
		return fmt.Errorf("configuration is missing or invalid")
	}

	db, err := backup.InitDB(bs.config.BackupDir)
	if err != nil {
		return err
	}
	
	bs.db = db
	return nil
}