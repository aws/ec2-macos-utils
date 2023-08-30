package diskutil

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	mock_diskutil "github.com/aws/ec2-macos-utils/internal/diskutil/mocks"
	"github.com/aws/ec2-macos-utils/internal/diskutil/types"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func init() {
	logrus.SetOutput(ioutil.Discard)
}

func TestGrowContainer_WithoutContainer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUtility := mock_diskutil.NewMockDiskUtil(ctrl)

	err := GrowContainer(context.Background(), mockUtility, nil)

	assert.Error(t, err, "shouldn't be able to grow container with nil container")
}

func TestGrowContainer_WithEmptyContainer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUtility := mock_diskutil.NewMockDiskUtil(ctrl)

	disk := types.DiskInfo{}

	err := GrowContainer(context.Background(), mockUtility, &disk)

	assert.Error(t, err, "shouldn't be able to grow container with empty container")
}

func TestGrowContainer_WithInfoErr(t *testing.T) {
	const testDiskID = "disk1"
	var ctx = context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUtility := mock_diskutil.NewMockDiskUtil(ctrl)
	mockUtility.EXPECT().Info(ctx, testDiskID).Return(nil, fmt.Errorf("error"))

	disk := types.DiskInfo{
		ContainerInfo: types.ContainerInfo{
			FilesystemType: "apfs",
		},
		ParentWholeDisk:   testDiskID,
		VirtualOrPhysical: "Virtual",
	}

	err := GrowContainer(context.Background(), mockUtility, &disk)

	assert.Error(t, err, "shouldn't be able to grow container with info error")
}

func TestGrowContainer_WithRepairDiskErr(t *testing.T) {
	const testDiskID = "disk1"
	var ctx = context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUtility := mock_diskutil.NewMockDiskUtil(ctrl)
	mockUtility.EXPECT().RepairDisk(ctx, testDiskID).Return("", fmt.Errorf("error"))

	disk := types.DiskInfo{
		APFSPhysicalStores: []types.APFSPhysicalStore{
			{DeviceIdentifier: testDiskID},
		},
		ContainerInfo: types.ContainerInfo{
			FilesystemType: "apfs",
		},
		ParentWholeDisk:   testDiskID,
		VirtualOrPhysical: "Physical",
	}

	err := GrowContainer(context.Background(), mockUtility, &disk)

	assert.Error(t, err, "shouldn't be able to grow container with repair disk error")
}

func TestGrowContainer_WithListError(t *testing.T) {
	const testDiskID = "disk1"
	var ctx = context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUtility := mock_diskutil.NewMockDiskUtil(ctrl)
	gomock.InOrder(
		mockUtility.EXPECT().RepairDisk(ctx, testDiskID).Return("", nil),
		mockUtility.EXPECT().List(ctx, nil).Return(nil, fmt.Errorf("error")),
	)

	disk := types.DiskInfo{
		APFSPhysicalStores: []types.APFSPhysicalStore{
			{DeviceIdentifier: testDiskID},
		},
		ContainerInfo: types.ContainerInfo{
			FilesystemType: "apfs",
		},
		ParentWholeDisk:   testDiskID,
		VirtualOrPhysical: "Physical",
	}

	err := GrowContainer(context.Background(), mockUtility, &disk)

	assert.Error(t, err, "shouldn't be able to grow container with list error")
}

func TestGrowContainer_WithoutFreeSpace(t *testing.T) {
	const (
		testDiskID = "disk1"
		// total disk size
		diskSize uint64 = 1_000_000
		// individual partition space occupied
		partSize uint64 = 500_000
		// expected amount of free space
		expectedFreeSpace = 0
	)
	var ctx = context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parts := types.SystemPartitions{
		AllDisksAndPartitions: []types.DiskPart{
			{
				DeviceIdentifier: testDiskID,
				Size:             diskSize,
				Partitions: []types.Partition{
					{Size: partSize},
					{Size: partSize},
				},
			},
		},
	}

	mockUtility := mock_diskutil.NewMockDiskUtil(ctrl)
	gomock.InOrder(
		mockUtility.EXPECT().RepairDisk(ctx, testDiskID).Return("", nil),
		mockUtility.EXPECT().List(ctx, nil).Return(&parts, nil),
	)

	disk := types.DiskInfo{
		APFSPhysicalStores: []types.APFSPhysicalStore{
			{DeviceIdentifier: testDiskID},
		},
		ContainerInfo: types.ContainerInfo{
			FilesystemType: "apfs",
		},
		ParentWholeDisk:   testDiskID,
		VirtualOrPhysical: "Physical",
	}

	expectedErr := fmt.Errorf("not enough space to resize container: %w", FreeSpaceError{expectedFreeSpace})

	actualErr := GrowContainer(context.Background(), mockUtility, &disk)

	assert.Error(t, actualErr, "shouldn't be able to grow container without free space")
	assert.Equal(t, expectedErr, actualErr, "should get FreeSpaceError since there's no free space")
}

