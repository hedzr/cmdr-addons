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
	OnRun(prog *Program, stopCh, doneCh chan struct{}, hotReloadListener net.Listener) (err error)
	OnStop(prog *Program) (err error)
	OnReload(prog *Program)
	OnStatus(prog *Program, p *os.Process) (err error)
	OnInstall(prog *Program) (err error)
	OnUninstall(prog *Program) (err error)

	// OnReadConfigFromCommandLine(root *cmdr.RootCommand)
	BeforeServiceStart(prog *Program, root *cmdr.Command) (err error)
	AfterServiceStop(prog *Program, root *cmdr.Command) (err error)
	OnCmdrPrepare(prog *Program, root *cmdr.RootCommand) (err error)
}

// HotReloadable enables hot-restart/hot-reload feature
type HotReloadable interface {
	OnHotReload(prog *Program) (err error)
}

var daemonImpl Daemon
