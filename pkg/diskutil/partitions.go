package diskutil

// SystemPartitions mirrors the output format of the command "diskutil list -plist" to store all disk
// and partition information.
type SystemPartitions struct {
	AllDisks              []string   `plist:"AllDisks"`
	AllDisksAndPartitions []DiskPart `plist:"AllDisksAndPartitions"`
	VolumesFromDisks      []string   `plist:"VolumesFromDisks"`
	WholeDisks            []string   `plist:"WholeDisks"`
}

// APFSPhysicalStore represents the physical device usually relating to synthesized virtual devices.
type APFSPhysicalStoreID struct {
	DeviceIdentifier string `plist:"DeviceIdentifier"`
}

// DiskPart represents a more condensed form of disk information and partitions as opposed to DiskInfo.
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
