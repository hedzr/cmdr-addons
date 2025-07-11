package service

import (
	"context"
	"fmt"
	"log/syslog"
	"os"
	"os/user"
	"runtime"
	"sync"
	"text/template"
	"time"

	"github.com/hedzr/is"
	"github.com/hedzr/is/dir"
	"github.com/hedzr/is/exec"
	"gopkg.in/hedzr/errors.v3"

	"github.com/hedzr/cmdr-addons/service/v2/systems"
	"github.com/hedzr/cmdr-addons/v2/tool/dbglog"
	logz "github.com/hedzr/logg/slog"
)

type launchD struct {
	Logger ZLogger
}

func (s *launchD) Close() {
	if s.Logger != nil {
		if c, ok := s.Logger.(interface{ Close() error }); ok {
			_ = c.Close()
		} else if c, ok := s.Logger.(interface{ Close() }); ok {
			c.Close()
		}
	}
}

func (s *launchD) Choose(ctx context.Context) (ok bool) {
	if systems.HasLaunchd {
		ok = hasLaunchD(ctx)
	}
	return
}

func hasLaunchD(ctx context.Context) (valid bool) {
	retCode, _, err := exec.RunWithOutput("launchctl", "list", "com.apple.lsd")
	if err != nil || retCode != 0 {
		valid = false
	}
	return true
}

func (s *launchD) IsValid(ctx context.Context) (valid bool) {
	valid = systems.HasLaunchd && runtime.GOOS == "darwin" && hasLaunchD(ctx)
	return
}

func (s *launchD) initSyslogger(ctx context.Context, config *Config, m *mgmtS, cmd Command) (err error) {
	if s.Logger != nil {
		return
	}

	errsCh := make(chan error, 1)
	go func() {
		defer close(errsCh)
		defer func() {
			// unregister syslog writer from our logz (logg/slog) containers
			m.NotifyLoggerDestroying(s.Logger)
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case e := <-errsCh:
				m.errs.Attach(e)
			}
		}
	}()
	if s.Logger, err = newSysLogger(config.ServiceName(), errsCh, syslog.LOG_INFO); err != nil {
		return
	}

	err = s.Logger.Infof("\n\n-------------------- service %s: %s\n", config.ServiceName(), cmd)
	m.NotifyLoggerCreated(s.Logger) // register syslog writer to our logz (logg/slog) containers
	dbglog.OKContext(ctx, "a OKLevel message should be mapped as syslog's NOTICE message.")
	dbglog.InfoContext(ctx, "_", "level", dbglog.GetLevel())
	dbglog.Printf("level: %v", dbglog.GetLevel())
	return
}

func (s *launchD) Control(ctx context.Context, config *Config, m *mgmtS, cmd Command) (err error) {
	if systems.HasLaunchd {
		if fn, ok := launchdCommands[cmd]; ok {
			if m.serviceMode {
				if err = s.initSyslogger(ctx, config, m, cmd); err != nil {
					return
				}
				s.Logger.Infof("control(%s)", cmd)
				dbglog.Infof("control(%s)", cmd)
			} else {
				s.Logger = dbglog.ZLogger()
			}
			return fn(ctx, config, m, s)
		}
		err = errors.New("unknown command %v (valid commands are in [%v, %v])", cmd, MinCommand+1, MaxCommand-1)

	}
	return
}

var launchdCommands = map[Command]func(ctx context.Context, config *Config, m *mgmtS, s *launchD) (err error){
	Info:      launchdInfo,
	Port:      launchdPort,
	Addr:      launchdAddr,
	Start:     launchdStart,
	Stop:      launchdStop,
	Status:    launchdStatus,
	Restart:   launchdRestart,
	HotReload: launchdHotReload,
	Install:   launchdInstall,
	Uninstall: launchdUninstall,
	Enable:    launchdEnable,
	Disable:   launchdDisable,
	ViewLog:   launchdViewLog,
}

func launchdInfo(ctx context.Context, config *Config, m *mgmtS, s *launchD) (err error) {
	if fn, ok := config.Entity.(EntityInfoAware); ok {
		println(fn.Info(ctx, config, s.Logger))
	}
	return
}

