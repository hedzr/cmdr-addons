// Copyright © 2020 Hedzr Yeh.

package dex

import (
	"fmt"
	"github.com/hedzr/cmdr"
	"github.com/hedzr/log"
	"github.com/hedzr/log/dir"
	"github.com/kardianos/service"
	"os"
	"path"
	"runtime"
)

// WithDaemon enables daemon plugin:
// - add daemon commands and sub-commands: start/run, stop, restart/reload, status, install/uninstall
// - pidfile
// -
func WithDaemon(daemonImplObject Daemon,
	// modifier func(daemonServerCommand *cmdr.Command) *cmdr.Command,
	opts ...Opt,
) cmdr.ExecOption {

	pd = &Program{
		Config: &service.Config{
			Name:        "the-daemon",
			DisplayName: "The Daemon",
			Description: "The Daemon/Service here",
		},
		daemon: daemonImplObject,
		log:    log.NewDummyLogger(),
		// Service: nil,
		// Logger:  nil,
		// Command: nil,
		// exit:    make(chan struct{}),
		// done:    make(chan struct{}),
	}

	if dio, ok := daemonImplObject.(interface{ Config() *service.Config }); ok {
		pd.Config = dio.Config()
	}

	// set appname with daemon name
	os.Setenv("APPNAME", pd.Config.Name)
	// pd.log.Printf("set appname with daemon name: %q", pd.Config.Name)

	if len(pd.Config.Arguments) == 0 {
		pd.Config.Arguments = []string{"server", "run", "--in-daemon"}
	}
	if pd.Config.Option == nil {
		pd.Config.Option = make(service.KeyValue)
	}

	for _, opt := range opts {
		opt()
	}

	return func(w *cmdr.ExecWorker) {
		w.AddOnBeforeXrefBuilding(func(root *cmdr.RootCommand, args []string) {

			if pd.modifier != nil {
				root.SubCommands = append(root.SubCommands, pd.modifier(DaemonServerCommand))
			} else {
				root.SubCommands = append(root.SubCommands, DaemonServerCommand)
			}

			// prefix = strings.Join(append(cmdr.RxxtPrefix, "server"), ".")
			// prefix = "server"

			attachPreAction(root, pd.preActions...)
			attachPostAction(root, pd.postActions...)

			if err := cmdrPrepare(pd.daemon, root); err != nil {
				pd.log.Fatalf("%v", err)
			}

		})
	}
}

func attachPostAction(root *cmdr.RootCommand, postActions ...func(cmd *cmdr.Command, args []string)) {
	if root.PostAction != nil {
		savedPostAction := root.PostAction
		root.PostAction = func(cmd *cmdr.Command, args []string) {
			for _, postAction := range postActions {
				if postAction != nil {
					postAction(cmd, args)
				}
			}
			// pidfile.Destroy()
			savedPostAction(cmd, args)
			return
		}
	} else {
		root.PostAction = func(cmd *cmdr.Command, args []string) {
			for _, postAction := range postActions {
				if postAction != nil {
					postAction(cmd, args)
				}
			}
			// pidfile.Destroy()
			return
		}
	}
}

func attachPreAction(root *cmdr.RootCommand, preActions ...func(cmd *cmdr.Command, args []string) (err error)) {
	if root.PreAction != nil {
		savedPreAction := root.PreAction
		root.PreAction = func(cmd *cmdr.Command, args []string) (err error) {
			// pidfile.Create(cmd)
			logger.Setup(cmd)

			if err := prepare(pd.daemon, root); err != nil {
				pd.log.Fatalf("%v", err)
			}

			if err = savedPreAction(cmd, args); err != nil {
				return
			}
			for _, preAction := range preActions {
				if preAction != nil {
					err = preAction(cmd, args)
				}
			}
			return
		}
	} else {
		root.PreAction = func(cmd *cmdr.Command, args []string) (err error) {
			// pidfile.Create(cmd)
			logger.Setup(cmd)

			if err := prepare(pd.daemon, root); err != nil {
				pd.log.Fatalf("%v", err)
			}

			for _, preAction := range preActions {
				if preAction != nil {
					err = preAction(cmd, args)
				}
			}
			return
		}
	}
}