func TestGrowContainer_WithResizeContainerError(t *testing.T) {
	const (
		testDiskID = "disk1"
		// total disk size
		diskSize uint64 = 3_000_000
		// individual partition space occupied
		partSize uint64 = 500_000
	)
	var ctx = context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parts := types.SystemPartitions{
		AllDisksAndPartitions: []types.DiskPart{
			{
				DeviceIdentifier: testDiskID,
				Size:             diskSize,
				Partitions: []types.Partition{
					{Size: partSize},
					{Size: partSize},
				},
			},
		},
	}

	mockUtility := mock_diskutil.NewMockDiskUtil(ctrl)
	gomock.InOrder(
		mockUtility.EXPECT().RepairDisk(ctx, testDiskID).Return("", nil),
		mockUtility.EXPECT().List(ctx, nil).Return(&parts, nil),
		mockUtility.EXPECT().ResizeContainer(ctx, testDiskID, "0").Return("", fmt.Errorf("error")),
	)

	disk := types.DiskInfo{
		APFSPhysicalStores: []types.APFSPhysicalStore{
			{DeviceIdentifier: testDiskID},
		},
		ContainerInfo: types.ContainerInfo{
			FilesystemType: "apfs",
		},
		DeviceIdentifier:  testDiskID,
		ParentWholeDisk:   testDiskID,
		VirtualOrPhysical: "Physical",
	}

	err := GrowContainer(context.Background(), mockUtility, &disk)

	assert.Error(t, err, "shouldn't be able to grow container with resize container error")
}

func TestGrowContainer_Success(t *testing.T) {
	const (
		testDiskID = "disk1"
		// total disk size
		diskSize uint64 = 3_000_000
		// individual partition space occupied
		partSize uint64 = 500_000
	)
	var ctx = context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parts := types.SystemPartitions{
		AllDisksAndPartitions: []types.DiskPart{
			{
				DeviceIdentifier: testDiskID,
				Size:             diskSize,
				Partitions: []types.Partition{
					{Size: partSize},
					{Size: partSize},
				},
			},
		},
	}

	mockUtility := mock_diskutil.NewMockDiskUtil(ctrl)
	gomock.InOrder(
		mockUtility.EXPECT().RepairDisk(ctx, testDiskID).Return("", nil),
		mockUtility.EXPECT().List(ctx, nil).Return(&parts, nil),
		mockUtility.EXPECT().ResizeContainer(ctx, testDiskID, "0").Return("", nil),
	)

	disk := types.DiskInfo{
		APFSPhysicalStores: []types.APFSPhysicalStore{
			{DeviceIdentifier: testDiskID},
		},
		ContainerInfo: types.ContainerInfo{
			FilesystemType: "apfs",
		},
		DeviceIdentifier:  testDiskID,
		ParentWholeDisk:   testDiskID,
		VirtualOrPhysical: "Physical",
	}

	err := GrowContainer(context.Background(), mockUtility, &disk)

	assert.NoError(t, err, "should be able to grow container")
}

