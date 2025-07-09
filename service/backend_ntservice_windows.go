package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hedzr/is"
	"github.com/hedzr/is/basics"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"

	"gopkg.in/hedzr/errors.v3"

	"github.com/hedzr/cmdr-addons/v2/service/systems"
	"github.com/hedzr/cmdr-addons/v2/tool/dbglog"
)

type ntServiceD struct {
	Logger    ZLogger
	once      sync.Once // initOnce
	inService bool
}

func (s *ntServiceD) Close() {
	if s.Logger != nil {
		if c, ok := s.Logger.(interface{ Close() error }); ok {
			_ = c.Close()
		} else if c, ok := s.Logger.(interface{ Close() }); ok {
			c.Close()
		}
	}
}

func (s *ntServiceD) Choose(ctx context.Context) (valid bool) {
	if systems.HasNTService {
		valid = hasNTServiceD(ctx)
	}
	return
}

func hasNTServiceD(ctx context.Context) bool {
	return true
}

func (s *ntServiceD) IsValid(ctx context.Context) (valid bool) {
	valid = systems.HasNTService && runtime.GOOS == "windows" && hasNTServiceD(ctx)
	return
}

func (s *ntServiceD) InService() bool { return s.inService }

var inTesting = is.InTesting()
var inBenching = isInBench()
var isDebugging = is.InDebugging()
var isDebug = is.DebugMode() || is.DebugBuild()

func isInBench() bool {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.bench") || strings.HasPrefix(arg, "-bench") {
			return true
		}
		// if strings.HasPrefix(arg, "-test.bench=") {
		// 	// ignore the benchmark name after an underscore
		// 	bench = strings.SplitN(arg[12:], "_", 2)[0]
		// 	break
		// }
	}
	return false
}

func (s *ntServiceD) initOnce(ctx context.Context, config *Config, m *mgmtS, cmd Command) (err error) {
	s.once.Do(func() {
		s.inService, _ = svc.IsWindowsService()

		if isDebug || isDebugging {
			// s.Logger = wlog(debug.New(svcName))
			s.Logger = dbglog.ZLogger()
		} else if cmd != Start && cmd != Restart && cmd != Stop {
			s.Logger = dbglog.ZLogger()
		} else {
			// var elog *eventlog.Log
			// elog, err = eventlog.Open(svcName)
			// if err != nil {
			// 	return
			// }
			//
			// errsCh := make(chan error, 1)
			// go func() {
			// 	defer close(errsCh)
			// 	defer func() {
			// 		// unregister syslog writer from our logz (logg/slog) containers
			// 		m.NotifyLoggerDestroying(s.Logger)
			// 	}()
			// 	for {
			// 		select {
			// 		case <-ctx.Done():
			// 			return
			// 		case e := <-errsCh:
			// 			m.errs.Attach(e)
			// 		}
			// 	}
			// }()
			// s.Logger = wlog(elog, errsCh)

			svcName := serviceName(config)
			s.Logger, err = makeWinLogger(ctx, svcName, m, s)
		}

		if s.Logger == nil {
			s.Logger = dbglog.ZLogger()
		}
	})
	return
}

func (s *ntServiceD) initSyslogger(ctx context.Context, config *Config, m *mgmtS, cmd Command) (err error) {
	if err = s.initOnce(ctx, config, m, cmd); err != nil {
		dbglog.ErrorContext(ctx, "[ntServiceD] logger bad", "err", err)
		return
	}

	err = s.Logger.Infof("\n\n-------------------- service %s: %s\n", config.ServiceName(), cmd)
	// dbglog.InfoContext(ctx, "[ntServiceD] logger ready", "err", err, "cmd", cmd)
	m.NotifyLoggerCreated(s.Logger) // register syslog writer to our logz (logg/slog) containers
	dbglog.OKContext(ctx, "a OKLevel message should be shawn.")
	return
}

func (s *ntServiceD) Control(ctx context.Context, config *Config, m *mgmtS, cmd Command) (err error) {
	if fn, ok := ntsCommands[cmd]; ok {
		if err = s.initSyslogger(ctx, config, m, cmd); err != nil {
			return
		}
		// dbglog.InfoContext(ctx, "[ntServiceD] logger ready", "fn", fn)
		return fn(ctx, config, m, s)
	}
	err = errors.New("unknown command %v (valid commands are in [%v, %v])", cmd, MinCommand+1, MaxCommand-1)
	return
}

