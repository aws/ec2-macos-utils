package diskutil

import (
	"reflect"
	"testing"
)

type args struct {
	input string
}

var (
	// containerTests is the test data used in the test functions TestDecodeContainer and TestDecoder_DecodeContainer.
	containerTests = []struct {
		name          string
		args          args
		wantContainer *ContainerInfo
		wantErr       bool
	}{
		{
			name:          "Bad case: empty input",
			args:          args{input: ""},
			wantContainer: &ContainerInfo{},
			wantErr:       false,
		},
		{
			name:          "Bad case: garbage input",
			args:          args{input: "abcdefghijklmnopqrstuvwxyz"},
			wantContainer: nil,
			wantErr:       true,
		},
		{
			name: "Bad case: improperly formatted input (missing header)",
			args: args{
				input: "\t<key>AESHardware</key>\n" +
					"\t<false/>\n" +
					"\t<key>APFSContainerFree</key>\n" +
					"\t<integer>41477054464</integer>\n" +
					"\t<key>APFSContainerReference</key>\n" +
					"\t<string>disk2</string>\n" +
					"\t<key>APFSContainerSize</key>\n" +
					"\t<integer>64214753280</integer>\n" +
					"\t<key>APFSPhysicalStores</key>\n" +
					"\t<array>\n" +
					"\t\t<dict>\n" +
					"\t\t\t<key>APFSPhysicalStore</key>\n" +
					"\t\t\t<string>disk0s2</string>\n" +
					"\t\t</dict>\n" +
					"\t</array>\n" +
					"</dict>\n" +
					"</plist>",
			},
			wantContainer: nil,
			wantErr:       true,
		},
		{
			name: "Good case: properly formatted (sparse) input",
			args: args{
				input: "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" +
					"<!DOCTYPE plist PUBLIC \"-//Apple//DTD PLIST 1.0//EN\" \"http://www.apple.com/DTDs/PropertyList-1.0.dtd\">\n" +
					"<plist version=\"1.0\">\n" +
					"<dict>\n" +
					"\t<key>AESHardware</key>\n" +
					"\t<false/>\n" +
					"\t<key>APFSContainerFree</key>\n" +
					"\t<integer>41477054464</integer>\n" +
					"\t<key>APFSContainerReference</key>\n" +
					"\t<string>disk2</string>\n" +
					"\t<key>APFSContainerSize</key>\n" +
					"\t<integer>64214753280</integer>\n" +
					"\t<key>APFSPhysicalStores</key>\n" +
					"\t<array>\n" +
					"\t\t<dict>\n" +
					"\t\t\t<key>APFSPhysicalStore</key>\n" +
					"\t\t\t<string>disk0s2</string>\n" +
					"\t\t</dict>\n" +
					"\t</array>\n" +
					"</dict>\n" +
					"</plist>",
			},
			wantContainer: &ContainerInfo{
				DiskInfo: DiskInfo{
					AESHardware:            false,
					APFSContainerReference: "disk2",
					APFSPhysicalStores:     []APFSPhysicalStore{{DeviceIdentifier: "disk0s2"}},
				},
				APFSContainerFree: 41477054464,
				APFSContainerSize: 64214753280,
			},
			wantErr: false,
		},
	}

	// diskTests is the test data used in the test functions TestDecodeDisk and TestDecoder_DecodeDisk.
	diskTests = []struct {
		name     string
		args     args
		wantDisk *DiskInfo
		wantErr  bool
	}{
		{
			name:     "Good case: empty input",
			args:     args{input: ""},
			wantDisk: &DiskInfo{},
			wantErr:  false,
		},
		{
			name:     "Bad case: garbage input",
			args:     args{input: "abcdefghijklmnopqrstuvwxyz"},
			wantDisk: nil,
			wantErr:  true,
		},
		{
			name: "Bad case: improperly formatted input (missing header)",
			args: args{
				input: "\t<key>AESHardware</key>\n" +
					"\t<false/>\n" +
					"\t<key>APFSContainerReference</key>\n" +
					"\t<string>disk2</string>\n" +
					"\t<key>APFSPhysicalStores</key>\n" +
					"\t<array>\n" +
					"\t\t<dict>\n" +
					"\t\t\t<key>APFSPhysicalStore</key>\n" +
					"\t\t\t<string>disk0s2\n" +
					"\t\t</dict>\n" +
					"\t</array>\n" +
					"\t<key>SMARTDeviceSpecificKeysMayVaryNotGuaranteed</key>\n" +
					"\t<dict>\n" +
					"\t\t<key>AVAILABLE_SPARE</key>\n" +
					"\t\t<integer>100</integer>\n" +
					"\t</dict>\n" +
					"</dict>\n" +
					"</plist>",
			},
			wantDisk: nil,
			wantErr:  true,
		},
		{
			name: "Good case: properly formatted (sparse) input",
			args: args{
				input: "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" +
					"<!DOCTYPE plist PUBLIC \"-//Apple//DTD PLIST 1.0//EN\" \"http://www.apple.com/DTDs/PropertyList-1.0.dtd\">\n" +
					"<plist version=\"1.0\">\n" +
					"<dict>\n" +
					"\t<key>AESHardware</key>\n" +
					"\t<false/>\n" +
					"\t<key>APFSContainerReference</key>\n" +
					"\t<string>disk2</string>\n" +
					"\t<key>APFSPhysicalStores</key>\n" +
					"\t<array>\n" +
					"\t\t<dict>\n" +
					"\t\t\t<key>APFSPhysicalStore</key>\n" +
					"\t\t\t<string>disk0s2</string>\n" +
					"\t\t</dict>\n" +
					"\t</array>\n" +
					"\t<key>SMARTDeviceSpecificKeysMayVaryNotGuaranteed</key>\n" +
					"\t<dict>\n" +
					"\t\t<key>AVAILABLE_SPARE</key>\n" +
					"\t\t<integer>100</integer>\n" +
					"\t</dict>\n" +
					"</dict>\n" +
					"</plist>",
			},
			wantDisk: &DiskInfo{
				AESHardware:            false,
				APFSContainerReference: "disk2",
				APFSPhysicalStores:     []APFSPhysicalStore{{DeviceIdentifier: "disk0s2"}},
				SMARTDeviceSpecificKeysMayVaryNotGuaranteed: &SmartDeviceInfo{AvailableSpare: 100},
			},
			wantErr: false,
		},
	}

	// listTests is the test data used in the test functions TestDecodeList and TestDecoder_DecodeList.
	listTests = []struct {
		name           string
		args           args
		wantPartitions *SystemPartitions
		wantErr        bool
	}{
		{
			name: "Bad case: improperly formatted input (missing header)",
			args: args{
				input: "\t<key>AllDisks</key>\n" +
					"\t<array>\n" +
					"\t\t<string>disk0</string>\n" +
					"\t</array>\n" +
					"\t<key>AllDisksAndPartitions</key>\n" +
					"\t<array>\n" +
					"\t\t<dict>\n" +
					"\t\t\t<key>DeviceIdentifier</key>\n" +
					"\t\t\t<string>disk0</string>\n" +
					"\t\t\t<key>Partitions</key>\n" +
					"\t\t\t<array>\n" +
					"\t\t\t\t<dict>\n" +
					"\t\t\t\t\t<key>DeviceIdentifier</key>\n" +
					"\t\t\t\t\t<string>disk0s1</string>\n" +
					"\t\t\t\t</dict>\n" +
					"\t\t\t\t<dict>\n" +
					"\t\t\t\t\t<key>DeviceIdentifier</key>\n" +
					"\t\t\t\t\t<string>disk0s2</string>\n" +
					"\t\t\t\t</dict>\n" +
					"\t\t\t</array>\n" +
					"\t\t\t<key>Size</key>\n" +
					"\t\t\t<integer>121332826112</integer>\n" +
					"\t\t</dict>\n" +
					"\t\t<dict>\n" +
					"\t\t\t<key>APFSPhysicalStores</key>\n" +
					"\t\t\t<array>\n" +
					"\t\t\t\t<dict>\n" +
					"\t\t\t\t\t<key>DeviceIdentifier</key>\n" +
					"\t\t\t\t\t<string>disk0s2</string>\n" +
					"\t\t\t\t</dict>\n" +
					"\t\t\t</array>\n" +
					"\t\t\t<key>APFSVolumes</key>\n" +
					"\t\t\t<array>\n" +
					"\t\t\t\t</dict>\n" +
					"\t\t\t\t<dict>\n" +
					"\t\t\t\t\t<key>DeviceIdentifier</key>\n" +
					"\t\t\t\t\t<string>disk2s4</string>\n" +
					"\t\t\t\t\t<key>MountedSnapshots</key>\n" +
					"\t\t\t\t\t<array>\n" +
					"\t\t\t\t\t\t<dict>\n" +
					"\t\t\t\t\t\t\t<key>SnapshotUUID</key>\n" +
					"\t\t\t\t\t\t\t<string>7CA60DB3-9063-4559-BC98-5BC05599DCF1</string>\n" +
					"\t\t\t\t\t\t</dict>\n" +
					"\t\t\t\t\t</array>\n" +
					"\t\t\t\t</dict>\n" +
					"\t\t\t</array>\n" +
					"\t\t</dict>\n" +
					"\t</array>\n" +
					"\t<key>VolumesFromDisks</key>\n" +
					"\t<array>\n" +
					"\t\t<string>Macintosh HD - Data</string>\n" +
					"\t</array>\n" +
					"\t<key>WholeDisks</key>\n" +
					"\t<array>\n" +
					"\t\t<string>disk0</string>\n" +
					"\t</array>\n" +
					"</dict>\n" +
					"</plist>",
			},
			wantPartitions: nil,
			wantErr:        true,
		},
		{
			name: "Good case: properly formatted (sparse) input",
			args: args{
				input: "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" +
					"<!DOCTYPE plist PUBLIC \"-//Apple//DTD PLIST 1.0//EN\" \"http://www.apple.com/DTDs/PropertyList-1.0.dtd\">\n" +
					"<plist version=\"1.0\">\n" +
					"<dict>\n" +
					"\t<key>AllDisks</key>\n" +
					"\t<array>\n" +
					"\t\t<string>disk0</string>\n" +
					"\t</array>\n" +
					"\t<key>AllDisksAndPartitions</key>\n" +
					"\t<array>\n" +
					"\t\t<dict>\n" +
					"\t\t\t<key>DeviceIdentifier</key>\n" +
					"\t\t\t<string>disk0</string>\n" +
					"\t\t\t<key>Partitions</key>\n" +
					"\t\t\t<array>\n" +
					"\t\t\t\t<dict>\n" +
					"\t\t\t\t\t<key>DeviceIdentifier</key>\n" +
					"\t\t\t\t\t<string>disk0s1</string>\n" +
					"\t\t\t\t</dict>\n" +
					"\t\t\t</array>\n" +
					"\t\t\t<key>Size</key>\n" +
					"\t\t\t<integer>121332826112</integer>\n" +
					"\t\t</dict>\n" +
					"\t\t<dict>\n" +
					"\t\t\t<key>APFSPhysicalStores</key>\n" +
					"\t\t\t<array>\n" +
					"\t\t\t\t<dict>\n" +
					"\t\t\t\t\t<key>DeviceIdentifier</key>\n" +
					"\t\t\t\t\t<string>disk0s2</string>\n" +
					"\t\t\t\t</dict>\n" +
					"\t\t\t</array>\n" +
					"\t\t\t<key>APFSVolumes</key>\n" +
					"\t\t\t<array>\n" +
					"\t\t\t\t<dict>\n" +
					"\t\t\t\t\t<key>DeviceIdentifier</key>\n" +
					"\t\t\t\t\t<string>disk2s4</string>\n" +
					"\t\t\t\t\t<key>MountedSnapshots</key>\n" +
					"\t\t\t\t\t<array>\n" +
					"\t\t\t\t\t\t<dict>\n" +
					"\t\t\t\t\t\t\t<key>SnapshotUUID</key>\n" +
					"\t\t\t\t\t\t\t<string>7CA60DB3-9063-4559-BC98-5BC05599DCF1</string>\n" +
					"\t\t\t\t\t\t</dict>\n" +
					"\t\t\t\t\t</array>\n" +
					"\t\t\t\t</dict>\n" +
					"\t\t\t</array>\n" +
					"\t\t</dict>\n" +
					"\t</array>\n" +
					"\t<key>VolumesFromDisks</key>\n" +
					"\t<array>\n" +
					"\t\t<string>Macintosh HD - Data</string>\n" +
					"\t</array>\n" +
					"\t<key>WholeDisks</key>\n" +
					"\t<array>\n" +
					"\t\t<string>disk0</string>\n" +
					"\t</array>\n" +
					"</dict>\n" +
					"</plist>",
			},
			wantPartitions: &SystemPartitions{
				AllDisks: []string{"disk0"},
				AllDisksAndPartitions: []DiskPart{
					{
						DeviceIdentifier: "disk0",
						Partitions: []Partition{
							{
								DeviceIdentifier: "disk0s1",
							},
						},
						Size: 121332826112,
					},
					{
						APFSPhysicalStores: []APFSPhysicalStoreID{
							{
								DeviceIdentifier: "disk0s2",
							},
						},
						APFSVolumes: []APFSVolume{
							{
								DeviceIdentifier: "disk2s4",
								MountedSnapshots: []Snapshot{
									{
										SnapshotUUID: "7CA60DB3-9063-4559-BC98-5BC05599DCF1",
									},
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
)

func TestDecoder_DecodeContainer(t *testing.T) {
	for _, tt := range containerTests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Decoder{}
			gotContainer, err := p.DecodeContainer(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeContainer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotContainer, tt.wantContainer) {
				t.Errorf("DecodeContainer() gotContainer = %#v, want %#v", gotContainer, tt.wantContainer)
			}
		})
	}
}

func TestDecoder_DecodeDisk(t *testing.T) {
	for _, tt := range diskTests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Decoder{}
			gotDisk, err := p.DecodeDisk(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeDisk() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotDisk, tt.wantDisk) {
				t.Errorf("DecodeDisk() gotDisk = %#v, want %#v", gotDisk, tt.wantDisk)
			}
		})
	}
}

func TestDecoder_DecodeList(t *testing.T) {
	for _, tt := range listTests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Decoder{}
			gotPartitions, err := p.DecodeList(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPartitions, tt.wantPartitions) {
				t.Errorf("DecodeList() gotPartitions = %#v, want %#v", gotPartitions, tt.wantPartitions)
			}
		})
	}
}
