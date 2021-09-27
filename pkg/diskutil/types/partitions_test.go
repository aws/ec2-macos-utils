package types

import "testing"

func TestSystemPartitions_availableDiskSpace(t *testing.T) {
	type fields struct {
		AllDisks              []string
		AllDisksAndPartitions []DiskPart
		VolumesFromDisks      []string
		WholeDisks            []string
	}
	type args struct {
		id string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantSize uint64
		wantErr  bool
	}{
		{
			name: "Bad case: ID not found in system partitions",
			fields: fields{
				AllDisks: nil,
				AllDisksAndPartitions: []DiskPart{
					{DeviceIdentifier: "disk0"},
					{DeviceIdentifier: "disk1"},
					{DeviceIdentifier: "disk2"},
				},
				VolumesFromDisks: nil,
				WholeDisks:       nil,
			},
			args: args{
				id: "disk3",
			},
			wantSize: 0,
			wantErr:  true,
		},
		{
			name: "Good case: ID found in system partitions and size matches",
			fields: fields{
				AllDisks: nil,
				AllDisksAndPartitions: []DiskPart{
					{DeviceIdentifier: "disk0"},
					{
						DeviceIdentifier: "disk1",
						Size:             2000000,
						Partitions: []Partition{
							{Size: 500000},
							{Size: 500000},
						},
					},
				},
				VolumesFromDisks: nil,
				WholeDisks:       nil,
			},
			args: args{
				id: "disk1",
			},
			wantSize: 1000000,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &SystemPartitions{
				AllDisks:              tt.fields.AllDisks,
				AllDisksAndPartitions: tt.fields.AllDisksAndPartitions,
				VolumesFromDisks:      tt.fields.VolumesFromDisks,
				WholeDisks:            tt.fields.WholeDisks,
			}
			gotSize, err := p.AvailableDiskSpace(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("AvailableDiskSpace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotSize != tt.wantSize {
				t.Errorf("AvailableDiskSpace() gotSize = %v, want %v", gotSize, tt.wantSize)
			}
		})
	}
}
