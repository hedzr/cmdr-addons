//go:build solaris || aix || freebsd || openbsd || netbsd
// +build solaris aix freebsd openbsd netbsd

// unix but not darwin and linux ...

package systems

const HasNTService = false
const HasLaunchd = false
const HasLinuxBackends = false
const HasUnixBackends = true
