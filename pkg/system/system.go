// Package system provides the functionality necessary for interacting with the macOS system.
package system

import (
	"fmt"
	"howett.net/plist"
	"io"
	"os"
)

const (
	// VersionPath is the path on the root filesystem to the SystemVersion plist
	VersionPath = "/System/Library/CoreServices/SystemVersion.plist"
)

// VersionInfo mirrors the raw data found in VersionPath
type VersionInfo struct {
	ProductBuildVersion       string `plist:"ProductBuildVersion"`
	ProductCopyright          string `plist:"ProductCopyright"`
	ProductName               string `plist:"ProductName"`
	ProductUserVisibleVersion string `plist:"ProductUserVisibleVersion"`
	ProductVersion            string `plist:"ProductVersion"`
	IOSSupportVersion         string `plist:"iOSSupportVersion"`
}

// Product determines the specific product that the VersionInfo.ProductVersion is associated with
func (v *VersionInfo) Product() (Product, error) {
	return ProductFromVersion(v.ProductVersion)
}

// NewVersionInfo attempts to decode the raw data from the reader into a new VersionInfo struct.
func NewVersionInfo(reader io.ReadSeeker) (version *VersionInfo, err error) {
	// Catch panics thrown by the Decode method
	defer func() {
		if panicErr := recover(); panicErr != nil {
			version = nil
			err = fmt.Errorf("system panic occured while decoding: %s", panicErr)
		}
	}()

	// Create a reader from the raw data and create a new decoder
	version = &VersionInfo{}
	decoder := plist.NewDecoder(reader)

	// Decode the system version plist into the VersionInfo struct
	err = decoder.Decode(version)
	if err != nil {
		return nil, fmt.Errorf("system failed to decode contents of reader: %w", err)
	}

	return version, nil
}

// ReadVersion opens the VersionPath and calls NewVersionInfo to read and parse the raw plist into a new VersionInfo.
func ReadVersion() (*VersionInfo, error) {
	// Open the SystemVersion.plist file
	versionReader, err := os.Open(VersionPath)
	if err != nil {
		return nil, err
	}

	// Get the VersionInfo from the reader
	version, err := NewVersionInfo(versionReader)
	if err != nil {
		return nil, err
	}

	return version, nil
}
