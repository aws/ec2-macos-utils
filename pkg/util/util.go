package util

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"
)

// CommandOutput wraps the output from an exec command as strings.
type CommandOutput struct {
	Stdout string
	Stderr string
}

// ExecuteCommand executes the command and returns Stdout and Stderr as strings.
func ExecuteCommand(c []string, runAsUser string, envVars []string, stdin io.ReadCloser) (output CommandOutput, err error) {
	// Separate name and args, plus catch a few error cases
	var name string
	var args []string

	// Check the empty struct case ([]string{}) for the command
	if len(c) == 0 {
		return CommandOutput{}, fmt.Errorf("must provide a command")
	}

	// Set the name of the command and check if args are also provided
	name = c[0]
	if len(c) > 1 {
		args = c[1:]
	}

	// Set command and create output buffers
	cmd := exec.Command(name, args...)
	var stdoutb, stderrb bytes.Buffer
	cmd.Stdout = &stdoutb
	cmd.Stderr = &stderrb

	// Set command stdin if the stdin parameter is provided
	if stdin != nil {
		cmd.Stdin = stdin
	}

	// Set runAsUser, if defined, otherwise will run as root
	if runAsUser != "" {
		uid, gid, err := getUIDandGID(runAsUser)
		if err != nil {
			return CommandOutput{Stdout: stdoutb.String(), Stderr: stderrb.String()}, fmt.Errorf("error looking up user: %s\n", err)
		}
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
	}

	// Append environment variables
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, envVars...)

	// Start the command's execution
	if err = cmd.Start(); err != nil {
		return CommandOutput{}, fmt.Errorf("error starting specified command: %w", err)
	}

	// Wait for the command to exit
	if err = cmd.Wait(); err != nil {
		return CommandOutput{}, fmt.Errorf("error waiting for specified command to exit: %w", err)
	}

	return CommandOutput{Stdout: stdoutb.String(), Stderr: stderrb.String()}, err
}

// ExecuteCommandYes wraps ExecuteCommand with the yes binary in order to bypass user input states in automation.
func ExecuteCommandYes(c []string, runAsUser string, envVars []string) (output CommandOutput, err error) {
	// Set exec commands, one for yes and another for the specified command
	cmdYes := exec.Command("/usr/bin/yes")

	// Pipe cmdYes into cmd
	stdin, err := cmdYes.StdoutPipe()
	if err != nil {
		return CommandOutput{}, fmt.Errorf("error creating pipe between commands")
	}

	// Start the command to run /usr/bin/yes
	if err = cmdYes.Start(); err != nil {
		return CommandOutput{}, fmt.Errorf("error starting /usr/bin/yes command: %w", err)
	}

	return ExecuteCommand(c, runAsUser, envVars, stdin)
}

// getUIDandGID takes a username and returns the uid and gid for that user.
// While testing UID/GID lookup for a user, it was found that the user.Lookup() function does not always return
// information for a new user on first boot. In the case that user.Lookup() fails, try dscacheutil, which has a
// higher success rate. If that fails, return an error. Any successful case returns the UID and GID as ints.
func getUIDandGID(username string) (uid int, gid int, err error) {
	var uidstr, gidstr string
	// Preference is user.Lookup(), if it works
	u, lookuperr := user.Lookup(username)
	if lookuperr != nil {
		// user.Lookup() has failed, second try by checking the DS cache
		out, cmderr := ExecuteCommand([]string{"dscacheutil", "-q", "user", "-a", "name", username}, "", []string{}, nil)
		if cmderr != nil {
			// dscacheutil has failed with an error
			return 0, 0, fmt.Errorf("error while looking up user %s: \n"+
				"user.Lookup() error: %s \ndscacheutil error: %s\ndscacheutil Stderr: %s\n",
				username, lookuperr, cmderr, out.Stderr)
		}
		// Check length of Stdout - dscacheutil returns nothing if user is not found
		if len(out.Stdout) > 0 { // dscacheutil has returned something
			// Command output from dscacheutil should look like:
			//   name: ec2-user
			//   password: ********
			//   uid: 501
			//   gid: 20
			//   dir: /Users/ec2-user
			//   shell: /bin/bash
			//   gecos: ec2-user
			dsSplit := strings.Split(out.Stdout, "\n") // split on newline to separate uid and gid
			for _, e := range dsSplit {
				eSplit := strings.Fields(e) // split into fields to separate tag with id
				// Find UID and GID and set them
				if strings.HasPrefix(e, "uid") {
					if len(eSplit) != 2 {
						// dscacheutil has returned some sort of weird output that can't be split
						return 0, 0, fmt.Errorf("error while splitting dscacheutil uid output for user %s: %s\n"+
							"user.Lookup() error: %s \ndscacheutil error: %s\ndscacheutil Stderr: %s\n",
							username, out.Stdout, lookuperr, cmderr, out.Stderr)
					}
					uidstr = eSplit[1]
				} else if strings.HasPrefix(e, "gid") {
					if len(eSplit) != 2 {
						// dscacheutil has returned some sort of weird output that can't be split
						return 0, 0, fmt.Errorf("error while splitting dscacheutil gid output for user %s: %s\n"+
							"user.Lookup() error: %s \ndscacheutil error: %s\ndscacheutil Stderr: %s\n",
							username, out.Stdout, lookuperr, cmderr, out.Stderr)
					}
					gidstr = eSplit[1]
				}
			}
		} else {
			// dscacheutil has returned nothing, user is not found
			return 0, 0, fmt.Errorf("user %s not found: \n"+
				"user.Lookup() error: %s \ndscacheutil error: %s\ndscacheutil Stderr: %s\n",
				username, lookuperr, cmderr, out.Stderr)
		}
	} else {
		// user.Lookup() was successful, use the returned UID/GID
		uidstr = u.Uid
		gidstr = u.Gid
	}

	// Convert UID and GID to int
	uid, err = strconv.Atoi(uidstr)
	if err != nil {
		return 0, 0, fmt.Errorf("error while converting UID to int: %s\n", err)
	}
	gid, err = strconv.Atoi(gidstr)
	if err != nil {
		return 0, 0, fmt.Errorf("error while converting GID to int: %s\n", err)
	}

	return uid, gid, nil
}