var ntsCommands = map[Command]func(ctx context.Context, config *Config, m *mgmtS, s *ntServiceD) (err error){
	Info:    ntserviceInfo,
	Port:    ntservicePort,
	Addr:    ntserviceAddr,
	Start:   ntserviceStart,
	Stop:    ntserviceStop,
	Status:  ntserviceStatus,
	Restart: ntserviceRestart,

	Install:   ntserviceInstall,
	Uninstall: ntserviceUninstall,

	// HotReload: ntserviceHotReload,
	// Enable:    ntserviceEnable,
	// Disable:   ntserviceDisable,
	// ViewLog:   ntserviceViewLog,

	Pause:    ntservicePause,
	Continue: ntserviceContinue,
}

func ntserviceInfo(ctx context.Context, config *Config, m *mgmtS, s *ntServiceD) (err error) {
	if fn, ok := config.Entity.(EntityInfoAware); ok {
		println(fn.Info(ctx, config, s.Logger))
	}
	return
}

func ntservicePort(ctx context.Context, config *Config, m *mgmtS, s *ntServiceD) (err error) {
	if fn, ok := config.Entity.(EntityPortAware); ok {
		println(fn.Port(ctx, config, s.Logger))
	}
	return
}

func ntserviceAddr(ctx context.Context, config *Config, m *mgmtS, s *ntServiceD) (err error) {
	if fn, ok := config.Entity.(EntityAddrAware); ok {
		println(fn.Addr(ctx, config, s.Logger))
	}
	return
}

func ntserviceIsRunning(ctx context.Context, config *Config, m *mgmtS, s *ntServiceD) (err error) {
	if m.IsRunning() {
		return nil
	}
	return ErrServiceIsNotRunning
}

func ntserviceIsEnabled(ctx context.Context, config *Config, m *mgmtS, s *ntServiceD) (err error) {
	if m.IsRunning() {
		return nil
	}
	return ErrServiceIsNotEnabled
}

func serviceName(config *Config) string { return config.Name }

func controlService(name string, c svc.Cmd, to svc.State) (err error) {
	var m *mgr.Mgr
	m, err = mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	var s *mgr.Service
	s, err = m.OpenService(name)
	if err != nil {
		return fmt.Errorf("could not access service: %v", err)
	}
	defer s.Close()

	var status svc.Status
	status, err = s.Control(c)
	if err != nil {
		return fmt.Errorf("could not send control=%d: %v", c, err)
	}

	timeout := time.Now().Add(10 * time.Second)
	for status.State != to {
		if timeout.Before(time.Now()) {
			return fmt.Errorf("timeout waiting for service to go to state=%d", to)
		}
		time.Sleep(300 * time.Millisecond)
		status, err = s.Query()
		if err != nil {
			return fmt.Errorf("could not retrieve service status: %v", err)
		}
	}
	return
}

func ntserviceStart(ctx context.Context, config *Config, m *mgmtS, s *ntServiceD) (err error) {
	dbglog.DebugContext(ctx, "start")

	if ntserviceIsRunning(ctx, config, m, s) == nil {
		err = ErrServiceIsRunning
		s.Logger.Error("service ran already", "err", err)
		return
	}

	_ = s.Logger.Infof("command-line is %q\n", config.CmdLines)

	// s.Logger.Infof("store is:\n%s\n", cmdr.Store().Dump())
	_ = s.Logger.Infof("fore: %v, sMode: %v\n", m.fore, m.serviceMode)
	// cs := cmdr.ParsedLastCmd().Store()
	// _ = s.Logger.Infof("fore: %v, sMode: %v, user: %v", cs.MustBool("foreground"), cs.MustBool("service"), cs.MustBool("user"))

	// call into Entity.Start if exists
	if fn, ok := config.Entity.(EntityStartAware); ok {
		dbglog.DebugContext(ctx, "start EntityStartAware")
		_ = s.Logger.Infof("start EntityStartAware\n")
		return fn.Start(ctx, config, s.Logger)
	}

	if m.fore || s.inService {
		svcName := serviceName(config)
		err = runService(ctx, m, s, config, svcName, isDebug || isDebugging)
		return
	}

	startService := func(name string) error {
		m, err1 := mgr.Connect()
		if err1 != nil {
			return err1
		}
		defer m.Disconnect()
		s, err := m.OpenService(name)
		if err != nil {
			return fmt.Errorf("could not access service: %v", err)
		}
		defer s.Close()
		err = s.Start("is", "manual-started")
		if err != nil {
			return fmt.Errorf("could not start service: %v", err)
		}
		return nil
	}

	svcName := serviceName(config)
	err = startService(svcName)

	pid, ppid := os.Getpid(), os.Getppid()
	dbglog.DebugContext(ctx, "[NTService] ends.", "pid", pid, "ppid", ppid, "service", svcName)
	_ = s.Logger.Infof("[NTService] start %s done: pid=%v, ppid=%v\n", svcName, pid, ppid)
	return
}

