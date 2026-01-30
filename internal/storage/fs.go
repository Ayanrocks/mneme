package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"mneme/internal/config"
	"mneme/internal/constants"
	"mneme/internal/core"
	"mneme/internal/logger"
	"mneme/internal/utils"
	"mneme/internal/version"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// CreateDir Create directory function
func CreateDir(path string) error {
	expandedPath, err := utils.ExpandFilePath(path)
	if err != nil {
		logger.Errorf("Error Expanding Path: %+v", err.Error())
		return err
	}
	return os.MkdirAll(expandedPath, os.ModePerm)
}

// CreateFile creates a file at the given path, expanding ~ to home directory
func CreateFile(path string) (*os.File, error) {
	expandedPath, err := utils.ExpandFilePath(path)
	if err != nil {
		logger.Errorf("Error expanding path: %+v", err.Error())
		return nil, err
	}
	return os.Create(expandedPath)
}

// FileExists checks if a file exists at the given path
func FileExists(path string) (bool, error) {
	expandedPath, err := utils.ExpandFilePath(path)
	if err != nil {
		logger.Errorf("Error expanding path: %+v", err.Error())
		return false, err
	}

	info, err := os.Stat(expandedPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Debugf("File does not exist: %s", expandedPath)
			return false, nil
		}
		logger.Errorf("Error stating file %s: %+v", expandedPath, err)
		return false, fmt.Errorf("failed to check file: %w", err)
	}

	if info.IsDir() {
		logger.Debugf("Path is a directory, not a file: %s", expandedPath)
		return false, nil
	}

	logger.Debugf("File exists: %s", expandedPath)
	return true, nil
}

// ReadVersionFile reads and returns the contents of the VERSION file
func ReadVersionFile() (string, error) {
	versionPath := filepath.Join(constants.DirPath, "VERSION")
	expandedPath, err := utils.ExpandFilePath(versionPath)
	if err != nil {
		logger.Errorf("Error expanding version path: %+v", err.Error())
		return "", err
	}

	file, err := os.Open(expandedPath)
	if err != nil {
		logger.Errorf("Error opening VERSION file: %+v", err)
		return "", fmt.Errorf("failed to open VERSION file: %w", err)
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			logger.Errorf("Error closing VERSION file: %+v", err)
		}
	}(file)

	var content strings.Builder
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content.WriteString(scanner.Text())
		content.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		logger.Errorf("Error reading VERSION file: %+v", err)
		return "", fmt.Errorf("failed to read VERSION file: %w", err)
	}

	return content.String(), nil
}

// ParseVersionFile parses the VERSION file content and extracts version information
func ParseVersionFile(content string) (storageVersion string, cliVersion string, err error) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "STORAGE_VERSION:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				storageVersion = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "MNEME_CLI_VERSION:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				cliVersion = strings.TrimSpace(parts[1])
			}
		}
	}

	if storageVersion == "" {
		return "", "", fmt.Errorf("STORAGE_VERSION not found in VERSION file")
	}

	return storageVersion, cliVersion, nil
}

// IsVersionCompatible checks if the existing storage version is compatible with current version
func IsVersionCompatible() (bool, error) {
	exists, err := FileExists(filepath.Join(constants.DirPath, "VERSION"))
	if err != nil {
		return false, err
	}
	if !exists {
		logger.Debug("VERSION file does not exist, initialization needed")
		return false, nil
	}

	content, err := ReadVersionFile()
	if err != nil {
		return false, err
	}

	existingStorageVersion, existingCliVersion, err := ParseVersionFile(content)
	if err != nil {
		logger.Errorf("Error parsing VERSION file: %+v", err)
		return false, err
	}

	currentStorageVersion := version.MnemeStorageEngineVersion
	currentCliVersion := version.MnemeVersion

	logger.Debugf("Existing storage version: %s, Current: %s", existingStorageVersion, currentStorageVersion)
	logger.Debugf("Existing CLI version: %s, Current: %s", existingCliVersion, currentCliVersion)

	// Check if versions match exactly
	if existingStorageVersion == currentStorageVersion && existingCliVersion == currentCliVersion {
		logger.Info("Storage is already initialized with compatible version")
		return true, nil
	}

	// For now, we consider it incompatible if versions don't match
	// In a real implementation, you might want to support version upgrades
	logger.Warnf("Storage version mismatch: existing=%s, current=%s", existingStorageVersion, currentStorageVersion)
	return false, nil
}

