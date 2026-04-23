package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseVersionSemver(t *testing.T) {
	t.Run("parse invalid version will return non nil error", func(t *testing.T) {
		v := "xx"
		sv, err := ParseVersion(v)
		assert.Error(t, err)
		assert.Nil(t, sv)
	})

	t.Run("parse valid version will return nil error", func(t *testing.T) {
		v := "1.0"
		sv, err := ParseVersion(v)
		assert.Nil(t, err)
		assert.Equal(t, sv.Major(), uint64(1))
		assert.Equal(t, sv.Minor(), uint64(0))
	})

	t.Run("parse valid version with prefix 'v' will return nil error", func(t *testing.T) {
		v := "v1.0"
		sv, err := ParseVersion(v)
		assert.Nil(t, err)
		assert.Equal(t, sv.Major(), uint64(1))
		assert.Equal(t, sv.Minor(), uint64(0))
	})
}

func TestIncreaseMinorVersion(t *testing.T) {
	t.Run("increase minor version of invalid version will return non nil error", func(t *testing.T) {
		v := "xx"
		sv, err := IncreaseMinorVersion(v)
		assert.Error(t, err)
		assert.Empty(t, sv)
	})

	t.Run("increase minor version of valid version will return nil error", func(t *testing.T) {
		v := "1.0"
		sv, err := IncreaseMinorVersion(v)
		assert.Nil(t, err)
		assert.Equal(t, "1.1", sv)
	})

	t.Run("increase minor version of valid version with prefix 'v' will return nil error", func(t *testing.T) {
		v := "v1.0"
		sv, err := IncreaseMinorVersion(v)
		assert.Nil(t, err)
		assert.Equal(t, "1.1", sv)
	})

	t.Run("increase minor version of empty version will return nil error", func(t *testing.T) {
		v := ""
		sv, err := IncreaseMinorVersion(v)
		assert.Nil(t, err)
		assert.Equal(t, "0.1", sv)
	})
}

func TestIsGreaterThan(t *testing.T) {
	t.Run("new version is greater than old version returns true", func(t *testing.T) {
		assert.True(t, IsGreaterThan("0.2", "0.1"))
	})

	t.Run("new version is equal to old version returns false", func(t *testing.T) {
		assert.False(t, IsGreaterThan("0.1", "0.1"))
	})

	t.Run("new version is less than old version returns false", func(t *testing.T) {
		assert.False(t, IsGreaterThan("0.1", "0.2"))
	})

	t.Run("empty new version is not greater than empty old version returns false", func(t *testing.T) {
		assert.False(t, IsGreaterThan("", ""))
	})

	t.Run("empty new version is not greater than valid old version returns false", func(t *testing.T) {
		assert.False(t, IsGreaterThan("", "0.1"))
	})

	t.Run("valid new version is greater than empty old version returns true", func(t *testing.T) {
		assert.True(t, IsGreaterThan("0.1", ""))
	})

	t.Run("invalid new version is treated as 0.0 and is not greater than valid old version returns false", func(t *testing.T) {
		assert.False(t, IsGreaterThan("xx", "0.1"))
	})

	t.Run("valid new version with prefix v is greater than old version returns true", func(t *testing.T) {
		assert.True(t, IsGreaterThan("v0.2", "0.1"))
	})
}
