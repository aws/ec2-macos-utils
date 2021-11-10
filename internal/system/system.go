// Package system provides the functionality necessary for interacting with the macOS system.
package system

import (
	"fmt"
	"io"
	"os"

	"howett.net/plist"
)

// versionPath is the path on the root filesystem to the SystemVersion plist
const versionPath = "/System/Library/CoreServices/SystemVersion.plist"

// System correlates VersionInfo with a Product.
type System struct {
	versionInfo *VersionInfo
	product     *Product
}

func (sys *System) Product() *Product {
	return sys.product
}

// Scan reads the VersionInfo and creates a new System struct from that and the associated Product.
func Scan() (*System, error) {
	version, err := readVersion()
	if err != nil {
		return nil, err
	}

	product, err := version.Product()
	if err != nil {
		return nil, err
	}

	system := &System{
		versionInfo: version,
		product:     product,
	}

	return system, nil
}

// VersionInfo mirrors the raw data found in the SystemVersion plist file.
type VersionInfo struct {
	ProductBuildVersion       string `plist:"ProductBuildVersion"`
	ProductCopyright          string `plist:"ProductCopyright"`
	ProductName               string `plist:"ProductName"`
	ProductUserVisibleVersion string `plist:"ProductUserVisibleVersion"`
	ProductVersion            string `plist:"ProductVersion"`
	IOSSupportVersion         string `plist:"iOSSupportVersion"`
}

// Product determines the specific product that the VersionInfo.ProductVersion is associated with.
func (v *VersionInfo) Product() (*Product, error) {
	return newProduct(v.ProductVersion)
}

// decodeVersionInfo attempts to decode the raw data from the reader into a new VersionInfo struct.
func decodeVersionInfo(reader io.ReadSeeker) (version *VersionInfo, err error) {
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

// readVersion opens the versionPath and calls decodeVersionInfo to read and parse the raw plist into a new VersionInfo.
func readVersion() (*VersionInfo, error) {
	// Open the SystemVersion.plist file
	versionReader, err := os.Open(versionPath)
	if err != nil {
		return nil, err
	}

	// Get the VersionInfo from the reader
	version, err := decodeVersionInfo(versionReader)
	if err != nil {
		return nil, err
	}

	return version, nil
}