// ShouldInitialize determines if storage initialization is needed
func ShouldInitialize() (bool, error) {
	// Check if storage directory exists
	exists, err := DirExists(constants.DirPath)
	if err != nil {
		return false, err
	}

	if !exists {
		logger.Debug("Storage directory does not exist, initialization needed")
		return true, nil
	}

	// Check if VERSION file exists and is compatible
	compatible, err := IsVersionCompatible()
	if err != nil {
		return false, err
	}

	if !compatible {
		logger.Debug("Storage version is older")
		return false, nil
	}

	logger.Debug("Storage is already initialized and compatible, skipping initialization")
	return false, nil
}

// ShouldInitializeConfig determines if config initialization is needed
func ShouldInitializeConfig() (bool, error) {
	// Check if config file exists
	exists, err := FileExists(constants.ConfigPath)
	if err != nil {
		return false, err
	}

	if !exists {
		logger.Debug("Config file does not exist, initialization needed")
		return true, nil
	}

	logger.Debug("Config file already exists, skipping initialization")
	return false, nil
}

func InitMnemeStorage() error {
	logger.Info("Initializing mneme storage...")

	// Check if initialization is needed
	shouldInit, err := ShouldInitialize()
	if err != nil {
		logger.Errorf("Error checking if initialization is needed: %+v", err)
		return err
	}

	if !shouldInit {
		logger.Info("Storage already initialized, skipping...")
		return nil
	}

	// Fetch the default directory and check if it exists
	exists, err := DirExists(filepath.Dir(constants.ConfigPath))
	if err != nil {
		logger.Errorf("Error checking if config directory exists: %+v", err)
		return err
	}

	if !exists {
		logger.Infof("Creating storage directory: %s", constants.DirPath)
		if err := CreateDir(constants.DirPath); err != nil {
			logger.Errorf("Error creating directory %s: %+v", constants.DirPath, err)
			return err
		}
	}

	// Now if dir exists, let's initialize the folder structure for mneme
	// Create the mneme directory first
	logger.Infof("Creating app directory: %s", constants.DirPath)
	err = CreateDir(constants.DirPath)
	if err != nil {
		logger.Errorf("Error creating app directory %s: %+v", constants.AppName, err)
		return err
	}

	// create internal directories
	logger.Info("Creating internal directories...")
	err = CreateDir(path.Join(constants.DirPath, "meta"))
	if err != nil {
		logger.Errorf("Error creating meta directory: %+v", err)
		return err
	}

	err = CreateDir(path.Join(constants.DirPath, "segments"))
	if err != nil {
		logger.Errorf("Error creating segments directory: %+v", err)
		return err
	}

	err = CreateDir(path.Join(constants.DirPath, "tombstones"))
	if err != nil {
		logger.Errorf("Error creating tombstones directory: %+v", err)
		return err
	}

	// Create an empty version file with the current version of the mneme that created the folder
	logger.Info("Creating version file...")
	file, err := CreateFile(path.Join(constants.DirPath, "VERSION"))
	if err != nil {
		logger.Errorf("Error creating VERSION file: %+v", err)
		return err
	}
	defer func() {
		err = file.Close()
		if err != nil {
			logger.Errorf("Error closing VERSION file: %+v", err)
		}
	}()

	_, err = file.WriteString(getVersionFileContents())
	if err != nil {
		logger.Errorf("Error writing to VERSION file: %+v", err)
		return err
	}

	logger.Info("Storage initialization completed successfully!")
	return nil
}