func launchdPort(ctx context.Context, config *Config, m *mgmtS, s *launchD) (err error) {
	if fn, ok := config.Entity.(EntityPortAware); ok {
		println(fn.Port(ctx, config, s.Logger))
	}
	return
}

func launchdAddr(ctx context.Context, config *Config, m *mgmtS, s *launchD) (err error) {
	if fn, ok := config.Entity.(EntityAddrAware); ok {
		println(fn.Addr(ctx, config, s.Logger))
	}
	return
}

func launchdIsRunning(ctx context.Context, config *Config, m *mgmtS, s *launchD) (err error) {
	if m.IsRunning() {
		return nil
	}
	return ErrServiceIsNotRunning
}

func launchdIsEnabled(ctx context.Context, config *Config, m *mgmtS, s *launchD) (err error) {
	if m.IsRunning() {
		return nil
	}
	return ErrServiceIsNotEnabled
}

func launchdStart(ctx context.Context, config *Config, m *mgmtS, s *launchD) (err error) {
	dbglog.DebugContext(ctx, "start")

	if launchdIsRunning(ctx, config, m, s) == nil {
		err = ErrServiceIsRunning
		s.Logger.Errorf("service ran already, err: %v", err)
		return
	}

	_ = s.Logger.Infof("command-line is %q\n", config.CmdLines)

	// s.Logger.Infof("store is:\n%s\n", cmdr.Store().Dump())
	_ = s.Logger.Infof("fore: %v, sMode: %v, user: %v\n", m.fore, m.serviceMode, config.User)
	// cs := cmdr.Store()
	// if cmdr.Parsed() {
	// 	cs = cmdr.ParsedLastCmd().Store()
	// }
	// _ = s.Logger.Infof("fore: %v, sMode: %v, user: %v", cs.MustBool("foreground"), cs.MustBool("service"), cs.MustBool("user"))

	if m.fore {
		// run service at foreground
		if prog, ok := config.Entity.(RunnableService); ok {
			prog.SetServiceMode(m.serviceMode)
			_ = s.Logger.Infof("run program...\n")
			err = prog.Run(ctx, config, s.Logger)
		}
		return enterLoop(ctx, config, m, s)
	} else if fn, ok := config.Entity.(EntityStartAware); ok {
		// call into Entity.Start if exists
		dbglog.DebugContext(ctx, "start EntityStartAware")
		_ = s.Logger.Infof("start EntityStartAware\n")
		if err = fn.Start(ctx, config, s.Logger); err != nil {
			return
		}
		// return enterLoop(ctx, config, m, s)
		return
	}

	var retCode int
	_, _, _, serviceName, file := serviceFilename(config)
	if !dir.FileExists(file) && m.force {
		// create service file at first
		if err = launchdInstall(ctx, config, m, s); err != nil {
			return
		}
	}
	if dir.FileExists(file) {
		sn := serviceName
		_ = s.Logger.Infof("launchctl start %q with %q\n", sn, file)
		dbglog.Infof("launchctl start %q with %q\n", sn, file)
		// retCode, _, err = exec.Sudo("launchctl", "start", sn)
		retCode, _, err = exec.Sudo("launchctl", "load", file)
		if err != nil || retCode != 0 {
			dbglog.DebugContext(ctx, "`sudo launchctl start service` failed", "service", config.ServiceName(), "err", err)
			// cmdr.App().SetSuggestRetCode(retCode)
			config.RetCode = retCode
			return
		}
	}

	pid, ppid := os.Getpid(), os.Getppid()
	// dbglog.DebugContext(ctx, "'sudo launchctl start service' ends.", "pid", pid, "ppid", ppid, "service", config.ServiceName())
	_ = s.Logger.Infof("launchctl start %s done: pid=%v, ppid=%v\n", config.ServiceName(), pid, ppid)
	dbglog.Infof("launchctl start %s done: pid=%v, ppid=%v\n", config.ServiceName(), pid, ppid)

	// return enterLoop(ctx, config, m, s)
	return
}

