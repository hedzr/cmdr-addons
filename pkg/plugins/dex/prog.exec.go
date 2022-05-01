// Copyright © 2020 Hedzr Yeh.

package dex

import (
	"bytes"
	"fmt"
	"github.com/hedzr/log/dir"
	"gopkg.in/hedzr/errors.v3"
	"io"
	"os/exec"
)

func shell(command string, arguments ...string) error {
	_, _, err := runCommand(command, false, arguments...)
	return err
}

func shellWithOutput(command string, arguments ...string) (int, string, error) {
	return runCommand(command, true, arguments...)
}

func sudo(command string, arguments ...string) (int, string, error) {
	sudocmd, err := exec.LookPath("sudo")
	if err != nil {
		return -1, "'sudo' not found", shell(command, arguments...)
	}

	rc, output, err1 := runCommand(sudocmd, true, append([]string{command}, arguments...)...)
	return rc, output, err1
}

func runCommand(command string, readStdout bool, arguments ...string) (int, string, error) {
	cmd := exec.Command(command, arguments...)

	var output string
	var stdout io.ReadCloser
	var err error

	if readStdout {
		// Connect pipe to read Stdout
		stdout, err = cmd.StdoutPipe()

		if err != nil {
			// Failed to connect pipe
			return 0, "", fmt.Errorf("%q failed to connect stdout pipe: %v", command, err)
		}
	}

	// Connect pipe to read Stderr
	stderr, err := cmd.StderrPipe()

	if err != nil {
		// Failed to connect pipe
		return 0, "", fmt.Errorf("%q failed to connect stderr pipe: %v", command, err)
	}

	// Do not use cmd.Run()
	if err := cmd.Start(); err != nil {
		// Problem while copying stdin, stdout, or stderr
		return 0, "", fmt.Errorf("%q failed: %v", command, err)
	}

	// Zero exit status
	// Darwin: launchctl can fail with a zero exit status,
	// so check for emtpy stderr
	if command == "launchctl" {
		slurp, _ := dir.ReadAll(stderr)
		if len(slurp) > 0 && !bytes.HasSuffix(slurp, []byte("Operation now in progress\n")) {
			return 0, "", fmt.Errorf("%q failed with stderr: %s", command, slurp)
		}
	}

	if readStdout {
		out, err := dir.ReadAll(stdout)
		if err != nil {
			return 0, "", fmt.Errorf("%q failed while attempting to read stdout: %v", command, err)
		} else if len(out) > 0 {
			output = string(out)
		}
	}

	if err := cmd.Wait(); err != nil {
		exitStatus, ok := isExitError(err)
		slurp, _ := dir.ReadAll(stderr)
		if ok {
			// Command didn't exit with a zero exit status.
			// return exitStatus, output, fmt.Errorf("%q failed: %w |\n  stderr: %s", command, err, slurp)
			return exitStatus, output, errors.New("%q failed: %s |\n  stderr: %s", command, err.Error(), slurp).WithErrors(err)
		}

		// An error occurred and there is no exit status.
		// return 0, output, fmt.Errorf("%q failed: %w |\n  stderr: %s", command, err, slurp)
		return 0, output, errors.New("%q failed: %s |\n  stderr: %s", command, err.Error(), slurp).WithErrors(err)
	}

	return 0, output, nil
}
