package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// Persistent flag variables
var ContainerID string // ContainerID is the identifier for the container to be resized or "root"

// resizeContainerCmd represents the resizeContainer command which provides functionality for resizing APFS containers.
// It has the following flags:
//   * id - the ID for the container to be resized (required)
//
// It has the following sub-commands:
//   * grow - resizes the specified container to its maximum size using diskutil
var resizeContainerCmd = &cobra.Command{
	Use:   "resizeContainer",
	Short: "Resizes APFS Containers",
	Long: `resizeContainer automates the manual steps for resizing
APFS Containers. This can be used to grow the APFS Container to
the full size of the EBS Volume attached to the instance.`,
	Args: cobra.MinimumNArgs(1),
}

// growCmd represents the grow command which provides functionality for growing APFS containers to their maximum size.
var growCmd = &cobra.Command{
	Use:   "grow",
	Short: "Resizes the container to its maximum size",
	Long: `grow attempts to resize the specified container to its 
maximum size using Apple's diskutil tool. The container can be
specified with its identifier (e.g. disk1 or /dev/disk1) or
with "root" if the target container is the one with the OS root.'

Note: if the EBS Volume size was changed and the instance hasn't 
been restarted yet, this command will fail to resize the container
until the instance has been restarted.`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		fmt.Printf("grow called with args %#v\n", args)

		// Check if the ContainerID flag is "root" or assume it's a container ID (e.g. /dev/disk1 or disk1)
		if strings.EqualFold(ContainerID, "root") {
			fmt.Println("Attempting to grow root container...")
			message, err := growRootContainer()
			if err != nil {
				return fmt.Errorf("failed to grow root container with message [%s], error [%v]", message, err)
			}

			fmt.Printf("Successfully grew root container with message: %s\n", message)
		} else {
			fmt.Printf("Attempting to grow container with ID [%s]...\n", ContainerID)
			message, err := growContainer(ContainerID)
			if err != nil {
				return fmt.Errorf("failed to grow container with message [%s], error [%v]", message, err)
			}

			fmt.Printf("Successfully grew container with message: %s\n", message)
		}

		return nil
	},
}

// init initializes the resizeContainer command, all sub-commands, and sets their respective flags.
func init() {
	// Define flags used in the resize container command
	resizeContainerCmd.PersistentFlags().StringVarP(&ContainerID, "id", "", "", "the ID of the APFS Container or \"root\" (required)")
	resizeContainerCmd.MarkPersistentFlagRequired("id")

	// Add the resize container command and sub-commands to the root command
	rootCmd.AddCommand(resizeContainerCmd)
	resizeContainerCmd.AddCommand(growCmd)
}

// growRootContainer finds the ID for the root container and grows the container to its maximum size.
func growRootContainer() (message string, err error) {
	// Attempt to find the ID for the root container
	fmt.Println("Searching for root container to resize...")
	rootID, err := getRootContainerID()
	if err != nil {
		message = "Failed to find the ID for the root container"
		return message, err
	}

	// Attempt to grow the root container
	fmt.Println("Attempting to grow the root container...")
	message, err = growContainer(rootID)
	if err != nil {
		message = "Failed to grow the root container"
		return message, err
	}

	return message, nil
}

// growContainer grows a container to its maximum size given an ID.
func growContainer(id string) (message string, err error) {
	// Check that the given container ID is valid
	fmt.Printf("Validating container ID [%s]...\n", id)
	valid, err := validateContainerID(id)
	if err != nil {
		message = fmt.Sprintf("Failed to validate container [%s]", id)
		return message, err
	}
	if !valid {
		message = fmt.Sprintf("Container ID [%s] is not valid", id)
		return message, err
	}

	// Get the size of the container
	rootSize, err := getContainerSize(id)
	if err != nil {
		message = fmt.Sprintf("Failed to determine current size of container [%s]", id)
		return message, err
	}
	fmt.Printf("Found [%s], size [%s]\n", id, rootSize)

	// Attempt to resize the container to its maximum size
	fmt.Printf("Resizing [%s] to use full partition...\n", id)
	newSize, err := resizeContainer(id, "0")
	if err != nil {
		message = fmt.Sprintf("Failed to grow the container [%s]", id)
		return message, err
	}

	message = fmt.Sprintf("Container [%s] is now size %s", id, newSize)
	return message, nil
}

// resizeContainer uses macOS's diskutil command to change the size of the specified container ID.
func resizeContainer(id, size string) (newSize string, err error) {

	return newSize, nil
}

// validateContainerID verifies if the provided ID is a valid container.
func validateContainerID(id string) (valid bool, err error) {

	return valid, nil
}

// getRootContainerID determines the ID for the container which is mounted as root.
func getRootContainerID() (id string, err error) {

	return id, nil
}

// getContainerSize returns the human-readable size of a container given a container ID.
func getContainerSize(id string) (size string, err error) {

	return size, nil
}