func enterLoop(ctx context.Context, config *Config, m *mgmtS, s *launchD) (err error) {
	closeChan := make(chan struct{}, 8)
	defer func() { close(closeChan) }()

	pid, ppid := os.Getpid(), os.Getppid()

	catcher := is.Signals().Catch()
	catcher.WithOnSignalCaught(func(ctx context.Context, sig os.Signal, wgShutdown *sync.WaitGroup) {
		println()
		fmt.Printf("signal %q caught...\n", sig)
		closeChan <- struct{}{}
	}).WaitFor(ctx, func(ctx context.Context, closer func()) {
		defer func() {
			dbglog.InfoContext(ctx, "stop launchd service", "pid", pid, "ppid", ppid, "service", config.ServiceName())
			err = launchdStop(ctx, config, m, s)
			if err != nil {
				dbglog.ErrorContext(ctx, "stop launchd service failed", "pid", pid, "ppid", ppid, "service", config.ServiceName(), "err", err)
			}
			closer()
		}()
		dbglog.InfoContext(ctx, "entering loop", "service", config.ServiceName())
		for {
			select {
			case <-ctx.Done():
				return
			case <-closeChan:
				return
			}
		}
	})
	// for {
	// 	select {
	// 	case <-ctx.Done():
	// 		dbglog.InfoContext(ctx, "stop launchd service", "pid", pid, "ppid", ppid, "service", config.ServiceName())
	// 		err = launchdStop(ctx, config, m, s)
	// 		if err != nil {
	// 			dbglog.ErrorContext(ctx, "stop launchd service failed", "pid", pid, "ppid", ppid, "service", config.ServiceName(), "err", err)
	// 		}
	// 		return
	// 	}
	// }
	return
}

func launchdStop(ctx context.Context, config *Config, m *mgmtS, s *launchD) (err error) {
	dbglog.DebugContext(ctx, "stop")

	if fn, ok := config.Entity.(EntityStopAware); ok {
		return fn.Stop(ctx, config, s.Logger)
	}

	var retCode int
	_, _, _, serviceName, file := serviceFilename(config)
	if dir.FileExists(file) {
		sn := serviceName
		_ = s.Logger.Infof("launchctl stop %q with %q\n", sn, file)
		var msg string
		// retCode, msg, err = exec.Sudo("launchctl", "stop", sn)
		retCode, msg, err = exec.Sudo("launchctl", "unload", file)
		if err != nil || retCode != 0 {
			err = errors.New("failed to stop service. The console outputs are:\n%v", msg).WithErrors(err)
			return
		}
	}

	return
}

func launchdStatus(ctx context.Context, config *Config, m *mgmtS, s *launchD) (err error) {
	if fn, ok := config.Entity.(EntityStatusAware); ok {
		return fn.Status(ctx, config, s.Logger)
	}

	var retCode int
	_, _, _, serviceName, file := serviceFilename(config)
	if dir.FileExists(file) {
		var msg string
		retCode, msg, err = exec.Sudo("launchctl", "status", serviceName)
		if err != nil || retCode != 0 {
			err = errors.New("failed to status service. The console outputs are:\n%v", msg).WithErrors(err)
			return
		}
	}

	return
}

func launchdRestart(ctx context.Context, config *Config, m *mgmtS, s *launchD) (err error) {
	if fn, ok := config.Entity.(EntityRestartAware); ok {
		return fn.Restart(ctx, config, s.Logger)
	}

	var retCode int
	_, _, _, serviceName, file := serviceFilename(config)
	if dir.FileExists(file) {
		var msg string
		retCode, msg, err = exec.Sudo("launchctl", "stop", serviceName)
		if err != nil || retCode != 0 {
			err = errors.New("failed to stop service. The console outputs are:\n%v", msg).WithErrors(err)
			return
		}

		time.Sleep(333 * time.Millisecond)
		retCode, msg, err = exec.Sudo("launchctl", "stop", serviceName)
		if err != nil || retCode != 0 {
			err = errors.New("failed to (re)start service. The console outputs are:\n%v", msg).WithErrors(err)
			return
		}
	}

	return
}

func launchdHotReload(ctx context.Context, config *Config, m *mgmtS, s *launchD) (err error) {
	if fn, ok := config.Entity.(EntityHotReloadAware); ok {
		return fn.HotReload(ctx, config, s.Logger)
	}

	var retCode int
	_, _, _, serviceName, file := serviceFilename(config)
	if dir.FileExists(file) {
		var msg string
		retCode, msg, err = exec.Sudo("launchctl", "reload", serviceName)
		if err != nil || retCode != 0 {
			err = errors.New("failed to hot-reload service. The console outputs are:\n%v", msg).WithErrors(err)
			return
		}
	}

	return
}

