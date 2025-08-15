package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	imdsTokenURL = "http://169.254.169.254/latest/api/token"
)

func checkCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "run various system checks",
		Long:  "run diagnostics and checks on various system components",
	}

	cmd.AddCommand(
		checkImdsCommand(),
	)

	return cmd
}

func checkImdsCommand() *cobra.Command {
	return &cobra.Command{
		Use:          "imds",
		Short:        "check IMDS connectivity",
		Long:         "verifies connectivity to the EC2 Instance Metadata Service",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheckIMDS(cmd.Context())
		},
	}
}

func runCheckIMDS(ctx context.Context) error {
	const dialerTimeout = 5 * time.Second // timeout for the dialed network connection to start
	const imdsTokenLifetime = "941"       // arbitrary short-lived token lifetime

	logrus.Info("Starting IMDS connectivity check")

	client := &http.Client{Timeout: dialerTimeout}

	req, err := http.NewRequestWithContext(ctx, "PUT", imdsTokenURL, nil)
	if err != nil {
		logrus.WithError(err).Error("Failed to create request")
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("X-aws-ec2-metadata-token-ttl-seconds", imdsTokenLifetime)

	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Error("Failed to connect to IMDS")
		return fmt.Errorf("failed to connect to IMDS: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Error("Failed to read IMDS response")
		return fmt.Errorf("failed to read IMDS response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logrus.WithField("statusCode", resp.StatusCode).Error("IMDS returned non-200 status code")
		return fmt.Errorf("IMDS returned non-200 status code: %d", resp.StatusCode)
	}

	logrus.Info("IMDS connectivity check passed")
	return nil
}
