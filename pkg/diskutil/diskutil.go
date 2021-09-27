// Package diskutil provides the functionality necessary for interacting with macOS's diskutil CLI.
package diskutil

//go:generate mockgen -source=diskutil.go -destination=mocks/mock_diskutil.go

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aws/ec2-macos-utils/pkg/diskutil/types"
	"github.com/aws/ec2-macos-utils/pkg/system"

	"github.com/Masterminds/semver"
	"github.com/dustin/go-humanize"
)

const (
	// MinimumGrowSpaceRequired defines the minimum amount of free space (in bytes) required to attempt running
	// diskutil's resize command.
	MinimumGrowSpaceRequired = 1000000
)

// MinimumGrowSpaceError defines an error to distinguish when there's not enough space to grow the specified container.
type MinimumGrowSpaceError struct {
	size uint64
}

// Error provides the implementation for the error interface.
func (e MinimumGrowSpaceError) Error() string {
	message := fmt.Sprintf("grow requires [%s] but got [%s]",
		humanize.Bytes(MinimumGrowSpaceRequired), humanize.Bytes(e.size))
	return message
}

// DiskUtil outlines the functionality necessary for wrapping macOS's diskutil tool.
type DiskUtil interface {
	APFS
	Info(id string) (*types.DiskInfo, error)
	List(args []string) (*types.SystemPartitions, error)
	RepairDisk(id string) (string, error)
}

// APFS outlines the functionality necessary for wrapping diskutil's APFS verb.
type APFS interface {
	ResizeContainer(id, size string) (string, error)
}

// ForProduct creates a new diskutil controller for the given product.
func ForProduct(p system.Product) (DiskUtil, error) {
	switch p.Release {
	case system.Mojave:
		return newMojave(p.Version)
	case system.Catalina:
		return newCatalina(p.Version)
	case system.BigSur:
		return newBigSur(p.Version)
	}

	return nil, errors.New("unknown release")
}

// newMojave configures the DiskUtil for the specified Mojave version.
func newMojave(version semver.Version) (*DiskUtilityMojave, error) {
	du := &DiskUtilityMojave{
		Util:    &DiskUtilityCmd{},
		Decoder: &PlistDecoder{},
	}

	return du, nil
}

// newCatalina configures the DiskUtil for the specified Catalina version.
func newCatalina(version semver.Version) (*DiskUtilityCatalina, error) {
	du := &DiskUtilityCatalina{
		Util:    &DiskUtilityCmd{},
		Decoder: &PlistDecoder{},
	}

	return du, nil
}

// newMojave configures the DiskUtil for the specified Big Sur version.
func newBigSur(version semver.Version) (*DiskUtilityBigSur, error) {
	du := &DiskUtilityBigSur{
		Util:    &DiskUtilityCmd{},
		Decoder: &PlistDecoder{},
	}

	return du, nil
}

// DiskUtilityMojave wraps all the functionality necessary for interacting with macOS's diskutil on Mojave. The
// major difference is that the raw plist data emitted by macOS's diskutil CLI doesn't include the physical store data.
// This requires a separate fetch to find the specific physical store information for the disk(s).
type DiskUtilityMojave struct {
	Util    UtilImpl
	Decoder Decoder
}

// List utilizes the UtilImpl.List method to fetch the raw list output from diskutil and returns the decoded
// output in a SystemPartitions struct. List also attempts to update each APFS Volume's physical store via a separate
// fetch method since the version of diskutil on Mojave doesn't provide that information in its List verb.
//
// It is possible for List to fail when updating the physical stores, but it will still return the original data
// that was decoded into the SystemPartitions struct.
func (d *DiskUtilityMojave) List(args []string) (*types.SystemPartitions, error) {
	partitions, err := list(d.Util, d.Decoder, args)
	if err != nil {
		return nil, err
	}

	err = updatePhysicalStores(partitions)
	if err != nil {
		return partitions, err
	}

	return partitions, nil
}

// Info utilizes the UtilImpl.Info method to fetch the raw disk output from diskutil and returns the decoded
// output in a DiskInfo struct. Info also attempts to update each APFS Volume's physical store via a separate
// fetch method since the version of diskutil on Mojave doesn't provide that information in its Info verb.
//
// It is possible for Info to fail when updating the physical stores, but it will still return the original data
// that was decoded into the DiskInfo struct.
func (d *DiskUtilityMojave) Info(id string) (*types.DiskInfo, error) {
	disk, err := info(d.Util, d.Decoder, id)
	if err != nil {
		return nil, err
	}

	err = updatePhysicalStore(disk)
	if err != nil {
		return disk, err
	}

	return disk, nil
}

