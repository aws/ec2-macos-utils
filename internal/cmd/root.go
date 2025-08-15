// Package cmd provides the functionality necessary for CLI commands in EC2 macOS Utils.
package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/aws/ec2-macos-utils/internal/build"
)

const shortLicenseText = "Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved."

// MainCommand provides the main program entrypoint that dispatches to utility subcommands.
func MainCommand() *cobra.Command {
	cmd := rootCommand()

	cmds := []*cobra.Command{
		growContainerCommand(),
		checkCommand(),
	}
	for i := range cmds {
		cmd.AddCommand(cmds[i])
	}

	return cmd
}

// rootCommand builds a root command object for program run.
func rootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ec2-macos-utils",
		Short: "utilities for EC2 macOS instances",
		Long: strings.TrimSpace(`
This command provides utilities for common tasks on EC2 macOS instances to simplify operation & administration.

This includes disk manipulation and system configuration helpers. Tasks are reached through subcommands, each with 
help text and usages that accompany them.
`),
		Version:           build.Version,
		SilenceUsage:      true,
		DisableAutoGenTag: true,
	}

	versionTemplate := "{{.Name}} {{.Version}} [%s]\n\n%s\n"
	cmd.SetVersionTemplate(fmt.Sprintf(versionTemplate, build.CommitDate, shortLicenseText))

	var verbose bool
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging output")

	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		level := logrus.InfoLevel
		if verbose {
			level = logrus.DebugLevel
		}
		setupLogging(level)

		return nil
	}

	return cmd
}

// setupLogging configures logrus to use the desired timestamp format and log level.
func setupLogging(level logrus.Level) {
	Formatter := &logrus.TextFormatter{}

	// Configure the formatter
	Formatter.TimestampFormat = time.RFC822
	Formatter.FullTimestamp = true

	// Set the desired log level
	logrus.SetLevel(level)

	logrus.SetFormatter(Formatter)
}

func hasRootPrivileges() bool {
	return os.Geteuid() == 0
}

// assertRootPrivileges checks if the command is running with root permissions.
// If the command doesn't have root permissions, a help message is logged with
// an example and an error is returned.
func assertRootPrivileges(cmd *cobra.Command, args []string) error {
	logrus.Debug("Checking user permissions...")
	ok := hasRootPrivileges()
	if !ok {
		logrus.Warn("Root privileges required")
		return errors.New("root privileges required, re-run command with sudo")
	}

	return nil
}
