package identifier

import (
	"regexp"
	"strings"
)

// diskIDExp is the regexp expression for device identifiers.
var diskIDExp = regexp.MustCompile("disk[0-9]+")

// ParseDiskID parses a supported disk identifier from a string.
func ParseDiskID(s string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}
	return diskIDExp.FindString(s)
}
