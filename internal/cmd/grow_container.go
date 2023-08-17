package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/aws/ec2-macos-utils/internal/contextual"
	"github.com/aws/ec2-macos-utils/internal/diskutil"
	"github.com/aws/ec2-macos-utils/internal/diskutil/identifier"
	"github.com/aws/ec2-macos-utils/internal/diskutil/types"
)

// growDefaultTimeout is the default maximum run duration of 5 minutes. This time limit should be sufficiently long
// to allow macOS's diskutil command to execute for a variety of disk sizes. Anything beyond this limit will be treated
// as unresponsive and the process will be terminated. This default time limit can be overridden with a flag.
const growDefaultTimeout = 5 * time.Minute

// growContainer is a struct for holding all information passed into the grow container command.
type growContainer struct {
	dryrun  bool
	id      string
	timeout time.Duration
}

// growContainerCommand creates a new command which grows APFS containers to their maximum size.
func growContainerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grow",
		Short: "resize container to max size",
		Long: strings.TrimSpace(`
grow resizes the container to its maximum size using
'diskutil'. The container to operate on can be specified
with its identifier (e.g. disk1 or /dev/disk1). The string
'root' may be provided to resize the OS's root volume.
		`),
	}

	// Set up the flags to be passed into the command
	growArgs := growContainer{}
	cmd.PersistentFlags().StringVar(&growArgs.id, "id", "", `container identifier to be resized or "root"`)
	cmd.PersistentFlags().BoolVar(&growArgs.dryrun, "dry-run", false, "run command without mutating changes")
	cmd.PersistentFlags().DurationVar(&growArgs.timeout, "timeout", growDefaultTimeout, "Set the timeout for the command (e.g. 30s, 1m, 1.5h), 0s will disable the timeout")
	cmd.MarkPersistentFlagRequired("id")

	// Set up the command's pre-run to check for root permissions.
	// This is necessary since diskutil repairDisk requires root permissions to run.
	cmd.PreRunE = assertRootPrivileges

	// Set up the command's run function
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if growArgs.timeout != 0 {
			ctx, _ = context.WithTimeout(ctx, growArgs.timeout)
		}

		product := contextual.Product(ctx)
		if product == nil {
			return errors.New("product required in context")
		}

		logrus.WithField("product", product).Info("Configuring diskutil for product")
		d, err := diskutil.ForProduct(product)
		if err != nil {
			return err
		}

		if growArgs.dryrun {
			d = diskutil.Dryrun(d)
		}

		logrus.WithField("args", growArgs).Debug("Running grow command with args")
		if err := run(ctx, d, growArgs); err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return errors.New("timeout exceeded")
			}

			return err
		}

		return nil
	}

	return cmd
}

// run attempts to grow the disk for the specified device identifier to its maximum size using diskutil.GrowContainer.
func run(ctx context.Context, utility diskutil.DiskUtil, args growContainer) error {
	di, err := getTargetDiskInfo(ctx, utility, args.id)
	if err != nil {
		return fmt.Errorf("cannot grow container: %w", err)
	}

	logrus.WithField("device_id", di.DeviceIdentifier).Info("Attempting to grow container...")
	if err := diskutil.GrowContainer(ctx, utility, di); err != nil {
		// Don't treat FreeSpaceErrors as fatal, instead exit quietly since there's nothing else to do.
		if errors.As(err, &diskutil.FreeSpaceError{}) {
			logrus.WithField("id", args.id).Info("Nothing to do without free space, stopping command")
			return nil
		}

		return err
	}

	logrus.WithField("device_id", di.ParentWholeDisk).Info("Fetching updated information for device...")
	updatedDi, err := getTargetDiskInfo(ctx, utility, di.ParentWholeDisk)
	if err != nil {
		logrus.WithError(err).Error("Error while fetching updated disk information")
		return err
	}
	logrus.WithFields(logrus.Fields{
		"device_id":  di.DeviceIdentifier,
		"total_size": humanize.Bytes(updatedDi.TotalSize),
	}).Info("Successfully grew device to maximum size")

	return nil
}

// getTargetDiskInfo retrieves the disk info for the specified target identifier. If the identifier is "root", simply
// return the disk information for "/". Otherwise, check if the identifier exists in the system partitions before
// returning the disk information.
func getTargetDiskInfo(ctx context.Context, du diskutil.DiskUtil, target string) (*types.DiskInfo, error) {
	if strings.EqualFold("root", target) {
		return du.Info(ctx, "/")
	}

	partitions, err := du.List(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot list partitions: %w", err)
	}

	if err := validateDeviceID(target, partitions); err != nil {
		return nil, fmt.Errorf("invalid target: %w", err)
	}

	return du.Info(ctx, target)
}

// validateDeviceID verifies if the provided ID is a valid device identifier or device node.
func validateDeviceID(id string, partitions *types.SystemPartitions) error {
	// Check if ID is provided
	if strings.TrimSpace(id) == "" {
		return errors.New("empty device id")
	}

	// Get the device identifier
	deviceID := identifier.ParseDiskID(id)
	if deviceID == "" {
		return errors.New("id does not match the expected device identifier format")
	}

	// Check the device directory for the given identifier
	for _, name := range partitions.AllDisks {
		if strings.EqualFold(name, deviceID) {
			return nil
		}
	}

	return errors.New("invalid device identifier")
}
