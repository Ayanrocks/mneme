package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigCmd(t *testing.T) {
	t.Run("has correct configuration", func(t *testing.T) {
		assert.Equal(t, "config", configCmd.Use)
		assert.Contains(t, configCmd.Short, "Configuration")
		assert.Contains(t, configCmd.Long, "configuration")
	})
}

func TestShowCmd(t *testing.T) {
	t.Run("has correct configuration", func(t *testing.T) {
		assert.Equal(t, "show", showCmd.Use)
		assert.Contains(t, showCmd.Short, "configuration")
		assert.Contains(t, showCmd.Long, "configuration")
	})

	t.Run("has Run function set", func(t *testing.T) {
		assert.NotNil(t, showCmd.Run)
	})
}

func TestAddCmd(t *testing.T) {
	t.Run("has correct configuration", func(t *testing.T) {
		assert.Equal(t, "add", addCmd.Use)
		assert.Contains(t, addCmd.Short, "Add")
		assert.Contains(t, addCmd.Long, "path")
	})

	t.Run("has Run function set", func(t *testing.T) {
		assert.NotNil(t, addCmd.Run)
	})
}

func TestRemoveCmd(t *testing.T) {
	t.Run("has correct configuration", func(t *testing.T) {
		assert.Equal(t, "remove", removeCmd.Use)
		assert.Contains(t, removeCmd.Short, "Remove")
		assert.Contains(t, removeCmd.Long, "path")
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
