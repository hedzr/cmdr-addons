package service

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"log/syslog"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/hedzr/is/dir"
	cmdrexec "github.com/hedzr/is/exec"
	"gopkg.in/hedzr/errors.v3"

	"github.com/hedzr/cmdr-addons/v2/service/systems"
	"github.com/hedzr/cmdr-addons/v2/tool/dbglog"
)

func dummy0() { dbglog.Debug("debug msg") }
func dummy1() { dbglog.Debug("debug msg") }
func dummy2() { slog.Debug("debug msg") }

type systemD struct {
	Logger ZLogger
}

func (s *systemD) Close() {
	if s.Logger != nil {
		if c, ok := s.Logger.(interface{ Close() error }); ok {
			_ = c.Close()
		} else if c, ok := s.Logger.(interface{ Close() }); ok {
			c.Close()
		}
	}
}

func (s *systemD) Choose(ctx context.Context) (ok bool) {
	if systems.HasLinuxBackends {
		dbglog.VerboseContext(ctx, "systemD.Choose: checking hasSystemd")
		ok = hasSystemd(ctx)
		dbglog.VerboseContext(ctx, "systemD.Choose: checking hasSystemd", "ok", ok)
	}
	return
}

func (s *systemD) IsValid(ctx context.Context) (valid bool) {
	if systems.HasLinuxBackends {
		valid = hasSystemd(ctx)
	}

	if valid {
		// systemd self-assertion
		retCode, _, err := cmdrexec.RunWithOutput("systemd-analyze")
		if err != nil || retCode != 0 {
			valid = false
		}

		// refresh services
		retCode, _, err = cmdrexec.Sudo("systemctl", "daemon-reload")
		if err != nil || retCode != 0 {
			valid = false
		}
	}
	return
}

func hasSystemd(ctx context.Context) bool {
	if _, err := os.Stat("/run/systemd/system"); err == nil {
		return true
	}
	if _, err := cmdrexec.LookPath("systemctl"); err != nil {
		return false
	}
	if _, err := os.Stat("/proc/1/comm"); err == nil {
		// https://superuser.com/questions/1017959/how-to-know-if-i-am-using-systemd-on-linux
		var f *os.File
		f, err = os.Open("/proc/1/comm")
		if err != nil {
			return false
		}
		defer f.Close()

		var buf bytes.Buffer
		_, err = buf.ReadFrom(f)
		contents := buf.String()

		if strings.Trim(contents, " \r\n") == "systemd" {
			return true
		}
	}
	return false
}

func (s *systemD) initSyslogger(ctx context.Context, config *Config, m *mgmtS, cmd Command) (err error) {
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
	return
}

func (s *systemD) Control(ctx context.Context, config *Config, m *mgmtS, cmd Command) (err error) {
	if fn, ok := systemdCommands[cmd]; ok {
		if err = s.initSyslogger(ctx, config, m, cmd); err != nil {
			return
		}
		return fn(ctx, config, m, s)
	}
	err = errors.New("unknown command %v (valid commands are in [%v, %v])", cmd, MinCommand+1, MaxCommand-1)
	return
}

var systemdCommands = map[Command]func(ctx context.Context, config *Config, m *mgmtS, s *systemD) (err error){
	Info:      systemdInfo,
	Port:      systemdPort,
	Addr:      systemdAddr,
	Start:     systemdStart,
	Stop:      systemdStop,
	Status:    systemdStatus,
	Restart:   systemdRestart,
	HotReload: systemdHotReload,
	Install:   systemdInstall,
	Uninstall: systemdUninstall,
	Enable:    systemdEnable,
	Disable:   systemdDisable,
	ViewLog:   systemdViewLog,
}

func systemdInfo(ctx context.Context, config *Config, m *mgmtS, s *systemD) (err error) {
	if fn, ok := config.Entity.(EntityInfoAware); ok {
		println(fn.Info(ctx, config, s.Logger))
	}
	return
}

func systemdPort(ctx context.Context, config *Config, m *mgmtS, s *systemD) (err error) {
	if fn, ok := config.Entity.(EntityPortAware); ok {
		println(fn.Port(ctx, config, s.Logger))
	}
	return
}

func systemdAddr(ctx context.Context, config *Config, m *mgmtS, s *systemD) (err error) {
	if fn, ok := config.Entity.(EntityAddrAware); ok {
		println(fn.Addr(ctx, config, s.Logger))
	}
	return
}

