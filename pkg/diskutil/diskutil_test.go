package diskutil

import (
	"fmt"
	"testing"

	"github.com/dustin/go-humanize"
)

func TestMinimumGrowSpaceError_Error(t *testing.T) {
	type fields struct {
		size uint64
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Good case: returns the expected string",
			fields: fields{
				size: 0,
			},
			want: fmt.Sprintf("grow requires [%s] but got [%s]", humanize.Bytes(MinimumGrowSpaceRequired), humanize.Bytes(0)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := MinimumGrowSpaceError{
				size: tt.fields.size,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