func launchdInstall(ctx context.Context, config *Config, m *mgmtS, s *launchD) (err error) {
	if fn, ok := config.Entity.(EntityInstallAware); ok {
		return fn.Install(ctx, config, s.Logger)
	}

	if launchdIsRunning(ctx, config, m, s) == nil {
		println("service was running, or installed.")
		return
	}

	forceReinstall, autoEnable, _, serviceName, file := serviceFilename(config)
	if dir.FileExists(file) {
		if !forceReinstall && !m.force {
			msg := `Service had been installed already.
	
	If you wanna reinstall it, try this command line:
	
		$ {{.AppName}} {{.DadCommandsText}} install --force
	
	`
			// msg = cmdr.App().ParsedState().Translate(msg)
			if config.Translate != nil {
				msg = config.Translate(msg)
			}
			dbglog.Warn(msg, "service-name", serviceName)
			err = errors.New("service already installed")
			return
		}
	}

	err = createServiceFile(ctx, config, file, autoEnable)
	if err != nil {
		return
	}

	// refresh systemd
	var retCode int
	var msg string
	retCode, msg, err = exec.Sudo("launchctl", "load", file)
	if err != nil || retCode != 0 {
		err = errors.New("failed to install service. The console outputs are:\n%v", msg).WithErrors(err)
		return
	}

	if autoEnable {
		err = launchdEnable(ctx, config, m, s)
	}

	if err == nil {
		println("[SUCCESS]", serviceName, "service created successfully.")
		// _ = s.Logger.Infof("Service created successfully.\n")
	}
	return
}

func createServiceFile(ctx context.Context, config *Config, file string, autoLoad bool) (err error) {
	var tmpFile *os.File
	if tmpFile, err = dir.TempFile("", dir.Basename(file)); err != nil {
		return
	}

	defer func() {
		if err == nil {
			_ = tmpFile.Close()
			// dir.DeleteFile(tmpFile.Name())
			_, _, err = exec.Sudo("mv", tmpFile.Name(), file)
			if err == nil {
				logz.DebugContext(ctx, "moved service file ok.", "target", file)
				// ctt, _ := os.ReadFile(file)
				// logz.DebugContext(ctx, "service file content", "content", string(ctt))
			} else {
				logz.ErrorContext(ctx, "moved service file failed.", "target", file, "err", err)
			}
		}
	}()

	var tmpl *template.Template
	tmplFile := fmt.Sprintf("%v/share/service.darwin.tpl", config.TemplateDir)
	if dir.FileExists(tmplFile) {
		tmpl, err = template.New("service.file").ParseFiles(tmplFile)
	} else {
		tmpl, err = template.New("service.file").Parse(tplLaunchdService)
	}
	if err != nil {
		return
	}

	if err = tmpl.Execute(tmpFile, struct {
		*Config
		AutoLoad bool
	}{config,
		autoLoad,
	}); err != nil {
		return
	}
	return
}

func serviceFilename(config *Config) (forceReinstall, autoEnable, userLevel bool, serviceName, filename string) {
	forceReinstall = config.ForceReinstall
	autoEnable = config.AutoEnable
	userLevel = config.UserLevel
	vendorDomainPrefix := config.VendorDomainPrefix

	targetDir := systemAgents
	if userLevel {
		targetDir = userAgents
	}

	serviceName = fmt.Sprintf("%s%s", vendorDomainPrefix, config.Name)
	filename = os.ExpandEnv(fmt.Sprintf("%s/%s%s.plist", targetDir, vendorDomainPrefix, config.Name))
	return
}

