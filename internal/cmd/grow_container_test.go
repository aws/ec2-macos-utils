package cmd

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

func TestRun_WithInfoErr(t *testing.T) {
	const (
		testDiskID    = "root"
		testDiskAlias = "/"
	)
	var ctx = context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := mock_diskutil.NewMockDiskUtil(ctrl)
	mock.EXPECT().Info(ctx, testDiskAlias).Return(nil, fmt.Errorf("error"))

	err := run(ctx, mock, growContainer{
		id: testDiskID,
	})

	assert.Error(t, err, `should fail to get disk information for /`)
}

func TestRun_WithoutDiskInfo(t *testing.T) {
	const (
		testDiskID    = "root"
		testDiskAlias = "/"
	)
	var ctx = context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	disk := types.DiskInfo{
		DeviceIdentifier: testDiskID,
	}

	mock := mock_diskutil.NewMockDiskUtil(ctrl)
	mock.EXPECT().Info(ctx, testDiskAlias).Return(&disk, nil)

	err := run(ctx, mock, growContainer{
		id: testDiskID,
	})

	assert.Error(t, err, "should fail to grow the container since the DiskInfo isn't populated")
}

func TestRun_WithoutFreeSpace(t *testing.T) {
	const (
		testDiskID        = "disk1"
		diskSize   uint64 = 3_000_000
		partSize   uint64 = 1_500_000
	)
	var ctx = context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parts := types.SystemPartitions{
		AllDisks: []string{testDiskID},
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

	mock := mock_diskutil.NewMockDiskUtil(ctrl)
	gomock.InOrder(
		mock.EXPECT().List(ctx, nil).Return(&parts, nil),
		mock.EXPECT().Info(ctx, testDiskID).Return(&disk, nil),
		mock.EXPECT().RepairDisk(ctx, testDiskID).Return("", nil),
		mock.EXPECT().List(ctx, nil).Return(&parts, nil),
	)

	err := run(ctx, mock, growContainer{
		id: testDiskID,
	})

	assert.NoError(t, err, "should exit quietly if there isn't enough free space to grow")
}

func TestRun_WithUpdatedInfoErr(t *testing.T) {
	const (
		testDiskID        = "disk1"
		diskSize   uint64 = 3_000_000
		partSize   uint64 = 500_000
	)
	var ctx = context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parts := types.SystemPartitions{
		AllDisks: []string{testDiskID},
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

	mock := mock_diskutil.NewMockDiskUtil(ctrl)
	gomock.InOrder(
		mock.EXPECT().List(ctx, nil).Return(&parts, nil),
		mock.EXPECT().Info(ctx, testDiskID).Return(&disk, nil),
		mock.EXPECT().RepairDisk(ctx, testDiskID).Return("", nil),
		mock.EXPECT().List(ctx, nil).Return(&parts, nil),
		mock.EXPECT().ResizeContainer(ctx, testDiskID, "0").Return("", nil),
		mock.EXPECT().List(ctx, nil).Return(nil, fmt.Errorf("error")),
	)

	err := run(ctx, mock, growContainer{
		id: testDiskID,
	})

	assert.Error(t, err, "should fail to get updated DiskInfo due to list error")
}

func TestRun_Success(t *testing.T) {
	const (
		testDiskID        = "disk1"
		diskSize   uint64 = 3_000_000
		partSize   uint64 = 500_000
	)
	var ctx = context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parts := types.SystemPartitions{
		AllDisks: []string{testDiskID},
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

	disk := types.DiskInfo{
		APFSPhysicalStores: []types.APFSPhysicalStore{
			{DeviceIdentifier: testDiskID},
		},
		ContainerInfo: types.ContainerInfo{
			FilesystemType: "apfs",
		},
		DeviceIdentifier:  testDiskID,
		ParentWholeDisk:   testDiskID,
		TotalSize:         diskSize,
		VirtualOrPhysical: "Physical",
	}

	mock := mock_diskutil.NewMockDiskUtil(ctrl)
	gomock.InOrder(
		mock.EXPECT().List(ctx, nil).Return(&parts, nil),
		mock.EXPECT().Info(ctx, testDiskID).Return(&disk, nil),
		mock.EXPECT().RepairDisk(ctx, testDiskID).Return("", nil),
		mock.EXPECT().List(ctx, nil).Return(&parts, nil),
		mock.EXPECT().ResizeContainer(ctx, testDiskID, "0").Return("", nil),
		mock.EXPECT().List(ctx, nil).Return(&parts, nil),
		mock.EXPECT().Info(ctx, testDiskID).Return(&disk, nil),
	)

	err := run(ctx, mock, growContainer{
		id: testDiskID,
	})

	assert.NoError(t, err, "should be able to grow container with valid data")
}

func TestGetTargetDiskInfo_WithRootInfoErr(t *testing.T) {
	const testDiskID = "root"
	var ctx = context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := mock_diskutil.NewMockDiskUtil(ctrl)
	mock.EXPECT().Info(ctx, "/").Return(nil, fmt.Errorf("error"))

	di, err := getTargetDiskInfo(ctx, mock, testDiskID)

	assert.Error(t, err, `should fail to get DiskInfo for /`)
	assert.Nil(t, di)
}

func TestGetTargetDiskInfo_WithListErr(t *testing.T) {
	const testDiskID = "disk1"
	var ctx = context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := mock_diskutil.NewMockDiskUtil(ctrl)
	mock.EXPECT().List(ctx, nil).Return(nil, fmt.Errorf("error"))

	di, err := getTargetDiskInfo(ctx, mock, testDiskID)

	assert.Error(t, err, "should fail to get partition information")
	assert.Nil(t, di)
}

func TestGetTargetDiskInfo_NoTargetDisk(t *testing.T) {
	const (
		testDiskID     = "disk1"
		testAllDisksID = "disk0"
	)
	var ctx = context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parts := types.SystemPartitions{
		AllDisks: []string{testAllDisksID},
	}

	mock := mock_diskutil.NewMockDiskUtil(ctrl)
	mock.EXPECT().List(ctx, nil).Return(&parts, nil)

	di, err := getTargetDiskInfo(ctx, mock, testDiskID)

	assert.Error(t, err, "should fail to find targetDiskID in partition information")
	assert.Nil(t, di, "should get nil data for invalid target disk")
}

func TestGetTargetDiskInfo_WithInfoErr(t *testing.T) {
	const testDiskID = "disk1"
	var ctx = context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parts := types.SystemPartitions{
		AllDisks: []string{testDiskID},
	}

	mock := mock_diskutil.NewMockDiskUtil(ctrl)
	gomock.InOrder(
		mock.EXPECT().List(ctx, nil).Return(&parts, nil),
		mock.EXPECT().Info(ctx, testDiskID).Return(nil, fmt.Errorf("error")),
	)

	di, err := getTargetDiskInfo(ctx, mock, testDiskID)

	assert.Error(t, err, "should fail to get disk information")
	assert.Nil(t, di, "should get nil data with info error")
}

func TestGetTargetDiskInfo_Success(t *testing.T) {
	const testDiskID = "disk1"
	var ctx = context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	parts := types.SystemPartitions{
		AllDisks: []string{testDiskID},
	}

	expectedDisk := &types.DiskInfo{}

	mock := mock_diskutil.NewMockDiskUtil(ctrl)
	gomock.InOrder(
		mock.EXPECT().List(ctx, nil).Return(&parts, nil),
		mock.EXPECT().Info(ctx, testDiskID).Return(expectedDisk, nil),
	)

	actualDisk, err := getTargetDiskInfo(ctx, mock, testDiskID)

	assert.NoError(t, err, "should be able succeeded with valid info")
	assert.Equal(t, expectedDisk, actualDisk, "should be able to get expected data from info")
}

func TestValidateDeviceID(t *testing.T) {
	type args struct {
		id         string
		partitions *types.SystemPartitions
	}
	tests := []struct {
		name      string
		args      args
		wantValid bool
		wantErr   bool
	}{
		{
			name: "without ID",
			args: args{
				id:         "",
				partitions: nil,
			},
			wantErr: true,
		},
		{
			name: "invalid ID prefix",
			args: args{
				id:         "bad",
				partitions: nil,
			},
			wantErr: true,
		},
		{
			name: "without disk number",
			args: args{
				id:         "disk",
				partitions: nil,
			},
			wantErr: true,
		},
		{
			name: "no target disk",
			args: args{
				id: "disk2",
				partitions: &types.SystemPartitions{
					AllDisks: []string{"disk0", "disk1"},
				},
			},
			wantErr: true,
		},
		{
			name: "success",
			args: args{
				id: "disk0",
				partitions: &types.SystemPartitions{
					AllDisks: []string{"disk0", "disk1"},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDeviceID(tt.args.id, tt.args.partitions)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
