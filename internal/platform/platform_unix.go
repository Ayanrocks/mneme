//go:build !windows

package platform

import (
	"os"
	"path/filepath"
)

// GetDataDir returns the platform-appropriate data directory for Unix systems
// Uses XDG_DATA_HOME or falls back to ~/.local/share/mneme
func GetDataDir() string {
	// Check XDG_DATA_HOME first
	xdgDataHome := os.Getenv("XDG_DATA_HOME")
	if xdgDataHome != "" {
		return filepath.Join(xdgDataHome, "mneme")
	}

	// Fall back to ~/.local/share/mneme
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/mneme" // Last resort fallback
	}
	return filepath.Join(home, ".local", "share", "mneme")
}

// GetConfigDir returns the platform-appropriate config directory for Unix systems
// Uses XDG_CONFIG_HOME or falls back to ~/.config/mneme
func GetConfigDir() string {
	// Check XDG_CONFIG_HOME first
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome != "" {
		return filepath.Join(xdgConfigHome, "mneme")
	}

	// Fall back to ~/.config/mneme
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/mneme" // Last resort fallback
	}
	return filepath.Join(home, ".config", "mneme")
}

// GetConfigPath returns the full path to the config file for Unix systems
func GetConfigPath() string {
	return filepath.Join(GetConfigDir(), "mneme.toml")
}
