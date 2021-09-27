package cmd

import (
	"bytes"
	"embed"
	"fmt"
	"path"
	"testing"

	"github.com/aws/ec2-macos-utils/pkg/diskutil"
	mock_diskutil "github.com/aws/ec2-macos-utils/pkg/diskutil/mocks"
	"github.com/aws/ec2-macos-utils/pkg/diskutil/types"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

//go:embed testdata/growContainer
var testDataFS embed.FS

const testDataDir = "testdata/growContainer"

func Test_rootContainer(t *testing.T) {
	// testPrefix is the prefix used to load test data files from testDataFS
	testPrefix := path.Join(testDataDir, "Test_rootContainer-")

	tests := []struct {
		name          string
		configure     func(utility *mock_diskutil.MockDiskUtil)
		wantContainer *types.DiskInfo
		wantErr       bool
	}{
		{
			name: "Bad case: utility failed to get the root container information",
			configure: func(utility *mock_diskutil.MockDiskUtil) {
				utility.EXPECT().Info(gomock.Eq("/")).Return(nil, fmt.Errorf("error"))
			},
			wantContainer: nil,
			wantErr:       true,
		},
		{
			name: "Good case: utility successfully got the root container information",
			configure: func(utility *mock_diskutil.MockDiskUtil) {
				rawInfoOutput, err := testDataFS.ReadFile(testPrefix + "good-Info.txt")
				assert.Nil(t, err)

				decoder := &diskutil.PlistDecoder{}
				info, err := decoder.DecodeDiskInfo(bytes.NewReader(rawInfoOutput))
				assert.Nil(t, err)

				utility.EXPECT().Info(gomock.Eq("/")).Return(info, nil)
			},
			wantContainer: &types.DiskInfo{
				APFSContainerReference: "disk2",
				DeviceIdentifier:       "disk2",
			},
			wantErr: false,
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

			gotContainer, err := rootContainer(mockUtility)
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
		invo growContainer
	}
	tests := []struct {
		name      string
		args      args
		configure func(utility *mock_diskutil.MockDiskUtil)
		wantErr   bool
	}{
		{
			name: "Bad case: utility failed to list all system partitions",
			args: args{
				invo: growContainer{id: ""},
			},
			configure: func(utility *mock_diskutil.MockDiskUtil) {
				var args []string
				utility.EXPECT().List(args).Return(nil, fmt.Errorf("error"))
			},
			wantErr: true,
		},
		{
			name: "Bad case: failed to find the root container",
			args: args{
				growContainer{id: "root"},
			},
			configure: func(utility *mock_diskutil.MockDiskUtil) {
				var args []string
				gomock.InOrder(
					utility.EXPECT().List(args).Return(&types.SystemPartitions{}, nil),
					utility.EXPECT().Info("/").Return(nil, fmt.Errorf("error")),
				)
			},
			wantErr: true,
		},
		{
			name: "Bad case: failed to validate the container ID - invalid ID format",
			args: args{
				growContainer{id: "/not/a/disk"},
			},
			configure: func(utility *mock_diskutil.MockDiskUtil) {
				var args []string
				utility.EXPECT().List(args).Return(&types.SystemPartitions{}, nil)
			},
			wantErr: true,
		},
		{
			name: "Bad case: failed to validate the container ID - no ID found",
			args: args{
				growContainer{id: "disk3"},
			},
			configure: func(utility *mock_diskutil.MockDiskUtil) {
				rawListOutput, err := testDataFS.ReadFile(testPrefix + "bad-List.txt")
				assert.Nil(t, err)

				decoder := &diskutil.PlistDecoder{}
				list, err := decoder.DecodeSystemPartitions(bytes.NewReader(rawListOutput))
				assert.Nil(t, err)

				var args []string
				utility.EXPECT().List(args).Return(list, nil)
			},
			wantErr: true,
		},
		{
			name: "Bad case: utility fails to get disk information",
			args: args{
				growContainer{id: "disk1"},
			},
			configure: func(utility *mock_diskutil.MockDiskUtil) {
				rawListOutput, err := testDataFS.ReadFile(testPrefix + "good-List.txt")
				assert.Nil(t, err)

				decoder := &diskutil.PlistDecoder{}
				list, err := decoder.DecodeSystemPartitions(bytes.NewReader(rawListOutput))
				assert.Nil(t, err)

				var args []string
				gomock.InOrder(
					utility.EXPECT().List(args).Return(list, nil),
					utility.EXPECT().Info("disk1").Return(nil, fmt.Errorf("error")),
				)
			},
			wantErr: true,
		},
		{
			name: "Bad case: utility fails to resize the container",
			args: args{
				growContainer{id: "disk1"},
			},
			configure: func(utility *mock_diskutil.MockDiskUtil) {
				rawListOutput, err := testDataFS.ReadFile(testPrefix + "good-List.txt")
				assert.Nil(t, err)

				rawInfoOutput, err := testDataFS.ReadFile(testPrefix + "good-Info.txt")
				assert.Nil(t, err)

				decoder := &diskutil.PlistDecoder{}
				list, err := decoder.DecodeSystemPartitions(bytes.NewReader(rawListOutput))
				assert.Nil(t, err)

				info, err := decoder.DecodeDiskInfo(bytes.NewReader(rawInfoOutput))
				assert.Nil(t, err)

				var args []string
				gomock.InOrder(
					utility.EXPECT().List(args).Return(list, nil),
					utility.EXPECT().Info("disk1").Return(info, nil),
					utility.EXPECT().RepairDisk("disk0").Return("success", nil),
					utility.EXPECT().ResizeContainer("disk1", "0").Return("error", fmt.Errorf("error")),
				)
			},
			wantErr: true,
		},
		{
			name: "Good case: utility successfully resizes the container",
			args: args{
				growContainer{id: "disk1"},
			},
			configure: func(utility *mock_diskutil.MockDiskUtil) {
				rawListOutput, err := testDataFS.ReadFile(testPrefix + "good-List.txt")
				assert.Nil(t, err)

				rawInfoOutput, err := testDataFS.ReadFile(testPrefix + "good-Info.txt")
				assert.Nil(t, err)

				rawUpdatedInfoOutput, err := testDataFS.ReadFile(testPrefix + "good-Info-2.txt")
				assert.Nil(t, err)

				decoder := &diskutil.PlistDecoder{}
				list, err := decoder.DecodeSystemPartitions(bytes.NewReader(rawListOutput))
				assert.Nil(t, err)

				info, err := decoder.DecodeDiskInfo(bytes.NewReader(rawInfoOutput))
				assert.Nil(t, err)

				updatedInfo, err := decoder.DecodeDiskInfo(bytes.NewReader(rawUpdatedInfoOutput))
				assert.Nil(t, err)

				var args []string
				gomock.InOrder(
					utility.EXPECT().List(args).Return(list, nil),
					utility.EXPECT().Info("disk1").Return(info, nil),
					utility.EXPECT().RepairDisk("disk0").Return("success", nil),
					utility.EXPECT().ResizeContainer("disk1", "0").Return("success", nil),
					utility.EXPECT().Info("disk1").Return(updatedInfo, nil),
				)
			},
			wantErr: false,
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

			if err := run(mockUtility, tt.args.invo); (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateDeviceID(t *testing.T) {
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
				partitions: &types.SystemPartitions{
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
				partitions: &types.SystemPartitions{
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
