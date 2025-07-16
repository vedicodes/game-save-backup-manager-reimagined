package backup

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Backup represents a single backup record.
type Backup struct {
	ID        int
	Name      string
	Path      string
	CreatedAt time.Time
}

// DB represents the backup database.
type DB struct {
	*sql.DB
}

// InitDB initializes the database in the backup directory.
func InitDB(backupDir string) (*DB, error) {
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, err
	}
	dbPath := filepath.Join(backupDir, "backups.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Create table if not exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS backups (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			path TEXT NOT NULL,
			created_at DATETIME NOT NULL
		)
	`)
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

// CreateBackup creates a new backup.
func (db *DB) CreateBackup(savePath, backupDir, backupName string) error {
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		return fmt.Errorf("save file not found: %s", savePath)
	}

	if backupName == "" {
		backupName = fmt.Sprintf("Backup_%s", time.Now().Format("2006-01-02_15-04-05"))
	}

	backupPath := filepath.Join(backupDir, backupName+".sav")
	// Ensure the backup name is unique
	counter := 1
	baseName := backupName
	for {
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			break
		}
		backupName = fmt.Sprintf("%s_%d", baseName, counter)
		backupPath = filepath.Join(backupDir, backupName+".sav")
		counter++
	}

	data, err := os.ReadFile(savePath)
	if err != nil {
		return err
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return err
	}

	// Add to database
	_, err = db.Exec("INSERT INTO backups (name, path, created_at) VALUES (?, ?, ?)", backupName, backupPath, time.Now())
	return err
}

// GetBackups retrieves all backups from the database.
func (db *DB) GetBackups() ([]Backup, error) {
	rows, err := db.Query("SELECT id, name, path, created_at FROM backups ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var backups []Backup
	for rows.Next() {
		var b Backup
		if err := rows.Scan(&b.ID, &b.Name, &b.Path, &b.CreatedAt); err != nil {
			return nil, err
		}
		backups = append(backups, b)
	}
	return backups, nil
}

// RestoreBackup restores a selected backup.
func (db *DB) RestoreBackup(b Backup, savePath string) error {
	data, err := os.ReadFile(b.Path)
	if err != nil {
		return err
	}
	return os.WriteFile(savePath, data, 0644)
}

// DeleteBackup deletes a backup.
func (db *DB) DeleteBackup(b Backup) error {
	if err := os.Remove(b.Path); err != nil {
		return err
	}
	_, err := db.Exec("DELETE FROM backups WHERE id = ?", b.ID)
	return err
}
// DeleteBackups deletes multiple backups in a single transaction.
func (db *DB) DeleteBackups(backups []Backup) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, b := range backups {
		// Delete the file
		if err := os.Remove(b.Path); err != nil {
			// Continue with other deletions even if one file fails
			// This handles cases where the file might already be deleted
			continue
		}
		
		// Delete from database
		_, err := tx.Exec("DELETE FROM backups WHERE id = ?", b.ID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}