const systemdScript = `[Unit]
Description={{.Description}}
# Documentation=man:sshd(8) man:sshd_config(5) man:{{.Name}}(1)
ConditionFileIsExecutable={{.Path|cmdEscape}}
{{range $i, $dep := .Dependencies}} 
{{$dep}} {{end}}

[Service]
LimitNOFILE=65535
StartLimitInterval=5
StartLimitBurst=10
ExecStart={{.Path|cmdEscape}} $GLOBAL_OPTIONS {{range .Arguments}} {{.|cmd}}{{end}}
{{if .Config.Option.stopArgs}}ExecStop={{.Path|cmdEscape}} $GLOBAL_OPTIONS {{.Config.Option.stopArgs}}{{end}}
{{if .Config.Option.reloadArgs}}ExecReload={{.Path|cmdEscape}} $GLOBAL_OPTIONS {{.Config.Option.reloadArgs}}{{end}}
#ExecReload=/bin/kill -HUP $MAINPID
{{if .ChRoot}}RootDirectory={{.ChRoot|cmd}}{{end}}
{{if .WorkingDirectory}}WorkingDirectory={{.WorkingDirectory|cmdEscape}}{{end}}
{{if .UserName}}User={{.UserName}}{{end}}
{{if .Config.Option.GroupName}}Group={{.Config.Option.GroupName}}{{end}}
{{if .ReloadSignal}}ExecReload=/bin/kill -{{.ReloadSignal}} "$MAINPID"{{end}}
{{if .PIDFile}}#PIDFile={{.PIDFile|cmd}}{{end}}
{{if .PIDFile}}PIDFile={{.PIDFile}}{{end}}
{{if and .LogOutput .HasOutputFileSupport -}}
# Works only in systemd v240 and newer! And, it can't work in most cases.
StandardOutput=append:/var/log/{{.Name}}/{{.Name}}.out
StandardError=append:/var/log/{{.Name}}/{{.Name}}.err
{{- end}}
{{if .LogOutput -}}{{if not .HasOutputFileSupport -}}
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier={{.AppName}}
{{- end}}{{- end}}
{{if .Restart}}Restart={{.Restart}}{{end}}
{{if .SuccessExitStatus}}SuccessExitStatus={{.SuccessExitStatus}}{{end}}
RestartSec=120
EnvironmentFile=-{{.Config.Option.envFileName}}

[Install]
WantedBy=multi-user.target
`

func cmdrPrepare(daemonImplObject Daemon, cmd *cmdr.RootCommand) (err error) {
	err = pd.daemon.OnCmdrPrepare(pd, cmd)
	return
}

func prepare(daemonImplObject Daemon, cmd *cmdr.RootCommand) (err error) {

	// set appname with daemon name
	os.Setenv("APPNAME", pd.Config.Name)
	pd.log.Printf("set appname with daemon name: %q", pd.Config.Name)

	err = pd.prepareAppDirs()
	if err != nil {
		pd.log.Fatalf("Cannot prepare the app directories: %+v", err)
		return
	}

	if len(pd.Config.WorkingDirectory) == 0 {
		workDir := path.Join("/var/lib", pd.Config.Name)
		if dir.FileExists(workDir) {
			pd.Config.WorkingDirectory = workDir
		} else {
			pd.Config.WorkingDirectory = dir.GetExecutableDir()
		}
	}

	if runtime.GOOS != "windows" {
		pd.Config.Option["PIDFile"] = pd.PidFileName()

		logDir := path.Dir(pd.LogStdoutFileName()) // "/var/log" // path.Join("/var", "log") // , conf.AppName)
		if dir.FileExists(logDir) {
			// logFile := path.Join(logDir, conf.AppName, ".out")
			// errFile := path.Join(logDir, conf.AppName, ".err")
			pd.Config.Option["LogOutput"] = true

			err = pd.prepareLogFiles()
			if err != nil {
				pd.log.Fatalf("Cannot prepare the logging files: %+v", err)
				return
			}

			if pd.ForwardLogToFile {
				pd.log.Debugf("All logging output will be forwarded to this file: %q.", pd.LogStdoutFileName())
				// TODO logrus.SetOutput(pd.fOut)
			}
		}

		pd.Config.Option["envFileName"] = pd.EnvFileName()

		pd.Config.Option["SystemdScript"] = systemdScript
	}

	// pd.Config.Option["stopArgs"] = "server stop -3"
	// pd.Config.Option["reloadArgs"] = "server reload"

	pd.Service, err = service.New(pd, pd.Config)
	if err != nil {
		return
	}

	errs := make(chan error, 5)
	pd.Logger, err = pd.Service.Logger(errs)
	if err != nil {
		return
	}

	if cmdr.GetBoolRP("server.start", "in-daemon") {
		sl := log.FromSystemdLogger(pd.Logger)
		log.SetLogger(sl)
		cmdr.SetLogger(sl)
		pd.Logger.Infof("daemonStart: log-level: %v", cmdr.GetLoggerLevel())
		sl.SetLevel(log.DebugLevel)
		log.Debugf("hello")
		cmdr.Logger.Debugf("hello [cmdr]")
	}

	// pd.daemon.OnReadConfigFromCommandLine(root)

	if err = pd.daemon.BeforeServiceStart(pd, &cmd.Command); err != nil {
		return
	}

	go func() {
		for {
			err := <-errs
			if err != nil {
				pd.log.Errorf("error: %v", err)
			}
		}
	}()

	// pd.Logger.Info("daemonex prepared.")
	log.Trace("daemonex prepared.")
	return
}

