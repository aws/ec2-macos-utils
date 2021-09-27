// Package cmd provides the functionality necessary for CLI commands in EC2 macOS Utils.
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/ec2-macos-utils/internal/build"
	"github.com/aws/ec2-macos-utils/pkg/system"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ec2-macos-utils",
	Short: "EC2 macOS Utils provides utilities for EC2 macOS instances.",
	Long: `EC2 macOS Utils provides utilities for EC2 macOS instances. 
These utilities provide quick access to a variety of automation steps 
for configuring macOS instances.`,
	Version:      build.Version,
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
		build.Version, build.CommitDate, build.GitHubLink,
	))

	// Set the persistent pre-run function to configure things before command execution
	rootCmd.PersistentPreRunE = configureUtils

	// Set persistent flags
	rootCmd.PersistentFlags().BoolVarP(&build.Verbose, "verbose", "v", false, "verbose output")
}

// configureUtils configures everything necessary before ec2-macos-utils runs.
func configureUtils(cmd *cobra.Command, args []string) error {
	setupLogger()

	logrus.Debug("Configuring the product version...")
	version, err := system.ReadVersion()
	logrus.WithField("version", version).Debug("Found version")
	if err != nil {
		return err
	}

	logrus.Debug("Configuring the product...")
	product, err := version.Product()
	logrus.WithField("version", version).Debug("Found product")
	if err != nil {
		return err
	}
	build.Product = *product

	logrus.WithField("product", build.Product).Debug("Configured ec2-macos-utils for product")

	return nil
}

// setupLogger configures logrus to use the desired timestamp format and log level.
func setupLogger() {
	Formatter := new(logrus.TextFormatter)

	// Configure the formatter
	Formatter.TimestampFormat = time.RFC822
	Formatter.FullTimestamp = true

	// Set the desired log level
	if build.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	logrus.SetFormatter(Formatter)
}
