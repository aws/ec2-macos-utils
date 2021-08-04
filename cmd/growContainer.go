package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/ec2-macos-utils/pkg/diskutil"
	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	// MinimumGrowSpaceRequired defines the minimum amount of free space (in bytes) required to attempt running
	// diskutil's resize command.
	MinimumGrowSpaceRequired = 1000000
)

var (
	// ContainerID is the identifier for the container to be resized or "root"
	ContainerID string
)

// growContainerCmd represents the grow command which provides functionality for growing APFS containers
// to their maximum size.
var growContainerCmd = &cobra.Command{
	Use:   "grow",
	Short: "Resizes a container to its maximum size",
	Long: `grow attempts to resize the specified container to its 
maximum size using Apple's diskutil tool. The container can be
specified with its identifier (e.g. disk1 or /dev/disk1) or
with "root" if the target container is the one with the OS root.'

Note: If the EBS Volume size was changed and the instance hasn't 
been restarted yet, this command will fail to resize the container
until the instance has been restarted.`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		// Create a new DiskUtil that is compatible with the current underlying product version
		utility, err := diskutil.NewDiskUtil(Product)
		if err != nil {
			return fmt.Errorf("error creating new DiskUtil for ProductVersion [%v]: %w", Product, err)
		}

		return run(utility, ContainerID)
	},
}

// MinimumGrowSpaceError defines an error to distinguish when there's not enough space to grow the specified container.
type MinimumGrowSpaceError struct {
	message string
}

// Error provides the implementation for the error interface.
func (e MinimumGrowSpaceError) Error() string {
	return e.message
}

// init initializes the resizeContainer command, all sub-commands, and sets their respective flags.
func init() {
	// Define flags used in the resize container command
	name := "id"
	shorthand := ""
	value := ""
	description := "the ID of the APFS Container or \"root\" (required)"
	growContainerCmd.PersistentFlags().StringVarP(&ContainerID, name, shorthand, value, description)
	growContainerCmd.MarkPersistentFlagRequired(name)

	// Add the resize container command and sub-commands to the root command
	rootCmd.AddCommand(growContainerCmd)
}

