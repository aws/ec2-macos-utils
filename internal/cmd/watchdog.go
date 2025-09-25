package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

func watchdogCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watchdog",
		Short: "monitor system health",
		Long: strings.TrimSpace(`
monitor system health and collect diagnostic data.
Contains subcommands for monitoring various aspects of system health.
        `),
	}

	cmd.AddCommand(newNetworkHealthMonitorCommand())
	return cmd
}
