package diskutil

import (
	"fmt"
	"regexp"

	"github.com/aws/ec2-macos-utils/internal/diskutil/types"
	"github.com/aws/ec2-macos-utils/internal/util"
)

// updatePhysicalStores provides separate functionality for fetching APFS physical stores for SystemPartitions.
func updatePhysicalStores(partitions *types.SystemPartitions) error {
	// Independently update all APFS disks' physical stores
	for i, part := range partitions.AllDisksAndPartitions {
		// Only do the update if the disk/partition is APFS
		if isAPFSVolume(part) {
			// Fetch the physical store for the disk/partition
			physicalStoreId, err := fetchPhysicalStore(part.DeviceIdentifier)
			if err != nil {
				return err
			}

			// Create a new physical store from the output
			physicalStore := types.APFSPhysicalStoreID{physicalStoreId}

			// Add the physical store to the DiskInfo
			partitions.AllDisksAndPartitions[i].APFSPhysicalStores = append(part.APFSPhysicalStores, physicalStore)
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
func fetchPhysicalStore(id string) (string, error) {
	// Create the command for running diskutil and parsing the output to retrieve the desired info (physical store)
	//   * list - specifies the diskutil 'list' verb for a specific device ID and returns the human-readable output
	cmdPhysicalStore := []string{"diskutil", "list", id}

	// Execute the command to parse output from diskutil list
	out, err := util.ExecuteCommand(cmdPhysicalStore, "", nil, nil)
	if err != nil {
		return "", fmt.Errorf("%s: %w", out.Stderr, err)
	}

	return parsePhysicalStoreId(out.Stdout)
}

// parsePhysicalStoreId searches a raw string for the string "Physical Store disk[0-9]+(s[0-9]+)*". The regular
// expression "disk[0-9]+(s[0-9]+)*" matches any disk ID without the "/dev/" prefix.
func parsePhysicalStoreId(raw string) (string, error) {
	physicalStoreExp := regexp.MustCompile("\\s*Physical Store disk[0-9]+(s[0-9]+)*")
	diskIdExp := regexp.MustCompile("disk[0-9]+(s[0-9]+)*")

	physicalStore := physicalStoreExp.FindString(raw)
	diskId := diskIdExp.FindString(physicalStore)
	if diskId == "" {
		return "", fmt.Errorf("physical store not found")
	}

	return diskId, nil
}

// updatePhysicalStore provides separate functionality for fetching APFS physical stores for DiskInfo.
func updatePhysicalStore(disk *types.DiskInfo) error {
	if isAPFSMedia(disk) {
		physicalStoreId, err := fetchPhysicalStore(disk.DeviceIdentifier)
		if err != nil {
			return err
		}

		physicalStore := types.APFSPhysicalStore{physicalStoreId}

		disk.APFSPhysicalStores = append(disk.APFSPhysicalStores, physicalStore)
	}

	return nil
}

// isAPFSMedia checks if the given DiskInfo is an APFS container or volume.
func isAPFSMedia(disk *types.DiskInfo) bool {
	return disk.FilesystemType == "apfs" || disk.IORegistryEntryName == "AppleAPFSMedia"
}
