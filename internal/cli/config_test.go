package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigCmd(t *testing.T) {
	t.Run("has correct configuration", func(t *testing.T) {
		assert.Equal(t, "config", configCmd.Use)
		assert.Equal(t, "Configuration commands", configCmd.Short)
		assert.Equal(t, "Configuration commands", configCmd.Long)
	})
}

func TestShowCmd(t *testing.T) {
	t.Run("has correct configuration", func(t *testing.T) {
		assert.Equal(t, "show", showCmd.Use)
		assert.Equal(t, "show configuration values", showCmd.Short)
		assert.Equal(t, "show configuration values", showCmd.Long)
	})

	t.Run("has Run function set", func(t *testing.T) {
		assert.NotNil(t, showCmd.Run)
	})
}

func TestAddCmd(t *testing.T) {
	t.Run("has correct configuration", func(t *testing.T) {
		assert.Equal(t, "add", addCmd.Use)
		assert.Equal(t, "add path to index", addCmd.Short)
		assert.Equal(t, "add path to index", addCmd.Long)
	})

	t.Run("has Run function set", func(t *testing.T) {
		assert.NotNil(t, addCmd.Run)
	})
}

func TestRemoveCmd(t *testing.T) {
	t.Run("has correct configuration", func(t *testing.T) {
		assert.Equal(t, "remove", removeCmd.Use)
		assert.Equal(t, "remove path from indexing", removeCmd.Short)
		assert.Equal(t, "remove path from indexing", removeCmd.Long)
	})

	t.Run("has Run function set", func(t *testing.T) {
		assert.NotNil(t, removeCmd.Run)
	})

	t.Run("has all flag registered", func(t *testing.T) {
		flag := removeCmd.Flags().Lookup("all")
		require.NotNil(t, flag, "--all flag should be registered")
		assert.Equal(t, "a", flag.Shorthand)
		assert.Equal(t, "false", flag.DefValue)
	})
}

func TestConfigSubcommandRegistration(t *testing.T) {
	t.Run("show command is registered under config", func(t *testing.T) {
		found := false
		for _, cmd := range configCmd.Commands() {
			if cmd.Use == "show" {
				found = true
				break
			}
		}
		assert.True(t, found, "show command should be registered under config")
	})

	t.Run("add command is registered under config", func(t *testing.T) {
		found := false
		for _, cmd := range configCmd.Commands() {
			if cmd.Use == "add" {
				found = true
				break
			}
		}
		assert.True(t, found, "add command should be registered under config")
	})

	t.Run("remove command is registered under config", func(t *testing.T) {
		found := false
		for _, cmd := range configCmd.Commands() {
			if cmd.Use == "remove" {
				found = true
				break
			}
		}
		assert.True(t, found, "remove command should be registered under config")
	})
}

func TestConfigCmdCount(t *testing.T) {
	t.Run("has exactly 3 subcommands", func(t *testing.T) {
		assert.Equal(t, 3, len(configCmd.Commands()), "ConfigCmd should have exactly 3 subcommands (show, add, remove)")
	})
}
