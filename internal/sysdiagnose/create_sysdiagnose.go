package sysdiagnose

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	// systemSysdiagnoseExecutable is the hardcoded path to the macOS sysdiagnose executable
	systemSysdiagnoseExecutable = "/usr/bin/sysdiagnose"
)

// Collect executes a full run of sysdiagnose and returns a handle to read the
// resulting archive. Callers should close the returned io.ReadCloser when
// finished. Sysdiagnose requires root privileges to collect system data and an
// error will be returned if called without root privileges.
func Collect(ctx context.Context, archiveName string) (io.ReadCloser, error) {
	// Validate archive name
	if archiveName == "" {
		return nil, errors.New("archive name required")
	}
	if filepath.Base(archiveName) != filepath.Clean(archiveName) {
		return nil, errors.New("archive name must be a valid path basename without directory parts")
	}

	workDir, err := os.MkdirTemp("", "sysdiagnose-helper*")
	if err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	if err = os.Chmod(workDir, 0700); err != nil {
		return nil, fmt.Errorf("unable to chmod sysdiagnose dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(workDir) }()

	archiveOutputPath := filepath.Join(workDir, fmt.Sprintf("%s.tar.gz", archiveName))
	args, err := sysdiagnoseArgs(archiveOutputPath)
	if err != nil {
		return nil, fmt.Errorf("error building sysdiagnose args: %w", err)
	}
	logrus.WithContext(ctx).WithField("args", args).Debug("preparing sysdiagnose collection")

	cmd := exec.CommandContext(ctx, systemSysdiagnoseExecutable, args...)

	logrus.WithContext(ctx).WithFields(logrus.Fields{
		"archive_name": archiveName,
		"command":      cmd.String(),
	}).Info("running sysdiagnose - this produces large archive file in a few minutes, usually 100s of MB")

	tStart := time.Now()
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("error running sysdiagnose: %w", err)
	}

	runtime := time.Since(tStart).Truncate(time.Second)
	logrus.WithContext(ctx).WithField("runtime_seconds", runtime.Seconds()).Info("sysdiagnose collected")

	// now open a handle to retain the IO stream and remove the filesystem entry
	// before returning to caller - this ensures the library call isn't the one
	// leaking data on disk.
	handle, err := os.OpenFile(archiveOutputPath, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("open result handle: %w", err)
	}

	// Note: The file is intentionally removed while keeping the handle open.
	// This is a common pattern in Unix systems - the file will remain accessible
	// through the handle until it's closed, but won't be visible in the filesystem.
	if err := os.Remove(archiveOutputPath); err != nil {
		logrus.WithContext(ctx).WithField("path", archiveOutputPath).Warn("sysdiagnose unable to remove temporary file")
	}

	return handle, nil
}

func sysdiagnoseArgs(outputFileFullPath string) ([]string, error) {
	if outputFileFullPath == "" {
		return nil, errors.New("output file path required")
	}

	return []string{
		"-f", filepath.Dir(outputFileFullPath), // output directory
		"-A", filepath.Base(outputFileFullPath), // archive name
		"-u", // without UI feedback
		"-b", // without showing Finder
	}, nil
}
