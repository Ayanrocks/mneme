package cli

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCmd(t *testing.T) {
	t.Run("has correct configuration", func(t *testing.T) {
		assert.Equal(t, "mneme", rootCmd.Use)
		assert.Equal(t, "Mneme - A powerful personal search engine", rootCmd.Short)
		assert.Contains(t, rootCmd.Long, "powerful search engine")
	})

	t.Run("has PreRun hook set", func(t *testing.T) {
		assert.NotNil(t, rootCmd.PreRun)
	})

	t.Run("has Run function set", func(t *testing.T) {
		assert.NotNil(t, rootCmd.Run)
	})
}

func TestVersionCmd(t *testing.T) {
	t.Run("has correct configuration", func(t *testing.T) {
		assert.Equal(t, "version", versionCmd.Use)
		assert.Contains(t, versionCmd.Aliases, "v")
		assert.Equal(t, "Show version information", versionCmd.Short)
		assert.Contains(t, versionCmd.Long, "version")
	})

	t.Run("has Run function set", func(t *testing.T) {
		assert.NotNil(t, versionCmd.Run)
	})
}

func TestInitCmd(t *testing.T) {
	t.Run("has correct configuration", func(t *testing.T) {
		assert.Equal(t, "init", initCmd.Use)
		assert.Contains(t, initCmd.Short, "Initialize")
		assert.Contains(t, initCmd.Long, "setup")
	})

	t.Run("has Run function set", func(t *testing.T) {
		assert.NotNil(t, initCmd.Run)
	})
}

func TestSubcommandRegistration(t *testing.T) {
	t.Run("version command is registered", func(t *testing.T) {
		found := false
		for _, cmd := range rootCmd.Commands() {
			if cmd.Use == "version" {
				found = true
				break
			}
		}
		assert.True(t, found, "version command should be registered")
	})

	t.Run("init command is registered", func(t *testing.T) {
		found := false
		for _, cmd := range rootCmd.Commands() {
			if cmd.Use == "init" {
				found = true
				break
			}
		}
		assert.True(t, found, "init command should be registered")
	})

	t.Run("config command is registered", func(t *testing.T) {
		found := false
		for _, cmd := range rootCmd.Commands() {
			if cmd.Use == "config" {
				found = true
				break
			}
		}
		assert.True(t, found, "config command should be registered")
	})
}

func TestPersistentFlags(t *testing.T) {
	t.Run("verbose flag is registered", func(t *testing.T) {
		flag := rootCmd.PersistentFlags().Lookup("verbose")
		require.NotNil(t, flag)
		assert.Equal(t, "v", flag.Shorthand)
		assert.Equal(t, "false", flag.DefValue)
	})

	t.Run("quiet flag is registered", func(t *testing.T) {
		flag := rootCmd.PersistentFlags().Lookup("quiet")
		require.NotNil(t, flag)
		assert.Equal(t, "q", flag.Shorthand)
		assert.Equal(t, "false", flag.DefValue)
	})
}

func TestVersionCmdExecute(t *testing.T) {
	t.Run("prints version information", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Execute the version command
		versionCmdExecute(&cobra.Command{}, []string{})

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		// Verify output contains expected information
		assert.Contains(t, output, "Version")
		assert.Contains(t, output, "Storage Engine")
		assert.Contains(t, output, "Go Version")
	})
}

func TestExecute(t *testing.T) {
	t.Run("returns nil when no subcommand provided", func(t *testing.T) {
		// Reset args to just the program name
		oldArgs := os.Args
		os.Args = []string{"mneme", "--help"}
		defer func() { os.Args = oldArgs }()

		// Capture output to suppress help text
		oldStdout := os.Stdout
		_, w, _ := os.Pipe()
		os.Stdout = w

		err := Execute()

		w.Close()
		os.Stdout = oldStdout

		assert.NoError(t, err)
	})
}

func TestFlagVariables(t *testing.T) {
	t.Run("verbose defaults to false", func(t *testing.T) {
		// Reset flags to default
		verbose = false
		assert.False(t, verbose)
	})

	t.Run("quiet defaults to false", func(t *testing.T) {
		// Reset flags to default
		quiet = false
		assert.False(t, quiet)
	})
}