func ntserviceStop(ctx context.Context, config *Config, m *mgmtS, s *ntServiceD) (err error) {
	dbglog.InfoContext(ctx, "stop")

	if fn, ok := config.Entity.(EntityStopAware); ok {
		return fn.Stop(ctx, config, s.Logger)
	}

	svcName := serviceName(config)
	err = controlService(svcName, svc.Stop, svc.Stopped)
	if err != nil {
		return
	}

	dbglog.InfoContext(ctx, "service stopped", "svcName", svcName)
	return
}

func ntserviceStatus(ctx context.Context, config *Config, m *mgmtS, s *ntServiceD) (err error) {
	if fn, ok := config.Entity.(EntityStatusAware); ok {
		return fn.Status(ctx, config, s.Logger)
	}

	// var retCode int
	// _, _, _, serviceName, file := serviceFilename(config)
	// if dir.FileExists(file) {
	// 	retCode, _, err = cmdrexec.Sudo("launchctl", "status", serviceName)
	// 	if err != nil || retCode != 0 {
	// 		return
	// 	}
	// }

	return
}

func ntserviceRestart(ctx context.Context, config *Config, m *mgmtS, s *ntServiceD) (err error) {
	if fn, ok := config.Entity.(EntityRestartAware); ok {
		return fn.Restart(ctx, config, s.Logger)
	}

	if ntserviceStop(ctx, config, m, s) == nil {
		time.Sleep(400 * time.Millisecond)

		if ntserviceStart(ctx, config, m, s) == nil {
			println("service was running, or installed.")
			return
		}
	}

	return
}

func ntserviceHotReload(ctx context.Context, config *Config, m *mgmtS, s *ntServiceD) (err error) {
	if fn, ok := config.Entity.(EntityHotReloadAware); ok {
		return fn.HotReload(ctx, config, s.Logger)
	}

	// var retCode int
	// _, _, _, serviceName, file := serviceFilename(config)
	// if dir.FileExists(file) {
	// 	retCode, _, err = cmdrexec.Sudo("launchctl", "reload", serviceName)
	// 	if err != nil || retCode != 0 {
	// 		return
	// 	}
	// }

	return
}

func ntserviceInstall(ctx context.Context, config *Config, m *mgmtS, s *ntServiceD) (err error) {
	if fn, ok := config.Entity.(EntityInstallAware); ok {
		return fn.Install(ctx, config, s.Logger)
	}

	if ntserviceIsRunning(ctx, config, m, s) == nil {
		println("service was running, or installed.")
		// dbglog.InfoContext(ctx, "[ntserviceInstall] service was running, or installed.")
		return
	}

	autoEnable := config.AutoEnable
	delayedAutoStart := config.DelayedAutoStart

	svcName := serviceName(config)
	dbglog.InfoContext(ctx, "[ntserviceInstall]", "serviceName", svcName, "autoEnable", autoEnable, "delayedAutoStart", delayedAutoStart)

	// _, err = adjustPrivilege(privilege, true)
	// if err != nil {
	// 	return err
	// }
	err = installService(svcName, config.DisplayName, config.Description, autoEnable, delayedAutoStart)
	if err == nil {
		println("\n", "\n", svcName, "service removed", "\n")
		// println(serviceName, "service created successfully.")
		// _ = s.Logger.Infof("Service created successfully.\n")
	}
	dbglog.ErrorContext(ctx, "[ntserviceInstall]", "err", err)
	return
}

