package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	gitHubLink = "https://github.com/aws/ec2-macos-utils"
)

// Build time variables
var CommitDate string
var Version string

// Persistent flag variables
var Verbose bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ec2-macos-utils",
	Short: "EC2 macOS Utils provides utilities for EC2 macOS instances.",
	Long: `EC2 macOS Utils provides utilities for EC2 macOS instances. 
These utilities provide quick access to a variety of automation steps 
for configuring mac1.metal instances.`,
	Version: Version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
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

	// Set persistent flags
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
}
