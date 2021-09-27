package diskutil

import (
	"bytes"
	"embed"
	"io"
	"path"
	"strings"
	"testing"

	"github.com/aws/ec2-macos-utils/pkg/diskutil/types"

	"github.com/stretchr/testify/assert"
)

//go:embed testdata/decoder
var decoderDataFS embed.FS

const decoderDataDir = "testdata/decoder"

func TestPlistDecoder_DecodeInfo(t *testing.T) {
	// testPrefix is the prefix used to load test data files from testDataFS
	testPrefix := path.Join(decoderDataDir, "TestPlistDecoder_DecodeInfo-")

	type args struct {
		rawDisk string
	}
	tests := []struct {
		name         string
		args         args
		useArgs      bool
		testFileName string
		wantDisk     *types.DiskInfo
		wantErr      bool
	}{
		{
			name:     "Good case: empty input",
			args:     args{rawDisk: ""},
			useArgs:  true,
			wantDisk: &types.DiskInfo{},
			wantErr:  false,
		},
		{
			name:     "Bad case: bad input",
			args:     args{rawDisk: "abcdefghijklmnopqrstuvwxyz"},
			useArgs:  true,
			wantDisk: nil,
			wantErr:  true,
		},
		{
			name:         "Bad case: improperly formatted disk input (missing header)",
			useArgs:      false,
			testFileName: testPrefix + "bad-Info.txt",
			wantDisk:     nil,
			wantErr:      true,
		},
		{
			name:         "Good case: properly formatted (sparse) disk input",
			useArgs:      false,
			testFileName: testPrefix + "good-Info.txt",
			wantDisk: &types.DiskInfo{
				AESHardware:            false,
				APFSContainerReference: "disk2",
				APFSPhysicalStores:     []types.APFSPhysicalStore{{DeviceIdentifier: "disk0s2"}},
				SMARTDeviceSpecificKeysMayVaryNotGuaranteed: &types.SmartDeviceInfo{AvailableSpare: 100},
			},
			wantErr: false,
		},
		{
			name:         "Bad case: improperly formatted container input (missing header)",
			useArgs:      false,
			testFileName: testPrefix + "bad-Container.txt",
			wantDisk:     nil,
			wantErr:      true,
		},
		{
			name:         "Good case: properly formatted (sparse) container input",
			useArgs:      false,
			testFileName: testPrefix + "good-Container.txt",
			wantDisk: &types.DiskInfo{
				ContainerInfo: types.ContainerInfo{
					APFSContainerFree: 41477054464,
					APFSContainerSize: 64214753280,
				},
				AESHardware:            false,
				APFSContainerReference: "disk2",
				APFSPhysicalStores:     []types.APFSPhysicalStore{{DeviceIdentifier: "disk0s2"}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &PlistDecoder{}

			var input io.ReadSeeker

			// Check to use test's args or read embedded data
			if tt.useArgs {
				input = strings.NewReader(tt.args.rawDisk)
			} else {
				read, err := decoderDataFS.ReadFile(tt.testFileName)
				assert.Nil(t, err)

				input = bytes.NewReader(read)
			}

			gotDisk, err := d.DecodeDiskInfo(input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeDiskInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.ObjectsAreEqualValues(gotDisk, tt.wantDisk) {
				t.Errorf("DecodeDiskInfo() gotDisk = %v, want %v", gotDisk, tt.wantDisk)
			}
		})
	}
}

func TestPlistDecoder_DecodeList(t *testing.T) {
	// testPrefix is the prefix used to load test data files from testDataFS
	testPrefix := path.Join(decoderDataDir, "TestPlistDecoder_DecodeList-")

	type args struct {
		rawList string
	}
	tests := []struct {
		name           string
		args           args
		useArgs        bool
		testFileName   string
		wantPartitions *types.SystemPartitions
		wantErr        bool
	}{
		{
			name:           "Bad case: improperly formatted input (missing header)",
			useArgs:        false,
			testFileName:   testPrefix + "bad-List.txt",
			wantPartitions: nil,
			wantErr:        true,
		},
		{
			name:         "Good case: properly formatted (sparse) input",
			useArgs:      false,
			testFileName: testPrefix + "good-List.txt",
			wantPartitions: &types.SystemPartitions{
				AllDisks: []string{"disk0"},
				AllDisksAndPartitions: []types.DiskPart{
					{
						DeviceIdentifier: "disk0",
						Partitions: []types.Partition{
							{DeviceIdentifier: "disk0s1"},
						},
						Size: 121332826112,
					},
					{
						APFSPhysicalStores: []types.APFSPhysicalStoreID{
							{DeviceIdentifier: "disk0s2"},
						},
						APFSVolumes: []types.APFSVolume{
							{
								DeviceIdentifier: "disk2s4",
								MountedSnapshots: []types.Snapshot{
									{SnapshotUUID: "7CA60DB3-9063-4559-BC98-5BC05599DCF1"},
								},
							},
						},
					},
				},
				VolumesFromDisks: []string{"Macintosh HD - Data"},
				WholeDisks:       []string{"disk0"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &PlistDecoder{}

			var input io.ReadSeeker

			// Check to use test's args or read embedded data
			if tt.useArgs {
				input = strings.NewReader(tt.args.rawList)
			} else {
				read, err := decoderDataFS.ReadFile(tt.testFileName)
				assert.Nil(t, err)

				input = bytes.NewReader(read)
			}

			gotPartitions, err := d.DecodeSystemPartitions(input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeSystemPartitions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.ObjectsAreEqualValues(gotPartitions, tt.wantPartitions) {
				t.Errorf("DecodeSystemPartitions() gotPartitions = %v, want %v", gotPartitions, tt.wantPartitions)
			}
		})
	}
}
