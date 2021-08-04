// Package diskutil provides the functionality necessary for interacting with macOS's diskutil CLI.
package diskutil

//go:generate mockgen -source=diskutil.go -destination=mocks/mock_diskutil.go

import (
	"fmt"
	"regexp"

	"github.com/aws/ec2-macos-utils/pkg/system"
	"github.com/aws/ec2-macos-utils/pkg/util"
)

// DiskUtil outlines the functionality necessary for wrapping macOS's diskutil tool.
type DiskUtil interface {
	List(args []string) (*SystemPartitions, error)
	Info(id string) (*DiskInfo, error)
	RepairDisk(id string) (out string, err error)
	APFS
}

// APFS outlines the functionality necessary for wrapping diskutil's APFS verb.
type APFS interface {
	ResizeContainer(id, size string) (out string, err error)
}

// NewDiskUtil creates a new diskutil controller for the given product type.
func NewDiskUtil(product system.Product) (DiskUtil, error) {
	switch product.(type) {
	case system.ProductMojave, *system.ProductMojave:
		utility := &DiskUtilityMojave{
			Util:    &DiskUtilityCmd{},
			Decoder: &PlistDecoder{},
		}
		return utility, nil
	case system.ProductCatalina, *system.ProductCatalina:
		utility := &DiskUtility{
			Util:    &DiskUtilityCmd{},
			Decoder: &PlistDecoder{},
		}
		return utility, nil
	case system.ProductBigSur, *system.ProductBigSur:
		utility := &DiskUtility{
			Util:    &DiskUtilityCmd{},
			Decoder: &PlistDecoder{},
		}
		return utility, nil
	default:
		return nil, fmt.Errorf("unknown product type")
	}
}

// DiskUtility wraps all of the functionality necessary for interacting with macOS's diskutil in GoLang.
type DiskUtility struct {
	Util    UtilImpl
	Decoder Decoder
}

// List utilizes the DiskUtil.List method to fetch the raw list output from diskutil and returns the decoded
// output in a SystemPartitions struct.
func (d *DiskUtility) List(args []string) (*SystemPartitions, error) {
	rawPartitions, err := d.Util.List(args)
	if err != nil {
		return nil, err
	}

	partitions, err := d.Decoder.DecodeList(rawPartitions)
	if err != nil {
		return nil, err
	}

	return partitions, nil
}

// Info utilizes the DiskUtil.Info method to fetch the raw disk output from diskutil and returns the decoded
// output in a DiskInfo struct.
func (d *DiskUtility) Info(id string) (*DiskInfo, error) {
	rawDisk, err := d.Util.Info(id)
	if err != nil {
		return nil, err
	}

	disk, err := d.Decoder.DecodeInfo(rawDisk)
	if err != nil {
		return nil, err
	}

	return disk, nil
}

// RepairDisk wraps DiskUtil.RepairDisk.
func (d *DiskUtility) RepairDisk(id string) (out string, err error) {
	return d.Util.RepairDisk(id)
}

// ResizeContainer wraps APFS.ResizeContainer.
func (d *DiskUtility) ResizeContainer(id, size string) (out string, err error) {
	return d.Util.ResizeContainer(id, size)
}

// DiskUtilityMojave wraps all of the functionality necessary for interacting with macOS's diskutil on Mojave.
type DiskUtilityMojave struct {
	Util    UtilImpl
	Decoder Decoder
}

// List utilizes the DiskUtil.List method to fetch the raw list output from diskutil and returns the decoded
// output in a SystemPartitions struct. List also attempts to update each APFS Volume's physical store via a separate
// fetch method since the version of diskutil on Mojave doesn't provide that information in its List verb.
//
// It is possible for List to fail when updating the physical stores but it will still return the original data
// that was decoded into the SystemPartitions struct.
func (d *DiskUtilityMojave) List(args []string) (*SystemPartitions, error) {
	rawPartitions, err := d.Util.List(args)
	if err != nil {
		return nil, err
	}

	partitions, err := d.Decoder.DecodeList(rawPartitions)
	if err != nil {
		return nil, err
	}

	err = updatePhysicalStores(partitions)
	if err != nil {
		return partitions, err
	}

	return partitions, nil
}

// updatePhysicalStores provides separate functionality for fetching APFS physical stores for SystemPartitions.
func updatePhysicalStores(partitions *SystemPartitions) error {
	for i, part := range partitions.AllDisksAndPartitions {
		if isAPFSVolume(part) {
			physicalStoreId, err := fetchPhysicalStore(part.DeviceIdentifier)
			if err != nil {
				return err
			}

			physicalStore := APFSPhysicalStoreID{physicalStoreId}

			partitions.AllDisksAndPartitions[i].APFSPhysicalStores = append(part.APFSPhysicalStores, physicalStore)
		}
	}

	return nil
}

// isAPFSVolume checks if a given DiskPart is an APFS container.
func isAPFSVolume(part DiskPart) bool {
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
	out, err := util.ExecuteCommand(cmdPhysicalStore, "", nil)
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

// Info utilizes the DiskUtil.Info method to fetch the raw disk output from diskutil and returns the decoded
// output in a DiskInfo struct. Info also attempts to update each APFS Volume's physical store via a separate
// fetch method since the version of diskutil on Mojave doesn't provide that information in its Info verb.
//
// It is possible for Info to fail when updating the physical stores but it will still return the original data
// that was decoded into the DiskInfo struct.
func (d *DiskUtilityMojave) Info(id string) (*DiskInfo, error) {
	rawDisk, err := d.Util.Info(id)
	if err != nil {
		return nil, err
	}

	disk, err := d.Decoder.DecodeInfo(rawDisk)
	if err != nil {
		return nil, err
	}

	err = updatePhysicalStore(disk)
	if err != nil {
		return disk, err
	}

	return disk, nil
}

// updatePhysicalStore provides separate functionality for fetching APFS physical stores for DiskInfo.
func updatePhysicalStore(disk *DiskInfo) error {
	if isAPFSMedia(disk) {
		physicalStoreId, err := fetchPhysicalStore(disk.DeviceIdentifier)
		if err != nil {
			return err
		}

		physicalStore := APFSPhysicalStore{physicalStoreId}

		disk.APFSPhysicalStores = append(disk.APFSPhysicalStores, physicalStore)
	}

	return nil
}

// isAPFSMedia checks if the given DiskInfo is an APFS container or volume.
func isAPFSMedia(disk *DiskInfo) bool {
	return disk.FilesystemType == "apfs" || disk.IORegistryEntryName == "AppleAPFSMedia"
}

// RepairDisk wraps DiskUtil.RepairDisk.
func (d *DiskUtilityMojave) RepairDisk(id string) (out string, err error) {
	return d.Util.RepairDisk(id)
}

// ResizeContainer wraps APFS.ResizeContainer.
func (d *DiskUtilityMojave) ResizeContainer(id, size string) (out string, err error) {
	return d.Util.ResizeContainer(id, size)
}
