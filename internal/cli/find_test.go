package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindCmdFlags(t *testing.T) {
	t.Run("regex flag is registered", func(t *testing.T) {
		flag := findCmd.Flags().Lookup("regex")
		assert.NotNil(t, flag)
		assert.Equal(t, "r", flag.Shorthand)
		assert.Equal(t, "", flag.DefValue)
	})
}