func daemonStart(cmd *cmdr.Command, args []string) (err error) {
	pd.Command, pd.Args = cmd, args
	pd.InvokedInDaemon = cmdr.GetBoolRP("server.start", "in-daemon")
	hotReloading := cmdr.GetBoolRP("server.start", "in-hot-reload")
	foreground := cmdr.GetBoolRP("server.start", "foreground")
	pd.Logger.Infof("daemonStart: foreground: %v, in-daemon: %v, hot-reload: %v, hit: %v", foreground, pd.InvokedInDaemon, hotReloading, cmd.GetHitStr())

	cmdr.Set("in-daemon", pd.InvokedInDaemon)

	// ctx := impl.GetContext(Command, Args, daemonImpl, onHotReloading)
	if hotReloading {
		err = daemonHotReload(cmd, args)
		return
	}

	if foreground || cmd.GetHitStr() == "run" {
		pd.InvokedDirectly = true
		err = run(cmd, args)
	} else {
		err = runAsDaemon(cmd, args)
	}

	pd.Logger.Info("daemonStart END.")
	if err != nil {
		pd.Logger.Errorf("  >> with error: %v", err)
	}
	if pd.InvokedInDaemon && runtime.GOOS == "windows" {
		err = pd.Service.Run()
	}
	return
}

func runAsDaemon(cmd *cmdr.Command, args []string) (err error) {
	err = service.Control(pd.Service, "start")
	if err != nil {
		pd.Logger.Errorf("Valid actions %q: %v\n", service.ControlAction, err)
		return // log.Fatal(err)
	}
	return
}

func run(cmd *cmdr.Command, args []string) (err error) {
	if runtime.GOOS != "windows" {
		pd.run()
		err = pd.err
	}

	// defer func() {
	// 	if service.Interactive() {
	// 		err = pd.Stop(pd.service)
	// 	} else {
	// 		err = pd.service.Stop()
	// 	}
	// }()
	// err = pd.daemon.OnRun(Command, Args, nil, nil, nil)
	return
}

func daemonStop(cmd *cmdr.Command, args []string) (err error) {
	pd.Command, pd.Args = cmd, args
	err = service.Control(pd.Service, "stop")
	if err != nil {
		pd.Logger.Errorf("Valid actions are %q, cannot 'stop' the service: %v\n", service.ControlAction, err)
		// return // log.Fatal(err)
	}

	if err = pd.daemon.AfterServiceStop(pd, cmd); err != nil {
		return
	}

	return
}

func daemonRestart(cmd *cmdr.Command, args []string) (err error) {
	pd.Command, pd.Args = cmd, args
	err = service.Control(pd.Service, "restart")
	if err != nil {
		pd.Logger.Errorf("Valid actions %q: %v\n", service.ControlAction, err)
		return // log.Fatal(err)
	}

	// getContext(Command, Args)
	//
	// p, err := daemonCtx.Search()
	// if err != nil {
	// 	fmt.Printf("%v is stopped.\n", Command.GetRoot().AppName)
	// } else {
	// 	if err = p.Signal(syscall.SIGHUP); err != nil {
	// 		return
	// 	}
	// }

	// ctx := impl.GetContext(Command, Args, daemonImpl, onHotReloading)
	// impl.Reload(Command.GetRoot().AppName, ctx)
	return
}

