package service

import (
	"context"
	"errors"
	"fmt"
	"io"
)

// const Version = "v0.1.0"
const DefaultVendorDomainPrefix = "com.examples."

type Manager interface {
	Close()
	Err() error

	WithEntity(entity Entity) Manager

	SetForegroundMode(b bool) // run in foreground-mode?
	SetServiceMode(b bool)    // run in service-mode?
	SetForceMode(b bool)      // force reinstall service?

	Run(ctx context.Context, config *Config) (err error)
	// Control the service's action.
	//
	// [Command] [service.Start] won't return till ctx cancelled.
	// But in foreground mode, Control will return right now.
	Control(ctx context.Context, config *Config, cmd Command) (err error)

	NotifyLoggerCreated(logger ZLogger)
}

type Backend interface {
	IsValid(ctx context.Context) (valid bool)
	Control(ctx context.Context, config *Config, m *mgmtS, cmd Command) (err error)
}

type Service interface {
	Entity() Entity
}

type Config struct {
	Name        string
	DisplayName string
	Description string

	WorkDir         string            // service main workDir, eg: /var/lib/service1/
	Executable      string            // eg: /var/lib/service1/bin/service1
	ArgsForInstall  []string          //
	Env             map[string]string //
	RunAs           string            // run as user
	User            string
	Group           string
	TimeoutStartSec string
	TimeoutStopSec  string
	RestartSec      string
	PIDFile         string

	Dependencies []string

	StandardOutPath   string // "/dev/null" is valid for darwin and linux
	StandardErrorPath string //

	Entity Entity

	Type string // for systemd: simple, forking, exec, oneshot, dbus, notify, idle

	ForceReinstall     bool
	AutoEnable         bool
	DelayedAutoStart   bool
	UserLevel          bool
	VendorDomainPrefix string

	CmdLines    []string
	TemplateDir string
	RunDir      string
	LogDir      string
	TempDir     string

	RetCode int // return error code to main func if non-zero

	// Translate allows formatting msg with your own translator.
	Translate func(string) string
}

type Chooser interface {
	Choose(ctx context.Context) (ok bool)
}

type EntityInfoAware interface {
	Info(ctx context.Context, config *Config, logger Logger) (text string)
}

type EntityPortAware interface {
	Port(ctx context.Context, config *Config, logger Logger) (port int)
}

type EntityAddrAware interface {
	Addr(ctx context.Context, config *Config, logger Logger) (addr string)
}

type EntityStartAware interface {
	Start(ctx context.Context, config *Config, logger Logger) (err error)
}

type EntityStopAware interface {
	Stop(ctx context.Context, config *Config, logger Logger) (err error)
}

type EntityStatusAware interface {
	Status(ctx context.Context, config *Config, logger Logger) (err error)
}

type EntityRestartAware interface {
	Restart(ctx context.Context, config *Config, logger Logger) (err error)
}

type EntityHotReloadAware interface {
	HotReload(ctx context.Context, config *Config, logger Logger) (err error)
}

type EntityInstallAware interface {
	Install(ctx context.Context, config *Config, logger Logger) (err error)
}

type EntityUninstallAware interface {
	Uninstall(ctx context.Context, config *Config, logger Logger) (err error)
}

type EntityEnableAware interface {
	Enable(ctx context.Context, config *Config, logger Logger) (err error)
}

type EntityDisableAware interface {
	Disable(ctx context.Context, config *Config, logger Logger) (err error)
}

type EntityPauseAware interface {
	Pause(ctx context.Context, config *Config, logger Logger) (err error)
}

type EntityContinueAware interface {
	Continue(ctx context.Context, config *Config, logger Logger) (err error)
}

type EntityViewLogAware interface {
	ViewLog(ctx context.Context, config *Config, logger Logger) (err error)
}

type RunnableService interface {
	SetServiceMode(serviceMode bool)
	Run(ctx context.Context, config *Config, logger Logger) (err error)
}

type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Infof(format string, a ...any) error
	Warnf(format string, a ...any) error
	Errorf(format string, a ...any) error
}

type ZLogger interface {
	Logger
	io.Writer
}

type Command int

func (c *Command) UnmarshalText(text []byte) error {
	for k, v := range mCommandStrings {
		if string(v) == string(text) {
			*c = k
			return nil
		}
	}
	return fmt.Errorf("unknown command %q", text)
}

func (c Command) MarshalText() (text []byte, err error) {
	text = []byte(c.String())
	return
}

func (c Command) String() string {
	if s, ok := mCommandStrings[c]; ok {
		return s
	}
	return fmt.Sprintf("Cmd{%d}", int(c))
}

var mCommandStrings = map[Command]string{
	MinCommand: "MIN",
	Info:       "Info",
	Port:       "Port",
	Addr:       "Addr",
	Start:      "Start",
	Stop:       "Stop",
	Status:     "Status",
	Restart:    "Restart",
	HotReload:  "HotReload",
	Install:    "Install",
	Uninstall:  "Uninstall",
	Enable:     "Enable",
	Disable:    "Disable",
	ViewLog:    "ViewLog",
	MaxCommand: "MAX",
}

const (
	MinCommand = Command(iota) //
	Info                       // = Status command.
	Port                       // main port if tcp server, -1 for socket server and others
	Addr                       // dialing address (like "localhost:8000") for client
	Start                      // server start
	Stop                       // server stop
	Status                     // server status
	Restart                    // = Reload Command, server restart
	HotReload                  // hot-reload, live-reload
	Install                    //
	Uninstall                  //
	Enable                     //
	Disable                    //

	Pause    // for windows
	Continue // for windows

	/** Advance Commands In The Future: */

	// ViewLog to show system log about this service
	ViewLog

	MaxCommand
)

// func init() {
// 	// /etc/app
// 	d := cmdr.ConfigDir()
// 	dbglog.Debug("config directory:", "dir", d)
// 	// /var/run/app
// 	d = cmdr.VarRunDir()
// }

var ErrServiceIsRunning = errors.New("service is running already")
var ErrServiceIsNotRunning = errors.New("service is not running")
var ErrServiceIsNotEnabled = errors.New("service is not enabled")
