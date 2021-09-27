package diskutil

import (
	"bytes"
	"embed"
	"fmt"
	"path"
	"testing"

	mock_diskutil "github.com/aws/ec2-macos-utils/pkg/diskutil/mocks"
	"github.com/aws/ec2-macos-utils/pkg/diskutil/types"

	"github.com/dustin/go-humanize"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

//go:embed testdata/grow
var growDataFS embed.FS

const (
	growDataDir = "testdata/grow"
)

func TestGrowContainer(t *testing.T) {
	// testPrefix is the prefix used to load test data files from testDataFS
	testPrefix := path.Join(growDataDir, "TestGrowContainer-")

	type args struct {
		disk       *types.DiskInfo
		partitions *types.SystemPartitions
	}
	tests := []struct {
		name        string
		args        args
		configure   func(utility *mock_diskutil.MockDiskUtil)
		wantMessage string
		wantErr     bool
		wantErrType error
	}{
		{
			name: "Bad case: not enough free space in parent disk to repair",
			args: args{
				disk: &types.DiskInfo{
					DeviceIdentifier: "disk1",
					APFSPhysicalStores: []types.APFSPhysicalStore{
						{DeviceIdentifier: "disk0s2"},
					},
				},
				partitions: &types.SystemPartitions{
					AllDisksAndPartitions: []types.DiskPart{
						{
							DeviceIdentifier: "disk0",
							Size:             MinimumGrowSpaceRequired,
							Partitions: []types.Partition{
								{Size: MinimumGrowSpaceRequired / 2},
							},
						},
						{DeviceIdentifier: "disk1"},
					},
				},
			},
			configure:   nil,
			wantMessage: "",
			wantErr:     true,
			wantErrType: MinimumGrowSpaceError{},
		},
		{
			name: "Bad case: failed to resize the container",
			args: args{
				disk: &types.DiskInfo{
					DeviceIdentifier: "disk1",
					APFSPhysicalStores: []types.APFSPhysicalStore{
						{DeviceIdentifier: "disk0s2"},
					},
				},
				partitions: &types.SystemPartitions{
					AllDisksAndPartitions: []types.DiskPart{
						{
							DeviceIdentifier: "disk0",
							Size:             MinimumGrowSpaceRequired * 2,
							Partitions: []types.Partition{
								{Size: MinimumGrowSpaceRequired / 2},
							},
						},
						{DeviceIdentifier: "disk1"},
					},
				},
			},
			configure: func(utility *mock_diskutil.MockDiskUtil) {
				gomock.InOrder(
					utility.EXPECT().RepairDisk("disk0").Return("", nil),
					utility.EXPECT().ResizeContainer("disk1", "0").
						Return("error", fmt.Errorf("error")),
				)
			},
			wantMessage: fmt.Sprint("error"),
			wantErr:     true,
			wantErrType: nil,
		},
		{
			name: "Bad case: failed to fetch updated disk information after repair",
			args: args{
				disk: &types.DiskInfo{
					DeviceIdentifier: "disk1",
					APFSPhysicalStores: []types.APFSPhysicalStore{
						{DeviceIdentifier: "disk0s2"},
					},
				},
				partitions: &types.SystemPartitions{
					AllDisksAndPartitions: []types.DiskPart{
						{
							DeviceIdentifier: "disk0",
							Size:             MinimumGrowSpaceRequired * 2,
							Partitions: []types.Partition{
								{Size: MinimumGrowSpaceRequired / 2},
							},
						},
						{DeviceIdentifier: "disk1"},
					},
				},
			},
			configure: func(utility *mock_diskutil.MockDiskUtil) {
				gomock.InOrder(
					utility.EXPECT().RepairDisk("disk0").Return("", nil),
					utility.EXPECT().ResizeContainer("disk1", "0").Return("success", nil),
					utility.EXPECT().Info("disk1").Return(nil, fmt.Errorf("error")),
				)
			},
			wantMessage: fmt.Sprintf("failed to fetch updated disk information for container [%s]", "disk1"),
			wantErr:     true,
			wantErrType: nil,
		},
		{
			name: "Good case: successfully resized the container and fetched the updated information",
			args: args{
				disk: &types.DiskInfo{
					DeviceIdentifier: "disk1",
					APFSPhysicalStores: []types.APFSPhysicalStore{
						{DeviceIdentifier: "disk0s2"},
					},
				},
				partitions: &types.SystemPartitions{
					AllDisksAndPartitions: []types.DiskPart{
						{
							DeviceIdentifier: "disk0",
							Size:             MinimumGrowSpaceRequired * 2,
							Partitions: []types.Partition{
								{Size: MinimumGrowSpaceRequired / 2},
							},
						},
						{DeviceIdentifier: "disk1"},
					},
				},
			},
			configure: func(utility *mock_diskutil.MockDiskUtil) {
				rawInfoOutput, err := growDataFS.ReadFile(testPrefix + "good-Info.txt")
				assert.Nil(t, err)

				decoder := &PlistDecoder{}
				info, err := decoder.DecodeDiskInfo(bytes.NewReader(rawInfoOutput))
				assert.Nil(t, err)

				gomock.InOrder(
					utility.EXPECT().RepairDisk("disk0").Return("", nil),
					utility.EXPECT().ResizeContainer("disk1", "0").Return("success", nil),
					utility.EXPECT().Info("disk1").Return(info, nil),
				)
			},
			wantMessage: fmt.Sprintf("grew container [%s] to size [%s]", "disk1", humanize.Bytes(1500000)),
			wantErr:     false,
			wantErrType: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUtility := mock_diskutil.NewMockDiskUtil(ctrl)

			// If the test has a configure function, initialize the mock utility and configure it for the test to use
			if tt.configure != nil {
				tt.configure(mockUtility)
			}

			gotMessage, err := GrowContainer(tt.args.disk, tt.args.partitions, mockUtility)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("growContainer() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.wantErrType != nil {
					if !assert.IsType(t, tt.wantErrType, err) {
						t.Errorf("growContainer() error = %T, wantErrType %T", err, tt.wantErrType)
					}
				}
			}
			if gotMessage != tt.wantMessage {
				t.Errorf("growContainer() gotMessage = %v, want %v", gotMessage, tt.wantMessage)
			}
		})
	}
}

