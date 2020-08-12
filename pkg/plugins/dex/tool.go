/*
 * Copyright © 2019 Hedzr Yeh.
 */

package dex

import (
	"bytes"
	"github.com/hedzr/cmdr-addons/pkg/plugins/dex/sig"
	"gopkg.in/hedzr/errors.v2"
	"net"
	"os"
	"os/exec"
	"syscall"
	"text/template"
)

func tplApply(tmpl string, data interface{}) string {
	var w = new(bytes.Buffer)
	var tpl = template.Must(template.New("y").Parse(tmpl))
	if err := tpl.Execute(w, data); err != nil {
		pd.log.Errorf("tpl execute error: %v", err)
	}
	return w.String()
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

func isRootUser() bool {
	return os.Getuid() == 0
}

func shellRunAuto(name string, arg ...string) error {
	output, err := shellRun(name, arg...)
	if err != nil {
		pd.log.Fatalf("shellRunAuto err: %v\n\noutput:\n%v", err, output.String())
	}
	return err
}

func shellRun(name string, arg ...string) (output bytes.Buffer, err error) {
	cmd := exec.Command(name, arg...)
	// Command.Stdin = strings.NewReader("some input")
	cmd.Stdout = &output
	err = cmd.Run()
	return
}

// IsRunningInDemonizedMode returns true if you are running under demonized mode.
// false means that you're running in normal console/tty mode.
func IsRunningInDemonizedMode() bool {
	// return cmdr.GetBoolR(DaemonizedKey)
	return sig.IsRunningInDemonizedMode()
}

// SetTermSignals allows an functor to provide a list of Signals
func SetTermSignals(sigfn func() []os.Signal) {
	// onSetTermHandler = sig
	sig.SetTermSignals(sigfn)
}

// SetSigEmtSignals allows an functor to provide a list of Signals
func SetSigEmtSignals(sigfn func() []os.Signal) {
	// onSetSigEmtHandler = sig
	sig.SetSigEmtSignals(sigfn)
}

// SetReloadSignals allows an functor to provide a list of Signals
func SetReloadSignals(sigfn func() []os.Signal) {
	// onSetReloadHandler = sig
	sig.SetReloadSignals(sigfn)
}

// SetHotReloadSignals allows an functor to provide a list of Signals
func SetHotReloadSignals(sigfn func() []os.Signal) {
	sig.SetHotReloadSignals(sigfn)
}

// SetOnGetListener returns tcp/http listener for daemon hot-restarting
func SetOnGetListener(fn func() net.Listener) {
	sig.SetOnGetListener(fn)
}

// // QuitSignal return a channel for quit signal raising up.
// func QuitSignal() chan os.Signal {
// 	return sig.QuitSignal()
// }
//
// // SendNilSig sends the POSIX NUL signal
// func SendNilSig(process *os.Process) error {
// 	return sig.SendNilSig(process)
// }
//
// // SendHUP sends the POSIX HUP signal
// func SendHUP(process *os.Process) error {
// 	return sig.SendHUP(process)
// }
//
// // SendUSR1 sends the POSIX USR1 signal
// func SendUSR1(process *os.Process) error {
// 	return sig.SendUSR1(process)
// }
//
// // SendUSR2 sends the POSIX USR2 signal
// func SendUSR2(process *os.Process) error {
// 	return sig.SendUSR2(process)
// }
//
// // SendTERM sends the POSIX TERM signal
// func SendTERM(process *os.Process) error {
// 	return sig.SendTERM(process)
// }
//
// // SendQUIT sends the POSIX QUIT signal
// func SendQUIT(process *os.Process) error {
// 	return sig.SendQUIT(process)
// }
//
// // SendKILL sends the POSIX KILL signal
// func SendKILL(process *os.Process) error {
// 	return sig.SendKILL(process)
// }
//
//
// // ServeSignals calls handlers for system signals.
// // before invoking ServeSignals(), you should run SetupSignals() at first.
// func ServeSignals() (err error) {
// 	return sig.ServeSignals()
// }
//
// // HandleSignalCaughtEvent is a shortcut to block the main business logic loop but break it if os signals caught.
// // `stop` channel will be trigger if any hooked os signal caught, such as os.Interrupt;
// // the main business logic loop should trigger `done` once `stop` holds.
// func HandleSignalCaughtEvent() bool {
// 	return sig.HandleSignalCaughtEvent()
// }
