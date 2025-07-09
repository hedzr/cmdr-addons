package service

import (
	"context"
	"syscall"
	"unsafe"
)

func ChooseBackend(ctx context.Context) (be Backend, err error) {
	for _, c := range choosers {
		if ok := c.Choose(ctx); ok {
			if be, ok = c.(Backend); ok {
				return
			}
		}
	}
	return
}

var (
	choosers = []Chooser{
		&ntServiceD{},
	}
)

//
//
//

// ntdll.dll
var (
	modntdll               = syscall.NewLazyDLL("ntdll.dll")
	procRtlAdjustPrivilege = modntdll.NewProc("RtlAdjustPrivilege")
	procNtShutdownSystem   = modntdll.NewProc("NtShutdownSystem")
)

// user32.dll
var (
	user32            = syscall.NewLazyDLL("user32.dll")
	procExitWindowsEx = user32.NewProc("ExitWindowsEx")
)

// 提权等级
const (
	privilege = 19
)

// adjustPrivilege 提权
func adjustPrivilege(privilege uintptr, enable bool) (bool, error) {
	var (
		enableInt  uintptr
		wasEnabled uintptr
	)
	if enable {
		enableInt = 1
	}
	status, _, err := procRtlAdjustPrivilege.Call(privilege, enableInt, 0, uintptr(unsafe.Pointer(&wasEnabled)))
	if status != 0 {
		return false, err
	}
	return wasEnabled > 0, nil
}
