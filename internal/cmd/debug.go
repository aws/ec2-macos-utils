package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/go-units"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/aws/ec2-macos-utils/internal/sysdiagnose"
)

const (
	// Timeouts
	sysdiagnoseDefaultTimeout = 15 * time.Minute
	sysdiagnoseMinTimeout     = 5 * time.Minute

	// Timestamp format
	sysdiagnoseTimestampFormat = "20060102_150405" // YYYYMMDD_HHMMSS
)

type sysdiagnoseArgs struct {
	outputDir string
	timeout   time.Duration
}

func debugCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug",
		Short: "debug utilities for EC2 macOS instances",
		Long:  "utilities and tools for debugging EC2 macOS instances",
	}

	cmd.AddCommand(createSysdiagnoseCommand())

	return cmd
}

func createSysdiagnoseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-sysdiagnose",
		Short: "create sysdiagnose archive",
		Long: strings.TrimSpace(`
creates a sysdiagnose archive including logs, system stats,
and other debug data. The resulting archive will be saved in the specified
output directory.

This command requires root privileges. Run with sudo if not running as root.
        `),
	}

	var args sysdiagnoseArgs
	cmd.Flags().StringVar(&args.outputDir, "output-dir", os.TempDir(), "directory where the sysdiagnose archive will be saved")
	cmd.Flags().DurationVar(&args.timeout, "timeout", sysdiagnoseDefaultTimeout, "set the timeout for creation (e.g. 10m, 30m, 1.5h)")

	cmd.RunE = func(cmd *cobra.Command, cmdArgs []string) error {
		if os.Geteuid() != 0 {
			return errors.New("root privileges required - run with sudo")
		}

		ctx := cmd.Context()

		if args.timeout < sysdiagnoseMinTimeout {
			return fmt.Errorf("timeout must be at least %v to ensure creation can complete", sysdiagnoseMinTimeout)
		}

		timeoutCtx, cancel := context.WithTimeout(ctx, args.timeout)
		defer cancel()
		ctx = timeoutCtx

		logrus.WithField("args", args).Debug("Running sysdiagnose")
		if err := runSysdiagnose(ctx, args); err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return errors.New("creation timeout exceeded")
			}
			return err
		}

		return nil
	}

	return cmd
}

func runSysdiagnose(ctx context.Context, args sysdiagnoseArgs) error {
	// Create output directory with owner-only permissions (rwx------) since it will contain sensitive diagnostic data
	if err := os.MkdirAll(args.outputDir, 0700); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	timestamp := time.Now().UTC().Format(sysdiagnoseTimestampFormat)
	archiveName := fmt.Sprintf("sysdiagnose_%s", timestamp)
	outputPath := filepath.Join(args.outputDir, archiveName+".tar.gz")

	logrus.WithFields(logrus.Fields{
		"output_path": outputPath,
	}).Info("Starting sysdiagnose creation")

	outputReader, err := sysdiagnose.Collect(ctx, archiveName)
	if err != nil {
		return fmt.Errorf("failed to create sysdiagnose: %w", err)
	}
	defer func() { _ = outputReader.Close() }()

	// Create output file with read-only permissions (r--------) since diagnostic data should not be modified
	output, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0400)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
	}
	defer func() { _ = output.Close() }()

	written, err := io.Copy(output, outputReader)
	if err != nil {
		// Ignore error from Remove() since:
		// 1. We're already in an error state from io.Copy
		// 2. If Remove() fails, the incomplete/corrupt file remaining is not critical
		_ = os.Remove(outputPath)
		return fmt.Errorf("failed to write sysdiagnose data: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"output_path": outputPath,
		"bytes":       written,
	}).Infof("Sysdiagnose creation completed (%s)", units.HumanSize(float64(written)))

	return nil
}
