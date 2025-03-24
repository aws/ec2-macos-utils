package diskutil

import (
	"context"
	"fmt"
	"regexp"

	"github.com/aws/ec2-macos-utils/internal/diskutil/types"
	"github.com/aws/ec2-macos-utils/internal/util"
)

// updatePhysicalStores provides separate functionality for fetching APFS physical stores for SystemPartitions.
func updatePhysicalStores(ctx context.Context, partitions *types.SystemPartitions) error {
	// Independently update all APFS disks' physical stores
	for i, part := range partitions.AllDisksAndPartitions {
		// Only do the update if the disk/partition is APFS
		if isAPFSVolume(part) {
			// Fetch the physical store for the disk/partition
			physicalStoreDeviceID, err := fetchPhysicalStore(ctx, part.DeviceIdentifier)
			if err != nil {
				return err
			}

			// Create a new physical store from the output
			physicalStoreElement := types.APFSPhysicalStoreID{
				DeviceIdentifier: physicalStoreDeviceID,
			}

			// Add the physical store to the DiskInfo
			partitions.AllDisksAndPartitions[i].APFSPhysicalStores = append(part.APFSPhysicalStores, physicalStoreElement)
		}
	}

	return nil
}

// isAPFSVolume checks if a given DiskPart is an APFS container.
func isAPFSVolume(part types.DiskPart) bool {
	return part.APFSVolumes != nil
}

// fetchPhysicalStore parses the human-readable output of the list verb for the given ID in order to fetch its
// physical store. This function is limited to returning only one physical store so the behavior might cause problems
// for fusion devices that have more than one APFS physical store.
func fetchPhysicalStore(ctx context.Context, id string) (string, error) {
	// Create the command for running diskutil and parsing the output to retrieve the desired info (physical store)
	//   * list - specifies the diskutil 'list' verb for a specific device ID and returns the human-readable output
	cmdPhysicalStore := []string{"diskutil", "list", id}

	// Execute the command to parse output from diskutil list
	out, err := util.ExecuteCommand(ctx, cmdPhysicalStore, "", nil, nil)
	if err != nil {
		return "", fmt.Errorf("%s: %w", out.Stderr, err)
	}

	return parsePhysicalStoreId(out.Stdout)
}

var (
	physicalStoreFieldTokenRegexp = regexp.MustCompile(`\s*Physical Store disk[0-9]+(s[0-9]+)*`)
	physicalStoreValueDiskIDRegexp = regexp.MustCompile("disk[0-9]+(s[0-9]+)*")
)

// parsePhysicalStoreId searches a raw string for the string "Physical Store disk[0-9]+(s[0-9]+)*". The regular
// expression "disk[0-9]+(s[0-9]+)*" matches any disk ID without the "/dev/" prefix.
func parsePhysicalStoreId(raw string) (string, error) {
	physicalStore := physicalStoreFieldTokenRegexp.FindString(raw)
	diskId := physicalStoreValueDiskIDRegexp.FindString(physicalStore)
	if diskId == "" {
		return "", fmt.Errorf("physical store not found")
	}

	return diskId, nil
}

// updatePhysicalStore provides separate functionality for fetching APFS physical stores for DiskInfo.
func updatePhysicalStore(ctx context.Context, disk *types.DiskInfo) error {
	if isAPFSMedia(disk) {
		physicalStoreDeviceID, err := fetchPhysicalStore(ctx, disk.DeviceIdentifier)
		if err != nil {
			return err
		}

		physicalStoreElement := types.APFSPhysicalStore{
			DeviceIdentifier: physicalStoreDeviceID,
		}

		disk.APFSPhysicalStores = append(disk.APFSPhysicalStores, physicalStoreElement)
	}

	return nil
}

// isAPFSMedia checks if the given DiskInfo is an APFS container or volume.
func isAPFSMedia(disk *types.DiskInfo) bool {
	return disk.FilesystemType == "apfs" || disk.IORegistryEntryName == "AppleAPFSMedia"
}
