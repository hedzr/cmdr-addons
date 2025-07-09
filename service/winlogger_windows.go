package service

import (
	"context"
	"fmt"
	"sync/atomic"

	"golang.org/x/sys/windows/svc/eventlog"
)

func makeWinLogger(ctx context.Context, svcName string, m *mgmtS, s *ntServiceD) (l ZLogger, err error) {
	var elog *eventlog.Log
	elog, err = eventlog.Open(svcName)
	if err != nil {
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
	l = wlog(elog, errsCh)
	return
}

func wlog(elog *eventlog.Log, errs chan<- error) ZLogger {
	return &wlogS{elog: elog, errs: errs}
}

type wlogS struct {
	elog *eventlog.Log
	errs chan<- error
	eidx uint32
}

func (s *wlogS) Close() {
	if s.elog != nil {
		s.elog.Close()
	}
}

func (s *wlogS) Info(msg string, args ...any) {
	if err := s.elog.Info(s.eidx, fmt.Sprintf(msg, args...)); err != nil {
		s.errs <- err
	}
	atomic.AddUint32(&s.eidx, 1)
}

func (s *wlogS) Warn(msg string, args ...any) {
	if err := s.elog.Warning(s.eidx, fmt.Sprintf(msg, args...)); err != nil {
		s.errs <- err
	}
	atomic.AddUint32(&s.eidx, 1)
}

func (s *wlogS) Error(msg string, args ...any) {
	if err := s.elog.Error(s.eidx, fmt.Sprintf(msg, args...)); err != nil {
		s.errs <- err
	}
	atomic.AddUint32(&s.eidx, 1)
}

func (s *wlogS) Infof(format string, a ...any) (err error) {
	if err = s.elog.Info(s.eidx, fmt.Sprintf(format, a...)); err != nil {
		s.errs <- err
	}
	atomic.AddUint32(&s.eidx, 1)
	return
}

func (s *wlogS) Warnf(format string, a ...any) (err error) {
	if err = s.elog.Warning(s.eidx, fmt.Sprintf(format, a...)); err != nil {
		s.errs <- err
	}
	atomic.AddUint32(&s.eidx, 1)
	return
}

func (s *wlogS) Errorf(format string, a ...any) (err error) {
	if err = s.elog.Error(s.eidx, fmt.Sprintf(format, a...)); err != nil {
		s.errs <- err
	}
	atomic.AddUint32(&s.eidx, 1)
	return
}

func (s *wlogS) Write(p []byte) (n int, err error) {
	if err = s.elog.Info(s.eidx, string(p)); err != nil {
		s.errs <- err
	}
	atomic.AddUint32(&s.eidx, 1)
	return
}
