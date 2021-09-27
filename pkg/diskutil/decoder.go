package diskutil

import (
	"fmt"
	"io"

	"github.com/aws/ec2-macos-utils/pkg/diskutil/types"

	"howett.net/plist"
)

// Decoder outlines the functionality necessary for decoding plist output from the macOS diskutil command.
type Decoder interface {
	DecodeSystemPartitions(reader io.ReadSeeker) (*types.SystemPartitions, error)
	DecodeDiskInfo(reader io.ReadSeeker) (*types.DiskInfo, error)
}

// PlistDecoder is an empty struct that provides the implementation for the Decoder interface.
type PlistDecoder struct{}

// DecodeSystemPartitions takes an io.ReadSeeker for the raw plist data of all disks and partition information
// and decodes it into a new SystemPartitions struct.
func (d *PlistDecoder) DecodeSystemPartitions(reader io.ReadSeeker) (partitions *types.SystemPartitions, err error) {
	// Catch panics thrown by the Decode method
	defer func() {
		if panicErr := recover(); panicErr != nil {
			partitions = nil
			err = fmt.Errorf("decoder panicked: %s", panicErr)
		}
	}()

	// Set up a new SystemPartitions and create a decoder from the reader
	partitions = &types.SystemPartitions{}
	decoder := plist.NewDecoder(reader)

	// Decode the plist output from diskutil into a SystemPartitions struct for easier access
	err = decoder.Decode(partitions)
	if err != nil {
		return nil, fmt.Errorf("error decoding list: %v", err)
	}

	return partitions, nil
}

// DecodeDiskInfo takes an io.ReadSeeker for the raw plist data of disk information and decodes it into
// a new DiskInfo struct.
func (d *PlistDecoder) DecodeDiskInfo(reader io.ReadSeeker) (disk *types.DiskInfo, err error) {
	// Catch panics thrown by the Decode method
	defer func() {
		if panicErr := recover(); panicErr != nil {
			disk = nil
			err = fmt.Errorf("decoder panicked: %s", panicErr)
		}
	}()

	// Set up a new DiskInfo and create a decoder from the reader
	disk = &types.DiskInfo{}
	decoder := plist.NewDecoder(reader)

	// Decode the plist output from diskutil into a DiskInfo struct for easier access
	err = decoder.Decode(disk)
	if err != nil {
		return nil, fmt.Errorf("error decoding disk info: %v", err)
	}

	return disk, nil
}