func systemdStart(ctx context.Context, config *Config, m *mgmtS, s *systemD) (err error) {
	dbglog.DebugContext(ctx, "start")

	if !hasSystemd(ctx) {
		return errors.Unavailable
	}

	if systemdIsRunning(ctx, config, m, s) == nil {
		err = ErrServiceIsRunning
		s.Logger.Error("service ran already", "err", err)
		return
	}

	_ = s.Logger.Infof("command-line is %q\n", config.CmdLines)
	// s.Logger.Infof("store is:\n%s\n", cmdr.Store().Dump())
	_ = s.Logger.Infof("fore: %v, sMode: %v\n", m.fore, m.serviceMode)
	// cs := cmdr.Store().WithPrefix("server.start")
	// _ = s.Logger.Infof("fore: %v, sMode: %v, user: %v", cs.MustBool("foreground"), cs.MustBool("service"), cs.MustBool("user"))

	// call into Entity.Start if exists
	if fn, ok := config.Entity.(EntityStartAware); ok {
		dbglog.DebugContext(ctx, "start EntityStartAware")
		_ = s.Logger.Infof("start EntityStartAware\n")
		return fn.Start(ctx, config, s.Logger)
	}

	if m.fore {
		if prog, ok := config.Entity.(RunnableService); ok {
			prog.SetServiceMode(m.serviceMode)
			_ = s.Logger.Infof("run program...\n")
			err = prog.Run(ctx, config, s.Logger)
		}
		return
	}

	// or, call systemd command to trigger the real serivce starting
	var retCode int
	var msg string
	_ = s.Logger.Infof("systemctl start %s\n", config.ServiceName())
	retCode, msg, err = cmdrexec.Sudo("systemctl", "start", config.ServiceName())
	if err != nil || retCode != 0 {
		// dbglog.DebugContext(ctx, "`sudo systemctl start service` failed", "service", config.ServiceName(), "err", err)
		// cmdr.App().SetSuggestRetCode(retCode)
		config.RetCode = retCode
		err = errors.New("failed to start service (%d). The console outputs are:\n%v", retCode, msg).WithErrors(err)
		return
	}

	pid, ppid := os.Getpid(), os.Getppid()
	dbglog.DebugContext(ctx, "'sudo systemctl start service' ends.", "pid", pid, "ppid", ppid, "service", config.ServiceName())
	_ = s.Logger.Infof("systemctl start %s done: pid=%v, ppid=%v\n", config.ServiceName(), pid, ppid)
	return
}

func systemdStop(ctx context.Context, config *Config, m *mgmtS, s *systemD) (err error) {
	if fn, ok := config.Entity.(EntityStopAware); ok {
		return fn.Stop(ctx, config, s.Logger)
	}

	if systemdIsRunning(ctx, config, m, s) != nil {
		println("service not running")
		return
	}

	var retCode int
	var msg string
	retCode, msg, err = cmdrexec.Sudo("systemctl", "stop", config.ServiceName())
	if err != nil || retCode != 0 {
		err = errors.New("failed to stop service. The console outputs are:\n%v", msg).WithErrors(err)
		return
	}

	return
}

func systemdStatus(ctx context.Context, config *Config, m *mgmtS, s *systemD) (err error) {
	if fn, ok := config.Entity.(EntityStatusAware); ok {
		return fn.Status(ctx, config, s.Logger)
	}

	var retCode int
	var msg string
	retCode, msg, err = cmdrexec.Sudo("systemctl", "status", config.ServiceName())
	if err != nil || retCode != 0 {
		err = errors.New("failed to status service. The console outputs are:\n%v", msg).WithErrors(err)
		return
	}

	return
}

func systemdRestart(ctx context.Context, config *Config, m *mgmtS, s *systemD) (err error) {
	if fn, ok := config.Entity.(EntityRestartAware); ok {
		return fn.Restart(ctx, config, s.Logger)
	}

	var retCode int
	var msg string
	retCode, msg, err = cmdrexec.Sudo("systemctl", "restart", config.ServiceName())
	if err != nil || retCode != 0 {
		err = errors.New("failed to restart service. The console outputs are:\n%v", msg).WithErrors(err)
		return
	}

	return
}

func systemdHotReload(ctx context.Context, config *Config, m *mgmtS, s *systemD) (err error) {
	if fn, ok := config.Entity.(EntityHotReloadAware); ok {
		return fn.HotReload(ctx, config, s.Logger)
	}

	var retCode int
	var msg string
	retCode, msg, err = cmdrexec.Sudo("systemctl", "reload", config.ServiceName())
	if err != nil || retCode != 0 {
		err = errors.New("failed to hot-reload service. The console outputs are:\n%v", msg).WithErrors(err)
		return
	}

	return
}

