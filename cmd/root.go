// Package cmd provides the functionality necessary for CLI commands in EC2 macOS Utils.
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/ec2-macos-utils/pkg/system"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	// gitHubLink is the static HTTPS URL for EC2 macOS Utils public GitHub repository.
	gitHubLink = "https://github.com/aws/ec2-macos-utils"
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

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ec2-macos-utils",
	Short: "EC2 macOS Utils provides utilities for EC2 macOS instances.",
	Long: `EC2 macOS Utils provides utilities for EC2 macOS instances. 
These utilities provide quick access to a variety of automation steps 
for configuring mac1.metal instances.`,
	Version:      Version,
	SilenceUsage: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// init initializes the root command, all sub-commands, and sets flags
func init() {
	// Set the template to be used when the version flag is provided
	rootCmd.SetVersionTemplate(fmt.Sprintf("\nEC2 macOS Utils\n"+
		"Version: %s [%s]\n"+
		"%s\n"+
		"Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.\n\n",
		Version, CommitDate, gitHubLink,
	))

	// Set the persistent pre-run function to configure things before command execution
	rootCmd.PersistentPreRunE = configureUtils

	// Set persistent flags
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
}

// configureUtils configures everything necessary before ec2-macos-utils runs.
func configureUtils(cmd *cobra.Command, args []string) error {
	// Configure the logger
	setupLogger()

	logrus.Debug("Configuring EC2 macOS Utils...")

	logrus.Debug("Configuring the product version...")
	// Set the running product version
	err := setupProductVersion()
	if err != nil {
		return err
	}
	logrus.WithField("product", Product).Debug("Configured the product version")

	logrus.Debug("Completed configuring EC2 macOS Utils")

	return nil
}

// setupProductVersion configures the package global variable ProductVersion.
func setupProductVersion() error {
	// Get the system version information
	version, err := system.ReadVersion()
	if err != nil {
		return err
	}

	// Determine the correct baseProduct from the version information
	Product, err = version.Product()
	if err != nil {
		return err
	}

	return nil
}

// setupLogger configures logrus to use the desired timestamp format and log level.
func setupLogger() {
	Formatter := new(logrus.TextFormatter)

	// Set the desired timestamp format and log level
	if Verbose {
		Formatter.TimestampFormat = time.RFC3339Nano
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		Formatter.TimestampFormat = time.RFC822
		logrus.SetLevel(logrus.InfoLevel)
	}

	Formatter.FullTimestamp = true

	logrus.SetFormatter(Formatter)
}
