package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/aws/ec2-macos-utils/internal/system"
)

const (
	networkMonitorDefaultInterval      = 5 * time.Minute
	networkMonitorDefaultStartupDelay  = 5 * time.Minute
	networkMonitorDefaultOutputBaseDir = "/private/var/db/ec2-macos-utils/sysdiagnose"
)

type networkHealthMonitorArgs struct {
	interval           time.Duration
	startupDelay       time.Duration
	outputDir          string
	sysdiagnoseTimeout time.Duration
}

func newNetworkHealthMonitorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "network-health-monitor",
		Short: "monitor network health",
		Long: strings.TrimSpace(`
monitor network health with periodic checks.
A sysdiagnose will be collected on first failure, after which the monitor will exit.

This command requires root privileges. Run with sudo if not running as root.
        `),
	}

	var args networkHealthMonitorArgs
	cmd.Flags().DurationVar(&args.interval, "interval", networkMonitorDefaultInterval, "interval between network checks")
	cmd.Flags().DurationVar(&args.startupDelay, "startup-delay", networkMonitorDefaultStartupDelay, "delay before starting checks")
	cmd.Flags().StringVar(&args.outputDir, "output-base-dir", networkMonitorDefaultOutputBaseDir, "base directory for sysdiagnose output")
	cmd.Flags().DurationVar(&args.sysdiagnoseTimeout, "sysdiagnose-timeout", sysdiagnoseDefaultTimeout, "timeout for sysdiagnose collection")

	cmd.PreRunE = func(cmd *cobra.Command, _ []string) error {
		if os.Geteuid() != 0 {
			return errors.New("root privileges required - run with sudo")
		}

		if args.interval < 0 {
			return errors.New("interval cannot be negative")
		}

		if args.startupDelay < 0 {
			return errors.New("startup delay cannot be negative")
		}

		if args.sysdiagnoseTimeout < sysdiagnoseMinTimeout {
			return fmt.Errorf("timeout must be at least %v to ensure creation can complete", sysdiagnoseMinTimeout)
		}

		return nil
	}

	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		// Get collection prefix for potential use later
		prefix, err := getCollectionPrefix()
		if err != nil {
			logrus.WithError(err).Warn("Failed to get prefix, using 'unknown'")
			prefix = "unknown"
		}

		// Create only the base output directory
		if err := os.MkdirAll(args.outputDir, 0700); err != nil {
			return fmt.Errorf("base output directory creation: %w", err)
		}

		// Check if sysdiagnose already exists in the prefix directory
		prefixDir := filepath.Join(args.outputDir, prefix)
		existing, err := filepath.Glob(filepath.Join(prefixDir, "sysdiagnose_*.tar.gz"))
		if err != nil {
			return fmt.Errorf("invalid glob pattern: %w", err)
		}
		if len(existing) > 0 {
			logrus.Warn("Monitor already captured sysdiagnose for failure, stopping watchdog")
			return nil
		}

		// Set the final output directory
		args.outputDir = prefixDir

		return runNetworkHealthMonitor(cmd.Context(), args)
	}

	return cmd
}

func runNetworkHealthMonitor(ctx context.Context, args networkHealthMonitorArgs) error {
	logrus.WithField("delay", args.startupDelay).Info("Waiting before starting network checks")

	// Handle startup delay
	select {
	case <-time.After(args.startupDelay):
	case <-ctx.Done():
		return ctx.Err()
	}

	timer := time.NewTimer(args.interval)
	defer timer.Stop()

	logrus.WithField("interval", args.interval).Info("Starting network health monitoring")

	sysdiagnoseCollectionArgs := sysdiagnoseArgs{
		outputDir: args.outputDir,
		timeout:   args.sysdiagnoseTimeout,
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			sysdiagnoseCollected, err := checkNetworkAndCollect(ctx, sysdiagnoseCollectionArgs)
			timer.Reset(args.interval)

			if err != nil {
				logrus.WithError(err).Error("Sysdiagnose collection failed")
				continue
			}
			if sysdiagnoseCollected {
				logrus.Info("Sysdiagnose collected, stopping watchdog")
				return nil
			}
		}
	}
}

func checkNetworkAndCollect(ctx context.Context, sysArgs sysdiagnoseArgs) (bool, error) {
	if err := runCheckIMDS(ctx); err != nil {
		logrus.WithError(err).Warn("IMDS check failed, collecting sysdiagnose")

		// Create the directory before collecting sysdiagnose
		if err := os.MkdirAll(sysArgs.outputDir, 0700); err != nil {
			return false, fmt.Errorf("sysdiagnose output directory creation: %w", err)
		}

		if err := runSysdiagnose(ctx, sysArgs); err != nil {
			return false, fmt.Errorf("sysdiagnose collection: %w", err)
		}

		return true, nil
	}

	return false, nil
}

func getCollectionPrefix() (string, error) {
	return system.GetHostIOPlatformUUID()
}
