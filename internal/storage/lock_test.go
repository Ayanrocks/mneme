package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAcquireLock(t *testing.T) {
	t.Run("acquires lock successfully", func(t *testing.T) {
		tempDir := t.TempDir()

		err := AcquireLock(tempDir)
		require.NoError(t, err)

		// Verify lock folder was created
		lockFolderPath := filepath.Join(tempDir, lockFolderName)
		info, err := os.Stat(lockFolderPath)
		require.NoError(t, err)
		assert.True(t, info.IsDir())

		// Verify lock file was created
		lockFilePath := filepath.Join(lockFolderPath, lockFileName)
		_, err = os.Stat(lockFilePath)
		require.NoError(t, err)
	})

	t.Run("lock file contains valid metadata", func(t *testing.T) {
		tempDir := t.TempDir()

		beforeAcquire := time.Now()
		err := AcquireLock(tempDir)
		require.NoError(t, err)
		afterAcquire := time.Now()

		// Read and verify metadata
		metadata, err := ReadLockMetadata(tempDir)
		require.NoError(t, err)

		assert.Equal(t, os.Getpid(), metadata.ProcessID)
		assert.True(t, metadata.AcquiredAt.After(beforeAcquire) || metadata.AcquiredAt.Equal(beforeAcquire))
		assert.True(t, metadata.AcquiredAt.Before(afterAcquire) || metadata.AcquiredAt.Equal(afterAcquire))
		assert.NotEmpty(t, metadata.Hostname)
	})

	t.Run("fails when lock already exists", func(t *testing.T) {
		tempDir := t.TempDir()

		// Acquire lock first time
		err := AcquireLock(tempDir)
		require.NoError(t, err)

		// Try to acquire again - should fail
		err = AcquireLock(tempDir)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "a lock has already been acquired")
		assert.Contains(t, err.Error(), "PID:")
	})

	t.Run("fails with generic message when lock folder exists but metadata is corrupted", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create lock folder manually without valid metadata
		lockFolderPath := filepath.Join(tempDir, lockFolderName)
		err := os.MkdirAll(lockFolderPath, 0755)
		require.NoError(t, err)

		// Write invalid JSON to lock file
		lockFilePath := filepath.Join(lockFolderPath, lockFileName)
		err = os.WriteFile(lockFilePath, []byte("invalid json"), 0644)
		require.NoError(t, err)

		// Try to acquire - should fail with generic message
		err = AcquireLock(tempDir)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "a lock has already been acquired")
	})

	t.Run("fails with generic message when lock folder exists but file is missing", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create lock folder manually without lock file
		lockFolderPath := filepath.Join(tempDir, lockFolderName)
		err := os.MkdirAll(lockFolderPath, 0755)
		require.NoError(t, err)

		// Try to acquire - should fail
		err = AcquireLock(tempDir)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "a lock has already been acquired")
	})
}

func TestReleaseLock(t *testing.T) {
	t.Run("releases lock successfully", func(t *testing.T) {
		tempDir := t.TempDir()

		// Acquire lock first
		err := AcquireLock(tempDir)
		require.NoError(t, err)

		// Release lock
		err = ReleaseLock(tempDir)
		require.NoError(t, err)

		// Verify lock folder was removed
		lockFolderPath := filepath.Join(tempDir, lockFolderName)
		_, err = os.Stat(lockFolderPath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("fails when no lock exists", func(t *testing.T) {
		tempDir := t.TempDir()

		err := ReleaseLock(tempDir)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no lock exists to release")
	})

	t.Run("can acquire lock after release", func(t *testing.T) {
		tempDir := t.TempDir()

		// Acquire, release, acquire again
		err := AcquireLock(tempDir)
		require.NoError(t, err)

		err = ReleaseLock(tempDir)
		require.NoError(t, err)

		err = AcquireLock(tempDir)
		require.NoError(t, err)
	})
}

func TestCheckLock(t *testing.T) {
	t.Run("returns nil when no lock exists", func(t *testing.T) {
		tempDir := t.TempDir()

		err := CheckLock(tempDir)
		assert.NoError(t, err)
	})

	t.Run("returns error when lock exists", func(t *testing.T) {
		tempDir := t.TempDir()

		// Acquire lock
		err := AcquireLock(tempDir)
		require.NoError(t, err)

		// Check lock - should return error
		err = CheckLock(tempDir)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "a lock has already been acquired")
		assert.Contains(t, err.Error(), "PID:")
	})

	t.Run("returns nil after lock is released", func(t *testing.T) {
		tempDir := t.TempDir()

		// Acquire and release lock
		err := AcquireLock(tempDir)
		require.NoError(t, err)

		err = ReleaseLock(tempDir)
		require.NoError(t, err)

		// Check lock - should return nil
		err = CheckLock(tempDir)
		assert.NoError(t, err)
	})

	t.Run("returns generic error when metadata is corrupted", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create lock folder with corrupted metadata
		lockFolderPath := filepath.Join(tempDir, lockFolderName)
		err := os.MkdirAll(lockFolderPath, 0755)
		require.NoError(t, err)

		lockFilePath := filepath.Join(lockFolderPath, lockFileName)
		err = os.WriteFile(lockFilePath, []byte("not json"), 0644)
		require.NoError(t, err)

		// Check lock - should return error without PID details
		err = CheckLock(tempDir)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "a lock has already been acquired")
	})
}