func systemdIsActive(config *Config) (text string, err error) {
	var retCode int
	retCode, text, err = cmdrexec.Sudo("systemctl", "is-active", config.ServiceName())
	if err != nil || retCode != 0 {
		return
	}
	text = strings.Trim(text, " \t\r\n")
	return
}

func systemdIsRunning(ctx context.Context, config *Config, m *mgmtS, s *systemD) (err error) {
	var text string
	text, err = systemdIsActive(config)
	if text != "activating" && text != "activated" {
		err = ErrServiceIsNotRunning
	}
	return
}

func systemdIsInactive(ctx context.Context, config *Config, m *mgmtS, s *systemD) (err error) {
	var text string
	text, err = systemdIsActive(config)
	if text != "inactive" {
		err = errors.New("service is not inactive")
	}
	return
}

func systemdIsStarted(ctx context.Context, config *Config, m *mgmtS, s *systemD) (err error) {
	var text string
	text, err = systemdIsActive(config)
	if text != "activating" {
		err = errors.New("service is not running")
	}
	return
}

func systemdIsEnabled(ctx context.Context, config *Config, m *mgmtS, s *systemD) (err error) {
	var retCode int
	var text string
	retCode, text, err = cmdrexec.Sudo("systemctl", "is-enabled", config.ServiceName())
	if err != nil || retCode != 0 {
		return
	}
	if strings.Trim(text, " \t\r\n") != "enabled" {
		err = ErrServiceIsNotEnabled
	}
	return
}

func systemdGetStatus(ctx context.Context, config *Config, m *mgmtS, s *systemD) (err error) {
	return
}

func createServiceFile(ctx context.Context, config *Config, svcfile string) (err error) {
	var tmpl *template.Template
	tmplFile := fmt.Sprintf("%v/share/service.tpl", config.TemplateDir)
	if dir.FileExists(tmplFile) {
		tmpl, err = template.New("service.file").ParseFiles(tmplFile)
	} else {
		tmpl, err = template.New("service.file").Parse(tplSystemdService)
	}
	if err != nil {
		return
	}

	var of *os.File
	tmpFile := fmt.Sprintf("%v/%v", config.TempDir, config.ServiceName())
	if err = dir.EnsureDir(path.Dir(tmpFile)); err != nil {
		return
	}
	if of, err = os.Create(tmpFile); err != nil {
		return
	}

	if config.Type == "" {
		config.Type = "exec"
	}

	defer func() {
		of.Close()
		dir.DeleteFile(tmpFile)
	}()
	if err = tmpl.Execute(of, config); err != nil {
		return
	}

	_, _, err = cmdrexec.Sudo("mv", tmpFile, svcfile)
	return
}

func createDefaultFile(ctx context.Context, config *Config, svcfile string) (err error) {
	var tmpl *template.Template
	tmplFile := fmt.Sprintf("%v/share/default.tpl", config.TemplateDir)
	if dir.FileExists(tmplFile) {
		tmpl, err = template.New("service.file").ParseFiles(tmplFile)
	} else {
		tmpl, err = template.New("service.file").Parse(tplEtcDefault)
	}
	if err != nil {
		return
	}

	var of *os.File
	tmpFile := fmt.Sprintf("%v/%v", config.TempDir, config.ServiceName())
	if err = dir.EnsureDir(path.Dir(tmpFile)); err != nil {
		return
	}
	if of, err = os.Create(tmpFile); err != nil {
		return
	}

	defer func() {
		of.Close()
		dir.DeleteFile(tmpFile)
	}()
	if err = tmpl.Execute(of, config); err != nil {
		return
	}

	cmdrexec.Sudo("mv", tmpFile, svcfile)
	return
}

