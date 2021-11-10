package identifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDiskID(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "with empty input",
			args: args{
				s: "",
			},
			want: "",
		},
		{
			name: "without device id",
			args: args{
				s: "this is not a device identifier",
			},
			want: "",
		},
		{
			name: "with device id",
			args: args{
				s: "disk1",
			},
			want: "disk1",
		},
		{
			name: "with full device id",
			args: args{
				s: "/dev/disk1",
			},
			want: "disk1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseDiskID(tt.args.s)

			assert.Equal(t, tt.want, got, "parsed id should match expected")
		})
	}
}