func Test_repairParentDisk(t *testing.T) {
	type args struct {
		disk       *types.DiskInfo
		partitions *types.SystemPartitions
	}
	tests := []struct {
		name        string
		args        args
		configure   func(utility *mock_diskutil.MockDiskUtil)
		wantMessage string
		wantErr     bool
		wantErrType error
	}{
		{
			name: "Bad case: error getting parent device ID from disk",
			args: args{
				disk: &types.DiskInfo{
					DeviceIdentifier:   "disk0",
					APFSPhysicalStores: nil,
				},
				partitions: nil,
			},
			configure:   nil,
			wantMessage: "failed to get the parent disk ID for container [disk0]",
			wantErr:     true,
			wantErrType: nil,
		},
		{
			name: "Bad case: error getting available free space for ID",
			args: args{
				disk: &types.DiskInfo{DeviceIdentifier: "disk2"},
				partitions: &types.SystemPartitions{
					AllDisksAndPartitions: []types.DiskPart{
						{DeviceIdentifier: "disk0"},
						{DeviceIdentifier: "disk1"},
					},
				},
			},
			configure:   nil,
			wantMessage: "failed to get the parent disk ID for container [disk2]",
			wantErr:     true,
			wantErrType: nil,
		},
		{
			name: "Bad case: not enough space on parent disk",
			args: args{
				disk: &types.DiskInfo{
					DeviceIdentifier: "disk1",
					APFSPhysicalStores: []types.APFSPhysicalStore{
						{DeviceIdentifier: "disk0s2"},
					},
				},
				partitions: &types.SystemPartitions{
					AllDisksAndPartitions: []types.DiskPart{
						{DeviceIdentifier: "disk0"},
						{
							DeviceIdentifier: "disk1",
							Size:             MinimumGrowSpaceRequired,
							Partitions: []types.Partition{
								{Size: MinimumGrowSpaceRequired / 2},
							},
						},
					},
				},
			},
			configure:   nil,
			wantMessage: "",
			wantErr:     true,
			wantErrType: MinimumGrowSpaceError{},
		},
		{
			name: "Bad case: error while attempting to repair the parent disk",
			args: args{
				disk: &types.DiskInfo{
					DeviceIdentifier: "disk1",
					APFSPhysicalStores: []types.APFSPhysicalStore{
						{DeviceIdentifier: "disk0s2"},
					},
				},
				partitions: &types.SystemPartitions{
					AllDisksAndPartitions: []types.DiskPart{
						{DeviceIdentifier: "disk0",
							Size: MinimumGrowSpaceRequired * 2,
							Partitions: []types.Partition{
								{Size: MinimumGrowSpaceRequired / 2},
							},
						},
						{DeviceIdentifier: "disk1"},
					},
				},
			},
			configure: func(utility *mock_diskutil.MockDiskUtil) {
				utility.EXPECT().RepairDisk("disk0").Return("error", fmt.Errorf("error"))
			},
			wantMessage: "error",
			wantErr:     true,
			wantErrType: nil,
		},
		{
			name: "Good case: successfully repair the parent disk",
			args: args{
				disk: &types.DiskInfo{
					DeviceIdentifier: "disk1",
					APFSPhysicalStores: []types.APFSPhysicalStore{
						{DeviceIdentifier: "disk0s2"},
					},
				},
				partitions: &types.SystemPartitions{
					AllDisksAndPartitions: []types.DiskPart{
						{
							DeviceIdentifier: "disk0",
							Size:             MinimumGrowSpaceRequired * 2,
							Partitions: []types.Partition{
								{Size: MinimumGrowSpaceRequired / 2},
							},
						},
						{DeviceIdentifier: "disk1"},
					},
				},
			},
			configure: func(utility *mock_diskutil.MockDiskUtil) {
				utility.EXPECT().RepairDisk("disk0").Return("success", nil)
			},
			wantMessage: "success",
			wantErr:     false,
			wantErrType: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUtility := mock_diskutil.NewMockDiskUtil(ctrl)

			// If the test has a configure function, initialize the mock utility and configure it for the test to use
			if tt.configure != nil {
				tt.configure(mockUtility)
			}

			gotMessage, err := repairParentDisk(tt.args.disk, tt.args.partitions, mockUtility)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("repairParentDisk() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.wantErrType != nil {
					if !assert.IsType(t, tt.wantErrType, err) {
						t.Errorf("repairParentDisk() error = %T, wantErrType %T", err, tt.wantErrType)
					}
				}
			}
			if gotMessage != tt.wantMessage {
				t.Errorf("repairParentDisk() gotMessage = %v, want %v", gotMessage, tt.wantMessage)
			}
		})
	}
}
