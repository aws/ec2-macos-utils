package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiskInfo_parentDeviceID(t *testing.T) {
	type args struct {
		disk *DiskInfo
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
				disk: &DiskInfo{
					APFSPhysicalStores: nil,
				},
			},
			wantId:  "",
			wantErr: true,
		},
		{
			name: "Bad case: more than 1 APFS physical store",
			args: args{
				disk: &DiskInfo{
					APFSPhysicalStores: []APFSPhysicalStore{
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
				disk: &DiskInfo{
					APFSPhysicalStores: []APFSPhysicalStore{
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
				disk: &DiskInfo{
					APFSPhysicalStores: []APFSPhysicalStore{
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
			gotId, err := tt.args.disk.ParentDeviceID()

			assert.Equal(t, tt.wantId, gotId, "should have matching parent device ID")

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