func installService(name, displayName, desc string, autoEnable, delayedAutoStart bool) error {
	exepath, err1 := exePath()
	if err1 != nil {
		return err1
	}
	m, err2 := mgr.Connect()
	if err2 != nil {
		return err2
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %s already exists", name)
	}

	startType := mgr.StartManual
	if autoEnable {
		startType = mgr.StartAutomatic
	}
	s, err = m.CreateService(name, exepath, mgr.Config{
		DisplayName:      displayName,
		Description:      desc,
		StartType:        uint32(startType),
		DelayedAutoStart: delayedAutoStart,
	})
	if err != nil {
		return err
	}
	defer s.Close()

	err = eventlog.InstallAsEventCreate(name, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		s.Delete()
		return fmt.Errorf("SetupEventLogSource() failed: %s", err)
	}
	return nil
}

func exePath() (string, error) {
	prog := os.Args[0]
	p, err := filepath.Abs(prog)
	if err != nil {
		return "", err
	}
	fi, err := os.Stat(p)
	if err == nil {
		if !fi.Mode().IsDir() {
			return p, nil
		}
		err = fmt.Errorf("%s is directory", p)
	}
	if filepath.Ext(p) == "" {
		p += ".exe"
		fi, err := os.Stat(p)
		if err == nil {
			if !fi.Mode().IsDir() {
				return p, nil
			}
			err = fmt.Errorf("%s is directory", p)
		}
	}
	return "", err
}

func removeService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("service %s is not installed", name)
	}
	defer s.Close()
	err = s.Delete()
	if err != nil {
		return err
	}
	err = eventlog.Remove(name)
	if err != nil {
		return fmt.Errorf("RemoveEventLogSource() failed: %s", err)
	}
	return nil
}

func ntserviceUninstall(ctx context.Context, config *Config, m *mgmtS, s *ntServiceD) (err error) {
	if fn, ok := config.Entity.(EntityUninstallAware); ok {
		return fn.Uninstall(ctx, config, s.Logger)
	}

	_ = ntserviceStop(ctx, config, m, s)

	// if err = ntserviceDisable(ctx, config, m, s); err != nil {
	// 	return
	// }

	svcName := serviceName(config)
	dbglog.InfoContext(ctx, "removing service...", "svcName", svcName)
	err = removeService(svcName)
	if err != nil {
		return
	}
	println("\n", "\n", svcName, "service removed", "\n")
	return
}

func ntservicePause(ctx context.Context, config *Config, m *mgmtS, s *ntServiceD) (err error) {
	if fn, ok := config.Entity.(EntityPauseAware); ok {
		return fn.Pause(ctx, config, s.Logger)
	}

	svcName := serviceName(config)
	err = controlService(svcName, svc.Pause, svc.Paused)
	if err != nil {
		return
	}
	println(svcName, "service paused")
	return
}

func ntserviceContinue(ctx context.Context, config *Config, m *mgmtS, s *ntServiceD) (err error) {
	if fn, ok := config.Entity.(EntityContinueAware); ok {
		return fn.Continue(ctx, config, s.Logger)
	}

	svcName := serviceName(config)
	err = controlService(svcName, svc.Continue, svc.Running)
	if err != nil {
		return
	}
	println(svcName, "service resumed")
	return
}

func ntserviceEnable(ctx context.Context, config *Config, m *mgmtS, s *ntServiceD) (err error) {
	if fn, ok := config.Entity.(EntityEnableAware); ok {
		return fn.Enable(ctx, config, s.Logger)
	}

	if ntserviceIsEnabled(ctx, config, m, s) == nil {
		println("service has been enabled.")
		return
	}

	// var retCode int
	// var sid string
	// _, _, userLevel, serviceName, _ := serviceFilename(config)
	// if userLevel {
	// 	var currentUser *user.User
	// 	currentUser, err = user.Current()
	// 	if err != nil {
	// 		return
	// 	}
	//
	// 	sid = fmt.Sprintf("gui/%s/%s", currentUser.Uid, serviceName)
	// } else {
	// 	sid = fmt.Sprintf("%s/%s", "system", serviceName)
	// }
	// retCode, _, err = cmdrexec.RunWithOutput("launchctl", "enable", sid)
	// if err != nil || retCode != 0 {
	// 	return
	// }
	//
	// println(serviceName, "service has been enabled.")
	return
}

