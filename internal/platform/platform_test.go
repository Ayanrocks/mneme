package platform

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestCurrent(t *testing.T) {
	current := Current()

	// Current should return one of the valid platform constants
	validPlatforms := map[string]bool{
		PlatformWindows:    true,
		PlatformWindowsWSL: true,
		PlatformLinux:      true,
		PlatformDarwin:     true,
	}

	if !validPlatforms[current] {
		t.Errorf("Current() returned invalid platform: %s", current)
	}

	// On this system, verify it matches the expected platform
	switch runtime.GOOS {
	case "windows":
		if current != PlatformWindows {
			t.Errorf("Expected %s on Windows, got %s", PlatformWindows, current)
		}
	case "darwin":
		if current != PlatformDarwin {
			t.Errorf("Expected %s on macOS, got %s", PlatformDarwin, current)
		}
	case "linux":
		// Could be linux or windows_wsl
		if current != PlatformLinux && current != PlatformWindowsWSL {
			t.Errorf("Expected %s or %s on Linux, got %s", PlatformLinux, PlatformWindowsWSL, current)
		}
	}
}

func TestIsWindows(t *testing.T) {
	result := IsWindows()
	expected := runtime.GOOS == "windows"

	if result != expected {
		t.Errorf("IsWindows() = %v, expected %v", result, expected)
	}
}

func TestIsUnix(t *testing.T) {
	result := IsUnix()
	expected := runtime.GOOS != "windows"

	if result != expected {
		t.Errorf("IsUnix() = %v, expected %v", result, expected)
	}
}

func TestIsWSL(t *testing.T) {
	result := IsWSL()

	// On non-Linux systems, WSL should always be false
	if runtime.GOOS != "linux" && result {
		t.Error("IsWSL() should return false on non-Linux systems")
	}

	// If WSL_DISTRO_NAME is set, should return true
	if runtime.GOOS == "linux" && os.Getenv("WSL_DISTRO_NAME") != "" && !result {
		t.Error("IsWSL() should return true when WSL_DISTRO_NAME is set")
	}
}

func TestGetDataDir(t *testing.T) {
	dir := GetDataDir()

	if dir == "" {
		t.Error("GetDataDir() returned empty string")
	}

	// Should contain "mneme"
	if !strings.Contains(dir, "mneme") {
		t.Errorf("GetDataDir() should contain 'mneme', got: %s", dir)
	}
}

func TestGetConfigDir(t *testing.T) {
	dir := GetConfigDir()

	if dir == "" {
		t.Error("GetConfigDir() returned empty string")
	}

	// Should contain "mneme"
	if !strings.Contains(dir, "mneme") {
		t.Errorf("GetConfigDir() should contain 'mneme', got: %s", dir)
	}
}

func TestGetConfigPath(t *testing.T) {
	path := GetConfigPath()

	if path == "" {
		t.Error("GetConfigPath() returned empty string")
	}

	// Should end with mneme.toml
	if !strings.HasSuffix(path, "mneme.toml") {
		t.Errorf("GetConfigPath() should end with 'mneme.toml', got: %s", path)
	}
}

func TestIsPlatformCompatible(t *testing.T) {
	tests := []struct {
		stored   string
		current  string
		expected bool
	}{
		{PlatformLinux, PlatformLinux, true},
		{PlatformWindows, PlatformWindows, true},
		{PlatformDarwin, PlatformDarwin, true},
		{PlatformWindowsWSL, PlatformWindowsWSL, true},
		{PlatformLinux, PlatformWindows, false},
		{PlatformWindows, PlatformLinux, false},
		{PlatformLinux, PlatformWindowsWSL, false},
		{PlatformDarwin, PlatformLinux, false},
	}

	for _, tt := range tests {
		t.Run(tt.stored+"_to_"+tt.current, func(t *testing.T) {
			result := IsPlatformCompatible(tt.stored, tt.current)
			if result != tt.expected {
				t.Errorf("IsPlatformCompatible(%s, %s) = %v, expected %v",
					tt.stored, tt.current, result, tt.expected)
			}
		})
	}
}
