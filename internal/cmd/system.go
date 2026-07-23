package cmd

import (
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/aws/ec2-macos-utils/internal/system"
)

func systemCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "system",
		Short: "system provisioning & helper utilities",
		Long: strings.TrimSpace(`
System provisioning and management utilities for EC2 Mac instances & image
building.
        `),
	}

	cmd.AddCommand(cleanupStateCommand())

	return cmd
}

func cleanupStateCommand() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "cleanup-state",
		Short: "remove OS state for instance imaging",
		Long: strings.TrimSpace(`
removes well-known macOS state from the running system that must not be
carried into an image (AMI) built from this instance.

This removes leftover OS state files, such as the network interface configuration
cache (NetworkInterfaces.plist), required when creating & after provisioning
derived images from an instance.

This command requires root privileges. Run with sudo if not running as root.
        `),
		PreRunE: assertRootPrivileges,
		RunE: func(cmd *cobra.Command, args []string) error {
			logrus.Info("Starting system state cleanup")

			cleaner, err := system.NewStateCleaner(dryRun)
			if err != nil {
				logrus.WithError(err).Error("Unable to load cleanup-state configuration")
				return err
			}

			if err := cleaner.Cleanup(cmd.Context()); err != nil {
				logrus.WithError(err).Error("System state cleanup failed")
				return err
			}

			logrus.Info("System state cleanup complete")
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "run command without mutating changes")

	return cmd
}
