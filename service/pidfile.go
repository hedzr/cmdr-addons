package service

import (
	"context"
	"io"
	"os"
	"path"
	"strconv"

	"github.com/hedzr/is"
	"github.com/hedzr/is/dir"

	"github.com/hedzr/cmdr-addons/service/v2/filelock"
	"github.com/hedzr/cmdr-addons/v2/tool/dbglog"
)

func newpidfile(ctx context.Context, s *mgmtS, c *Config) (p *pidFileS, err error) {
	p = new(pidFileS)
	err = p.init(ctx, s, c)
	return
}

type pidFileS struct {
	file string
	f    *os.File
	flck *filelock.Filelock
}

func (p *pidFileS) Close() {
	if f := p.f; f != nil {
		if err := f.Close(); err != nil {
			dbglog.Error("Failed to close pidfile", "err", err)
		}
		if err := os.Remove(p.file); err != nil {
			dbglog.Error("Failed to remove pidfile", "err", err)
		} else {
			dbglog.Info("pid file removed", "file", p.file)
		}
	}
	if p.flck != nil {
		if err := p.flck.Close(); err != nil {
			dbglog.Error("Failed to close pidfile (flck)", "err", err)
		}
	}
}

func (p *pidFileS) init(ctx context.Context, s *mgmtS, c *Config) (err error) {
	p.file = path.Join(c.RunDir, c.Name+".pid")
	if dir.FileExists(path.Dir(p.file)) {
		var locked bool
		f := filelock.New(p.file)
		locked, err = f.TryLock()
		if err != nil {
			return
		}

		if !locked {
			if err = f.Unlock(); err != nil {
				return
			}
			if p.f, err = os.Create(p.file); err != nil {
				return
			}
			if _, err = p.f.WriteString(strconv.Itoa(os.Getpid())); err != nil {
				return
			}
			is.Closers().RegisterPeripheral(p)
			dbglog.InfoContext(ctx, "pid file created", "file", p.file, "pid", os.Getpid())
		} else {
			// update pid
			p.flck = f
			if _, err = p.flck.WriteString(strconv.Itoa(os.Getpid())); err != nil {
				return
			}
			is.Closers().RegisterPeripheral(p)
			dbglog.WarnContext(ctx, "pid file updated", "file", p.file, "pid", os.Getpid())
		}
	}
	return
}

func (p *pidFileS) isRunning() (running bool) {
	if p != nil {
		pid, err := p.ReadPid()
		if err != nil {
			return
		}

		var proc *os.Process
		if proc, err = os.FindProcess(pid); err != nil {
			return
		}
		running = proc != nil
	}
	return
}

func (p *pidFileS) ReadPid() (pid int, err error) {
	if p != nil {
		if p.f == nil {
			var data []byte
			data, err = io.ReadAll(p.f)
			if err != nil {
				return
			}

			if pid, err = strconv.Atoi(string(data)); err == nil {
				return
			}
		}

		if p.flck != nil {
			var data string
			data, err = p.flck.ReadAll()
			if err != nil {
				return
			}

			if pid, err = strconv.Atoi(data); err == nil {
				return
			}
		}
	}
	return
}
