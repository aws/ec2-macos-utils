package types

import (
	"fmt"
	"strings"
)

// SystemPartitions mirrors the output format of the command "diskutil list -plist" to store all disk
// and partition information.
type SystemPartitions struct {
	AllDisks              []string   `plist:"AllDisks"`
	AllDisksAndPartitions []DiskPart `plist:"AllDisksAndPartitions"`
	VolumesFromDisks      []string   `plist:"VolumesFromDisks"`
	WholeDisks            []string   `plist:"WholeDisks"`
}

// AvailableDiskSpace calculates the amount of unallocated disk space for a specific device id.
func (p *SystemPartitions) AvailableDiskSpace(id string) (size uint64, err error) {
	var diskPart *DiskPart

	// Loop through all the partitions in the system and attempt to find the struct with a matching ID
	for i, disk := range p.AllDisksAndPartitions {
		if strings.EqualFold(disk.DeviceIdentifier, id) {
			diskPart = &p.AllDisksAndPartitions[i]
			break
		}
	}

	// Ensure a DiskPart struct was found
	if diskPart == nil {
		return 0, fmt.Errorf("no partition information found for ID [%s]", id)
	}

	// diskPart.size is the disk's maximum size so it will be used as a starting point to subtract
	// the sizes of its individual p.
	size = diskPart.Size

	// Iterate through all of the disk's partitions and subtract their total size from the disk's maximum size.
	// At the end of this loop, size will be the amount of remaining free space the disk has.
	for _, part := range diskPart.Partitions {
		size -= part.Size
	}

	return size, nil
}

// APFSPhysicalStoreID represents the physical device usually relating to synthesized virtual devices.
type APFSPhysicalStoreID struct {
	DeviceIdentifier string `plist:"DeviceIdentifier"`
}

// DiskPart represents a subset of information from DiskInfo.
type DiskPart struct {
	APFSPhysicalStores []APFSPhysicalStoreID `plist:"APFSPhysicalStores"`
	APFSVolumes        []APFSVolume          `plist:"APFSVolumes"`
	Content            string                `plist:"Content"`
	DeviceIdentifier   string                `plist:"DeviceIdentifier"`
	OSInternal         bool                  `plist:"OSInternal"`
	Partitions         []Partition           `plist:"Partitions"`
	Size               uint64                `plist:"Size"`
}

// Partition stores relevant information about a partition in macOS.
type Partition struct {
	Content          string `plist:"Content"`
	DeviceIdentifier string `plist:"DeviceIdentifier"`
	DiskUUID         string `plist:"DiskUUID"`
	Size             uint64 `plist:"Size"`
	VolumeName       string `plist:"VolumeName"`
	VolumeUUID       string `plist:"VolumeUUID"`
}

// APFSVolume represents a macOS APFS Volume with relevant information.
type APFSVolume struct {
	DeviceIdentifier string     `plist:"DeviceIdentifier"`
	DiskUUID         string     `plist:"DiskUUID"`
	MountPoint       string     `plist:"MountPoint"`
	MountedSnapshots []Snapshot `plist:"MountedSnapshots"`
	OSInternal       bool       `plist:"OSInternal"`
	Size             uint64     `plist:"Size"`
	VolumeName       string     `plist:"VolumeName"`
	VolumeUUID       string     `plist:"VolumeUUID"`
}

// Snapshot stores relevant information about a snapshot in macOS.
type Snapshot struct {
	Sealed             string `plist:"Sealed"`
	SnapshotBSD        string `plist:"SnapshotBSD"`
	SnapshotMountPoint string `plist:"SnapshotMountPoint"`
	SnapshotName       string `plist:"SnapshotName"`
	SnapshotUUID       string `plist:"SnapshotUUID"`
}
