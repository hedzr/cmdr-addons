package service

import (
	"context"

	"github.com/hedzr/is"
	"github.com/hedzr/is/basics"
	"gopkg.in/hedzr/errors.v3"

	"github.com/hedzr/cmdr-addons/v2/tool/dbglog"

	"github.com/hedzr/cmdr-addons/service/v2/systems"
)

// New creates an instance of service Manager, so that you can run it or manage it.
//
// A sample:
//
//	svc := service.New(context.TODO())
//	err := svc.Run(&Config{
//		Name:           "demo",
//		DisplayName:    "demo service",
//		Description:    "demo",
//		WorkDir:        "",
//		Executable:     "",
//		ArgsForInstall: nil,
//		Env:            nil,
//		RunAs:          "",
//		Dependencies:   nil,
//	})
func New(ctx context.Context) Manager {
	s := &mgmtS{force: true, errs: errors.New()}
	if err := s.init(ctx); err != nil {
		dbglog.Error("Service manager init() failed", "err", err)
	}
	return s
}

type mgmtS struct {
	errs        errors.Error
	closeSelfCB func()
	pidfile     *pidFileS

	fore          bool
	serviceMode   bool
	colorModeSave bool
	force         bool
}

func (s *mgmtS) SetForegroundMode(b bool) { s.fore = b }
func (s *mgmtS) SetServiceMode(b bool)    { s.serviceMode = b }
func (s *mgmtS) SetForceMode(b bool)      { s.force = b }

func (s *mgmtS) Err() error { return s.errs }

func (s *mgmtS) init(ctx context.Context) (err error) { return s.initSelf(ctx) }
func (s *mgmtS) Close()                               { s.closeSelfCB() }

func (s *mgmtS) closeSelf(ctx context.Context) (err error) {
	return
}

func (s *mgmtS) initSelf(ctx context.Context) (err error) {
	s.closeSelfCB = func() { s.errs.Attach(s.closeSelf(ctx)) }
	return
}

func (s *mgmtS) WithEntity(entity Entity) Manager {
	// s.entity = entity
	return s
}

func (s *mgmtS) Run(ctx context.Context, config *Config) (err error) {
	err = s.Control(ctx, config, Start)
	return
}

func (s *mgmtS) IsRunning() bool {
	if s.pidfile == nil {
		return s.pidfile.isRunning()
	}
	return false
}

func (s *mgmtS) Control(ctx context.Context, config *Config, cmd Command) (err error) {
	var be Backend
	// dbglog.ErrorContext(ctx, "[mgmtS] tip is safe")
	if be, err = s.chooseBackend(ctx); err == nil && be != nil {
		dbglog.DebugContext(ctx, "[mgmtS] backend chose", "backend", be)

		if be.IsValid(ctx) {
			dbglog.DebugContext(ctx, "[mgmtS] backend is valid", "backend", be)

			// if a backend needs to be cleanup at shutting down...
			if c, ok := be.(basics.Peripheral); ok {
				basics.RegisterPeripheral(c) // register it into closers
				defer basics.Close()
			}

			if cmd > MinCommand && cmd < MaxCommand {
				dbglog.DebugContext(ctx, "[mgmtS] execute control command", "command", cmd, "backend", be)

				config.makeSafety()

				if cmd == Start && s.serviceMode && !systems.HasNTService {
					s.pidfile, err = newpidfile(ctx, s, config)
				}

				if systems.HasNTService {
					dbglog.InfoContext(ctx, "[mgmtS] control backend", "backend", be, "cmd", cmd)
				}
				err = be.Control(ctx, config, s, cmd)
				if err != nil {
					dbglog.ErrorContext(ctx, "[mgmtS] execute control command failed", "command", cmd, "err", err)
					if config.RetCode == 0 {
						config.RetCode = 3
					}
				} else {
					dbglog.InfoContext(ctx, "Service operations sent ok.", "cmd", cmd)
				}
				return
			}

			err = errors.New("Invalid service command: %v, the valid commands are in [%v, %v]", cmd, MinCommand+1, MaxCommand-1)
			dbglog.ErrorContext(ctx, "Invalid service command", "cmd", cmd, "err", err)
			return
		}

		dbglog.WarnContext(ctx, "[mgmtS] backend is invalid", "backend", be)
		return
	}

	dbglog.WarnContext(ctx, "[mgmtS] no backend chose")
	return
}

func (s *mgmtS) chooseBackend(ctx context.Context) (be Backend, err error) {
	return ChooseBackend(ctx)
}

func (s *mgmtS) NotifyLoggerCreated(logger ZLogger) {
	s.colorModeSave = dbglog.RawLogger().ColorMode()
	if is.InDebugging() || is.DebugBuild() || is.DebugMode() || is.Windows() {
		// avoid adding recursive writer | dbglog.ZLogger() is now a static, unique pointer
		if logger != dbglog.ZLogger() {
			dbglog.RawLogger().AddWriter(logger)
			dbglog.RawLogger().AddErrorWriter(logger)
		}
	} else if logger != dbglog.ZLogger() {
		// dbglog.RawLogger().SetColorMode(false)
		// dbglog.RawLogger().SetWriter(logger)
		// dbglog.RawLogger().SetErrorWriter(logger)
		dbglog.RawLogger().AddWriter(logger)
		dbglog.RawLogger().AddErrorWriter(logger)
	}
	if s.serviceMode {
		dbglog.SetColorMode(false)
	}
	dbglog.SetColorMode(false)
}

func (s *mgmtS) NotifyLoggerDestroying(logger ZLogger) {
	dbglog.RawLogger().ResetWriters()
	dbglog.SetColorMode(s.colorModeSave)
}

type NotifyLogger interface {
	NotifyLoggerCreated(logger ZLogger)
	NotifyLoggerDestroying(logger ZLogger)
}

// func chooseBackend(ctx context.Context) (be Backend, err error) {
// 	return ChooseBackend(ctx)
// }
