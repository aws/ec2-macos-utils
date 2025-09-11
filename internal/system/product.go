package system

import (
	"fmt"

	"github.com/Masterminds/semver"
)

// Release is used to define macOS releases in an enumerated constant (e.g. Mojave, Catalina, BigSur)
type Release uint8

const (
	Unknown Release = iota
	Mojave
	Catalina
	BigSur
	Monterey
	Ventura
	Sonoma
	Sequoia
	Tahoe
	CompatMode
)

func (r Release) String() string {
	switch r {
	case Mojave:
		return "Mojave"
	case Catalina:
		return "Catalina"
	case BigSur:
		return "Big Sur"
	case Monterey:
		return "Monterey"
	case Ventura:
		return "Ventura"
	case Sonoma:
		return "Sonoma"
	case Sequoia:
		return "Sequoia"
	case Tahoe:
		return "Tahoe"
	case CompatMode:
		return "Compatability Mode"
	default:
		return "unknown"
	}
}

var (
	// mojaveConstraints are the constraints used to identify Mojave versions (10.14.x).
	mojaveConstraints = mustInitConstraint(semver.NewConstraint("~10.14"))
	// catalinaConstraints are the constraints used to identify Catalina versions (10.15.x).
	catalinaConstraints = mustInitConstraint(semver.NewConstraint("~10.15"))
	// bigSurConstraints are the constraints used to identify BigSur versions (11.x.x).
	bigSurConstraints = mustInitConstraint(semver.NewConstraint("~11"))
	// montereyConstraints are the constraints used to identify Monterey versions (12.x.x).
	montereyConstraints = mustInitConstraint(semver.NewConstraint("~12"))
	// venturaConstraints are the constraints used to identify Ventura versions (13.x.x).
	venturaConstraints = mustInitConstraint(semver.NewConstraint("~13"))
	// sonomaConstraints are the constraints used to identify Sonoma versions (14.x.x).
	sonomaConstraints = mustInitConstraint(semver.NewConstraint("~14"))
	// sequoiaConstraints are the constraints used to identify Sequoia versions (15.x.x).
	sequoiaConstraints = mustInitConstraint(semver.NewConstraint("~15"))
	// tahoeConstraints are the constraints used to identify Tahoe versions (26.x.x).
	tahoeConstraints = mustInitConstraint(semver.NewConstraint("~26"))
	// compatModeConstraints are the constraints used to identify macOS Big Sur and later. This version is returned
	// when the system is in compat mode (SYSTEM_VERSION_COMPAT=1).
	compatModeConstraints = mustInitConstraint(semver.NewConstraint("~10.16"))
)

// mustInitConstraint ensures that a semver.Constraints can be initialized and used.
func mustInitConstraint(c *semver.Constraints, err error) *semver.Constraints {
	if err != nil {
		panic(fmt.Errorf("must initialize semver constraint: %w", err))
	}
	return c
}

// Product identifies a macOS release and product version (e.g. Big Sur 11.x).
type Product struct {
	Release
	Version semver.Version
}

func (p Product) String() string {
	return fmt.Sprintf("macOS %s %s", p.Release, p.Version.String())
}

// newProduct initializes a new Product given the version string as input. It attempts to parse the version into a new
// semver.Version and then checks the version's constraints to identify the Release.
func newProduct(version string) (*Product, error) {
	ver, err := semver.NewVersion(version)
	if err != nil {
		return nil, err
	}

	release := getVersionRelease(*ver)

	product := &Product{
		Release: release,
		Version: *ver,
	}

	return product, nil
}

// getVersionRelease checks all known release constraints to determine which Release the version belongs to.
func getVersionRelease(version semver.Version) Release {
	switch {
	case mojaveConstraints.Check(&version):
		return Mojave
	case catalinaConstraints.Check(&version):
		return Catalina
	case bigSurConstraints.Check(&version):
		return BigSur
	case montereyConstraints.Check(&version):
		return Monterey
	case venturaConstraints.Check(&version):
		return Ventura
	case sonomaConstraints.Check(&version):
		return Sonoma
	case sequoiaConstraints.Check(&version):
		return Sequoia
	case tahoeConstraints.Check(&version):
		return Tahoe
	case compatModeConstraints.Check(&version):
		return CompatMode
	default:
		return Unknown
	}
}
