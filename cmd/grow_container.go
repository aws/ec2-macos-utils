package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/ec2-macos-utils/internal/build"
	"github.com/aws/ec2-macos-utils/pkg/diskutil"
	"github.com/aws/ec2-macos-utils/pkg/diskutil/types"

	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const diskIDRegex = "disk[0-9]+"

// growContainer is a struct for holding all information passed into the grow container command.
type growContainer struct {
	id string
}

// NewGrowCommand creates a new command which grows APFS containers to their maximum size.
func NewGrowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grow",
		Short: "resize container to max size",
		Long: strings.TrimSpace(`
			grow resizes the container to its maximum size using
			'diskutil'. The container to operate on can be specified
			with its identifier (e.g. disk1 or /dev/disk1). The string
			'root' may be provided to resize the OS's root volume.
			
			NOTE: instances must be rebooted after resizing an EBS volume
		`),
	}

	// Set up the flags to be passed into the command
	growArgs := growContainer{}
	cmd.PersistentFlags().StringVarP(&growArgs.id, "id", "", "", "container identifier to be resized")
	cmd.MarkPersistentFlagRequired("id")

	// Set up the command's run function
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		logrus.WithField("product", build.Product).Info("Configuring diskutil for product")
		d, err := diskutil.ForProduct(build.Product)
		if err != nil {
			return err
		}

		logrus.WithField("args", growArgs).Debug("Running grow command with args")
		if err := run(d, growArgs); err != nil {
			return err
		}

		return nil
	}

	return cmd
}

// init initializes the resizeContainer command, all sub-commands, and sets their respective flags.
func init() {
	// Add the resize container command and sub-commands to the root command
	rootCmd.AddCommand(NewGrowCommand())
}

// run performs the following operations:
//   1. Fetch the full list of system disks and partitions.
//   2. Validate the provided id exists.
//   3. Fetch the disk information for the provided id.
//   4. Fetch the container's parent disk information.
//   5. Check if there's enough available space to execute diskutil's resizeContainer command.
//   6. Attempt to repair the container's parent disk.
//   7. Attempt to resize the container to use all available free space.
//   8. Fetch the latest disk information for the container to output its new size.
func run(utility diskutil.DiskUtil, args growContainer) error {
	// Get the list of all disks and partitions in the system
	var listArgs []string
	logrus.Info("Fetching all disk and partition information...")
	partitions, err := utility.List(listArgs)
	if err != nil {
		return fmt.Errorf("failed to fetch all disk and partition information: %w", err)
	}
	logrus.WithField("partitions", partitions).Debug("Found partition information")

	// Set up the disk pointer to be initialized based on the contents of the provided disk id
	var container *types.DiskInfo

	// Check if the id flag is "root", an identifier (e.g. disk1), or node (e.g. /dev/disk1)
	logrus.WithField("id", args.id).Debug("Checking if device ID is \"root\"")
	if strings.EqualFold(args.id, "root") {
		logrus.Info("Searching for root container to resize...")
		container, err = rootContainer(utility)
		if err != nil {
			return err
		}
		logrus.WithField("container", container).Debug("Found container information")
	} else {
		// Check that the given container ID is valid
		logrus.WithField("id", args.id).Info("Validating container ID...")
		valid, err := validateDeviceID(args.id, partitions)
		if err != nil {
			return err
		}
		if !valid {
			logrus.WithField("id", args.id).Warn("Container ID is not a valid device ID")
			return fmt.Errorf("error container ID [%s] is not valid", args.id)
		}
		logrus.WithField("id", args.id).Info("Container ID is valid")

		// Get the disk information for the container
		logrus.Info("Fetching container information...")
		container, err = utility.Info(args.id)
		if err != nil {
			return err
		}
		logrus.WithField("container", container).Debug("Found container information")
	}
	logrus.WithFields(logrus.Fields{
		"id":   container.DeviceIdentifier,
		"size": humanize.Bytes(container.Size),
	}).Info("Successfully loaded disk information")

	// Attempt to resize the container
	logrus.Info("Attempting to grow container...")
	message, err := diskutil.GrowContainer(container, partitions, utility)
	if err != nil {
		// Check if the error is a MinimumGrowSpaceError and return without an error if it is
		if _, ok := err.(diskutil.MinimumGrowSpaceError); ok {
			logrus.WithError(err).Warn("Could not grow the container")
			return nil
		}

		logrus.WithField("message", message).Warn("Error growing the container", message)
		return fmt.Errorf("error growing the container: %w", err)
	}
	logrus.Infof("Successfully completed with message: %s", message)

	return nil
}

// validateDeviceID verifies if the provided ID is a valid device identifier or device node.
func validateDeviceID(id string, partitions *types.SystemPartitions) (valid bool, err error) {
	// Check if ID is provided
	if len(id) == 0 {
		return false, fmt.Errorf("no ID provided")
	}

	// Check if the ID is a device node or device identifier
	if !strings.HasPrefix(id, "/dev/disk") && !strings.HasPrefix(id, "disk") {
		return false, fmt.Errorf("ID [%s] does not start with \"/dev/disk\" or \"disk\"", id)
	}

	// Get the device identifier
	idRegex := regexp.MustCompile(diskIDRegex)
	deviceID := idRegex.FindString(id)
	if deviceID == "" {
		return false, fmt.Errorf("ID [%s] does not contain the expected expression \"%s\"", id, diskIDRegex)
	}

	// Check the device directory for the given identifier
	for _, name := range partitions.AllDisks {
		if strings.EqualFold(name, deviceID) {
			return true, nil
		}
	}

	return false, nil
}

// rootContainer determines the ID for the container which is mounted as root.
func rootContainer(utility diskutil.DiskUtil) (container *types.DiskInfo, err error) {
	// Get the disk information for the root file system
	container, err = utility.Info("/")
	if err != nil {
		return nil, err
	}

	// Replace the root disk's DeviceIdentifier with the identifier for the container reference.
	// This is necessary since the growContainer() function utilizes the DeviceIdentifier field and expects
	// a container reference. The function expects a DeviceIdentifier matching the format "disk2" but the
	// DeviceIdentifier returned from the call getDiskInformation("/") looks like "disk2s4s1" which will cause
	// growContainer() to fail.
	if container.APFSContainerReference != "" {
		container.DeviceIdentifier = container.APFSContainerReference
	} else {
		idRegex := regexp.MustCompile(diskIDRegex)
		id := idRegex.FindString(container.DeviceIdentifier)
		container.DeviceIdentifier = id
	}

	return container, nil
}
