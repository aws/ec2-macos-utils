package diskutil

import (
	"fmt"
	"io"

	"howett.net/plist"
)

// Decoder outlines the functionality necessary for decoding plist output from the macOS diskutil command.
type Decoder interface {
	DecodeList(reader io.ReadSeeker) (*SystemPartitions, error)
	DecodeInfo(reader io.ReadSeeker) (*DiskInfo, error)
}

// PlistDecoder is an empty struct that provides the implementation for the Decoder interface.
type PlistDecoder struct{}

// DecodeList takes a string containing the raw plist data for all disks and partition information
// and decodes it into a new SystemPartitions struct.
func (d *PlistDecoder) DecodeList(reader io.ReadSeeker) (partitions *SystemPartitions, err error) {
	// Catch panics thrown by the Decode method
	defer func() {
		if panicErr := recover(); panicErr != nil {
			partitions = nil
			err = fmt.Errorf("diskutil: panic occured while decoding: %s", panicErr)
		}
	}()

	// Create a reader from the raw data and create a new plist Decoder
	partitions = &SystemPartitions{}
	decoder := plist.NewDecoder(reader)

	// Decode the plist output from diskutil into a SystemPartitions struct for easier access
	err = decoder.Decode(partitions)
	if err != nil {
		return nil, fmt.Errorf("diskutil: failed to decode diskutil list disks output: %v", err)
	}

	return partitions, nil
}

// DecodeInfo takes a string containing the raw plist data for disk information and decodes it into
// a new DiskInfo struct.
func (d *PlistDecoder) DecodeInfo(reader io.ReadSeeker) (disk *DiskInfo, err error) {
	// Catch panics thrown by the Decode method
	defer func() {
		if panicErr := recover(); panicErr != nil {
			disk = nil
			err = fmt.Errorf("diskutil: panic occured while decoding: %s", panicErr)
		}
	}()

	// Create a reader from the raw data and create a new PlistDecoder
	disk = &DiskInfo{}
	decoder := plist.NewDecoder(reader)

	// Decode the plist output from diskutil into a DiskInfo struct for easier access
	err = decoder.Decode(disk)
	if err != nil {
		return nil, fmt.Errorf("diskutil: failed to decode diskutil disk info output: %v", err)
	}

	return disk, nil
}