func systemdInstall(ctx context.Context, config *Config, m *mgmtS, s *systemD) (err error) {
	if fn, ok := config.Entity.(EntityInstallAware); ok {
		return fn.Install(ctx, config, s.Logger)
	}

	if !hasSystemd(ctx) {
		return errors.Unavailable
	}

	if systemdIsRunning(ctx, config, m, s) == nil {
		s.Logger.Infof("service was running, or installed.")
		return
	}

	// cmdstore := cmdr.Store()
	// forceReinstall := cmdstore.MustBool("server.install.force")

	file := fmt.Sprintf("%s/%s", systemdDir, config.ServiceName())
	if dir.FileExists(file) {
		if !config.ForceReinstall {
			msg := `Service had been installed already.

If you wanna reinstall it, try this command line:

	$ {{.AppName}} {{.DadCommandsText}} install --force

`
			if config.Translate != nil {
				msg = config.Translate(msg)
			}
			dbglog.WarnContext(ctx, msg, "service-name", config.ServiceName())
			err = errors.New("service already installed")
			return
		}
	}

	err = createServiceFile(ctx, config, file)
	if err != nil {
		return
	}

	file = fmt.Sprintf("%s/%s", defaultsDir, config.Name)
	fileExist := dir.FileExists(file)
	if fileExist && !config.ForceReinstall {
		//
	} else {
		err = createDefaultFile(ctx, config, file)
		if err != nil {
			dbglog.WarnContext(ctx, "something's wrong.", "err", err)
			err = nil
			return
		}
	}

	// refresh systemd
	var retCode int
	var msg string
	retCode, msg, err = cmdrexec.Sudo("systemctl", "daemon-reload")
	if err != nil || retCode != 0 {
		err = errors.New("failed to refresh services list. The console outputs are:\n%v", msg).WithErrors(err)
		return
	}

	if autoEnable := config.AutoEnable; autoEnable {
		err = systemdEnable(ctx, config, m, s)
	}

	if err == nil {
		println("Service created successfully.")
		s.Logger.Infof("Service created successfully.\n")
	}
	return
}

func systemdUninstall(ctx context.Context, config *Config, m *mgmtS, s *systemD) (err error) {
	if fn, ok := config.Entity.(EntityUninstallAware); ok {
		return fn.Uninstall(ctx, config, s.Logger)
	}

	if !hasSystemd(ctx) {
		return errors.Unavailable
	}

	if err = systemdStop(ctx, config, m, s); err != nil {
		dbglog.WarnContext(ctx, "systemd stop command failed.", "err", err)
		// return
	}

	if err = systemdDisable(ctx, config, m, s); err != nil {
		dbglog.WarnContext(ctx, "systemd disable command failed.", "err", err)
		// return
	}

	anyExist := false
	file := fmt.Sprintf("%s/%s", systemdDir, config.ServiceName())
	if dir.FileExists(file) {
		var retCode int
		var msg string
		retCode, msg, err = cmdrexec.Sudo("mv", file, os.TempDir())
		if err != nil || retCode != 0 {
			err = errors.New("failed to uninstall service. The console outputs are:\n%v", msg).WithErrors(err)
			dbglog.WarnContext(ctx, "something's wrong.", "err", err)
			err = nil
			return
		}
		anyExist = true
	}

	file = fmt.Sprintf("%s/%s", defaultsDir, config.Name)
	if dir.FileExists(file) {
		var retCode int
		var msg string
		retCode, msg, err = cmdrexec.Sudo("mv", file, os.TempDir())
		if err != nil || retCode != 0 {
			err = errors.New("failed to mv service file to trashbin. The console outputs are:\n%v", msg).WithErrors(err)
			dbglog.WarnContext(ctx, "something's wrong.", "err", err)
			err = nil
			return
		}
		anyExist = true
	}

	if !anyExist {
		println("nothing needs to be done.")
		return
	}

	// refresh systemd
	var retCode int
	retCode, _, err = cmdrexec.Sudo("systemctl", "daemon-reload")
	if err != nil || retCode != 0 {
		return
	}

	println("service uninstalled")
	return
}

func systemdEnable(ctx context.Context, config *Config, m *mgmtS, s *systemD) (err error) {
	if fn, ok := config.Entity.(EntityEnableAware); ok {
		return fn.Enable(ctx, config, s.Logger)
	}

	if systemdIsEnabled(ctx, config, m, s) == nil {
		println("service has been enabled.")
		return
	}

	var retCode int
	var msg string
	retCode, msg, err = cmdrexec.Sudo("systemctl", "enable", config.ServiceName())
	if err != nil || retCode != 0 {
		err = errors.New("failed to enable service. The console outputs are:\n%v", msg).WithErrors(err)
		return
	}

	println("service has been enabled.")
	return
}

