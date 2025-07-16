package validation

import (
	"os"
	"path/filepath"
)

// IsWritable checks if a directory path is writable.
// It will attempt to create the directory if it doesn't exist.
func IsWritable(dir string) error {
	// Try to create the directory
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Check write permissions by creating a temporary file
	tmpFile := filepath.Join(dir, ".tmp_write_test")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		return err
	}

	// Clean up the temporary file
	return os.Remove(tmpFile)
}
