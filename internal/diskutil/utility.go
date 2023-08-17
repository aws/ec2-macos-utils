package diskutil

import (
	"context"
	"fmt"

	"github.com/aws/ec2-macos-utils/internal/util"
)

// UtilImpl outlines the functionality necessary for wrapping macOS's diskutil tool. The methods are intentionally
// named to correspond to diskutil(8)'s subcommand names as its API.
type UtilImpl interface {
	// APFSImpl outlines the functionality necessary for wrapping diskutil's APFS verb.
	APFSImpl
	// Info fetches raw disk information for the specified device identifier.
	Info(ctx context.Context, id string) (string, error)
	// List fetches all disk and partition information for the system.
	// This output will be filtered based on the args provided.
	List(ctx context.Context, args []string) (string, error)
	// RepairDisk attempts to repair the disk for the specified device identifier.
	// This process requires root access.
	RepairDisk(ctx context.Context, id string) (string, error)
}

// APFSImpl outlines the functionality necessary for wrapping diskutil's APFS verb.
type APFSImpl interface {
	// ResizeContainer attempts to grow the APFS container with the given device identifier
	// to the specified size. If the given size is 0, ResizeContainer will attempt to grow
	// the disk to its maximum size.
	ResizeContainer(ctx context.Context, id string, size string) (string, error)
}

// DiskUtilityCmd is an empty struct that provides the implementation for the DiskUtility interface.
type DiskUtilityCmd struct{}

// List uses the macOS diskutil list command to list disks and partitions in a plist format by passing the -plist arg.
// List also appends any given args to fully support the diskutil list verb.
func (d *DiskUtilityCmd) List(ctx context.Context, args []string) (string, error) {
	// Create the diskutil command for retrieving all disk and partition information
	//   * -plist converts diskutil's output from human-readable to the plist format
	cmdListDisks := []string{"diskutil", "list", "-plist"}

	// Append arguments to the diskutil list verb
	if len(args) > 0 {
		cmdListDisks = append(cmdListDisks, args...)
	}

	// Execute the diskutil list command and store the output
	cmdOut, err := util.ExecuteCommand(ctx, cmdListDisks, "", nil, nil)
	if err != nil {
		return cmdOut.Stdout, fmt.Errorf("diskutil: failed to run diskutil command to list all disks, stderr: [%s]: %w", cmdOut.Stderr, err)
	}

	return cmdOut.Stdout, nil
}

// Info uses the macOS diskutil info command to get detailed information about a disk, partition, or container
// format by passing the -plist arg.
func (d *DiskUtilityCmd) Info(ctx context.Context, id string) (string, error) {
	// Create the diskutil command for retrieving disk information given a device identifier
	//   * -plist converts diskutil's output from human-readable to the plist format
	//   * id - the device identifier for the disk to be fetched
	cmdDiskInfo := []string{"diskutil", "info", "-plist", id}

	// Execute the diskutil info command and store the output
	cmdOut, err := util.ExecuteCommand(ctx, cmdDiskInfo, "", nil, nil)
	if err != nil {
		return cmdOut.Stdout, fmt.Errorf("diskutil: failed to run diskutil command to fetch disk information, stderr: [%s]: %w", cmdOut.Stderr, err)
	}

	return cmdOut.Stdout, nil
}

// RepairDisk uses the macOS diskutil diskRepair command to repair the specified volume and get updated information
// (e.g. amount of free space).
func (d *DiskUtilityCmd) RepairDisk(ctx context.Context, id string) (string, error) {
	// cmdRepairDisk represents the command used for executing macOS's diskutil to repair a disk.
	// The repairDisk command requires interactive-input ("yes"/"no") but is automated with util.ExecuteCommandYes.
	//   * repairDisk - indicates that a disk is going to be repaired (used to fetch amount of free space)
	//   * id - the device identifier for the disk to be repaired
	cmdRepairDisk := []string{"diskutil", "repairDisk", id}

	// Execute the diskutil repairDisk command and store the output
	cmdOut, err := util.ExecuteCommandYes(ctx, cmdRepairDisk, "", []string{})
	if err != nil {
		return cmdOut.Stdout, fmt.Errorf("diskutil: failed to run repairDisk command, stderr: [%s]: %w", cmdOut.Stderr, err)
	}

	return cmdOut.Stdout, nil
}

// ResizeContainer uses the macOS diskutil apfs resizeContainer command to change the size of the specific container ID.
func (d *DiskUtilityCmd) ResizeContainer(ctx context.Context, id string, size string) (string, error) {
	// cmdResizeContainer represents the command used for executing macOS's diskutil to resize a container
	//   * apfs - specifies that a virtual APFS volume is going to be modified
	//   * resizeContainer - indicates that a container is going to be resized
	//   * id - the device identifier for the container
	//   * size - the size which can be in a human-readable format (e.g. "0", "110g", and "1.5t")
	cmdResizeContainer := []string{"diskutil", "apfs", "resizeContainer", id, size}

	// Execute the diskutil apfs resizeContainer command and store the output
	cmdOut, err := util.ExecuteCommand(ctx, cmdResizeContainer, "", nil, nil)
	if err != nil {
		return cmdOut.Stdout, fmt.Errorf("diskutil: failed to run diskutil command to resize the container, stderr [%s]: %w", cmdOut.Stderr, err)
	}

	return cmdOut.Stdout, nil
}