func TestCanAPFSResize(t *testing.T) {
	type args struct {
		container *types.DiskInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "WithoutContainer",
			args: args{
				container: nil,
			},
			wantErr: true,
		},
		{
			name: "WithEmptyContainer",
			args: args{
				container: &types.DiskInfo{
					ContainerInfo: types.ContainerInfo{},
				},
			},
			wantErr: true,
		},
		{
			name: "WithoutAPFSFilesystem",
			args: args{
				container: &types.DiskInfo{
					ContainerInfo: types.ContainerInfo{
						FilesystemType: "not apfs",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "WithoutAPFSReference",
			args: args{
				container: &types.DiskInfo{},
			},
			wantErr: true,
		},
		{
			name: "WithoutAPFSPhysicalStores",
			args: args{
				container: &types.DiskInfo{
					APFSContainerReference: "disk1",
				},
			},
			wantErr: true,
		},
		{
			name: "SuccessAPFSFilesystem",
			args: args{
				container: &types.DiskInfo{
					ContainerInfo: types.ContainerInfo{
						FilesystemType: "apfs",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "SuccessAPFSContainer",
			args: args{
				container: &types.DiskInfo{
					APFSContainerReference: "disk1",
					APFSPhysicalStores: []types.APFSPhysicalStore{
						{DeviceIdentifier: "disk0"},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := canAPFSResize(tt.args.container)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetDiskFreeSpace_WithListErr(t *testing.T) {
	const expectedSize uint64 = 0
	var ctx = context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUtility := mock_diskutil.NewMockDiskUtil(ctrl)
	mockUtility.EXPECT().List(ctx, nil).Return(nil, fmt.Errorf("error"))

	disk := types.DiskInfo{}

	actual, err := getDiskFreeSpace(context.Background(), mockUtility, &disk)

	assert.Error(t, err, "shouldn't be able to get free space with list error")
	assert.Equal(t, expectedSize, actual, "shouldn't get size due to list error")
}

func TestGetDiskFreeSpace_WithNilSystemPartitions(t *testing.T) {
	const expectedSize uint64 = 0
	var ctx = context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUtility := mock_diskutil.NewMockDiskUtil(ctrl)
	mockUtility.EXPECT().List(ctx, nil).Return(nil, nil)

	disk := types.DiskInfo{}

	actual, err := getDiskFreeSpace(context.Background(), mockUtility, &disk)

	assert.Error(t, err, "shouldn't be able to get free space for nil partitions")
	assert.Equal(t, expectedSize, actual, "shouldn't get size due to nil partitions")
}

func TestGetDiskFreeSpace_WithoutFreeSpace(t *testing.T) {
	const (
		testDiskID = "disk1"
		// total disk size
		diskSize uint64 = 1_000_000
		// individual partition space occupied
		partSize uint64 = 500_000
		// should see: diskSize - (2 * partSize)
		expectedFreeSpace uint64 = 0
	)
	var ctx = context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUtility := mock_diskutil.NewMockDiskUtil(ctrl)

	parts := types.SystemPartitions{
		AllDisksAndPartitions: []types.DiskPart{
			{
				DeviceIdentifier: testDiskID,
				Size:             diskSize,
				Partitions: []types.Partition{
					{Size: partSize},
					{Size: partSize},
				},
			},
		},
	}
	mockUtility.EXPECT().List(ctx, nil).Return(&parts, nil)

	disk := types.DiskInfo{
		APFSPhysicalStores: []types.APFSPhysicalStore{
			{DeviceIdentifier: testDiskID},
		},
	}

	actual, err := getDiskFreeSpace(context.Background(), mockUtility, &disk)

	assert.NoError(t, err, "should be able to calculate free space with valid data")
	assert.Equal(t, expectedFreeSpace, actual, "should have calculated free space based on partitions")
}

func TestGetDiskFreeSpace_FreeSpace(t *testing.T) {
	const (
		testDiskID = "disk1"
		// total disk size
		diskSize uint64 = 2_000_000
		// individual partition space occupied
		partSize uint64 = 500_000
		// should see: diskSize - (2 * partSize)
		expectedFreeSpace uint64 = 1_000_000
	)
	var ctx = context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUtility := mock_diskutil.NewMockDiskUtil(ctrl)

	parts := types.SystemPartitions{
		AllDisksAndPartitions: []types.DiskPart{
			{
				DeviceIdentifier: testDiskID,
				Size:             diskSize,
				Partitions: []types.Partition{
					{Size: partSize},
					{Size: partSize},
				},
			},
		},
	}
	mockUtility.EXPECT().List(ctx, nil).Return(&parts, nil)

	disk := types.DiskInfo{
		APFSPhysicalStores: []types.APFSPhysicalStore{
			{DeviceIdentifier: testDiskID},
		},
	}

	actual, err := getDiskFreeSpace(context.Background(), mockUtility, &disk)

	assert.NoError(t, err, "should be able to calculate free space with valid data")
	assert.Equal(t, expectedFreeSpace, actual, "should have calculated free space based on partitions")
}

func TestRepairParentDisk_WithoutDiskInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUtility := mock_diskutil.NewMockDiskUtil(ctrl)

	disk := types.DiskInfo{}
	expectedMessage := fmt.Sprintf("failed to get the parent disk ID for container [%s]", disk.DeviceIdentifier)

	actualMessage, err := repairParentDisk(context.Background(), mockUtility, &disk)

	assert.Error(t, err, "shouldn't be able to repair disk without disk info")
	assert.Equal(t, expectedMessage, actualMessage, "should see error message for device")
}

func TestRepairParentDisk_WithRepairDiskErr(t *testing.T) {
	const testDiskID = "disk0"
	var ctx = context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUtility := mock_diskutil.NewMockDiskUtil(ctrl)
	mockUtility.EXPECT().RepairDisk(ctx, testDiskID).Return("error", fmt.Errorf("error"))

	disk := types.DiskInfo{
		APFSPhysicalStores: []types.APFSPhysicalStore{
			{DeviceIdentifier: testDiskID},
		},
	}
	expectedMessage := "error"

	actualMessage, err := repairParentDisk(context.Background(), mockUtility, &disk)

	assert.Error(t, err, "shouldn't be able to repair parent disk with repair disk error")
	assert.Equal(t, expectedMessage, actualMessage, "should see error message for device")
}

func TestRepairParentDisk_Success(t *testing.T) {
	const (
		testDiskID      = "disk0"
		expectedMessage = ""
	)
	var ctx = context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUtility := mock_diskutil.NewMockDiskUtil(ctrl)
	mockUtility.EXPECT().RepairDisk(ctx, testDiskID).Return("", nil)

	disk := types.DiskInfo{
		APFSPhysicalStores: []types.APFSPhysicalStore{
			{DeviceIdentifier: testDiskID},
		},
	}

	actualMessage, err := repairParentDisk(context.Background(), mockUtility, &disk)

	assert.NoError(t, err, "should be able to repair parent with valid data")
	assert.Equal(t, expectedMessage, actualMessage, "should see expected message")
}