func TestReadLockMetadata(t *testing.T) {
	t.Run("reads metadata successfully", func(t *testing.T) {
		tempDir := t.TempDir()

		// Acquire lock
		err := AcquireLock(tempDir)
		require.NoError(t, err)

		// Read metadata
		metadata, err := ReadLockMetadata(tempDir)
		require.NoError(t, err)
		require.NotNil(t, metadata)

		assert.Equal(t, os.Getpid(), metadata.ProcessID)
		assert.False(t, metadata.AcquiredAt.IsZero())
	})

	t.Run("returns error when lock file does not exist", func(t *testing.T) {
		tempDir := t.TempDir()

		metadata, err := ReadLockMetadata(tempDir)
		require.Error(t, err)
		assert.Nil(t, metadata)
		assert.Contains(t, err.Error(), "failed to read lock file")
	})

	t.Run("returns error when metadata is invalid JSON", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create lock folder with invalid JSON
		lockFolderPath := filepath.Join(tempDir, lockFolderName)
		err := os.MkdirAll(lockFolderPath, 0755)
		require.NoError(t, err)

		lockFilePath := filepath.Join(lockFolderPath, lockFileName)
		err = os.WriteFile(lockFilePath, []byte("invalid json content"), 0644)
		require.NoError(t, err)

		metadata, err := ReadLockMetadata(tempDir)
		require.Error(t, err)
		assert.Nil(t, metadata)
		assert.Contains(t, err.Error(), "failed to parse lock metadata")
	})

	t.Run("parses all metadata fields correctly", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create lock with known metadata
		lockFolderPath := filepath.Join(tempDir, lockFolderName)
		err := os.MkdirAll(lockFolderPath, 0755)
		require.NoError(t, err)

		expectedTime := time.Date(2026, 1, 30, 10, 0, 0, 0, time.UTC)
		metadata := LockMetadata{
			ProcessID:  12345,
			AcquiredAt: expectedTime,
			Hostname:   "test-host",
		}

		data, err := json.MarshalIndent(metadata, "", "  ")
		require.NoError(t, err)

		lockFilePath := filepath.Join(lockFolderPath, lockFileName)
		err = os.WriteFile(lockFilePath, data, 0644)
		require.NoError(t, err)

		// Read and verify
		readMetadata, err := ReadLockMetadata(tempDir)
		require.NoError(t, err)

		assert.Equal(t, 12345, readMetadata.ProcessID)
		assert.Equal(t, expectedTime, readMetadata.AcquiredAt)
		assert.Equal(t, "test-host", readMetadata.Hostname)
	})
}

func TestIsLockStale(t *testing.T) {
	t.Run("returns false for current process lock", func(t *testing.T) {
		tempDir := t.TempDir()

		// Acquire lock with current process
		err := AcquireLock(tempDir)
		require.NoError(t, err)

		isStale, err := IsLockStale(tempDir)
		require.NoError(t, err)
		assert.False(t, isStale)
	})

	t.Run("returns error when no lock exists", func(t *testing.T) {
		tempDir := t.TempDir()

		_, err := IsLockStale(tempDir)
		require.Error(t, err)
	})

	t.Run("returns true for non-existent process", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create lock with a non-existent PID
		lockFolderPath := filepath.Join(tempDir, lockFolderName)
		err := os.MkdirAll(lockFolderPath, 0755)
		require.NoError(t, err)

		// Use a very high PID that's unlikely to exist
		metadata := LockMetadata{
			ProcessID:  999999999,
			AcquiredAt: time.Now(),
			Hostname:   "test-host",
		}

		data, err := json.MarshalIndent(metadata, "", "  ")
		require.NoError(t, err)

		lockFilePath := filepath.Join(lockFolderPath, lockFileName)
		err = os.WriteFile(lockFilePath, data, 0644)
		require.NoError(t, err)

		isStale, err := IsLockStale(tempDir)
		require.NoError(t, err)
		assert.True(t, isStale)
	})
}

func TestLockIntegration(t *testing.T) {
	t.Run("full lock lifecycle", func(t *testing.T) {
		tempDir := t.TempDir()

		// Initially no lock
		err := CheckLock(tempDir)
		require.NoError(t, err)

		// Acquire lock
		err = AcquireLock(tempDir)
		require.NoError(t, err)

		// Check lock exists
		err = CheckLock(tempDir)
		require.Error(t, err)

		// Read metadata
		metadata, err := ReadLockMetadata(tempDir)
		require.NoError(t, err)
		assert.Equal(t, os.Getpid(), metadata.ProcessID)

		// Check not stale
		isStale, err := IsLockStale(tempDir)
		require.NoError(t, err)
		assert.False(t, isStale)

		// Release lock
		err = ReleaseLock(tempDir)
		require.NoError(t, err)

		// Lock no longer exists
		err = CheckLock(tempDir)
		require.NoError(t, err)
	})

	t.Run("lock file structure is correct", func(t *testing.T) {
		tempDir := t.TempDir()

		err := AcquireLock(tempDir)
		require.NoError(t, err)

		// Check folder structure
		lockFolderPath := filepath.Join(tempDir, "lock")
		lockFilePath := filepath.Join(lockFolderPath, "mneme.lock")

		// Lock folder exists
		info, err := os.Stat(lockFolderPath)
		require.NoError(t, err)
		assert.True(t, info.IsDir())

		// Lock file exists
		info, err = os.Stat(lockFilePath)
		require.NoError(t, err)
		assert.False(t, info.IsDir())

		// Lock file is valid JSON
		data, err := os.ReadFile(lockFilePath)
		require.NoError(t, err)

		var metadata LockMetadata
		err = json.Unmarshal(data, &metadata)
		require.NoError(t, err)
	})
}