// RepairDisk wraps DiskUtil.RepairDisk.
func (d *DiskUtilityMojave) RepairDisk(id string) (out string, err error) {
	return repairDisk(d.Util, id)
}

// ResizeContainer wraps APFS.ResizeContainer.
func (d *DiskUtilityMojave) ResizeContainer(id, size string) (out string, err error) {
	return resizeContainer(d.Util, id, size)
}

// DiskUtilityCatalina wraps all the functionality necessary for interacting with macOS's diskutil in GoLang.
type DiskUtilityCatalina struct {
	Util    UtilImpl
	Decoder Decoder
}

// List utilizes the UtilImpl.List method to fetch the raw list output from diskutil and returns the decoded
// output in a SystemPartitions struct.
func (d *DiskUtilityCatalina) List(args []string) (*types.SystemPartitions, error) {
	return list(d.Util, d.Decoder, args)
}

// Info utilizes the UtilImpl.Info method to fetch the raw disk output from diskutil and returns the decoded
// output in a DiskInfo struct.
func (d *DiskUtilityCatalina) Info(id string) (*types.DiskInfo, error) {
	return info(d.Util, d.Decoder, id)
}

// RepairDisk wraps UtilImpl.RepairDisk to provide the functionality within DiskUtil. No decoding of the output is
// necessary since it outputs human-readable lines.
func (d *DiskUtilityCatalina) RepairDisk(id string) (string, error) {
	return repairDisk(d.Util, id)
}

// ResizeContainer wraps APFS.ResizeContainer to provide the functionality within DiskUtil. No decoding of the output
// is necessary since it outputs human-readable lines.
func (d *DiskUtilityCatalina) ResizeContainer(id, size string) (string, error) {
	return resizeContainer(d.Util, id, size)
}

// DiskUtilityBigSur wraps all the functionality necessary for interacting with macOS's diskutil in GoLang.
type DiskUtilityBigSur struct {
	Util    UtilImpl
	Decoder Decoder
}

// List utilizes the UtilImpl.List method to fetch the raw list output from diskutil and returns the decoded
// output in a SystemPartitions struct.
func (d *DiskUtilityBigSur) List(args []string) (*types.SystemPartitions, error) {
	return list(d.Util, d.Decoder, args)
}

// Info utilizes the UtilImpl.Info method to fetch the raw disk output from diskutil and returns the decoded
// output in a DiskInfo struct.
func (d *DiskUtilityBigSur) Info(id string) (*types.DiskInfo, error) {
	return info(d.Util, d.Decoder, id)
}

// RepairDisk wraps UtilImpl.RepairDisk to provide the functionality within DiskUtil. No decoding of the output is
// necessary since it outputs human-readable lines.
func (d *DiskUtilityBigSur) RepairDisk(id string) (string, error) {
	return repairDisk(d.Util, id)
}

// ResizeContainer wraps APFS.ResizeContainer to provide the functionality within DiskUtil. No decoding of the output
// is necessary since it outputs human-readable lines.
func (d *DiskUtilityBigSur) ResizeContainer(id, size string) (string, error) {
	return resizeContainer(d.Util, id, size)
}

func list(util UtilImpl, decoder Decoder, args []string) (*types.SystemPartitions, error) {
	// Fetch the raw list information from the util
	rawPartitions, err := util.List(args)
	if err != nil {
		return nil, err
	}

	// Create a reader for the raw data
	reader := strings.NewReader(rawPartitions)

	// Decode the raw data into a more usable SystemPartitions struct
	partitions, err := decoder.DecodeSystemPartitions(reader)
	if err != nil {
		return nil, err
	}

	return partitions, nil
}

func info(util UtilImpl, decoder Decoder, id string) (*types.DiskInfo, error) {
	// Fetch the raw disk information from the util
	rawDisk, err := util.Info(id)
	if err != nil {
		return nil, err
	}

	// Create a reader for the raw data
	reader := strings.NewReader(rawDisk)

	// Decode the raw data into a more usable DiskInfo struct
	disk, err := decoder.DecodeDiskInfo(reader)
	if err != nil {
		return nil, err
	}

	return disk, nil
}

func repairDisk(util UtilImpl, id string) (string, error) {
	return util.RepairDisk(id)
}

func resizeContainer(util UtilImpl, id, size string) (string, error) {
	return util.ResizeContainer(id, size)
}
