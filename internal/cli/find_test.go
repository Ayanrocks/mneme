package cli

import (
	"testing"
)

func TestFindCommand(t *testing.T) {
	t.Run("find command has correct usage", func(t *testing.T) {
		if findCmd.Use != "find" {
			t.Errorf("Expected use 'find', got '%s'", findCmd.Use)
		}
	})

	t.Run("find command has short description", func(t *testing.T) {
		if findCmd.Short == "" {
			t.Error("Expected non-empty short description")
		}
	})

	t.Run("find command has long description", func(t *testing.T) {
		if findCmd.Long == "" {
			t.Error("Expected non-empty long description")
		}
	})

	t.Run("find command has run function", func(t *testing.T) {
		if findCmd.Run == nil {
			t.Error("Expected non-nil run function")
		}
	})
}

func TestFindCmdExecute_EdgeCases(t *testing.T) {
	t.Run("findCmdExecute does not panic with nil args", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("findCmdExecute panicked: %v", r)
			}
		}()

		// Should handle nil args
		findCmdExecute(findCmd, nil)
	})

	t.Run("findCmdExecute does not panic with empty args", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("findCmdExecute panicked: %v", r)
			}
		}()

		// Should handle empty args (expects at least one query arg)
		findCmdExecute(findCmd, []string{})
	})

	t.Run("findCmdExecute does not panic with single arg", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("findCmdExecute panicked: %v", r)
			}
		}()

		findCmdExecute(findCmd, []string{"test"})
	})

	t.Run("findCmdExecute does not panic with multiple args", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("findCmdExecute panicked: %v", r)
			}
		}()

		findCmdExecute(findCmd, []string{"test", "query", "string"})
	})

	t.Run("findCmdExecute handles special characters in query", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("findCmdExecute panicked: %v", r)
			}
		}()

		findCmdExecute(findCmd, []string{"test@example.com", "special$chars"})
	})
}