func launchdUninstall(ctx context.Context, config *Config, m *mgmtS, s *launchD) (err error) {
	if fn, ok := config.Entity.(EntityUninstallAware); ok {
		return fn.Uninstall(ctx, config, s.Logger)
	}

	if err = launchdStop(ctx, config, m, s); err != nil {
		dbglog.WarnContext(ctx, "Failed to stop service: "+err.Error())
		// return
	}

	if err = launchdDisable(ctx, config, m, s); err != nil {
		dbglog.WarnContext(ctx, "Failed to disable service: "+err.Error())
		// return
	}

	var retCode int
	anyExist := false
	_, _, userLevel, serviceName, file := serviceFilename(config)
	if dir.FileExists(file) {
		dbglog.InfoContext(ctx, "uninstalling service file: "+file)
		var msg string
		if userLevel {
			retCode, msg, err = exec.RunWithOutput("launchctl", "unload", file)
		} else {
			retCode, msg, err = exec.Sudo("launchctl", "unload", file)
		}
		if err != nil || retCode != 0 {
			err = errors.New("failed to uninstall service. The console outputs are:\n%v", msg).WithErrors(err)
			return
		}

		retCode, msg, err = exec.Sudo("mv", file, os.TempDir())
		if err != nil || retCode != 0 {
			err = errors.New("failed to mv service file to trashbin. The console outputs are:\n%v", msg).WithErrors(err)
			return
		}
		anyExist = true
	}

	if !anyExist {
		println("nothing needs to be done.")
		return
	}

	println(serviceName, "[SUCCESS] service uninstalled.")
	return
}

func launchdEnable(ctx context.Context, config *Config, m *mgmtS, s *launchD) (err error) {
	if fn, ok := config.Entity.(EntityEnableAware); ok {
		return fn.Enable(ctx, config, s.Logger)
	}

	if launchdIsEnabled(ctx, config, m, s) == nil {
		println("service has been enabled.")
		return
	}

	var retCode int
	var sid string
	var msg string
	_, _, userLevel, serviceName, _ := serviceFilename(config)
	if userLevel {
		var currentUser *user.User
		currentUser, err = user.Current()
		if err != nil {
			return
		}

		sid = fmt.Sprintf("gui/%s/%s", currentUser.Uid, serviceName)
	} else {
		sid = fmt.Sprintf("%s/%s", "system", serviceName)
	}
	retCode, msg, err = exec.RunWithOutput("launchctl", "enable", sid)
	if err != nil || retCode != 0 {
		err = errors.New("failed to enable service. The console outputs are:\n%v", msg).WithErrors(err)
		return
	}

	println(serviceName, "service has been enabled.")
	return
}

func launchdDisable(ctx context.Context, config *Config, m *mgmtS, s *launchD) (err error) {
	if fn, ok := config.Entity.(EntityDisableAware); ok {
		return fn.Disable(ctx, config, s.Logger)
	}

	if launchdIsEnabled(ctx, config, m, s) != nil {
		println("service has not been enabled.")
		return
	}

	var retCode int
	var msg string
	_, _, _, serviceName, _ := serviceFilename(config)
	retCode, msg, err = exec.Sudo("launchctl", "disable", serviceName)
	if err != nil || retCode != 0 {
		err = errors.New("failed to disable service. The console outputs are:\n%v", msg).WithErrors(err)
		return
	}

	println(serviceName, "service has been disabled.")
	return
}

func launchdViewLog(ctx context.Context, config *Config, m *mgmtS, s *launchD) (err error) {
	if fn, ok := config.Entity.(EntityViewLogAware); ok {
		return fn.ViewLog(ctx, config, s.Logger)
	}
	return
}

const (
	tplLaunchdService = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>{{.Name}}</string>
    <key>Program</key>
    <string>{{.ExecutablePath}}</string>
    <key>ProgramArguments</key>
    <array>
        {{range $k, $v := .ArgsForInstall -}}
        <string>{{$v}}</string>
        {{end -}}
    </array>
    <key>KeepAlive</key>
    <true/>
    <key>RunAtLoad</key>
	{{if .AutoLoad}}
	<true/>
	{{else}}
    <false/>
	{{end}}
    <key>OnDemand</key>
    <true/>
    <key>LaunchOnlyOnce</key>
    <false/>
    <key>StandardOutPath</key>
    <string>{{.StandardOutPath}}</string>
    <key>StandardErrorPath</key>
    <string>{{.StandardErrorPath}}</string>
</dict>
</plist>`

	systemDaemons = "/Library/LaunchDaemons"
	systemAgents  = "/Library/LaunchAgents"
	userAgents    = "$HOME/Library/LaunchAgents"
)
