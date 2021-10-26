package diskutil

import (
	"fmt"
	"io"

	"github.com/aws/ec2-macos-utils/pkg/diskutil/types"

	"howett.net/plist"
)

// Decoder outlines the functionality necessary for decoding plist output from the macOS diskutil command.
type Decoder interface {
	// DecodeSystemPartitions takes an io.ReadSeeker for the raw plist data of all disks and partition information
	// and decodes it into a new types.SystemPartitions struct.
	DecodeSystemPartitions(reader io.ReadSeeker) (*types.SystemPartitions, error)

	// DecodeDiskInfo takes an io.ReadSeeker for the raw plist data of disk information and decodes it into
	// a new types.DiskInfo struct.
	DecodeDiskInfo(reader io.ReadSeeker) (*types.DiskInfo, error)
}

// PlistDecoder provides the plist Decoder implementation.
type PlistDecoder struct{}

// DecodeSystemPartitions assumes the io.ReadSeeker it's given contains raw plist data and attempts to decode that.
func (d *PlistDecoder) DecodeSystemPartitions(reader io.ReadSeeker) (*types.SystemPartitions, error) {
	// Set up a new SystemPartitions and create a decoder from the reader
	partitions := &types.SystemPartitions{}
	decoder := plist.NewDecoder(reader)

	// Decode the plist output from diskutil into a SystemPartitions struct for easier access
	err := decoder.Decode(partitions)
	if err != nil {
		return nil, fmt.Errorf("error decoding list: %w", err)
	}

	return partitions, nil
}

// DecodeDiskInfo assumes the io.ReadSeeker it's given contains raw plist data and attempts to decode that.
func (d *PlistDecoder) DecodeDiskInfo(reader io.ReadSeeker) (*types.DiskInfo, error) {
	// Set up a new DiskInfo and create a decoder from the reader
	disk := &types.DiskInfo{}
	decoder := plist.NewDecoder(reader)

	// Decode the plist output from diskutil into a DiskInfo struct for easier access
	err := decoder.Decode(disk)
	if err != nil {
		return nil, fmt.Errorf("error decoding disk info: %w", err)
	}

	return disk, nil
}
