package system

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadStateEntries_EmbeddedDefault(t *testing.T) {
	entries, err := loadStateEntries(filepath.Join(t.TempDir(), "absent.toml"))

	require.NoError(t, err)
	require.NotEmpty(t, entries)
	assert.Equal(t, networkInterfacesPlist, entries[0].Path, "default should target the interface cache")
}

func TestLoadStateEntries_ExternalConfigReplacesDefault(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cleanup-state.toml")
	require.NoError(t, os.WriteFile(path, []byte(`
[[state]]
path        = "Library/Caches/custom"
description = "custom state"
rationale   = "test"
recursive   = true
`), 0644))

	entries, err := loadStateEntries(path)

	require.NoError(t, err)
	require.Len(t, entries, 1, "external config should replace, not extend, the default")
	assert.Equal(t, "Library/Caches/custom", entries[0].Path)
	assert.True(t, entries[0].Recursive)
}

func TestLoadStateEntries_MalformedConfigErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cleanup-state.toml")
	require.NoError(t, os.WriteFile(path, []byte("this = is = not = toml"), 0644))

	_, err := loadStateEntries(path)

	assert.Error(t, err, "a malformed config must be an error, not a silent fallback")
}

func TestLoadStateEntries_UnreadableConfigErrors(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission-based test cannot run as root")
	}

	path := filepath.Join(t.TempDir(), "cleanup-state.toml")
	require.NoError(t, os.WriteFile(path, []byte("[[state]]\n"), 0000))

	_, err := loadStateEntries(path)

	assert.Error(t, err, "an unreadable config must be an error, not a silent fallback")
}