func daemonHotReload(cmd *cmdr.Command, args []string) (err error) {
	pd.Command, pd.Args = cmd, args
	// ctx := impl.GetContext(Command, Args, daemonImpl, onHotReloading)
	// impl.HotReload(Command.GetRoot().AppName, ctx)

	if runtime.GOOS == "linux" {
		// TODO impl hot reloading for linux
	}
	return
}

func onHotReloading() (err error) {
	if hr, ok := pd.Service.(HotReloadable); ok {
		err = hr.OnHotReload(pd)
	}
	return
}

func daemonStatus(cmd *cmdr.Command, args []string) (err error) {
	pd.Command, pd.Args = cmd, args
	pd.Logger.Infof("Args: %v", args)
	// err = service.Control(pd.service, "status")
	// if err != nil {
	// 	pd.log.Errorf("Valid actions: %q\n", service.ControlAction)
	// 	return // log.Fatal(err)
	// }
	var st service.Status
	st, err = pd.Service.Status()
	var sst string
	switch st {
	case service.StatusStopped:
		sst = "Stopped"
	case service.StatusRunning:
		sst = "Running"
	default:
		if err == service.ErrNotInstalled {
			sst = "Not Installed"
		} else {
			sst = "Unknown"
		}
	}
	fmt.Printf("Status: %v\n", sst)
	if pd.daemon != nil {
		err = pd.daemon.OnStatus(pd, nil)
	}

	// getContext(Command, Args)
	//
	// p, err := daemonCtx.Search()
	// if err != nil {
	// 	fmt.Printf("%v is stopped.\n", Command.GetRoot().AppName)
	// } else {
	// 	fmt.Printf("%v is running as %v.\n", Command.GetRoot().AppName, p.Pid)
	// }
	//
	// if daemonImpl != nil {
	// 	err = daemonImpl.OnStatus(&Context{Context: *daemonCtx}, Command, p)
	// }

	// ctx := impl.GetContext(Command, Args, daemonImpl, onHotReloading)
	// present, process := impl.FindDaemonProcess(ctx)
	// if present && daemonImpl != nil {
	// 	err = daemonImpl.OnStatus(ctx, Command, process)
	// }
	return
}

func daemonInstall(cmd *cmdr.Command, args []string) (err error) {
	pd.Command, pd.Args = cmd, args
	pd.Config.UserName = cmdr.GetStringRP(cmd.GetDottedNamePath(), "user")
	pd.Config.Option["GroupName"] = cmdr.GetStringRP(cmd.GetDottedNamePath(), "group")
	err = service.Control(pd.Service, "install")
	if err != nil {
		pd.Logger.Errorf("Valid actions %q: %v\n", service.ControlAction, err)
		return // log.Fatal(err)
	}

	// ctx := impl.GetContext(Command, Args, daemonImpl, onHotReloading)
	//
	// err = runInstaller(Command, Args)
	// if err != nil {
	// 	return
	// }
	// if daemonImpl != nil {
	// 	err = daemonImpl.OnInstall(ctx /*&Context{Context: *daemonCtx}*/, Command, Args)
	// }
	return
}

func daemonUninstall(cmd *cmdr.Command, args []string) (err error) {
	pd.Command, pd.Args = cmd, args
	err = service.Control(pd.Service, "uninstall")
	if err != nil {
		pd.Logger.Errorf("Valid actions %q: %v\n", service.ControlAction, err)
		return // log.Fatal(err)
	}

	// ctx := impl.GetContext(Command, Args, daemonImpl, onHotReloading)
	//
	// err = runUninstaller(Command, Args)
	// if err != nil {
	// 	return
	// }
	// if daemonImpl != nil {
	// 	err = daemonImpl.OnUninstall(ctx /*&Context{Context: *daemonCtx}*/, Command, Args)
	// }
	return
}

//
//
//
