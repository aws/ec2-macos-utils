package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSystemPartitions_AvailableDiskSpace_WithoutTargetDisk(t *testing.T) {
	const (
		testDiskID = "disk3"
		// should see 0 since testDiskID isn't in AllDisksAndPartitions
		expectedAvailableSize uint64 = 0
	)

	p := &SystemPartitions{
		AllDisksAndPartitions: []DiskPart{
			{DeviceIdentifier: "disk0"},
			{DeviceIdentifier: "disk1"},
			{DeviceIdentifier: "disk2"},
		},
	}

	actual, err := p.AvailableDiskSpace(testDiskID)

	assert.Error(t, err, "shouldn't be able to find disk in partitions")
	assert.Equal(t, expectedAvailableSize, actual, "shouldn't return anything since the disk doesn't exist")
}

func TestSystemPartitions_AvailableDiskSpace_GoodDisk(t *testing.T) {
	const (
		testDiskID = "disk1"
		// total disk size
		diskSize uint64 = 2_000_000
		// individual partition space occupied
		partSize uint64 = 250_000
		// should see: diskSize - (2 * partSize)
		expectedAvailableSize uint64 = 1_500_000
	)

	p := &SystemPartitions{
		AllDisksAndPartitions: []DiskPart{
			// Non-targeted disk, should be skipped
			{DeviceIdentifier: "disk0"},
			{
				DeviceIdentifier: "disk1",
				Size:             diskSize,
				Partitions: []Partition{
					{Size: partSize},
					{Size: partSize},
				},
			},
		},
	}

	actual, err := p.AvailableDiskSpace(testDiskID)

	assert.NoError(t, err, "should be able to calculate free space with valid data")
	assert.Equal(t, expectedAvailableSize, actual, "should have calculated free space based on partitions")
}
