// Copyright © 2020 Hedzr Yeh.

package dex

import (
	"fmt"
	"github.com/hedzr/cmdr"
	"github.com/hedzr/cmdr-addons/pkg/plugins/dex/sig"
	"github.com/hedzr/cmdr/conf"
	"github.com/hedzr/log"
	"github.com/hedzr/log/dir"
	"github.com/kardianos/service"
	"os"
	"os/user"
	"path"
	"runtime"
	"strings"
)

var pd *Program

type Program struct {
	daemon  Daemon // the daemon object customized by user
	Config  *service.Config
	Service service.Service
	Logger  service.Logger

	log log.Logger

	// the arguments of cmdr entry.
	Command *cmdr.Command
	Args    []string
	Env     []string

	// InvokedInDaemon will be set to true if this daemon service is running under service/daemon manager.
	// For Windows, it's available if the service app has been starting from serivces.msc or by system automatically.
	// For macOS, the Launchctl starts it.
	// For Linux, systemd/upstart/sysv does it.
	//
	// The underlying detector
	InvokedInDaemon bool
	// InvokedDirectly means that administrator is running it from a tty/console/terminal.
	InvokedDirectly bool
	// ForwardLogToFile enables logging forward to /var/log if systemd mode enabled.
	ForwardLogToFile bool

	modifier    func(daemonServerCommand *cmdr.Command) *cmdr.Command
	preActions  []func(cmd *cmdr.Command, args []string) (err error)
	postActions []func(cmd *cmdr.Command, args []string)

	err     error
	pidFile *pidFileStruct
	fOut    *os.File
	fErr    *os.File
	// exit            chan struct{}
	// done            chan struct{}
	// Command     *exec.Cmd
}

func (p *Program) prepareAppDirs() (err error) {

	if runtime.GOOS == "windows" {
		return
	}

	var currUser *user.User
	currUser, err = user.Current()

	for _, pdir := range []string{"/var/lib", "/var/log", "/var/run"} {
		d := path.Join(pdir, p.Config.Name)
		if !dir.FileExists(d) {
			fmt.Printf(`The directory %q needs be created via sudo priviledge, 
so the OS account password will be prompted via:
'sudo mkdir %q':`, d, d)
			_, _, err = sudo("mkdir", d)
			_, _, err = sudo("chown", "-R", currUser.Username, d)
			_, _, err = sudo("chmod", "-R", "0770", d)
		}
	}

	if !dir.FileExists(p.EnvFileName()) && runtime.GOOS == "linux" {
		// var f *os.File
		// f, err = os.OpenFile(p.EnvFileName(), os.O_CREATE|os.O_WRONLY, 0770)
		// if err != nil {
		// 	return
		// }
		// defer f.Close()
		// _, _ = f.WriteString(fmt.Sprintf("# default config file for daemon %q",p.Config.Name))
		fmt.Printf(`The env-file %q should be present, 
we will 'touch' it via sudo priviledge, 
so the OS account password will be prompted via:
'sudo mkdir %q':`, p.EnvFileName(), p.EnvFileName())
		_, _, err = sudo("touch", p.EnvFileName())
	}

	return
}

