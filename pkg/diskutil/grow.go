package diskutil

import (
	"errors"
	"fmt"

	"github.com/aws/ec2-macos-utils/pkg/diskutil/types"

	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"
)

// GrowContainer grows a container to its maximum size by performing the following operations:
// 		1. Verify that the given types.DiskInfo is an APFS container that can be resized.
//		2. Fetch the types.DiskInfo for the underlying physical disk (if the container isn't a physical device).
//		3. Repair the parent disk to force the kernel to get the latest GPT information for the disk.
//		4. Check if there's enough free space on the disk to perform an APFS.ResizeContainer.
//		5. Resize the container to its maximum size.
func GrowContainer(u DiskUtil, container *types.DiskInfo) error {
	if container == nil {
		return fmt.Errorf("unable to resize nil container")
	}

	logrus.WithField("device_id", container.DeviceIdentifier).Info("Checking if device can be APFS resized...")
	if err := canAPFSResize(container); err != nil {
		return fmt.Errorf("unable to resize container: %w", err)
	}
	logrus.Info("Device can be resized")

	// We'll need to mutate the container's underlying physical disk, so resolve that if that's not what we have
	// (which is basically guaranteed to not have physical disk for container resizes, should be the virtual APFS
	// container).
	phy := container
	if !phy.IsPhysical() {
		parent, err := u.Info(phy.ParentWholeDisk)
		if err != nil {
			return fmt.Errorf("unable to determine physical disk: %w", err)
		}
		// using the parent disk of provided disk (probably a container)
		phy = parent
	}

	// Capture any free space on a resized disk
	logrus.Info("Repairing the parent disk...")
	_, err := repairParentDisk(u, phy)
	if err != nil {
		return fmt.Errorf("cannot update free space on disk: %w", err)
	}
	logrus.Info("Successfully repaired the parent disk")

	// Minimum free space to resize required - bail if we don't have enough.
	logrus.WithField("device_id", phy.DeviceIdentifier).Info("Fetching amount of free space on device...")
	totalFree, err := getDiskFreeSpace(u, phy)
	if err != nil {
		return fmt.Errorf("cannot determine available space on disk: %w", err)
	}
	logrus.WithField("freed_bytes", humanize.Bytes(totalFree)).Trace("updated free space on disk")
	if totalFree < minimumGrowFreeSpace {
		logrus.WithFields(logrus.Fields{
			"total_free":       humanize.Bytes(totalFree),
			"required_minimum": humanize.Bytes(minimumGrowFreeSpace),
		}).Warn("Available free space does not meet required minimum to grow")
		return fmt.Errorf("not enough space to resize container: %w", FreeSpaceError{totalFree})
	}

	logrus.WithFields(logrus.Fields{
		"device_id":  phy.DeviceIdentifier,
		"free_space": humanize.Bytes(totalFree),
	}).Info("Resizing container to maximum size...")
	out, err := u.ResizeContainer(phy.DeviceIdentifier, "0")
	logrus.WithField("out", out).Debug("Resize output")

	return err
}

// canAPFSResize does some basic checking on a types.DiskInfo to see if it matches the criteria necessary for
// APFS.ResizeContainer to succeed. It checks that the types.ContainerInfo is not empty and that the
// types.ContainerInfo's FilesystemType is "apfs".
func canAPFSResize(container *types.DiskInfo) error {
	if container == nil {
		return errors.New("no disk information")
	}

	if (container.ContainerInfo == types.ContainerInfo{}) {
		return errors.New("no container information")
	}

	if container.FilesystemType != "apfs" {
		return errors.New("disk is not apfs")
	}

	return nil
}

// getDiskFreeSpace calculates the amount of free space a disk has available by summing the sizes of each partition
// and then subtracting that from the total size. See types.SystemPartitions for more information.
func getDiskFreeSpace(util DiskUtil, disk *types.DiskInfo) (uint64, error) {
	partitions, err := util.List(nil)
	if err != nil {
		return 0, err
	}

	parentDiskID, err := disk.ParentDeviceID()
	if err != nil {
		return 0, err
	}

	return partitions.AvailableDiskSpace(parentDiskID)
}

// repairParentDisk attempts to find and repair the parent device for the given disk in order to update the current
// amount of free space available.
func repairParentDisk(utility DiskUtil, disk *types.DiskInfo) (message string, err error) {
	// Get the device identifier for the parent disk
	parentDiskID, err := disk.ParentDeviceID()
	if err != nil {
		return fmt.Sprintf("failed to get the parent disk ID for container [%s]", disk.DeviceIdentifier), err
	}

	// Attempt to repair the container's parent disk
	logrus.WithField("parent_id", parentDiskID).Info("Found parent disk ID")
	out, err := utility.RepairDisk(parentDiskID)
	logrus.WithField("out", out).Debug("RepairDisk output")
	if err != nil {
		return out, err
	}

	return out, nil
}
