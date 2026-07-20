package system

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/sirupsen/logrus"
)

// StateEntry describes a well-known piece of OS state that must be removed
// from an instance before it is suitable for imaging (e.g. AMI creation).
type StateEntry struct {
	// Path is the path of the state relative to the target system root.
	Path string
	// Description is a short, human-readable name for the state.
	Description string
	// Rationale explains why the state must be removed.
	Rationale string
	// Recursive controls whether a directory and all of its contents are removed.
	Recursive bool
}

// cleanupStateEntries is the set of well-known OS state that is removed
// during state cleanup.
var cleanupStateEntries = []StateEntry{
	{
		Path:        "Library/Preferences/SystemConfiguration/NetworkInterfaces.plist",
		Description: "macOS cached interface configuration",
		Rationale:   "affects startup health of networking for EC2 Mac instances",
	},
}

// StateCleaner removes well-known OS state beneath a target system root.
//
// The target root is an internal abstraction to permit future use against
// other targets (e.g. an external volume); commands always target the
// running system's root.
type StateCleaner struct {
	// targetRoot is the root path of the system whose state is cleaned.
	targetRoot string
	// entries is the OS state removed by Cleanup.
	entries []StateEntry
	// dryRun reports cleanup actions without removing state.
	dryRun bool
}

// NewStateCleaner returns a StateCleaner that cleans up the running system.
func NewStateCleaner(dryRun bool) *StateCleaner {
	return newStateCleaner("/", dryRun)
}

// newStateCleaner returns a StateCleaner that cleans up the system rooted at
// the given target root.
func newStateCleaner(targetRoot string, dryRun bool) *StateCleaner {
	return &StateCleaner{
		targetRoot: targetRoot,
		entries:    slices.Clone(cleanupStateEntries),
		dryRun:     dryRun,
	}
}

// Cleanup removes each well-known OS state entry beneath the target root.
// State that does not exist is skipped. Removal continues through errors so
// each entry is attempted; any errors encountered are joined and returned.
// A canceled context stops cleanup before the next entry is attempted.
func (c *StateCleaner) Cleanup(ctx context.Context) error {
	var errs []error

	for _, entry := range c.entries {
		if err := ctx.Err(); err != nil {
			errs = append(errs, err)
			break
		}

		if err := c.remove(entry); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// remove removes a single state entry beneath the target root, logging the
// action taken. Absent state is not an error.
func (c *StateCleaner) remove(entry StateEntry) error {
	path := filepath.Join(c.targetRoot, entry.Path)

	log := logrus.WithFields(logrus.Fields{
		"path":  path,
		"state": entry.Description,
	})

	if _, err := os.Lstat(path); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			log.Info("State not present, nothing to remove")
			return nil
		}
		return fmt.Errorf("inspect %q: %w", path, err)
	}

	if c.dryRun {
		log.Infof("Dry run: would remove state (%s: %s)", entry.Description, entry.Rationale)
		return nil
	}

	var removeErr error
	if entry.Recursive {
		removeErr = os.RemoveAll(path)
	} else {
		removeErr = os.Remove(path)
	}
	if removeErr != nil {
		return fmt.Errorf("remove %q: %w", path, removeErr)
	}

	log.Infof("Removed state (%s: %s)", entry.Description, entry.Rationale)
	return nil
}
