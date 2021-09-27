package build

import "github.com/aws/ec2-macos-utils/pkg/system"

const (
	// GitHubLink is the static HTTPS URL for EC2 macOS Utils public GitHub repository.
	GitHubLink = "https://github.com/aws/ec2-macos-utils"
)

var (
	// CommitDate is the date of the latest commit in the repository. This variable gets set at build-time.
	CommitDate string

	// Version is the latest version of the utility. This variable gets set at build-time.
	Version string

	// Product is the type used to define what product version EC2 macOS Utils is running on.
	Product system.Product

	// Verbose is a persistent flag that determines the level of output to be logged.
	Verbose bool
)
