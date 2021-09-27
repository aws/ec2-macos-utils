package diskutil

import (
	"fmt"

	"github.com/aws/ec2-macos-utils/pkg/diskutil/types"

	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"
)

// GrowContainer grows a container to its maximum size given an ID.
func GrowContainer(disk *types.DiskInfo, partitions *types.SystemPartitions, utility DiskUtil) (message string, err error) {
	// Attempt to repair the parent disk in order to get updated amount of free space
	logrus.Info("Attempting to repair the parent disk...")
	message, err = repairParentDisk(disk, partitions, utility)
	if err != nil {
		if _, ok := err.(MinimumGrowSpaceError); ok {
			return message, err
		}

		logrus.WithError(err).Warn("Failed to repair the parent disk, attempting to continue anyways...")
	}

	// Attempt to resize the container to its maximum size
	logrus.WithField("id", disk.DeviceIdentifier).Info("Resizing the container use full partition...")
	out, err := utility.ResizeContainer(disk.DeviceIdentifier, "0")
	logrus.WithField("out", out).Debug("Resize output")
	if err != nil {
		return out, err
	}
	logrus.Info("Successfully resized the container")

	// Get updated container size
	logrus.Info("Fetching updated container information...")
	updatedDisk, err := utility.Info(disk.DeviceIdentifier)
	if err != nil {
		return fmt.Sprintf("failed to fetch updated disk information for container [%s]", disk.DeviceIdentifier), err
	}
	logrus.WithField("updated_disk", updatedDisk).Debug("Updated disk")
	logrus.WithFields(logrus.Fields{
		"id":   updatedDisk.DeviceIdentifier,
		"size": humanize.Bytes(updatedDisk.Size),
	}).Info("Successfully loaded updated disk information")

	return fmt.Sprintf("grew container [%s] to size [%s]", disk.DeviceIdentifier, humanize.Bytes(updatedDisk.Size)), nil
}

// repairParentDisk attempts to find and repair the parent device for the given disk in order to update the current
// amount of free space available.
func repairParentDisk(disk *types.DiskInfo, partitions *types.SystemPartitions, utility DiskUtil) (message string, err error) {
	// Get the device identifier for the parent disk
	logrus.Info("Searching for a parent device...")
	parentDiskID, err := disk.ParentDeviceID()
	if err != nil {
		return fmt.Sprintf("failed to get the parent disk ID for container [%s]", disk.DeviceIdentifier), err
	}

	// Get the amount of available free space for the parent disk
	logrus.WithField("id", parentDiskID).Info("Checking for available space in parent disk...")
	availableSpace, err := partitions.AvailableDiskSpace(parentDiskID)
	if err != nil {
		return fmt.Sprintf("failed to get available space for disk ID [%s]", parentDiskID), err
	}
	logrus.WithField("free_space", humanize.Bytes(availableSpace)).Info("Successfully found remaining space")

	// Check if there's enough space to resize the container
	if availableSpace < MinimumGrowSpaceRequired {
		return "", MinimumGrowSpaceError{size: availableSpace}
	}

	// Attempt to repair the container's parent disk
	logrus.Info("Repairing the parent disk...")
	out, err := utility.RepairDisk(parentDiskID)
	logrus.WithField("out", out).Debug("RepairDisk output")
	if err != nil {
		return out, err
	}
	logrus.Info("Successfully repaired the parent disk")

	return out, nil
}
