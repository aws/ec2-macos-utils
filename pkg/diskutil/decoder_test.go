package diskutil

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/aws/ec2-macos-utils/pkg/diskutil/types"

	"github.com/stretchr/testify/assert"
)

var (
	//go:embed testdata/decoder/broken_disk_info.plist
	// decoderBrokenDiskInfo contains a disk plist file that is missing the plist header.
	decoderBrokenDiskInfo string

	//go:embed testdata/decoder/disk_info.plist
	// decoderDiskInfo contains a disk plist file that is properly formatted (but is also sparse).
	decoderDiskInfo string

	//go:embed testdata/decoder/broken_container_info.plist
	// decoderBrokenContainerInfo contains a container plist file that is missing the plist header.
	decoderBrokenContainerInfo string

	//go:embed testdata/decoder/container_info.plist
	// decoderContainerInfo contains a container plist file that is properly formatted (but is also sparse).
	decoderContainerInfo string

	//go:embed testdata/decoder/broken_list.plist
	// decoderBrokenList contains a container plist file that is missing the plist header.
	decoderBrokenList string

	//go:embed testdata/decoder/list.plist
	// decoderList contains a container plist file that is properly formatted (but is also sparse).
	decoderList string
)

func TestPlistDecoder_DecodeDiskInfo_WithoutInput(t *testing.T) {
	d := &PlistDecoder{}
	reader := strings.NewReader("")

	expectedDisk := &types.DiskInfo{}

	actualDisk, err := d.DecodeDiskInfo(reader)

	assert.NoError(t, err, "should be able to decode empty input")
	assert.ObjectsAreEqualValues(expectedDisk, actualDisk)
}

func TestPlistDecoder_DecodeDiskInfo_WithoutPlistInput(t *testing.T) {
	d := &PlistDecoder{}
	reader := strings.NewReader("this is not a plist")

	actualDisk, err := d.DecodeDiskInfo(reader)

	assert.Error(t, err, "shouldn't be able to decode non-plist input")
	assert.Nil(t, actualDisk, "should get nil since decode failed")
}

func TestPlistDecoder_DecodeDiskInfo_WithBrokenDiskInfoPlist(t *testing.T) {
	d := &PlistDecoder{}
	reader := strings.NewReader(decoderBrokenDiskInfo)

	actualDisk, err := d.DecodeDiskInfo(reader)

	assert.Error(t, err, "shouldn't be able to decode broken plist data")
	assert.Nil(t, actualDisk, "should get nil since decode failed")
}

func TestPlistDecoder_DecodeDiskInfo_DiskSuccess(t *testing.T) {
	const (
		testDiskID              = "disk2"
		testPhysicalStoreID     = "disk0s2"
		availableSpare      int = 100
	)

	d := &PlistDecoder{}
	reader := strings.NewReader(decoderDiskInfo)

	expectedDisk := &types.DiskInfo{
		AESHardware:            false,
		APFSContainerReference: testDiskID,
		APFSPhysicalStores:     []types.APFSPhysicalStore{{DeviceIdentifier: testPhysicalStoreID}},
		SMARTDeviceSpecificKeysMayVaryNotGuaranteed: &types.SmartDeviceInfo{AvailableSpare: availableSpare},
	}

	actualDisk, err := d.DecodeDiskInfo(reader)

	assert.NoError(t, err, "should be able to decode valid disk plist data")
	assert.ObjectsAreEqualValues(expectedDisk, actualDisk)
}

func TestPlistDecoder_DecodeDiskInfo_WithImproperContainerPlistInput(t *testing.T) {
	d := &PlistDecoder{}
	reader := strings.NewReader(decoderBrokenContainerInfo)

	actualDisk, err := d.DecodeDiskInfo(reader)

	assert.Error(t, err, "shouldn't be able to decode broken plist data")
	assert.Nil(t, actualDisk, "should get nil since decode failed")
}

