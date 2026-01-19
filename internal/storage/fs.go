package storage

import (
	"errors"
	"fmt"
	"mneme/internal/logger"
	"mneme/internal/version"
	"os"
	"path"
	"path/filepath"
)

const (
	DirPath    = "~/.local/share/mneme"
	ConfigPath = "~/.config/mneme.yaml"
	AppName    = "mneme"
)

// CreateDir Create directory function
func CreateDir(path string) error {
	expandedPath, err := expandPath(path)
	if err != nil {
		logger.Errorf("Error Expanding Path: %+v", err.Error())
		return err
	}
	return os.MkdirAll(expandedPath, os.ModePerm)
}

// CreateFile creates a file at the given path, expanding ~ to home directory
func CreateFile(path string) (*os.File, error) {
	expandedPath, err := expandPath(path)
	if err != nil {
		logger.Errorf("Error expanding path: %+v", err.Error())
		return nil, err
	}
	return os.Create(expandedPath)
}

func InitMnemeStorage() error {
	// Fetch the default directory and check if it exists
	logger.Info("Initializing mneme storage...")

	exists, err := DirExists(filepath.Dir(ConfigPath))
	if err != nil {
		logger.Errorf("Error checking if config directory exists: %+v", err)
		return err
	}

	if !exists {
		logger.Infof("Creating storage directory: %s", DirPath)
		if err := CreateDir(DirPath); err != nil {
			logger.Errorf("Error creating directory %s: %+v", DirPath, err)
			return err
		}
	}

	// Now if dir exists, let's initialize the folder structure for mneme
	// Create the mneme directory first
	logger.Infof("Creating app directory: %s", DirPath)
	err = CreateDir(DirPath)
	if err != nil {
		logger.Errorf("Error creating app directory %s: %+v", AppName, err)
		return err
	}

	// create internal directories
	logger.Info("Creating internal directories...")
	err = CreateDir(path.Join(DirPath, "meta"))
	if err != nil {
		logger.Errorf("Error creating meta directory: %+v", err)
		return err
	}

	err = CreateDir(path.Join(DirPath, "segments"))
	if err != nil {
		logger.Errorf("Error creating segments directory: %+v", err)
		return err
	}

	err = CreateDir(path.Join(DirPath, "tombstones"))
	if err != nil {
		logger.Errorf("Error creating tombstones directory: %+v", err)
		return err
	}

	// Create an empty version file with the current version of the mneme that created the folder
	logger.Info("Creating version file...")
	file, err := CreateFile(path.Join(DirPath, "VERSION"))
	if err != nil {
		logger.Errorf("Error creating VERSION file: %+v", err)
		return err
	}
	
	_, err = file.WriteString(getVersionFileContents())
	if err != nil {
		logger.Errorf("Error writing to VERSION file: %+v", err)
		return err
	}

	err = file.Close()
	if err != nil {
		logger.Errorf("Error closing VERSION file: %+v", err)
		return err
	}

	logger.Info("Storage initialization completed successfully!")
	return nil
}

// DirExists checks if a directory exists at the given path
func DirExists(path string) (bool, error) {
	// Expand ~ to home directory
	expandedPath, err := expandPath(path)
	if err != nil {
		logger.Errorf("Error Expanding Path: %+v", err.Error())
		return false, err
	}

	info, err := os.Stat(expandedPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Debugf("Directory does not exist: %s", expandedPath)
			return false, nil
		}
		logger.Errorf("Error stating directory %s: %+v", expandedPath, err)
		return false, fmt.Errorf("failed to check directory: %w", err)
	}

	if info.IsDir() {
		logger.Debugf("Directory exists: %s", expandedPath)
	}
	return info.IsDir(), nil
}

// expandPath expands ~ to the user's home directory
func expandPath(path string) (string, error) {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			logger.Errorf("Error getting user home directory: %+v", err)
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		return filepath.Join(home, path[1:]), nil
	}
	return path, nil
}

func getVersionFileContents() string {
	return fmt.Sprintf(`
	STORAGE_VERSION: %s
	MNEME_CLI_VERSION: %s
`, version.MnemeStorageEngineVersion, version.MnemeVersion)
}
