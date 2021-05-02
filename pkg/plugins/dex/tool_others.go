//+build !plan9
//+build !nacl

// Copyright © 2020 Hedzr Yeh.

package dex

import (
	"gopkg.in/hedzr/errors.v2"
	"os"
	"os/exec"
	"syscall"
)

func isExitError(err error) (int, bool) {
	if e, ok := err.(*exec.ExitError); ok {
		if status, ok := e.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus(), true
		}
	}

	return 0, false
}

// IsErrorAddressAlreadyInUse tests if "bind: address already in use" found
func IsErrorAddressAlreadyInUse(err error) bool {
	if e, ok := errors.Unwrap(err).(*os.SyscallError); ok {
		if errno, ok := e.Err.(syscall.Errno); ok {
			return errno == syscall.EADDRINUSE
		}
	}
	return false
}
