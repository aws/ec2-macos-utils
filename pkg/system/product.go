package system

import (
	"errors"
	"fmt"

	"github.com/Masterminds/semver"
)

// Release is used to define macOS releases in an enumerated constant (e.g. Mojave, Catalina, BigSur)
type Release uint8

const (
	Mojave Release = iota + 1
	Catalina
	BigSur
	MaxRelease
	Unknown Release = 0
)

func (r Release) String() string {
	switch r {
	case Mojave:
		return "Mojave"
	case Catalina:
		return "Catalina"
	case BigSur:
		return "Big Sur"
	default:
		return "unknown"
	}
}

var (
	mojaveConstraints   *semver.Constraints
	catalinaConstraints *semver.Constraints
	bigsurConstraints   *semver.Constraints
)

func init() {
	mojaveConstraints, _ = semver.NewConstraint("~10.14")
	catalinaConstraints, _ = semver.NewConstraint("~10.15")
	bigsurConstraints, _ = semver.NewConstraint("~11 || ~10.16")
}

// Product identifies a macOS release and product version (e.g. Big Sur 11.x).
type Product struct {
	Release
	Version semver.Version
}

func (p Product) String() string {
	return fmt.Sprintf("macOS %s %s", p.Release, p.Version.String())
}

// NewProduct initializes a new Product given the version string as input. It attempts to parse the version into a new
// semver.Version and then checks the version's constraints to identify the Release.
func NewProduct(version string) (*Product, error) {
	ver, err := semver.NewVersion(version)
	if err != nil {
		return nil, err
	}

	release, err := getVersionRelease(*ver)
	if err != nil {
		return nil, err
	}

	product := &Product{
		Release: release,
		Version: *ver,
	}

	return product, nil
}

// getVersionRelease checks all known release constraints to determine which Release the version belongs to.
func getVersionRelease(version semver.Version) (Release, error) {
	switch {
	case mojaveConstraints.Check(&version):
		return Mojave, nil
	case catalinaConstraints.Check(&version):
		return Catalina, nil
	case bigsurConstraints.Check(&version):
		return BigSur, nil
	}

	return Unknown, errors.New("unknown system version")
}
