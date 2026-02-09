//go:build windows

package platform

import (
	"os"
	"path/filepath"
)

// GetDataDir returns the platform-appropriate data directory for Windows
// Uses %LOCALAPPDATA%\mneme
func GetDataDir() string {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		// Fallback to user home + AppData/Local
		home, err := os.UserHomeDir()
		if err != nil {
			return `C:\mneme` // Last resort fallback
		}
		localAppData = filepath.Join(home, "AppData", "Local")
	}
	return filepath.Join(localAppData, "mneme")
}

// GetConfigDir returns the platform-appropriate config directory for Windows
// Uses %APPDATA%\mneme
func GetConfigDir() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		// Fallback to user home + AppData/Roaming
		home, err := os.UserHomeDir()
		if err != nil {
			return `C:\mneme` // Last resort fallback
		}
		appData = filepath.Join(home, "AppData", "Roaming")
	}
	return filepath.Join(appData, "mneme")
}

// GetConfigPath returns the full path to the config file for Windows
func GetConfigPath() string {
	return filepath.Join(GetConfigDir(), "mneme.toml")
}
