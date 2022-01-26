/*
 * Copyright © 2019 Hedzr Yeh.
 */

package dex

import (
	"fmt"
	"github.com/hedzr/cmdr"
	"github.com/hedzr/cmdr-addons/pkg/plugins/dex/sig"
	"github.com/hedzr/log/dir"
	"gopkg.in/hedzr/errors.v2"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

type pidFileStruct struct {
	Path string
}

// var pidfile = &pidFileStruct{}

func newPidFile(filename string) *pidFileStruct {
	return &pidFileStruct{
		Path: filename,
	}
}

func (pf *pidFileStruct) Create(cmd *cmdr.Command) {
	// if cmdr.GetBoolR("server.start.in-daemon") {
	//
	// }
	f, err := os.OpenFile(pf.Path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0770)
	if err != nil {
		panic(errors.New("Failed to create pid file %q", pf.Path).Attach(err))
	}
	defer f.Close()
	f.WriteString(fmt.Sprintf("%v", os.Getpid()))
}

func (pf *pidFileStruct) Destroy() {
	// if cmdr.GetBoolR("server.start.in-daemon") {
	//	//
	// }
	if dir.FileExists(pf.Path) {
		err := os.RemoveAll(pf.Path)
		if err != nil {
			panic(errors.New("Failed to destroy pid file %q", pf.Path).Attach(err))
		}
	}
}

type loggerStruct struct {
}

var logger = &loggerStruct{}

func (l *loggerStruct) Setup(cmd *cmdr.Command) {
	//
}

func pidExistsDeep(pid int) (bool, error) {
	// pid, err := strconv.ParseInt(p, 0, 64)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	process, err := os.FindProcess(int(pid))
	if err != nil {
		// fmt.Printf("Failed to find process: %s\n", err)
		return false, nil
	}

	err = sig.SendNilSig(process)
	log.Printf("process.Signal on pid %d returned: %v\n", pid, err)
	return err == nil, err
}

// IsPidFileExists checks if the pid file exists or not
func IsPidFileExists() bool {
	// check if daemon already running.
	if _, err := os.Stat(pd.PidFileName()); err == nil {
		return true

	}
	return false
}

// FindDaemonProcess locates the daemon process if running
func FindDaemonProcess() (present bool, process *os.Process) {
	if IsPidFileExists() {
		s, _ := ioutil.ReadFile(pd.PidFileName())
		pid, err := strconv.ParseInt(string(s), 0, 64)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("cat %v ... pid = %v", pd.PidFileName(), pid)

		process, err = os.FindProcess(int(pid))
		if err == nil {
			present = true
		}
	} else {
		log.Printf("cat %v ... app stopped", pd.PidFileName())
	}
	return
}

const nullDev = "/dev/null"
