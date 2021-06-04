package cmd

import (
	"embed"
	"fmt"
	"path"
	"testing"

	"github.com/aws/ec2-macos-utils/pkg/diskutil"
	"github.com/aws/ec2-macos-utils/pkg/diskutil/mocks"
	"github.com/dustin/go-humanize"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

//go:embed testdata/growContainer
var testDataFS embed.FS
var testDataDir = "testdata/growContainer"

func TestMinimumGrowSpaceError_Error(t *testing.T) {
	type fields struct {
		message string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Good case: returns the expected string",
			fields: fields{
				message: fmt.Sprintf("Nothing to do, at least [%s] of free space is required to grow the container", humanize.Bytes(MinimumGrowSpaceRequired)),
			},
			want: fmt.Sprintf("Nothing to do, at least [%s] of free space is required to grow the container", humanize.Bytes(MinimumGrowSpaceRequired)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := MinimumGrowSpaceError{
				message: tt.fields.message,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_availableDiskSpace(t *testing.T) {
	type args struct {
		id         string
		partitions *diskutil.SystemPartitions
	}
	tests := []struct {
		name     string
		args     args
		wantSize uint64
		wantErr  bool
	}{
		{
			name: "Bad case: ID not found in system partitions",
			args: args{
				id: "disk3",
				partitions: &diskutil.SystemPartitions{
					AllDisksAndPartitions: []diskutil.DiskPart{
						{DeviceIdentifier: "disk0"},
						{DeviceIdentifier: "disk1"},
						{DeviceIdentifier: "disk2"},
					},
				},
			},
			wantSize: 0,
			wantErr:  true,
		},
		{
			name: "Good case: ID found in system partitions and size matches",
			args: args{
				id: "disk1",
				partitions: &diskutil.SystemPartitions{
					AllDisksAndPartitions: []diskutil.DiskPart{
						{DeviceIdentifier: "disk0"},
						{
							DeviceIdentifier: "disk1",
							Size:             2000000,
							Partitions: []diskutil.Partition{
								{Size: 500000},
								{Size: 500000},
							},
						},
					},
				},
			},
			wantSize: 1000000,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSize, err := availableDiskSpace(tt.args.id, tt.args.partitions)
			if (err != nil) != tt.wantErr {
				t.Errorf("availableDiskSpace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotSize != tt.wantSize {
				t.Errorf("availableDiskSpace() gotSize = %v, want %v", gotSize, tt.wantSize)
			}
		})
	}
}

func Test_growContainer(t *testing.T) {
	// testPrefix is the prefix used to load test data files from testDataFS
	testPrefix := path.Join(testDataDir, "Test_growContainer-")

	type args struct {
		disk       *diskutil.DiskInfo
		partitions *diskutil.SystemPartitions
		utility    diskutil.DiskUtil
	}
	tests := []struct {
		name        string
		args        args
		configure   func(utility *mock_diskutil.MockDiskUtility)
		wantMessage string
		wantErr     bool
		wantErrType error
	}{
		{
			name: "Bad case: unable to repair parent disk since there's not enough free space",
			args: args{
				disk: &diskutil.DiskInfo{
					DeviceIdentifier: "disk1",
					APFSPhysicalStores: []diskutil.APFSPhysicalStore{
						{DeviceIdentifier: "disk0s2"},
					},
				},
				partitions: &diskutil.SystemPartitions{
					AllDisksAndPartitions: []diskutil.DiskPart{
						{
							DeviceIdentifier: "disk0",
							Size:             MinimumGrowSpaceRequired,
							Partitions: []diskutil.Partition{
								{Size: MinimumGrowSpaceRequired / 2},
							},
						},
						{DeviceIdentifier: "disk1"},
					},
				},
				utility: diskutil.DiskUtil{},
			},
			configure: nil,
			wantMessage: fmt.Sprintf("Nothing to do, at least [%s] of free space is "+
				"required to grow the container",
				humanize.Bytes(MinimumGrowSpaceRequired)),
			wantErr:     true,
			wantErrType: MinimumGrowSpaceError{},
		},
		{
			name: "Bad case: failed to repair the parent disk",
			args: args{
				disk: &diskutil.DiskInfo{
					DeviceIdentifier: "disk1",
					APFSPhysicalStores: []diskutil.APFSPhysicalStore{
						{DeviceIdentifier: "disk0s2"},
						{DeviceIdentifier: "disk2s2"},
					},
				},
				partitions: &diskutil.SystemPartitions{
					AllDisksAndPartitions: []diskutil.DiskPart{
						{
							DeviceIdentifier: "disk0",
							Size:             MinimumGrowSpaceRequired,
							Partitions: []diskutil.Partition{
								{Size: MinimumGrowSpaceRequired / 2},
							},
						},
						{DeviceIdentifier: "disk1"},
					},
				},
				utility: diskutil.DiskUtil{},
			},
			configure:   nil,
			wantMessage: fmt.Sprint("failed to get the parent disk ID for container [disk1]"),
			wantErr:     true,
			wantErrType: nil,
		},
		{
			name: "Bad case: failed to resize the container",
			args: args{
				disk: &diskutil.DiskInfo{
					DeviceIdentifier: "disk1",
					APFSPhysicalStores: []diskutil.APFSPhysicalStore{
						{DeviceIdentifier: "disk0s2"},
					},
				},
				partitions: &diskutil.SystemPartitions{
					AllDisksAndPartitions: []diskutil.DiskPart{
						{
							DeviceIdentifier: "disk0",
							Size:             MinimumGrowSpaceRequired * 2,
							Partitions: []diskutil.Partition{
								{Size: MinimumGrowSpaceRequired / 2},
							},
						},
						{DeviceIdentifier: "disk1"},
					},
				},
				utility: diskutil.DiskUtil{
					Utility: nil,
					Decoder: &diskutil.Decoder{},
				},
			},
			configure: func(utility *mock_diskutil.MockDiskUtility) {
				gomock.InOrder(
					utility.EXPECT().RepairDisk("disk0").Return("", nil),
					utility.EXPECT().ResizeContainer("disk1", "0").Return("error",
						fmt.Errorf("error")),
				)
			},
			wantMessage: fmt.Sprint("error"),
			wantErr:     true,
			wantErrType: nil,
		},
		{
			name: "Bad case: failed to fetch updated disk information after repair",
			args: args{
				disk: &diskutil.DiskInfo{
					DeviceIdentifier: "disk1",
					APFSPhysicalStores: []diskutil.APFSPhysicalStore{
						{DeviceIdentifier: "disk0s2"},
					},
				},
				partitions: &diskutil.SystemPartitions{
					AllDisksAndPartitions: []diskutil.DiskPart{
						{
							DeviceIdentifier: "disk0",
							Size:             MinimumGrowSpaceRequired * 2,
							Partitions: []diskutil.Partition{
								{Size: MinimumGrowSpaceRequired / 2},
							},
						},
						{DeviceIdentifier: "disk1"},
					},
				},
				utility: diskutil.DiskUtil{
					Utility: nil,
					Decoder: &diskutil.Decoder{},
				},
			},
			configure: func(utility *mock_diskutil.MockDiskUtility) {
				gomock.InOrder(
					utility.EXPECT().RepairDisk("disk0").Return("", nil),
					utility.EXPECT().ResizeContainer("disk1", "0").Return("success", nil),
					utility.EXPECT().Info("disk1").Return("", fmt.Errorf("error")),
				)
			},
			wantMessage: fmt.Sprintf("failed to fetch updated disk information for container [%s]", "disk1"),
			wantErr:     true,
			wantErrType: nil,
		},
		{
			name: "Good case: successfully resized the container and fetched the updated information",
			args: args{
				disk: &diskutil.DiskInfo{
					DeviceIdentifier: "disk1",
					APFSPhysicalStores: []diskutil.APFSPhysicalStore{
						{DeviceIdentifier: "disk0s2"},
					},
				},
				partitions: &diskutil.SystemPartitions{
					AllDisksAndPartitions: []diskutil.DiskPart{
						{
							DeviceIdentifier: "disk0",
							Size:             MinimumGrowSpaceRequired * 2,
							Partitions: []diskutil.Partition{
								{Size: MinimumGrowSpaceRequired / 2},
							},
						},
						{DeviceIdentifier: "disk1"},
					},
				},
				utility: diskutil.DiskUtil{
					Utility: nil,
					Decoder: &diskutil.Decoder{},
				},
			},
			configure: func(utility *mock_diskutil.MockDiskUtility) {
				rawInfoOutput, err := testDataFS.ReadFile(testPrefix + "good-Info.txt")
				assert.Nil(t, err)

				gomock.InOrder(
					utility.EXPECT().RepairDisk("disk0").Return("", nil),
					utility.EXPECT().ResizeContainer("disk1", "0").Return("success", nil),
					utility.EXPECT().Info("disk1").Return(string(rawInfoOutput), nil),
				)
			},
			wantMessage: fmt.Sprintf("Success, grew container [%s] to size [%s]", "disk1", humanize.Bytes(1500000)),
			wantErr:     false,
			wantErrType: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// If the test has a configure function, initialize the mock utility and configure it for the test to use
			if tt.configure != nil {
				mockUtility := mock_diskutil.NewMockDiskUtility(ctrl)
				tt.configure(mockUtility)
				tt.args.utility.Utility = mockUtility
			}

			gotMessage, err := growContainer(tt.args.disk, tt.args.partitions, tt.args.utility)
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

func Test_parentDeviceID(t *testing.T) {
	type args struct {
		disk *diskutil.DiskInfo
	}
	tests := []struct {
		name    string
		args    args
		wantId  string
		wantErr bool
	}{
		{
			name: "Bad case: no APFS physical stores",
			args: args{
				disk: &diskutil.DiskInfo{
					APFSPhysicalStores: nil,
				},
			},
			wantId:  "",
			wantErr: true,
		},
		{
			name: "Bad case: more than 1 APFS physical store",
			args: args{
				disk: &diskutil.DiskInfo{
					APFSPhysicalStores: []diskutil.APFSPhysicalStore{
						{DeviceIdentifier: "disk0s2"},
						{DeviceIdentifier: "disk1s2"},
					},
				},
			},
			wantId:  "",
			wantErr: true,
		},
		{
			name: "Bad case: APFS physical store doesn't have expected device identifier format",
			args: args{
				disk: &diskutil.DiskInfo{
					APFSPhysicalStores: []diskutil.APFSPhysicalStore{
						{DeviceIdentifier: "device0s2"},
					},
					DeviceIdentifier: "disk2",
				},
			},
			wantId:  "",
			wantErr: true,
		},
		{
			name: "Good case: one APFS physical store with expected device identifier format",
			args: args{
				disk: &diskutil.DiskInfo{
					APFSPhysicalStores: []diskutil.APFSPhysicalStore{
						{DeviceIdentifier: "disk0s2"},
					},
					DeviceIdentifier: "disk2",
				},
			},
			wantId:  "disk0",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotId, err := parentDeviceID(tt.args.disk)
			if (err != nil) != tt.wantErr {
				t.Errorf("parentDeviceID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotId != tt.wantId {
				t.Errorf("parentDeviceID() gotId = %v, want %v", gotId, tt.wantId)
			}
		})
	}
}

func Test_repairParentDisk(t *testing.T) {
	type args struct {
		disk       *diskutil.DiskInfo
		partitions *diskutil.SystemPartitions
		utility    diskutil.DiskUtil
	}
	tests := []struct {
		name        string
		args        args
		configure   func(utility *mock_diskutil.MockDiskUtility)
		wantMessage string
		wantErr     bool
		wantErrType error
	}{
		{
			name: "Bad case: error getting parent device ID from disk",
			args: args{
				disk: &diskutil.DiskInfo{
					DeviceIdentifier:   "disk0",
					APFSPhysicalStores: nil,
				},
				partitions: nil,
				utility:    diskutil.DiskUtil{},
			},
			configure:   nil,
			wantMessage: "failed to get the parent disk ID for container [disk0]",
			wantErr:     true,
			wantErrType: nil,
		},
		{
			name: "Bad case: error getting available free space for ID",
			args: args{
				disk: &diskutil.DiskInfo{DeviceIdentifier: "disk2"},
				partitions: &diskutil.SystemPartitions{
					AllDisksAndPartitions: []diskutil.DiskPart{
						{DeviceIdentifier: "disk0"},
						{DeviceIdentifier: "disk1"},
					},
				},
				utility: diskutil.DiskUtil{},
			},
			configure:   nil,
			wantMessage: "failed to get the parent disk ID for container [disk2]",
			wantErr:     true,
			wantErrType: nil,
		},
		{
			name: "Bad case: not enough space on parent disk",
			args: args{
				disk: &diskutil.DiskInfo{
					DeviceIdentifier: "disk1",
					APFSPhysicalStores: []diskutil.APFSPhysicalStore{
						{DeviceIdentifier: "disk0s2"},
					},
				},
				partitions: &diskutil.SystemPartitions{
					AllDisksAndPartitions: []diskutil.DiskPart{
						{DeviceIdentifier: "disk0"},
						{
							DeviceIdentifier: "disk1",
							Size:             MinimumGrowSpaceRequired,
							Partitions: []diskutil.Partition{
								{Size: MinimumGrowSpaceRequired / 2},
							},
						},
					},
				},
				utility: diskutil.DiskUtil{},
			},
			configure:   nil,
			wantMessage: "",
			wantErr:     true,
			wantErrType: MinimumGrowSpaceError{},
		},
		{
			name: "Bad case: error while attempting to repair the parent disk",
			args: args{
				disk: &diskutil.DiskInfo{
					DeviceIdentifier: "disk1",
					APFSPhysicalStores: []diskutil.APFSPhysicalStore{
						{DeviceIdentifier: "disk0s2"},
					},
				},
				partitions: &diskutil.SystemPartitions{
					AllDisksAndPartitions: []diskutil.DiskPart{
						{DeviceIdentifier: "disk0",
							Size: MinimumGrowSpaceRequired * 2,
							Partitions: []diskutil.Partition{
								{Size: MinimumGrowSpaceRequired / 2},
							},
						},
						{DeviceIdentifier: "disk1"},
					},
				},
				utility: diskutil.DiskUtil{
					Utility: nil,
					Decoder: &diskutil.Decoder{},
				},
			},
			configure: func(utility *mock_diskutil.MockDiskUtility) {
				utility.EXPECT().RepairDisk("disk0").Return("error", fmt.Errorf("error"))
			},
			wantMessage: "error",
			wantErr:     true,
			wantErrType: nil,
		},
		{
			name: "Good case: successfully repair the parent disk",
			args: args{
				disk: &diskutil.DiskInfo{
					DeviceIdentifier: "disk1",
					APFSPhysicalStores: []diskutil.APFSPhysicalStore{
						{DeviceIdentifier: "disk0s2"},
					},
				},
				partitions: &diskutil.SystemPartitions{
					AllDisksAndPartitions: []diskutil.DiskPart{
						{
							DeviceIdentifier: "disk0",
							Size:             MinimumGrowSpaceRequired * 2,
							Partitions: []diskutil.Partition{
								{Size: MinimumGrowSpaceRequired / 2},
							},
						},
						{DeviceIdentifier: "disk1"},
					},
				},
				utility: diskutil.DiskUtil{
					Utility: nil,
					Decoder: &diskutil.Decoder{},
				},
			},
			configure: func(utility *mock_diskutil.MockDiskUtility) {
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

			// If the test has a configure function, initialize the mock utility and configure it for the test to use
			if tt.configure != nil {
				mockUtility := mock_diskutil.NewMockDiskUtility(ctrl)
				tt.configure(mockUtility)
				tt.args.utility.Utility = mockUtility
			}

			gotMessage, err := repairParentDisk(tt.args.disk, tt.args.partitions, tt.args.utility)
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

func Test_rootContainer(t *testing.T) {
	// testPrefix is the prefix used to load test data files from testDataFS
	testPrefix := path.Join(testDataDir, "Test_rootContainer-")

	type args struct {
		utility diskutil.DiskUtil
	}
	tests := []struct {
		name          string
		args          args
		configure     func(utility *mock_diskutil.MockDiskUtility)
		wantContainer *diskutil.ContainerInfo
		wantErr       bool
	}{
		{
			name: "Bad case: utility failed to get the root container information",
			args: args{
				utility: diskutil.DiskUtil{
					Utility: nil,
					Decoder: &diskutil.Decoder{},
				},
			},
			configure: func(utility *mock_diskutil.MockDiskUtility) {
				utility.EXPECT().Info(gomock.Eq("/")).Return("", fmt.Errorf("error"))
			},
			wantContainer: nil,
			wantErr:       true,
		},
		{
			name: "Good case: utility successfully got the root container information",
			args: args{
				utility: diskutil.DiskUtil{
					Utility: nil,
					Decoder: &diskutil.Decoder{},
				},
			},
			configure: func(utility *mock_diskutil.MockDiskUtility) {
				rawInfoOutput, err := testDataFS.ReadFile(testPrefix + "good-Info.txt")
				assert.Nil(t, err)

				utility.EXPECT().Info(gomock.Eq("/")).Return(string(rawInfoOutput), nil)
			},
			wantContainer: &diskutil.ContainerInfo{
				DiskInfo: diskutil.DiskInfo{
					APFSContainerReference: "disk2",
					DeviceIdentifier:       "disk2",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// If the test has a configure function, initialize the mock utility and configure it for the test to use
			if tt.configure != nil {
				mockUtility := mock_diskutil.NewMockDiskUtility(ctrl)
				tt.configure(mockUtility)
				tt.args.utility.Utility = mockUtility
			}

			gotContainer, err := rootContainer(tt.args.utility)
			if (err != nil) != tt.wantErr {
				t.Errorf("rootContainer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.ObjectsAreEqualValues(gotContainer, tt.wantContainer) {
				t.Errorf("rootContainer() gotContainer = %v, want %v", gotContainer, tt.wantContainer)
			}
		})
	}
}

func Test_run(t *testing.T) {
	// testPrefix is the prefix used to load test data files from testDataFS
	testPrefix := path.Join(testDataDir, "Test_run-")

	type args struct {
		utility diskutil.DiskUtil
		id      string
	}
	tests := []struct {
		name      string
		args      args
		configure func(utility *mock_diskutil.MockDiskUtility)
		wantErr   bool
	}{
		{
			name: "Bad case: utility failed to list all system partitions",
			args: args{
				utility: diskutil.DiskUtil{
					Utility: nil,
					Decoder: nil,
				},
				id: "",
			},
			configure: func(utility *mock_diskutil.MockDiskUtility) {
				var args []string
				utility.EXPECT().List(args).Return("", fmt.Errorf("error"))
			},
			wantErr: true,
		},
		{
			name: "Bad case: failed to find the root container",
			args: args{
				utility: diskutil.DiskUtil{
					Utility: nil,
					Decoder: &diskutil.Decoder{},
				},
				id: "root",
			},
			configure: func(utility *mock_diskutil.MockDiskUtility) {
				var args []string
				gomock.InOrder(
					utility.EXPECT().List(args).Return("", nil),
					utility.EXPECT().Info("/").Return("", fmt.Errorf("error")),
				)
			},
			wantErr: true,
		},
		{
			name: "Bad case: failed to validate the container ID - invalid ID format",
			args: args{
				utility: diskutil.DiskUtil{
					Utility: nil,
					Decoder: &diskutil.Decoder{},
				},
				id: "/not/a/disk",
			},
			configure: func(utility *mock_diskutil.MockDiskUtility) {
				var args []string
				utility.EXPECT().List(args).Return("", nil)
			},
			wantErr: true,
		},
		{
			name: "Bad case: failed to validate the container ID - no ID found",
			args: args{
				utility: diskutil.DiskUtil{
					Utility: nil,
					Decoder: &diskutil.Decoder{},
				},
				id: "disk3",
			},
			configure: func(utility *mock_diskutil.MockDiskUtility) {
				rawListOutput, err := testDataFS.ReadFile(testPrefix + "bad-List.txt")
				assert.Nil(t, err)

				var args []string
				utility.EXPECT().List(args).Return(string(rawListOutput), nil)
			},
			wantErr: true,
		},
		{
			name: "Bad case: utility fails to get disk information",
			args: args{
				utility: diskutil.DiskUtil{
					Utility: nil,
					Decoder: &diskutil.Decoder{},
				},
				id: "disk1",
			},
			configure: func(utility *mock_diskutil.MockDiskUtility) {
				rawListOutput, err := testDataFS.ReadFile(testPrefix + "good-List.txt")
				assert.Nil(t, err)

				var args []string
				gomock.InOrder(
					utility.EXPECT().List(args).Return(string(rawListOutput), nil),
					utility.EXPECT().Info("disk1").Return("", fmt.Errorf("error")),
				)
			},
			wantErr: true,
		},
		{
			name: "Bad case: utility fails to repair the parent disk",
			args: args{
				utility: diskutil.DiskUtil{
					Utility: nil,
					Decoder: &diskutil.Decoder{},
				},
				id: "disk1",
			},
			configure: func(utility *mock_diskutil.MockDiskUtility) {
				rawListOutput, err := testDataFS.ReadFile(testPrefix + "good-List.txt")
				assert.Nil(t, err)

				rawInfoOutput, err := testDataFS.ReadFile(testPrefix + "good-Info.txt")
				assert.Nil(t, err)

				var args []string
				gomock.InOrder(
					utility.EXPECT().List(args).Return(string(rawListOutput), nil),
					utility.EXPECT().Info("disk1").Return(string(rawInfoOutput), nil),
					utility.EXPECT().RepairDisk("disk0").Return("error", fmt.Errorf("error")),
				)
			},
			wantErr: true,
		},
		{
			name: "Bad case: utility fails to resize the container",
			args: args{
				utility: diskutil.DiskUtil{
					Utility: nil,
					Decoder: &diskutil.Decoder{},
				},
				id: "disk1",
			},
			configure: func(utility *mock_diskutil.MockDiskUtility) {
				rawListOutput, err := testDataFS.ReadFile(testPrefix + "good-List.txt")
				assert.Nil(t, err)

				rawInfoOutput, err := testDataFS.ReadFile(testPrefix + "good-Info.txt")
				assert.Nil(t, err)

				var args []string
				gomock.InOrder(
					utility.EXPECT().List(args).Return(string(rawListOutput), nil),
					utility.EXPECT().Info("disk1").Return(string(rawInfoOutput), nil),
					utility.EXPECT().RepairDisk("disk0").Return("success", nil),
					utility.EXPECT().ResizeContainer("disk1", "0").Return("error", fmt.Errorf("error")),
				)
			},
			wantErr: true,
		},
		{
			name: "Good case: utility successfully resizes the container",
			args: args{
				utility: diskutil.DiskUtil{
					Utility: nil,
					Decoder: &diskutil.Decoder{},
				},
				id: "disk1",
			},
			configure: func(utility *mock_diskutil.MockDiskUtility) {
				rawListOutput, err := testDataFS.ReadFile(testPrefix + "good-List.txt")
				assert.Nil(t, err)

				rawInfoOutput, err := testDataFS.ReadFile(testPrefix + "good-Info.txt")
				assert.Nil(t, err)

				rawUpdatedInfoOutput, err := testDataFS.ReadFile(testPrefix + "good-Info-2.txt")
				assert.Nil(t, err)

				var args []string
				gomock.InOrder(
					utility.EXPECT().List(args).Return(string(rawListOutput), nil),
					utility.EXPECT().Info("disk1").Return(string(rawInfoOutput), nil),
					utility.EXPECT().RepairDisk("disk0").Return("success", nil),
					utility.EXPECT().ResizeContainer("disk1", "0").Return("success", nil),
					utility.EXPECT().Info("disk1").Return(string(rawUpdatedInfoOutput), nil),
				)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// If the test has a configure function, initialize the mock utility and configure it for the test to use
			if tt.configure != nil {
				mockUtility := mock_diskutil.NewMockDiskUtility(ctrl)
				tt.configure(mockUtility)
				tt.args.utility.Utility = mockUtility
			}

			if err := run(tt.args.utility, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateDeviceID(t *testing.T) {
	type args struct {
		id         string
		partitions *diskutil.SystemPartitions
	}
	tests := []struct {
		name      string
		args      args
		wantValid bool
		wantErr   bool
	}{
		{
			name: "Bad case: no ID provided",
			args: args{
				id:         "",
				partitions: nil,
			},
			wantValid: false,
			wantErr:   true,
		},
		{
			name: "Bad case: ID doesn't start with /dev/disk or disk",
			args: args{
				id:         "bad",
				partitions: nil,
			},
			wantValid: false,
			wantErr:   true,
		},
		{
			name: "Bad case: ID doesn't have disk number",
			args: args{
				id:         "disk",
				partitions: nil,
			},
			wantValid: false,
			wantErr:   true,
		},
		{
			name: "Bad case: ID is of expected format but isn't in system partitions",
			args: args{
				id: "disk2",
				partitions: &diskutil.SystemPartitions{
					AllDisks: []string{"disk0", "disk1"},
				},
			},
			wantValid: false,
			wantErr:   false,
		},
		{
			name: "Good case: ID is expected format and is in system partitions",
			args: args{
				id: "disk0",
				partitions: &diskutil.SystemPartitions{
					AllDisks: []string{"disk0", "disk1"},
				},
			},
			wantValid: true,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValid, err := validateDeviceID(tt.args.id, tt.args.partitions)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDeviceID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotValid != tt.wantValid {
				t.Errorf("validateDeviceID() gotValid = %v, want %v", gotValid, tt.wantValid)
			}
		})
	}
}
