// Package system provides the functionality necessary for interacting with the macOS system.
package system

import (
	"fmt"
	"io"
	"os"

	"howett.net/plist"
)

const (
	// versionPath is the path on the root filesystem to the SystemVersion plist
	versionPath = "/System/Library/CoreServices/SystemVersion.plist"

	// dotVersionPath is the path to the symlink that directly references versionPath and bypasses the compatibility
	// mode that was introduced with macOS 11.0.
	dotVersionPath = "/System/Library/CoreServices/.SystemVersionPlatform.plist"

	// dotVersionSwitch is the product version number returned by macOS when the system is in compat mode
	// (SYSTEM_VERSION_COMPAT=1). If this version is returned, dotVersionPath should be read to bypass compat mode.
	dotVersionSwitch = "10.16"
)

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

// readVersion reads the SystemVersion plist data from disk (versionPath). If "SYSTEM_VERSION_COMPAT" is enabled, it will instead
// read from dotVersionPath to bypass macOS's compat mode.
func readVersion() (*VersionInfo, error) {
	// Read the version info from the standard file path
	version, err := readProductVersionFile(versionPath)
	if err != nil {
		return nil, err
	}

	// If the returned product version is in compat mode, read the version info from the dot file to bypass compat mode.
	if version.ProductVersion == dotVersionSwitch {
		return readProductVersionFile(dotVersionPath)
	}

	return version, nil
}

// readProductVersion opens the given file and attempts to decode it as VersionInfo.
func readProductVersionFile(path string) (*VersionInfo, error) {
	// Open the SystemVersion.plist file
	versionFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer versionFile.Close()

	// Get the VersionInfo from the reader
	version, err := decodeVersionInfo(versionFile)
	if err != nil {
		return nil, err
	}

	return version, nil
}
