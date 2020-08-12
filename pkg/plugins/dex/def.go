/*
 * Copyright © 2019 Hedzr Yeh.
 */

package dex

import (
	"github.com/hedzr/cmdr"
	"net"
	"os"
)

// Daemon interface should be implemented when you are using `daemon.Enable()`.
type Daemon interface {
	// Config() (config *service.Config)

	// OnRun will be invoked when daemon being started, run/fork at foreground, hot reload ...
	OnRun(program *Program, stopCh, doneCh chan struct{}, hotReloadListener net.Listener) (err error)
	OnStop(program *Program) (err error)
	OnReload(program *Program)
	OnStatus(program *Program, p *os.Process) (err error)
	OnInstall(program *Program) (err error)
	OnUninstall(program *Program) (err error)

	// OnReadConfigFromCommandLine(root *cmdr.RootCommand)
	BeforeServiceStart(program *Program, root *cmdr.Command) (err error)
	AfterServiceStop(program *Program, root *cmdr.Command) (err error)
	OnCmdrPrepare(program *Program, root *cmdr.RootCommand) (err error)
}

// HotReloadable enables hot-restart/hot-reload feature
type HotReloadable interface {
	OnHotReload(program *Program) (err error)
}

var daemonImpl Daemon