// run is the the main runner for the grow command. It performs the following operations:
//   1. Fetch the full list of system disks and partitions.
//   2. Validate the given id exists.
//   3. Fetch the disk information for the provided id.
//   4. Fetch the container's parent disk information.
//   5. Check if there's enough available space to execute diskutil's resizeContainer command.
//   6. Attempt to repair the container's parent disk.
//   7. Attempt to resize the container to use all available free space.
//   8. Fetch the latest disk information for the container to output its new size.
func run(utility diskutil.DiskUtil, id string) (err error) {
	// Get the list of all disks and partitions in the system
	var args []string
	logrus.Info("Fetching all disk and partition information...")
	partitions, err := utility.List(args)
	if err != nil {
		return fmt.Errorf("failed to fetch all disk and partition information: %w", err)
	}
	logrus.WithField("partitions", partitions).Debug("Found partition information")

	// Setup the disk pointer to be initialized based on the contents of id
	var container *diskutil.DiskInfo

	// Check if the id flag is "root", an identifier (e.g. disk1), or node (e.g. /dev/disk1)
	logrus.WithField("id", id).Debug("Checking if device ID is \"root\"")
	if strings.EqualFold(id, "root") {
		logrus.Info("Searching for root container to resize...")
		container, err = rootContainer(utility)
		if err != nil {
			return err
		}
		logrus.WithField("container", container).Debug("Found container information")
	} else {
		// Check that the given container ID is valid
		logrus.WithField("id", id).Info("Validating container ID...")
		valid, err := validateDeviceID(id, partitions)
		if err != nil {
			return err
		}
		if !valid {
			logrus.WithField("id", id).Warn("Container ID is not a valid device ID")
			return fmt.Errorf("error container ID [%s] is not valid", id)
		}
		logrus.WithField("id", id).Info("Container ID is valid")

		// Get the disk information for the container
		logrus.Info("Fetching disk information...")
		container, err = utility.Info(id)
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
	message, err := growContainer(container, partitions, utility)
	if err != nil {
		// Check if the error is a MinimumGrowSpaceError and return without an error if it is
		if _, ok := err.(MinimumGrowSpaceError); ok {
			logrus.WithError(err).Warn("Could not grow the container")
			return nil
		}

		logrus.WithField("message", message).Warn("Error growing the container", message)
		return fmt.Errorf("error growing the container: %w", err)
	}
	logrus.Infof("Successfully completed with message: %s", message)

	return nil
}

// growContainer grows a container to its maximum size given an ID.
func growContainer(disk *diskutil.DiskInfo, partitions *diskutil.SystemPartitions, utility diskutil.DiskUtil) (message string, err error) {
	// Attempt to repair the parent disk in order to get updated amount of free space
	logrus.Info("Attempting to repair the parent disk...")
	message, err = repairParentDisk(disk, partitions, utility)
	if err != nil {
		if _, ok := err.(MinimumGrowSpaceError); ok {
			return message, err
		}

		logrus.WithError(err).Warn("Failed to repair the parent disk, attempting to continue anyways...")
	} else {
		logrus.WithField("disk_id", disk.DeviceIdentifier).Info("No Physical Stores found for disk, attempting to repair...")
	}

	// Attempt to resize the container to its maximum size
	logrus.WithField("id", disk.DeviceIdentifier).Info("Resizing the container use full partition...")
	out, err := utility.ResizeContainer(disk.DeviceIdentifier, "0")
	if err != nil {
		return out, err
	}
	logrus.WithField("out", out).Debug("Resize output")
	logrus.Info("Successfully resized the container")

	// Get updated container size
	logrus.Info("Fetching updated disk information...")
	updatedDisk, err := utility.Info(disk.DeviceIdentifier)
	if err != nil {
		return fmt.Sprintf("failed to fetch updated disk information for container [%s]", disk.DeviceIdentifier), err
	}
	logrus.WithField("updatedDisk", updatedDisk).Debug("Updated disk")
	logrus.WithFields(logrus.Fields{
		"id":   updatedDisk.DeviceIdentifier,
		"size": humanize.Bytes(updatedDisk.Size),
	}).Info("Successfully loaded updated disk information")

	return fmt.Sprintf("grew container [%s] to size [%s]", disk.DeviceIdentifier, humanize.Bytes(updatedDisk.Size)), nil
}

// repairParentDisk attempts to find and repair the parent device for the given disk in order to update the current
// amount of free space available.
func repairParentDisk(disk *diskutil.DiskInfo, partitions *diskutil.SystemPartitions, utility diskutil.DiskUtil) (message string, err error) {
	// Get the device identifier for the parent disk
	logrus.Info("Searching for a parent device...")
	parentDiskID, err := parentDeviceID(disk)
	if err != nil {
		return fmt.Sprintf("failed to get the parent disk ID for container [%s]", disk.DeviceIdentifier), err
	}

	// Get the amount of available free space for the parent disk
	logrus.WithField("id", parentDiskID).Info("Checking for available space in parent disk...")
	availableSpace, err := availableDiskSpace(parentDiskID, partitions)
	if err != nil {
		return fmt.Sprintf("failed to get available space for disk ID [%s]", parentDiskID), err
	}
	logrus.WithField("free_space", humanize.Bytes(availableSpace)).Info("Successfully found remaining space")

	// Check if there's enough space to resize the container
	if availableSpace < MinimumGrowSpaceRequired {
		err = MinimumGrowSpaceError{
			message: fmt.Sprintf("at least [%s] of free space is required to grow the container", humanize.Bytes(MinimumGrowSpaceRequired)),
		}
		return "", err
	}

	// Attempt to repair the container's parent disk
	logrus.Info("Repairing the parent disk...")
	out, err := utility.RepairDisk(parentDiskID)
	if err != nil {
		return out, err
	}
	logrus.WithField("out", out).Debug("RepairDisk output")
	logrus.Info("Successfully repaired the parent disk")

	return out, nil
}

// validateDeviceID verifies if the provided ID is a valid device identifier or device node.
func validateDeviceID(id string, partitions *diskutil.SystemPartitions) (valid bool, err error) {
	// Check if ID is provided
	if len(id) == 0 {
		return false, fmt.Errorf("no ID provided")
	}

	// Check if the ID is a device node or device identifier
	if !strings.HasPrefix(id, "/dev/disk") && !strings.HasPrefix(id, "disk") {
		return false, fmt.Errorf("ID [%s] does not start with \"/dev/disk\" or \"disk\"", id)
	}

	// Get the device identifier
	diskIDRegex := regexp.MustCompile("disk[0-9]+")
	deviceID := diskIDRegex.FindString(id)
	if deviceID == "" {
		return false, fmt.Errorf("ID [%s] does not contain the expected expression \"disk[0-9]+\"", id)
	}

	// Check the device directory for the given identifier
	for _, name := range partitions.AllDisks {
		if strings.EqualFold(name, deviceID) {
			return true, nil
		}
	}

	return false, nil
}

// availableDiskSpace calculates the amount of unallocated disk space for a specific device id.
func availableDiskSpace(id string, partitions *diskutil.SystemPartitions) (size uint64, err error) {
	var diskPart *diskutil.DiskPart

	// Loop through all of the partitions in the system and attempt to find the struct with a matching ID
	for i, disk := range partitions.AllDisksAndPartitions {
		if strings.EqualFold(disk.DeviceIdentifier, id) {
			diskPart = &partitions.AllDisksAndPartitions[i]
			break
		}
	}

	// Ensure a DiskPart struct was found
	if diskPart == nil {
		return 0, fmt.Errorf("no partition information found for ID [%s]", id)
	}

	// diskPart.size is the disk's maximum size so it will be used as a starting point to subtract
	// the sizes of its individual partitions.
	size = diskPart.Size

	// Iterate through all of the disk's partitions and subtract their total size from the disk's maximum size.
	// At the end of this loop, size will be the amount of remaining free space the disk has.
	for _, part := range diskPart.Partitions {
		size -= part.Size
	}

	return size, nil
}

// rootContainer determines the ID for the container which is mounted as root.
func rootContainer(utility diskutil.DiskUtil) (container *diskutil.DiskInfo, err error) {
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
		diskIDRegex := regexp.MustCompile("disk[0-9]+")
		id := diskIDRegex.FindString(container.DeviceIdentifier)
		container.DeviceIdentifier = id
	}

	return container, nil
}

