package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// LockMetadata contains information about the acquired lock
type LockMetadata struct {
	ProcessID  int       `json:"process_id"`
	AcquiredAt time.Time `json:"acquired_at"`
	Hostname   string    `json:"hostname,omitempty"`
}

const (
	lockFolderName = "lock"
	lockFileName   = "mneme.lock"
)

// AcquireLock acquires a lock in the data directory by creating a lock folder
// and a lock file with metadata about when the lock was acquired and by which process
func AcquireLock(dataDir string) error {
	lockFolderPath := filepath.Join(dataDir, lockFolderName)
	lockFilePath := filepath.Join(lockFolderPath, lockFileName)

	// Check if lock folder already exists
	if _, err := os.Stat(lockFolderPath); err == nil {
		// Lock folder exists, try to read existing lock metadata for better error message
		metadata, readErr := ReadLockMetadata(dataDir)
		if readErr == nil {
			return fmt.Errorf("a lock has already been acquired in the data directory (PID: %d, acquired at: %s)",
				metadata.ProcessID, metadata.AcquiredAt.Format(time.RFC3339))
		}
		return fmt.Errorf("a lock has already been acquired in the data directory")
	}

	// Create the lock folder
	if err := os.MkdirAll(lockFolderPath, 0755); err != nil {
		return fmt.Errorf("failed to create lock folder: %w", err)
	}

	// Get hostname (optional, for debugging)
	hostname, _ := os.Hostname()

	// Create lock metadata
	metadata := LockMetadata{
		ProcessID:  os.Getpid(),
		AcquiredAt: time.Now(),
		Hostname:   hostname,
	}

	// Marshal metadata to JSON
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		// Clean up the folder if we fail to create metadata
		os.RemoveAll(lockFolderPath)
		return fmt.Errorf("failed to marshal lock metadata: %w", err)
	}

	// Write the lock file
	if err := os.WriteFile(lockFilePath, data, 0644); err != nil {
		// Clean up the folder if we fail to write the file
		os.RemoveAll(lockFolderPath)
		return fmt.Errorf("failed to write lock file: %w", err)
	}

	return nil
}

// ReleaseLock releases the lock in the data directory by deleting the lock file and folder
func ReleaseLock(dataDir string) error {
	lockFolderPath := filepath.Join(dataDir, lockFolderName)

	// Check if lock folder exists
	if _, err := os.Stat(lockFolderPath); os.IsNotExist(err) {
		return fmt.Errorf("no lock exists to release")
	}

	// Remove the entire lock folder (including the lock file)
	if err := os.RemoveAll(lockFolderPath); err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	return nil
}

// CheckLock checks if a lock has been acquired in the data directory
func CheckLock(dataDir string) error {
	lockFolderPath := filepath.Join(dataDir, lockFolderName)

	if _, err := os.Stat(lockFolderPath); err == nil {
		// Lock exists, try to provide more context
		metadata, readErr := ReadLockMetadata(dataDir)
		if readErr == nil {
			return fmt.Errorf("a lock has already been acquired in the data directory (PID: %d, acquired at: %s)",
				metadata.ProcessID, metadata.AcquiredAt.Format(time.RFC3339))
		}
		return fmt.Errorf("a lock has already been acquired in the data directory")
	}

	return nil
}

// ReadLockMetadata reads the lock metadata from the lock file
func ReadLockMetadata(dataDir string) (*LockMetadata, error) {
	lockFilePath := filepath.Join(dataDir, lockFolderName, lockFileName)

	data, err := os.ReadFile(lockFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read lock file: %w", err)
	}

	var metadata LockMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse lock metadata: %w", err)
	}

	return &metadata, nil
}

// IsLockStale checks if the lock might be stale (the process that created it no longer exists)
// Note: This only works reliably on Unix-like systems and when the lock was created on the same machine
func IsLockStale(dataDir string) (bool, error) {
	metadata, err := ReadLockMetadata(dataDir)
	if err != nil {
		return false, err
	}

	// Check if the process still exists by sending signal 0
	// On Unix, FindProcess always succeeds, so we use signal 0 to verify
	process, err := os.FindProcess(metadata.ProcessID)
	if err != nil {
		return true, nil
	}

	// Signal 0 doesn't actually send a signal, but checks if the process exists
	// and if we have permission to send signals to it
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		// Process doesn't exist or we can't signal it - lock might be stale
		return true, nil
	}

	return false, nil
}