func (p *Program) prepareLogFiles() (err error) {
	if p.fErr == nil {
		p.fErr, err = os.OpenFile(p.LogStderrFileName(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
		if err != nil {
			p.Logger.Warningf("Failed to open/create std err %q: %v", p.LogStderrFileName(), err)
			return
		}
	}

	if p.fOut == nil {
		p.fOut, err = os.OpenFile(p.LogStdoutFileName(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
		if err != nil {
			// logger.Warningf("Failed to open std out %q: %v", p.Stdout, err)
			p.Logger.Errorf("Failed to open/create std out %q: %v\n", p.LogStdoutFileName(), err)
			return
		}
	}

	return
}

func (p *Program) GetLogFileHandlers() (fOut, fErr *os.File) {
	return p.fOut, p.fErr
}

func (p *Program) SocketFileName() string {
	if runtime.GOOS == "windows" {
		return ""
	}
	// if runtime.GOOS == "darwin" {
	// 	return ""
	// }
	// p.log.Infof("env.APPNAME = %v, conf.AppName = %v", os.Getenv("APPNAME"), conf.AppName)
	return cmdr.GetStringR("server.start.socket-file",
		os.ExpandEnv("/var/run/$APPNAME/$APPNAME.sock"))
}

func (p *Program) PidFileName() string {
	if runtime.GOOS == "windows" {
		return ""
	}
	// if runtime.GOOS == "darwin" {
	// 	return ""
	// }
	p.log.Infof("env.APPNAME = %v, conf.AppName = %v", os.Getenv("APPNAME"), conf.AppName)
	return os.ExpandEnv("/var/run/$APPNAME/$APPNAME.pid")
}

func (p *Program) LogStdoutFileName() string {
	if runtime.GOOS == "windows" {
		return os.ExpandEnv("/tmp/$APPNAME.out")
	}
	// if runtime.GOOS == "darwin" {
	// 	return os.ExpandEnv("/tmp/$APPNAME.out")
	// }
	return os.ExpandEnv("/var/log/$APPNAME/$APPNAME.out")
}

func (p *Program) LogStderrFileName() string {
	if runtime.GOOS == "windows" {
		return os.ExpandEnv("/tmp/$APPNAME.err")
	}
	// if runtime.GOOS == "darwin" {
	// 	return os.ExpandEnv("/tmp/$APPNAME.err")
	// }
	return os.ExpandEnv("/var/log/$APPNAME/$APPNAME.err")
}

func (p *Program) WorkDirName() string {
	if runtime.GOOS == "windows" {
		return os.ExpandEnv("/tmp")
	}
	// if runtime.GOOS == "darwin" {
	// 	return os.ExpandEnv("/tmp")
	// }
	return os.ExpandEnv("/var/lib/$APPNAME")
}

func (p *Program) EnvFileName() string {
	// if runtime.GOOS == "windows" {
	// 	return os.ExpandEnv("/tmp/$APPNAME")
	// }
	// if runtime.GOOS == "darwin" {
	// 	return os.ExpandEnv("/tmp/$APPNAME")
	// }
	if dir.FileExists("/etc/default") {
		return os.ExpandEnv("/etc/default/$APPNAME")
	}
	return os.ExpandEnv("/etc/sysconfig/$APPNAME")
}

func isUbuntu() bool {
	if runtime.GOOS == "linux" {
		_, o, _ := shellWithOutput("lsb_release", "-a")
		// fmt.Println(o)
		if strings.Contains(o, "Ubuntu") {
			// fmt.Println("ubuntu hasOutputFileSupport = true")
			return true
		}
	}
	return false
}

func (p *Program) Start(s service.Service) error {
	p.Logger.Infof("xx.pp.start; Args: %v;", os.Args)

	// Start should not block. Do the actual work async.

	go p.run()
	return p.err
}

func (p *Program) run() {
	p.runIt(p.Command, p.Args)
}

func (p *Program) runIt(cmd *cmdr.Command, args []string) {
	p.Logger.Infof("xx.pp.runIt; Args: %v;", os.Args)

	// logger.Info("Starting ", p.DisplayName)

	// if runtime.GOOS == "windows" {
	//	defer func() {
	//		if service.Interactive() {
	//			p.err = p.Stop(p.service)
	//		} else {
	//			p.err = p.service.Stop()
	//		}
	//	}()
	// }

	p.err = p.prepareLogFiles()
	if p.err != nil {
		panic(p.err)
	}

	if p.InvokedInDaemon {
		// go func() {
		p.pidFile = newPidFile(p.PidFileName())
		p.pidFile.Create(cmd)
		stop, done := sig.GetChs()
		p.err = pd.daemon.OnRun(p, stop, done, nil)
		// }()

	} else if p.InvokedDirectly {
		p.pidFile = newPidFile(p.PidFileName())
		p.pidFile.Create(cmd)
		stop, done := sig.GetChs()
		p.err = pd.daemon.OnRun(p, stop, done, nil)

	} else {
		stop, done := sig.GetChs()
		p.err = pd.daemon.OnRun(p, stop, done, nil)

	}
	return
}

func (p *Program) Stop(s service.Service) (err error) {
	p.Logger.Infof("xx.pp.stop; Args: %v;", os.Args)

	err = pd.daemon.OnStop(p)

	// Stop should not block. Return with a few seconds.
	// <-time.After(time.Second * 13)

	stop, done := sig.GetChs()
	close(stop)

	if p.pidFile != nil {
		p.pidFile.Destroy()
	}

	if p.fErr != nil {
		err = p.fErr.Close()
	}

	if p.fOut != nil {
		err = p.fOut.Close()
	}

	// logger.Info("Stopping ", p.DisplayName)
	// if p.Command.ProcessState.Exited() == false {
	// 	err = p.Command.Process.Kill()
	// }
	if service.Interactive() {
		os.Exit(0)
	}
	close(done)
	return
}