func ntserviceDisable(ctx context.Context, config *Config, m *mgmtS, s *ntServiceD) (err error) {
	if fn, ok := config.Entity.(EntityDisableAware); ok {
		return fn.Disable(ctx, config, s.Logger)
	}

	if ntserviceIsEnabled(ctx, config, m, s) != nil {
		println("service has not been enabled.")
		return
	}

	// var retCode int
	// _, _, _, serviceName, _ := serviceFilename(config)
	// retCode, _, err = cmdrexec.Sudo("launchctl", "disable", serviceName)
	// if err != nil || retCode != 0 {
	// 	return
	// }
	//
	// println(serviceName, "service has been disabled.")
	return
}

func ntserviceViewLog(ctx context.Context, config *Config, m *mgmtS, s *ntServiceD) (err error) {
	if fn, ok := config.Entity.(EntityViewLogAware); ok {
		return fn.ViewLog(ctx, config, s.Logger)
	}
	return
}

//

//

//

func runService(ctx context.Context, m *mgmtS, s *ntServiceD, config *Config, svcName string, isDebug bool) (err error) {
	// var err error
	//
	// if isDebug {
	// 	elog = debug.New(svcName)
	// } else {
	// 	elog, err = eventlog.Open(svcName)
	// 	if err != nil {
	// 		return
	// 	}
	// }
	// defer elog.Close()

	if !m.serviceMode {

	} //

	prog, ok := config.Entity.(RunnableService)
	if ok {
		prog.SetServiceMode(m.serviceMode)
		_ = s.Logger.Infof("run program...\n")
		err = prog.Run(ctx, config, s.Logger)
		if err != nil {
			return
		}
	}

	_ = s.Logger.Infof("starting %s service", svcName)
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	err = run(svcName, &serviceStub{s: s, config: config, done: ctx.Done()})
	if err != nil {
		_ = s.Logger.Errorf("%s service failed: %v", svcName, err)
		return
	}

	_ = s.Logger.Infof("%s service stopped", svcName)
	return
}

type serviceStub struct {
	s      *ntServiceD
	config *Config
	done   <-chan struct{}
	prog   RunnableService
}

func (m *serviceStub) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue

	changes <- svc.Status{State: svc.StartPending}

	fasttick := time.Tick(500 * time.Millisecond)
	slowtick := time.Tick(2 * time.Second)
	tick := fasttick

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	defer func() {
		changes <- svc.Status{State: svc.StopPending}
	}()

	firstCloser := func() { // an earlier closer to prefer to shut themselves down.
		m.s.Logger.Info("the exit signal received, shutting the prior entities down at first.")
		// shutting the program (config.Entity) down at first.
		// This would stop the http/2 server running in the program entity.
		if c, ok := m.prog.(basics.Peripheral); ok {
			c.Close()
		}
	}

	prog, ok := m.config.Entity.(RunnableService)
	if ok {
		// prog.SetServiceMode(m.serviceMode)
		_ = m.s.Logger.Infof("run program...\n")
		ctx := context.Background()
		err := prog.Run(ctx, m.config, m.s.Logger)
		if err != nil {
			return
		}
	}

	for {
		select {
		case <-tick:
			// beep()
			// m.s.Logger.Info("beep")

		case <-m.done:
			firstCloser()
			testOutput := strings.Join(args, "-")
			testOutput += " <-done (exit)"
			m.s.Logger.Info(testOutput)
			return

		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
				// Testing deadlock from https://code.google.com/p/winsvc/issues/detail?id=4
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				firstCloser()
				// golang.org/x/sys/windows/svc.TestExample is verifying this output.
				testOutput := strings.Join(args, "-")
				testOutput += fmt.Sprintf("-%d", c.Context)
				m.s.Logger.Info(testOutput)
				return
			case svc.Pause:
				changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
				tick = slowtick
			case svc.Continue:
				changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
				tick = fasttick
			default:
				m.s.Logger.Errorf("unexpected control request #%d", c)
			}
		}
	}
	return
}

// BUG(brainman): MessageBeep Windows api is broken on Windows 7,
// so this example does not beep when runs as service on Windows 7.

var (
	beepFunc = syscall.MustLoadDLL("user32.dll").MustFindProc("MessageBeep")
)

func beep() {
	beepFunc.Call(0xffffffff)
}
