package diskutil

// DiskUtil wraps all of the functionality necessary for interacting with macOS's diskutil in GoLang.
type DiskUtil struct {
	Utility DiskUtility
	Decoder PlistDecoder
}

// List utilizes the DiskUtility.List method to fetch the raw list output from diskutil and returns the decoded
// output in a SystemPartitions struct.
func (d *DiskUtil) List(args []string) (partitions *SystemPartitions, err error) {
	rawPartitions, err := d.Utility.List(args)
	if err != nil {
		return nil, err
	}

	partitions, err = d.Decoder.DecodeList(rawPartitions)
	if err != nil {
		return nil, err
	}

	return partitions, nil
}

// Info wraps DiskUtilityCmd.Info.
func (d *DiskUtil) Info(id string) (out string, err error) {
	return d.Utility.Info(id)
}

// InfoDisk utilizes the DiskUtility.Info method to fetch the raw disk output from diskutil and returns the decoded
// output in a DiskInfo struct.
func (d *DiskUtil) InfoDisk(id string) (disk *DiskInfo, err error) {
	rawDisk, err := d.Utility.Info(id)
	if err != nil {
		return nil, err
	}

	disk, err = d.Decoder.DecodeDisk(rawDisk)
	if err != nil {
		return nil, err
	}

	return disk, nil
}

// InfoContainer utilizes the DiskUtility.Info method to fetch the raw container output from diskutil and returns the
// decoded output in a ContainerInfo struct.
func (d *DiskUtil) InfoContainer(id string) (container *ContainerInfo, err error) {
	rawContainer, err := d.Utility.Info(id)
	if err != nil {
		return nil, err
	}

	container, err = d.Decoder.DecodeContainer(rawContainer)
	if err != nil {
		return nil, err
	}

	return container, nil
}

// RepairDisk wraps DiskUtility.RepairDisk.
func (d *DiskUtil) RepairDisk(id string) (out string, err error) {
	return d.Utility.RepairDisk(id)
}

// ResizeContainer wraps APFS.ResizeContainer.
func (d *DiskUtil) ResizeContainer(id, size string) (out string, err error) {
	return d.Utility.ResizeContainer(id, size)
}
