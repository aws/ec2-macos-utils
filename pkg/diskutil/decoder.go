package diskutil

import (
	"bytes"
	"fmt"

	"howett.net/plist"
)

// PlistDecoder outlines the functionality necessary for decoding plist output from macOS's diskutil.
type PlistDecoder interface {
	DecodeList(rawDiskList string) (partitions *SystemPartitions, err error)
	DecodeDisk(rawDisk string) (diskInfo *DiskInfo, err error)
	DecodeContainer(rawContainer string) (containerInfo *ContainerInfo, err error)
}

// Decoder is an empty struct that provides the implementation for the PlistDecoder interface.
type Decoder struct{}

// DecodeList takes a string containing the raw plist data for all disks and partition information
// and decodes it into a new SystemPartitions struct.
func (d *Decoder) DecodeList(rawList string) (partitions *SystemPartitions, err error) {
	// Catch panics thrown by the Decode method
	defer func() {
		if panicErr := recover(); panicErr != nil {
			partitions = nil
			err = fmt.Errorf("panic occured while decoding: %s", panicErr)
		}
	}()

	// Create a reader from the raw data and create a new decoder
	partitions = &SystemPartitions{}
	outputReader := bytes.NewReader([]byte(rawList))
	decoder := plist.NewDecoder(outputReader)

	// Decode the plist output from diskutil into a SystemPartitions struct for easier access
	err = decoder.Decode(partitions)
	if err != nil {
		return nil, fmt.Errorf("failed to decode diskutil list disks output: %v", err)
	}

	return partitions, nil
}

// DecodeDisk takes a string containing the raw plist data for disk information and decodes it into
// a new DiskInfo struct.
func (d *Decoder) DecodeDisk(rawDisk string) (disk *DiskInfo, err error) {
	// Catch panics thrown by the Decode method
	defer func() {
		if panicErr := recover(); panicErr != nil {
			disk = nil
			err = fmt.Errorf("panic occured while decoding: %s", panicErr)
		}
	}()

	// Create a reader from the raw data and create a new decoder
	disk = &DiskInfo{}
	outputReader := bytes.NewReader([]byte(rawDisk))
	decoder := plist.NewDecoder(outputReader)

	// Decode the plist output from diskutil into a DiskInfo struct for easier access
	err = decoder.Decode(disk)
	if err != nil {
		return nil, fmt.Errorf("failed to decode diskutil disk info output: %v", err)
	}

	return disk, nil
}

// DecodeContainer takes a string containing the raw plist data for container/file system information
// and decodes it into a new ContainerInfo struct.
func (d *Decoder) DecodeContainer(rawContainer string) (container *ContainerInfo, err error) {
	// Catch panics thrown by the Decode method
	defer func() {
		if panicErr := recover(); panicErr != nil {
			container = nil
			err = fmt.Errorf("panic occured while decoding: %s", panicErr)
		}
	}()

	// Create a reader from the raw data and create a new decoder
	container = &ContainerInfo{}
	outputReader := bytes.NewReader([]byte(rawContainer))
	decoder := plist.NewDecoder(outputReader)

	// Decode the plist output from diskutil into a DiskInfo struct for easier access
	err = decoder.Decode(container)
	if err != nil {
		return nil, fmt.Errorf("failed to decode diskutil container info output: %v", err)
	}

	return container, err
}

// defaultDecoder provides package functions for the PlistDecoder interface.
var defaultDecoder Decoder

// DecodeList takes a string containing the raw plist data for all disks and partition information
// and decodes it into a new SystemPartitions struct.
func DecodeList(rawList string) (partitions *SystemPartitions, err error) {
	return defaultDecoder.DecodeList(rawList)
}

// DecodeDisk takes a string containing the raw plist data for disk information and decodes it into
// a new DiskInfo struct.
func DecodeDisk(rawDisk string) (disk *DiskInfo, err error) {
	return defaultDecoder.DecodeDisk(rawDisk)
}

// DecodeContainer takes a string containing the raw plist data for container/file system information
// and decodes it into a new ContainerInfo struct.
func DecodeContainer(rawContainer string) (container *ContainerInfo, err error) {
	return defaultDecoder.DecodeContainer(rawContainer)
}