func InitMnemeConfigStorage() error {
	logger.Info("Initializing mneme configuration storage...")

	shouldInit, err := ShouldInitializeConfig()
	if err != nil {
		logger.Errorf("Error checking if initialization is needed: %+v", err)
		return err
	}

	if !shouldInit {
		logger.Info("Config already initialized, skipping...")
		return nil
	}

	exists, err := DirExists(filepath.Dir(constants.ConfigPath))
	if err != nil {
		logger.Errorf("Error checking if config directory exists: %+v", err)
		return err
	}

	if !exists {
		logger.Infof("Creating storage directory: %s", filepath.Dir(constants.ConfigPath))
		if err := CreateDir(filepath.Dir(constants.ConfigPath)); err != nil {
			logger.Errorf("Error creating config directory: %+v", err)
			return err
		}
	}

	logger.Info("Creating default config")

	// Get default config as TOML string
	configContent, err := config.DefaultConfigWriter()
	if err != nil {
		logger.Errorf("Error generating default config: %+v", err)
		return err
	}

	// Create config file
	file, err := CreateFile(constants.ConfigPath)
	if err != nil {
		logger.Errorf("Error creating config file: %+v", err)
		return err
	}
	defer func() {
		err = file.Close()
		if err != nil {
			logger.Errorf("Error closing VERSION file: %+v", err)
		}
	}()

	// Write config to file
	_, err = file.WriteString(configContent)
	if err != nil {
		logger.Errorf("Error writing to config file: %+v", err)
		return err
	}

	logger.Info("Config initialization completed successfully!")
	return nil
}

// DirExists checks if a directory exists at the given path
func DirExists(path string) (bool, error) {
	// Expand ~ to home directory
	expandedPath, err := utils.ExpandFilePath(path)
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

func getVersionFileContents() string {
	return fmt.Sprintf(`
STORAGE_VERSION: %s
MNEME_CLI_VERSION: %s
`, version.MnemeStorageEngineVersion, version.MnemeVersion)
}

// ReadFileContents reads and returns the contents of the file at the given path
func ReadFileContents(path string) ([]string, error) {
	expandedPath, err := utils.ExpandFilePath(path)
	if err != nil {
		logger.Errorf("Error expanding path: %+v", err)
		return nil, err
	}

	file, err := os.Open(expandedPath)
	if err != nil {
		logger.Errorf("Error opening file %s: %+v", expandedPath, err)
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		logger.Errorf("Error reading file %s: %+v", expandedPath, err)
		return nil, err
	}

	return lines, nil
}

func SaveSegmentIndex(segmentIndex *core.Segment) error {
	logger.Info("Saving segment index...")

	// Define the path for the segment JSON file
	segmentPath := filepath.Join(constants.DirPath, "segments", "segment.json")

	// Expand the file path (handles ~ expansion)
	expandedPath, err := utils.ExpandFilePath(segmentPath)
	if err != nil {
		logger.Errorf("Error expanding segment path: %+v", err)
		return fmt.Errorf("failed to expand segment path: %w", err)
	}

	// Marshal the segment index to JSON with indentation for readability
	jsonData, err := json.MarshalIndent(segmentIndex, "", "  ")
	if err != nil {
		logger.Errorf("Error marshaling segment index to JSON: %+v", err)
		return fmt.Errorf("failed to marshal segment index: %w", err)
	}

	// Write the JSON data to the file
	err = os.WriteFile(expandedPath, jsonData, 0644)
	if err != nil {
		logger.Errorf("Error writing segment index to file %s: %+v", expandedPath, err)
		return fmt.Errorf("failed to write segment index: %w", err)
	}

	logger.Infof("Segment index saved successfully to %s", expandedPath)
	return nil
}