func TestPlistDecoder_DecodeDiskInfo_ContainerSuccess(t *testing.T) {
	const (
		testDiskID                 = "disk2"
		testPhysicalStoreID        = "disk0s2"
		containerSize       uint64 = 6_000_000
		freeSize            uint64 = 4_000_000
	)

	d := &PlistDecoder{}
	reader := strings.NewReader(decoderContainerInfo)

	expectedDisk := &types.DiskInfo{
		ContainerInfo: types.ContainerInfo{
			APFSContainerFree: freeSize,
			APFSContainerSize: containerSize,
		},
		AESHardware:            false,
		APFSContainerReference: testDiskID,
		APFSPhysicalStores:     []types.APFSPhysicalStore{{DeviceIdentifier: testPhysicalStoreID}},
	}

	actualDisk, err := d.DecodeDiskInfo(reader)

	assert.NoError(t, err, "should be able to decode valid container plist data")
	assert.ObjectsAreEqualValues(expectedDisk, actualDisk)
}

func TestPlistDecoder_DecodeSystemPartitions_WithoutInput(t *testing.T) {
	d := &PlistDecoder{}
	reader := strings.NewReader("")

	wantParts := &types.SystemPartitions{}

	gotParts, err := d.DecodeSystemPartitions(reader)

	assert.NoError(t, err, "should be able to decode empty input")
	assert.ObjectsAreEqualValues(wantParts, gotParts)
}

func TestPlistDecoder_DecodeSystemPartitions_WithoutPlistInput(t *testing.T) {
	d := &PlistDecoder{}
	reader := strings.NewReader("this is not a plist")

	gotParts, err := d.DecodeSystemPartitions(reader)

	assert.Error(t, err, "shouldn't be able to decode non-plist input")
	assert.Nil(t, gotParts, "should get nil since decode failed")
}

func TestPlistDecoder_DecodeSystemPartitions_WithBrokenDiskInfoPlist(t *testing.T) {
	d := &PlistDecoder{}
	reader := strings.NewReader(decoderBrokenList)

	gotParts, err := d.DecodeSystemPartitions(reader)

	assert.Error(t, err, "shouldn't be able to decode broken plist data")
	assert.Nil(t, gotParts, "should get nil since decode failed")
}

func TestPlistDecoder_DecodeSystemPartitions_Success(t *testing.T) {
	const (
		testDiskID                 = "disk0"
		testPartID                 = "disk0s1"
		diskSize            uint64 = 1_000_000
		testPhysicalStoreID        = "disk0s2"
		testVolumeID               = "disk2s4"
		testSnapshotUUID           = "AAAAAAAA-BBBB-CCCC-DDDD-FFFFFFFFFFFF"
		testVolumeName             = "Macintosh HD - Data"
	)

	d := &PlistDecoder{}
	reader := strings.NewReader(decoderList)

	wantParts := &types.SystemPartitions{
		AllDisks: []string{testDiskID},
		AllDisksAndPartitions: []types.DiskPart{
			{
				DeviceIdentifier: testDiskID,
				Partitions: []types.Partition{
					{DeviceIdentifier: testPartID},
				},
				Size: diskSize,
			},
			{
				APFSPhysicalStores: []types.APFSPhysicalStoreID{
					{DeviceIdentifier: testPhysicalStoreID},
				},
				APFSVolumes: []types.APFSVolume{
					{
						DeviceIdentifier: testVolumeID,
						MountedSnapshots: []types.Snapshot{
							{SnapshotUUID: testSnapshotUUID},
						},
					},
				},
			},
		},
		VolumesFromDisks: []string{testVolumeName},
		WholeDisks:       []string{testDiskID},
	}

	gotParts, err := d.DecodeSystemPartitions(reader)

	assert.NoError(t, err, "should be able to decode valid list plist data")
	assert.ObjectsAreEqualValues(wantParts, gotParts)
}
