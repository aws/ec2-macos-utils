package diskutil

import (
	"embed"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed testdata/decoder
var testDataFS embed.FS

const testDataDir = "testdata/decoder"

func TestPlistDecoder_DecodeInfo(t *testing.T) {
	// testPrefix is the prefix used to load test data files from testDataFS
	testPrefix := path.Join(testDataDir, "TestPlistDecoder_DecodeInfo-")

	type args struct {
		rawDisk string
	}
	tests := []struct {
		name         string
		args         args
		useArgs      bool
		testFileName string
		wantDisk     *DiskInfo
		wantErr      bool
	}{
		{
			name:     "Good case: empty input",
			args:     args{rawDisk: ""},
			useArgs:  true,
			wantDisk: &DiskInfo{},
			wantErr:  false,
		},
		{
			name:     "Bad case: garbage input",
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
			wantDisk: &DiskInfo{
				AESHardware:            false,
				APFSContainerReference: "disk2",
				APFSPhysicalStores:     []APFSPhysicalStore{{DeviceIdentifier: "disk0s2"}},
				SMARTDeviceSpecificKeysMayVaryNotGuaranteed: &SmartDeviceInfo{AvailableSpare: 100},
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
			wantDisk: &DiskInfo{
				ContainerInfo: ContainerInfo{
					APFSContainerFree: 41477054464,
					APFSContainerSize: 64214753280,
				},
				AESHardware:            false,
				APFSContainerReference: "disk2",
				APFSPhysicalStores:     []APFSPhysicalStore{{DeviceIdentifier: "disk0s2"}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &PlistDecoder{}

			var input string

			// Check to use test's args or read embedded data
			if tt.useArgs {
				input = tt.args.rawDisk
			} else {
				read, err := testDataFS.ReadFile(tt.testFileName)
				assert.Nil(t, err)

				input = string(read)
			}

			gotDisk, err := d.DecodeInfo(input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.ObjectsAreEqualValues(gotDisk, tt.wantDisk) {
				t.Errorf("DecodeInfo() gotDisk = %v, want %v", gotDisk, tt.wantDisk)
			}
		})
	}
}

func TestPlistDecoder_DecodeList(t *testing.T) {
	// testPrefix is the prefix used to load test data files from testDataFS
	testPrefix := path.Join(testDataDir, "TestPlistDecoder_DecodeList-")

	type args struct {
		rawList string
	}
	tests := []struct {
		name           string
		args           args
		useArgs        bool
		testFileName   string
		wantPartitions *SystemPartitions
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
			wantPartitions: &SystemPartitions{
				AllDisks: []string{"disk0"},
				AllDisksAndPartitions: []DiskPart{
					{
						DeviceIdentifier: "disk0",
						Partitions: []Partition{
							{DeviceIdentifier: "disk0s1"},
						},
						Size: 121332826112,
					},
					{
						APFSPhysicalStores: []APFSPhysicalStoreID{
							{DeviceIdentifier: "disk0s2"},
						},
						APFSVolumes: []APFSVolume{
							{
								DeviceIdentifier: "disk2s4",
								MountedSnapshots: []Snapshot{
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

			var input string

			// Check to use test's args or read embedded data
			if tt.useArgs {
				input = tt.args.rawList
			} else {
				read, err := testDataFS.ReadFile(tt.testFileName)
				assert.Nil(t, err)

				input = string(read)
			}

			gotPartitions, err := d.DecodeList(input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.ObjectsAreEqualValues(gotPartitions, tt.wantPartitions) {
				t.Errorf("DecodeList() gotPartitions = %v, want %v", gotPartitions, tt.wantPartitions)
			}
		})
	}
}
