/*
 * Copyright © 2019 Hedzr Yeh.
 */

package svr

import (
	"fmt"
	"github.com/hedzr/cmdr"
	"github.com/hedzr/cmdr-addons/pkg/plugins/dex"
	"github.com/kardianos/service"
	"golang.org/x/crypto/acme/autocert"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// NewDaemon creates an `daemon.Daemon` object
func NewDaemon(opts ...Opt) dex.Daemon {
	return NewDaemonWithConfig(&service.Config{
		Name:        "my-daemon",
		DisplayName: "My Daemon",
		Description: "My Daemon/Service here",
	}, opts...)
}

// NewDaemonWithConfig creates an `daemon.Daemon` object
func NewDaemonWithConfig(config *service.Config, opts ...Opt) dex.Daemon {
	d := &daemonImpl{
		// exit:   make(chan struct{}),
		config: config,
		Type:   typeGin,
	}

	for _, opt := range opts {
		opt(d)
	}
	return d
}

type Opt func(d *daemonImpl)

func WithBackendType(typ muxType) Opt {
	return func(d *daemonImpl) {
		d.Type = typ
	}
}

func WithRouterImpl(r RouterMux) Opt {
	return func(d *daemonImpl) {
		d.routerImpl = r
	}
}

type daemonImpl struct {
	config *service.Config
	// service service.Service
	// logger  service.Logger
	// cmd     *exec.Cmd
	// exit    chan struct{}

	appTag      string
	certManager *autocert.Manager
	Type        muxType
	mux         *http.ServeMux
	routerImpl  RouterMux
	// router      *gin.Engine
	// irisApp     *iris.Application
}

func (d *daemonImpl) Config() (config *service.Config) {
	if dio, ok := d.routerImpl.(interface{ Config() *service.Config }); ok {
		d.config = dio.Config()
	}
	return d.config
}

func (d *daemonImpl) RouterMux() RouterMux {
	return d.routerImpl
}

func (d *daemonImpl) OnRun(prg *dex.Program, stopCh, doneCh chan struct{}, hotReloadListener net.Listener) (err error) {
	serverType := cmdr.GetStringR("server.start.Server-Type")

	prg.Logger.Infof("demo daemon OnRun (Server-Type = %q), pid = %v, ppid = %v", serverType, os.Getpid(), os.Getppid())

	if serverType == "h2-server" {
		err = d.onRunHttp2Server(prg, stopCh, doneCh, hotReloadListener)
		if err == nil {
			err = d.enterLoop(prg, stopCh, doneCh, hotReloadListener)
		}
		return
	}

	worker(prg, stopCh, doneCh)
	return
}

func worker(prg *dex.Program, stopCh, doneCh chan struct{}) {
	fullExec, err := exec.LookPath("git")
	if err != nil {
		prg.Logger.Errorf("Failed to find executable %q: %v", "git --version", err)
	}

	var args []string = []string{"--version"}
	var env []string

	ticker := time.NewTicker(5 * time.Second)
	defer func() {
		ticker.Stop()
		if doneCh != nil && prg.InvokedInDaemon {
			doneCh <- struct{}{}
		}
	}()

LOOP:
	for {
		// time.Sleep(3 * time.Second) // this is work to be done by worker.
		select {
		case <-stopCh:
			break LOOP
		case tc := <-ticker.C:
			cmd := exec.Command(fullExec, args...)
			cmd.Dir = prg.WorkDirName()
			cmd.Env = append(os.Environ(), env...)
			cmd.Stdout, cmd.Stderr = prg.GetLogFileHandlers()

			pwd, _ := os.Getwd()
			prg.Logger.Infof("demo running at %d [dir: %q], inDaemon: %v, tick: %v, OS=%v\n", os.Getpid(), pwd, prg.InvokedInDaemon, tc, runtime.GOOS)
			_ = cmd.Run()
			cmd.Wait()
			if !prg.InvokedInDaemon {
				return
			}
		}
	}
}

func (*daemonImpl) OnStop(prg *dex.Program) (err error) {
	prg.Logger.Infof("demo daemon OnStop")
	return
}

func (*daemonImpl) OnReload(prg *dex.Program) {
	prg.Logger.Infof("demo daemon OnReload")
}

func (*daemonImpl) OnStatus(prg *dex.Program, p *os.Process) (err error) {
	fmt.Printf("%v v%v\n", prg.Command.GetRoot().AppName, prg.Command.GetRoot().Version)
	// fmt.Printf("PID=%v\nLOG=%v\n", cxt.PidFileName, cxt.LogFileName)
	return
}

func (*daemonImpl) OnInstall(prg *dex.Program) (err error) {
	prg.Logger.Infof("demo daemon OnInstall")
	return
	// panic("implement me")
}

func (*daemonImpl) OnUninstall(prg *dex.Program) (err error) {
	prg.Logger.Infof("demo daemon OnUninstall")
	return
	// panic("implement me")
}
