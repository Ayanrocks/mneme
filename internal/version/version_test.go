package version

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMnemeVersion(t *testing.T) {
	t.Run("follows semver format", func(t *testing.T) {
		// Simple semver pattern: major.minor.patch
		semverPattern := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
		assert.True(t, semverPattern.MatchString(MnemeVersion), "MnemeVersion should follow semver format (x.y.z)")
	})

	t.Run("is not empty", func(t *testing.T) {
		assert.NotEmpty(t, MnemeVersion)
	})
}

func TestMnemeStorageEngineVersion(t *testing.T) {
	t.Run("follows semver format", func(t *testing.T) {
		// Simple semver pattern: major.minor.patch
		semverPattern := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
		assert.True(t, semverPattern.MatchString(MnemeStorageEngineVersion), "MnemeStorageEngineVersion should follow semver format (x.y.z)")
	})

	t.Run("is not empty", func(t *testing.T) {
		assert.NotEmpty(t, MnemeStorageEngineVersion)
	})
}

func TestVersionConsistency(t *testing.T) {
	t.Run("versions are defined", func(t *testing.T) {
		// Both versions should be defined
		assert.NotEqual(t, "", MnemeVersion)
		assert.NotEqual(t, "", MnemeStorageEngineVersion)
	})
}
