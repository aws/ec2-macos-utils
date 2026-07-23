package system

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	logrus.SetOutput(io.Discard)
}

const networkInterfacesPlist = "Library/Preferences/SystemConfiguration/NetworkInterfaces.plist"

// writeFile creates a file (and its parent directories) beneath root.
func writeFile(t *testing.T, root string, relPath string) string {
	t.Helper()

	path := filepath.Join(root, relPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
	require.NoError(t, os.WriteFile(path, []byte("test data"), 0644))

	return path
}

// newTestCleaner builds a cleaner that uses the embedded default config (no
// external config path) rooted at root.
func newTestCleaner(t *testing.T, root string, dryRun bool) *StateCleaner {
	t.Helper()

	cleaner, err := newStateCleaner(root, "", dryRun)
	require.NoError(t, err)

	return cleaner
}

func TestCleanup_RemovesKnownState(t *testing.T) {
	root := t.TempDir()
	path := writeFile(t, root, networkInterfacesPlist)

	err := newTestCleaner(t, root, false).Cleanup(context.Background())

	assert.NoError(t, err)
	assert.NoFileExists(t, path, "well-known state should be removed")
}

func TestCleanup_AbsentStateIsNotError(t *testing.T) {
	root := t.TempDir()

	err := newTestCleaner(t, root, false).Cleanup(context.Background())

	assert.NoError(t, err, "cleanup should succeed when state is already absent")
}

func TestCleanup_IsIdempotent(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, networkInterfacesPlist)

	cleaner := newTestCleaner(t, root, false)

	assert.NoError(t, cleaner.Cleanup(context.Background()))
	assert.NoError(t, cleaner.Cleanup(context.Background()), "second cleanup should also succeed")
}

func TestCleanup_LeavesUnrelatedStateIntact(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, networkInterfacesPlist)
	unrelated := writeFile(t, root, "Library/Preferences/SystemConfiguration/preferences.plist")

	err := newTestCleaner(t, root, false).Cleanup(context.Background())

	assert.NoError(t, err)
	assert.FileExists(t, unrelated, "unrelated state should not be removed")
}

func TestCleanup_DryRunLeavesStateIntact(t *testing.T) {
	root := t.TempDir()
	path := writeFile(t, root, networkInterfacesPlist)

	err := newTestCleaner(t, root, true).Cleanup(context.Background())

	assert.NoError(t, err)
	assert.FileExists(t, path, "dry-run should not remove well-known state")
}

func TestCleanup_RemovesSymlinkWithoutRemovingTarget(t *testing.T) {
	for _, recursive := range []bool{false, true} {
		t.Run(map[bool]string{false: "non-recursive", true: "recursive"}[recursive], func(t *testing.T) {
			root := t.TempDir()
			target := writeFile(t, root, "target/state")
			link := filepath.Join(root, networkInterfacesPlist)
			require.NoError(t, os.MkdirAll(filepath.Dir(link), 0755))
			require.NoError(t, os.Symlink(target, link))

			cleaner := newTestCleaner(t, root, false)
			cleaner.entries = []StateEntry{
				{Path: networkInterfacesPlist, Description: "linked state", Recursive: recursive},
			}

			err := cleaner.Cleanup(context.Background())

			assert.NoError(t, err)
			assert.NoFileExists(t, link, "cleanup should remove the symlink")
			assert.FileExists(t, target, "cleanup should not remove a symlink's target")
		})
	}
}

func TestCleanup_RecursiveControlsDirectoryRemoval(t *testing.T) {
	tests := []struct {
		name      string
		recursive bool
		wantError bool
	}{
		{name: "non-recursive", recursive: false, wantError: true},
		{name: "recursive", recursive: true, wantError: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			const directory = "Library/Caches/test-state"
			writeFile(t, root, filepath.Join(directory, "nested/state"))

			cleaner := newTestCleaner(t, root, false)
			cleaner.entries = []StateEntry{
				{Path: directory, Description: "directory state", Recursive: tt.recursive},
			}

			err := cleaner.Cleanup(context.Background())

			if tt.wantError {
				assert.Error(t, err, "non-recursive cleanup should reject a non-empty directory")
				assert.DirExists(t, filepath.Join(root, directory))
				return
			}

			assert.NoError(t, err)
			assert.NoDirExists(t, filepath.Join(root, directory), "recursive cleanup should remove the directory tree")
		})
	}
}

func TestCleanup_CanceledContextLeavesStateIntact(t *testing.T) {
	root := t.TempDir()
	path := writeFile(t, root, networkInterfacesPlist)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := newTestCleaner(t, root, false).Cleanup(ctx)

	assert.ErrorIs(t, err, context.Canceled)
	assert.FileExists(t, path, "canceled cleanup should not remove state")
}

func TestCleanup_ContinuesThroughErrors(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission-based test cannot run as root")
	}

	root := t.TempDir()

	// First entry's removal fails: its parent directory is read-only.
	blocked := writeFile(t, root, "blocked/state")
	require.NoError(t, os.Chmod(filepath.Dir(blocked), 0555))
	t.Cleanup(func() { _ = os.Chmod(filepath.Dir(blocked), 0755) })

	// Second entry should still be removed.
	removable := writeFile(t, root, networkInterfacesPlist)

	cleaner := newTestCleaner(t, root, false)
	cleaner.entries = []StateEntry{
		{Path: "blocked/state", Description: "blocked state"},
		{Path: networkInterfacesPlist, Description: "removable state"},
	}

	err := cleaner.Cleanup(context.Background())

	assert.Error(t, err, "blocked removal should be reported")
	assert.NoFileExists(t, removable, "cleanup should continue past a failed entry")
}

func TestNewStateCleaner_TargetsRunningSystem(t *testing.T) {
	cleaner, err := NewStateCleaner(false)
	require.NoError(t, err)

	assert.Equal(t, "/", cleaner.targetRoot)
	assert.NotEmpty(t, cleaner.entries)
	assert.False(t, cleaner.dryRun)
}

func TestNewStateCleaner_IndependentEntries(t *testing.T) {
	first := newTestCleaner(t, t.TempDir(), false)
	second := newTestCleaner(t, t.TempDir(), false)

	original := second.entries[0].Path
	first.entries[0].Path = "mutated"

	assert.Equal(t, original, second.entries[0].Path, "cleaners should not share entry storage")
}