func systemdDisable(ctx context.Context, config *Config, m *mgmtS, s *systemD) (err error) {
	if fn, ok := config.Entity.(EntityDisableAware); ok {
		return fn.Disable(ctx, config, s.Logger)
	}

	if systemdIsEnabled(ctx, config, m, s) != nil {
		println("service has not been enabled.")
		return
	}

	var retCode int
	var msg string
	retCode, msg, err = cmdrexec.Sudo("systemctl", "disable", config.ServiceName())
	if err != nil || retCode != 0 {
		err = errors.New("failed to disable service. The console outputs are:\n%v", msg).WithErrors(err)
		return
	}

	println("service has been disabled.")
	return
}

func systemdViewLog(ctx context.Context, config *Config, m *mgmtS, s *systemD) (err error) {
	if fn, ok := config.Entity.(EntityViewLogAware); ok {
		return fn.ViewLog(ctx, config, s.Logger)
	}
	return
}

const (
	tplEtcDefault = `### {{.ScreenName}} configurations
### executable: {{.ExecutablePath}}

# PORT=3211

# OPTIONS="--port 3211"

#
# the service startup command line is like:
#
#	$ service-app [global-options] server start [options]
#
GLOBAL_OPTIONS=""
OPTIONS=""

`

	// tplSystemService template file for creating a new systemd service.
	//
	//
	//
	//
	// ref: to-d-o
	tplSystemdService = `### {{.ScreenName}} services
### {{.ServiceName}}
### executable: {{.ExecutablePath}}

[Unit]
Description={{.ScreenName}} Service for %i - {{.Desc}}
# Documentation=man:sshd(8) man:sshd_config(5) man:{{.Name}}(1)
After=network.target
# Wants=syslog.service
ConditionPathExists={{.ExecutablePath}}

[Install]
WantedBy=multi-user.target

[Service]
Type={{.Type}}
{{if .User}}User={{.User}}{{else}}# User=%i{{end}}
{{if .Group}}Group={{.Group}}{{else}}# Group=%i{{end}}
LimitNOFILE=65535
{{if .TimeoutStartSec}}TimeoutStartSec={{.TimeoutStartSec}}{{else}}TimeoutStartSec=60s{{end}}
{{if .TimeoutStopSec}}TimeoutStopSec={{.TimeoutStopSec}}{{else}}TimeoutStopSec=60s{{end}}
{{if .PIDFile}}PIDFile={{.PIDFile}}{{else}}PIDFile=/run/{{.Name}}/{{.Name}}.pid{{end}}

KillMode=process
Restart=on-failure
{{if .RestartSec}}RestartSec={{.RestartSec}}{{else}}RestartSec=23s{{end}}
# RestartLimitIntervalSec=60

EnvironmentFile=/etc/sysconfig/{{.Name}}
{{range $k, $v := .Env -}}
Environment={{$k}}={{$v}}
{{end -}}

{{if .WorkDir}}WorkingDirectory={{.WorkDir}}{{else}}WorkingDirectory=%h{{end}}

#          start: --addr, --port,
#           todo: --pid
# global options: --verbose, --debug,
ExecStart={{.ExecutablePath}} $GLOBAL_OPTIONS server start --foreground --service $OPTIONS
#           stop: -1/--hup, -9/--kill,
### TODO ExecStop={{.ExecutablePath}} $GLOBAL_OPTIONS server stop -1
### TODO ExecReload=/bin/kill -HUP $MAINPID
ExecStop={{.ExecutablePath}} $GLOBAL_OPTIONS server stop -3
ExecReload={{.ExecutablePath}} $GLOBAL_OPTIONS server restart

# # make sure log directory exists and owned by syslog
#PermissionsStartOnly=true
ExecStartPre=-/bin/mkdir /run/{{.Name}}
ExecStartPre=-/bin/mkdir /var/lib/{{.Name}}
ExecStartPre=-/bin/mkdir /var/log/{{.Name}}
ExecStartPre=-/bin/chown -R %i: /var/run/{{.Name}} /var/lib/{{.Name}}
# ExecStartPre=-/bin/chown -R syslog:adm /var/log/{{.Name}}
ExecStartPre=-/bin/chown -R %i: /var/log/{{.Name}}

# # enable coredump
# ExecStartPre=ulimit -c unlimited

SyslogIdentifier={{.Name}}
StandardOutput=append:{{.StandardOutPath}}
StandardError=append:{{.StandardErrorPath}}




`

	systemdDir = "/etc/systemd/system"

	defaultsDir = "/etc/sysconfig"
	// defaultsDir = "/etc/default"
)
