package system

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// StateEntry describes a well-known piece of OS state that must be removed
// from an instance before it is suitable for imaging (e.g. AMI creation).
type StateEntry struct {
	// Path is the path of the state relative to the target system root.
	Path string `toml:"path"`
	// Description is a short, human-readable name for the state.
	Description string `toml:"description"`
	// Rationale explains why the state must be removed.
	Rationale string `toml:"rationale"`
	// Recursive controls whether a directory and all of its contents are removed.
	Recursive bool `toml:"recursive"`
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
// State entries are loaded from the external config if present, otherwise
// from the embedded default; a config that cannot be loaded is an error.
func NewStateCleaner(dryRun bool) (*StateCleaner, error) {
	return newStateCleaner("/", defaultConfigPath, dryRun)
}

// newStateCleaner returns a StateCleaner that cleans up the system rooted at
// the given target root, loading state entries from configPath (or the
// embedded default when configPath is absent).
func newStateCleaner(targetRoot string, configPath string, dryRun bool) (*StateCleaner, error) {
	entries, err := loadStateEntries(configPath)
	if err != nil {
		return nil, err
	}

	return &StateCleaner{
		targetRoot: targetRoot,
		entries:    entries,
		dryRun:     dryRun,
	}, nil
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