// parentDeviceID gets the parent's device identifier for a physical store
func parentDeviceID(disk *diskutil.DiskInfo) (id string, err error) {
	// Check if there are any Physical Stores the disk is a child of. APFS Containers and Volumes are virtualized
	// and should point back to some physical disk
	if disk.APFSPhysicalStores == nil {
		return "", fmt.Errorf("no physical stores found in disk")
	}

	// Check if there's more than one Physical Store in the disk's info. Having more than one APFS Physical Store
	// is unexpected and the common case shouldn't violate this.
	//
	// Note: more than one physical store can indicate a fusion drive - https://support.apple.com/en-us/HT202574.
	if len(disk.APFSPhysicalStores) == 1 {
		id = disk.APFSPhysicalStores[0].DeviceIdentifier
	} else {
		return "", fmt.Errorf("expected 1 physical store but got [%d]", len(disk.APFSPhysicalStores))
	}

	// Match the disk ID from the Physical Store's device identifier and remove extra partition information
	// from it (e.g. "s4s1")
	diskIDRegex := regexp.MustCompile("disk[0-9]+")
	id = diskIDRegex.FindString(id)
	if id == "" {
		return "", fmt.Errorf("physical store [%s] does not contain the expected expression \"disk[0-9]+\"", disk.APFSPhysicalStores[0].DeviceIdentifier)
	}

	return id, nil
}
