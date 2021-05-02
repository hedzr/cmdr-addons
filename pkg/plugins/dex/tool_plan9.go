//+build plan9

// Copyright © 2020 Hedzr Yeh.

package dex

import "os/exec"

func isExitError(err error) (int, bool) {
	if ee, ok := err.(*exec.ExitError); ok {
		return ee.ExitCode(), true
	}

	return 0, false
}

// IsErrorAddressAlreadyInUse tests if "bind: address already in use" found
func IsErrorAddressAlreadyInUse(err error) bool {
	//if e, ok := errors.Unwrap(err).(*os.SyscallError); ok {
	//	if errno, ok := e.Err.(syscall.Errno); ok {
	//		return errno == syscall.EADDRINUSE
	//	}
	//}
	return false
}
