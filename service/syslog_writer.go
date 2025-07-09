//go:build !windows && !plan9
// +build !windows,!plan9

// +//g/o:build linux || darwin
// +// +/build linux darwin

package service

import (
	"fmt"
	"log/syslog"

	logzorig "github.com/hedzr/logg/slog"
)

func newSysLogger(name string, errs chan<- error, level syslog.Priority) (ZLogger, error) {
	w, err := syslog.New(level, name)
	if err != nil {
		return nil, err
	}
	return syslogWriter{sysLogger{w, errs}, level}, nil
}

type syslogWriter struct {
	sysLogger
	level syslog.Priority
}

func (s syslogWriter) convertLevel(level logzorig.Level) syslog.Priority {
	switch level {
	case logzorig.FailLevel:
		return syslog.LOG_CRIT
	case logzorig.SuccessLevel:
		return syslog.LOG_NOTICE
	case logzorig.OKLevel:
		return syslog.LOG_NOTICE
	case logzorig.AlwaysLevel:
		return syslog.LOG_INFO
	case logzorig.OffLevel:
		return syslog.LOG_DEBUG
	case logzorig.TraceLevel:
		return syslog.LOG_DEBUG
	case logzorig.DebugLevel:
		return syslog.LOG_DEBUG
	case logzorig.InfoLevel:
		return syslog.LOG_INFO
	case logzorig.WarnLevel:
		return syslog.LOG_WARNING
	case logzorig.ErrorLevel:
		return syslog.LOG_ERR
	case logzorig.FatalLevel:
		return syslog.LOG_EMERG
	case logzorig.PanicLevel:
		return syslog.LOG_CRIT
	default:
		return syslog.LOG_INFO
	}
}

// SetLevel implements [logg/slog.LevelSettable] so that the
// syslog.Priority can be updated lively before writing.
func (s syslogWriter) SetLevel(level logzorig.Level) { s.level = s.convertLevel(level) }

func (s syslogWriter) Write(data []byte) (n int, err error) {
	switch s.level {
	case syslog.LOG_EMERG:
		_ = s.send(s.Writer.Emerg(string(data)))
	case syslog.LOG_ALERT:
		_ = s.send(s.Writer.Alert(string(data)))
	case syslog.LOG_CRIT:
		_ = s.send(s.Writer.Crit(string(data)))
	case syslog.LOG_ERR:
		_ = s.send(s.Writer.Err(string(data)))
	case syslog.LOG_WARNING:
		_ = s.send(s.Writer.Warning(string(data)))
	case syslog.LOG_NOTICE:
		_ = s.send(s.Writer.Notice(string(data)))
	case syslog.LOG_INFO:
		_ = s.send(s.Writer.Info(string(data)))
	}
	// println(dbglog.GetLevel().String())
	// dbglog.Println(string(data), "level", dbglog.GetLevel())
	return len(data), nil
}

type sysLogger struct {
	*syslog.Writer
	errs chan<- error
}

func (s sysLogger) send(err error) error {
	if err != nil && s.errs != nil {
		s.errs <- err
	}
	return err
}

func (s sysLogger) W(l syslog.Priority, msg string) error {
	// dbglog.Print(msg, "level", dbglog.GetLevel())
	switch l {
	case syslog.LOG_EMERG:
		return s.send(s.Writer.Emerg(msg))
	case syslog.LOG_ALERT:
		return s.send(s.Writer.Alert(msg))
	case syslog.LOG_CRIT:
		return s.send(s.Writer.Crit(msg))
	case syslog.LOG_ERR:
		return s.send(s.Writer.Err(msg))
	case syslog.LOG_WARNING:
		return s.send(s.Writer.Warning(msg))
	case syslog.LOG_NOTICE:
		return s.send(s.Writer.Notice(msg))
	case syslog.LOG_INFO:
		return s.send(s.Writer.Info(msg))
	}
	return nil
}

func (s sysLogger) Wf(l syslog.Priority, msg string, args ...any) error {
	// dbglog.Info("level", "level", dbglog.GetLevel())
	// dbglog.Infof(msg, args...)
	switch l {
	case syslog.LOG_EMERG:
		return s.send(s.Writer.Emerg(fmt.Sprintf(msg, args...)))
	case syslog.LOG_ALERT:
		return s.send(s.Writer.Alert(fmt.Sprintf(msg, args...)))
	case syslog.LOG_CRIT:
		return s.send(s.Writer.Crit(fmt.Sprintf(msg, args...)))
	case syslog.LOG_ERR:
		return s.send(s.Writer.Err(fmt.Sprintf(msg, args...)))
	case syslog.LOG_WARNING:
		return s.send(s.Writer.Warning(fmt.Sprintf(msg, args...)))
	case syslog.LOG_NOTICE:
		return s.send(s.Writer.Notice(fmt.Sprintf(msg, args...)))
	case syslog.LOG_INFO:
		return s.send(s.Writer.Info(fmt.Sprintf(msg, args...)))
	}
	return nil
}

// Error is a slog like api.
func (s sysLogger) Error(msg string, args ...any) { _ = s.Wf(syslog.LOG_ERR, msg, args...) }
func (s sysLogger) Warn(msg string, args ...any)  { _ = s.Wf(syslog.LOG_WARNING, msg, args...) }
func (s sysLogger) Info(msg string, args ...any)  { _ = s.Wf(syslog.LOG_INFO, msg, args...) }

// Errorf is a log like api
func (s sysLogger) Errorf(m string, a ...any) error { return s.Wf(syslog.LOG_ERR, m, a...) }
func (s sysLogger) Warnf(m string, a ...any) error  { return s.Wf(syslog.LOG_WARNING, m, a...) }
func (s sysLogger) Infof(m string, a ...any) error  { return s.Wf(syslog.LOG_INFO, m, a...) }

// Emergf is a wrapper to [syslog.Writer.Emerg]
func (s sysLogger) Emergf(m string, a ...any)  { _ = s.Wf(syslog.LOG_EMERG, m, a...) }
func (s sysLogger) Alertf(m string, a ...any)  { _ = s.Wf(syslog.LOG_ALERT, m, a...) }
func (s sysLogger) Critf(m string, a ...any)   { _ = s.Wf(syslog.LOG_CRIT, m, a...) }
func (s sysLogger) Noticef(m string, a ...any) { _ = s.Wf(syslog.LOG_NOTICE, m, a...) }
func (s sysLogger) Debugf(m string, a ...any)  { _ = s.Wf(syslog.LOG_DEBUG, m, a...) }
