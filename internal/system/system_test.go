package system

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseIOPlatformUUID(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    string
		wantErr bool
	}{
		{
			name: "valid UUID",
			input: []byte(`
                "IOPlatformUUID" = "ABCD1234-5678-90EF-GHIJ-KLMNOPQRSTUV"
            `),
			want:    "ABCD1234-5678-90EF-GHIJ-KLMNOPQRSTUV",
			wantErr: false,
		},
		{
			name:    "empty input",
			input:   []byte(``),
			want:    "",
			wantErr: true,
		},
		{
			name: "malformed line - missing equals",
			input: []byte(`
                "IOPlatformUUID" "ABCD1234-5678-90EF-GHIJ-KLMNOPQRSTUV"
            `),
			want:    "",
			wantErr: true,
		},
		{
			name: "malformed line - extra fields",
			input: []byte(`
                "IOPlatformUUID" = "ABCD1234-5678-90EF-GHIJ-KLMNOPQRSTUV" extra
            `),
			want:    "",
			wantErr: true,
		},
		{
			name: "empty UUID value",
			input: []byte(`
                "IOPlatformUUID" = ""
            `),
			want:    "",
			wantErr: true,
		},
		{
			name: "wrong key",
			input: []byte(`
                "WrongKey" = "ABCD1234-5678-90EF-GHIJ-KLMNOPQRSTUV"
            `),
			want:    "",
			wantErr: true,
		},
		{
			name: "actual ioreg format",
			input: []byte(`+-o Root  <class IORegistryEntry, id 0x100000100, retain 12>
                +-o IOPlatformExpertDevice  <class IOPlatformExpertDevice, id 0x100000110, registered, matched, active, busy 0 (0 ms), retain 35>
                {
                    "IOPlatformUUID" = "ABCD1234-5678-90EF-GHIJ-KLMNOPQRSTUV"
                }
            `),
			want:    "ABCD1234-5678-90EF-GHIJ-KLMNOPQRSTUV",
			wantErr: false,
		},
		{
			name: "ioreg format with multiple entries",
			input: []byte(`+-o J314sAP  <class IOPlatformExpertDevice, id 0x100000256, registered, matched, active, busy 0 (1032632 ms), retain 41>
                {
                    "manufacturer" = <"Apple Inc.">
                    "model" = <"MacBookPro18,3">
                    "IOPlatformSerialNumber" = "ABCDE1FGHI"
                    "IOPlatformUUID" = "ABCD1234-5678-90EF-GHIJ-KLMNOPQRSTUV"
                    "device_type" = <"bootrom">
                }
            `),
			want:    "ABCD1234-5678-90EF-GHIJ-KLMNOPQRSTUV",
			wantErr: false,
		},
		{
			name: "malformed ioreg format - concatenated entries",
			input: []byte(`+-o J314sAP  <class IOPlatformExpertDevice, id 0x100000256, registered, matched, active, busy 0 (1032632 ms), retain 41>
                {
                    "manufacturer" = <"Apple Inc.">
                    "model" = <"MacBookPro18,3">
                    "IOPlatformSerialNumber" = "ABCDE1FGHI""IOPlatformUUID" = "ABCD1234-5678-90EF-GHIJ-KLMNOPQRSTUV""device_type" = <"bootrom">
                }
            `),
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseIOPlatformUUID(tt.input)

			// Handle error expectations
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			// Handle success expectations
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
