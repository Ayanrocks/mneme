package platform

import (
	"os"
	"runtime"
	"strings"
)

// Platform constants - used in VERSION file and for detection
const (
	PlatformWindows    = "windows"
	PlatformWindowsWSL = "windows_wsl" // Linux running under WSL
	PlatformLinux      = "linux"
	PlatformDarwin     = "darwin" // macOS
)

// Current returns the current platform identifier
func Current() string {
	if runtime.GOOS == "windows" {
		return PlatformWindows
	}
	if runtime.GOOS == "darwin" {
		return PlatformDarwin
	}
	// Linux - check if running under WSL
	if IsWSL() {
		return PlatformWindowsWSL
	}
	return PlatformLinux
}

// IsWindows returns true for native Windows only
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsUnix returns true for Unix-like systems (Linux, macOS, WSL)
func IsUnix() bool {
	return runtime.GOOS != "windows"
}

// IsWSL detects if running under Windows Subsystem for Linux
func IsWSL() bool {
	// Quick check: WSL sets this environment variable
	if os.Getenv("WSL_DISTRO_NAME") != "" {
		return true
	}

	// Fallback: check /proc/version for WSL indicators
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}

	version := strings.ToLower(string(data))
	return strings.Contains(version, "microsoft") || strings.Contains(version, "wsl")
}

// IsPlatformCompatible checks if two platform strings are compatible
// For now, platforms must match exactly for compatibility
func IsPlatformCompatible(stored, current string) bool {
	return stored == current
}
