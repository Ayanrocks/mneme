package storage

import (
	"bufio"
	"errors"
	"fmt"
	"mneme/internal/core"
	"mneme/internal/logger"
	"mneme/internal/version"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const (
	DirPath    = "~/.local/share/mneme"
	ConfigPath = "~/.config/mneme/mneme.toml"
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

// FileExists checks if a file exists at the given path
func FileExists(path string) (bool, error) {
	expandedPath, err := expandPath(path)
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
	versionPath := filepath.Join(DirPath, "VERSION")
	expandedPath, err := expandPath(versionPath)
	if err != nil {
		logger.Errorf("Error expanding version path: %+v", err.Error())
		return "", err
	}

	file, err := os.Open(expandedPath)
	if err != nil {
		logger.Errorf("Error opening VERSION file: %+v", err)
		return "", fmt.Errorf("failed to open VERSION file: %w", err)
	}
	defer file.Close()

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
	exists, err := FileExists(filepath.Join(DirPath, "VERSION"))
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
	exists, err := DirExists(DirPath)
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
	exists, err := FileExists(ConfigPath)
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

	exists, err := DirExists(filepath.Dir(ConfigPath))
	if err != nil {
		logger.Errorf("Error checking if config directory exists: %+v", err)
		return err
	}

	if !exists {
		logger.Infof("Creating storage directory: %s", filepath.Dir(ConfigPath))
		if err := CreateDir(filepath.Dir(ConfigPath)); err != nil {
			logger.Errorf("Error creating config directory: %+v", err)
			return err
		}
	}

	logger.Info("Creating default config")

	// Get default config as TOML string
	configContent, err := defaultConfigWriter()
	if err != nil {
		logger.Errorf("Error generating default config: %+v", err)
		return err
	}

	// Create config file
	file, err := CreateFile(ConfigPath)
	if err != nil {
		logger.Errorf("Error creating config file: %+v", err)
		return err
	}
	defer file.Close()

	// Write config to file
	_, err = file.WriteString(configContent)
	if err != nil {
		logger.Errorf("Error writing to config file: %+v", err)
		return err
	}

	logger.Info("Config initialization completed successfully!")
	return nil
}

func defaultConfigWriter() (string, error) {
	defaultConfig := core.DefaultConfig{
		Version: 1,
		Index: core.IndexConfig{
			SegmentSize:          500,
			MaxTokensPerDocument: 10000,
			ReindexOnModify:      true,
			SkipBinaryFiles:      true,
		},
		Sources: core.SourcesConfig{
			Paths:             []string{},
			IncludeExtensions: []string{},
			Ignore:            []string{".git", "node_modules", ".vscode", ".idea", "vendor", ".cache"},
		},
		Watcher: core.WatcherConfig{
			Enabled:    true,
			DebounceMS: 500,
		},
		Search: core.SearchConfig{
			DefaultLimit: 20,
			UseStopwords: true,
			Language:     "en",
		},
		Ranking: core.RankingConfig{
			TFIDFWeight:         1,
			RecencyWeight:       0.3,
			TitleBoost:          1.5,
			PathBoost:           1.2,
			RecencyHalfLifeDays: 30,
		},
		Logging: core.LoggingConfig{
			Level: "info",
			JSON:  true,
		},
	}

	configBytes, err := toml.Marshal(defaultConfig)
	if err != nil {
		logger.Errorf("Error marshaling default config to TOML: %+v", err)
		return "", fmt.Errorf("failed to marshal default config: %w", err)
	}

	return string(configBytes), nil
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
