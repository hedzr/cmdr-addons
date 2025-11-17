package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/user"
	"path"
	"strconv"
	"syscall"

	"github.com/hedzr/is"
	"github.com/hedzr/is/dir"
	"github.com/hedzr/is/exec"
	"gopkg.in/hedzr/errors.v3"

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
	_ = s

	currentUser, err := user.Current()
	if err != nil {
		dbglog.Error("Error getting current user", "err", err)
	}

	needSudo := currentUser.Username == "root"
	pid := os.Getpid()

	if dir.FileExists(path.Dir(p.file)) {
		var locked bool
		f := filelock.New(p.file)
		locked, err = f.TryLock()
		if err != nil {
			var pass bool
			var ep *os.PathError
			if errors.As(err, &ep) {
				if ep.Err == syscall.EACCES {
					dbglog.DebugContext(ctx, "[pidFileS] retry pidfile initializing with sudo", "pidfile", p.file)
					exec.Sudo("rm", "-f", p.file)
					needSudo, pass = true, true
					err = nil
				}
			}
			if !pass {
				dbglog.Error("Failed to init (and lock) pidfile (flck)", "err", err, "pidfile", p.file)
				return
			}
		}

		if !locked {
			if err = f.Unlock(); err != nil {
				return
			}
			if needSudo {
				exec.Sudo("sh", "-c", fmt.Sprintf("echo > %s && chown %s: %s", p.file, currentUser.Username, p.file))
				p.f, err = os.Open(p.file)
				if err != nil {
					dbglog.Error("Error open pidfile", "err", err)
				}
			} else {
				if p.f, err = os.Create(p.file); err != nil {
					return
				}
			}
			if _, err = p.f.WriteString(strconv.Itoa(pid)); err != nil {
				dbglog.Error("Error write pidfile", "err", err)
				return
			}
			is.Closers().RegisterPeripheral(p)
			dbglog.InfoContext(ctx, "pid file created", "file", p.file, "pid", pid)
		} else {
			// update pid
			p.flck = f
			if _, err = p.flck.WriteString(strconv.Itoa(pid)); err != nil {
				dbglog.Error("Error write pidfile", "err", err)
				return
			}
			is.Closers().RegisterPeripheral(p)
			dbglog.InfoContext(ctx, "pid file updated", "file", p.file, "pid", pid)
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
