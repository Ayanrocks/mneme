package constants

import "mneme/internal/platform"

var (
	// DirPath is the data directory path (platform-aware)
	DirPath = platform.GetDataDir()
	// ConfigPath is the full config file path (platform-aware)
	ConfigPath = platform.GetConfigPath()
)

const (
	AppName          = "mneme"
	ConfigFileName   = "mneme.toml"
	
)
