package service

import (
	"fmt"
	"os"
	"strings"

	"github.com/hedzr/is/dir"
)

func (e *Config) String() string {
	return e.Name
}

func (e *Config) BaseName() string {
	if e.Name != "" {
		return e.Name
	}
	if e.Entity != nil {
		return e.Entity.Name()
	}
	return ""
}

func (e *Config) ScreenName() string {
	if e.DisplayName != "" {
		return e.DisplayName
	}
	if e.Entity != nil {
		return e.Entity.ScreenName()
	}
	return ""
}

func (e *Config) ServiceName() string {
	n := e.Name
	if n == "" {
		n = e.DisplayName
	}
	if n == "" {
		if e.Entity != nil {
			n = e.Entity.ServiceName()
		}
	}
	if !strings.HasSuffix(n, ".service") {
		n += ".service"
	}
	return n
}

func (e *Config) Desc() string {
	if e.Description != "" {
		return e.Description
	}
	if e.Entity != nil {
		return e.Entity.Desc()
	}
	return ""
}

func (e *Config) ExecutablePath() string {
	if e.Executable != "" {
		return e.Executable
	}
	if e.Entity != nil {
		if str := e.Entity.ExecutablePath(); str != "" {
			return str
		}
	}
	return ""
}

func (e *Config) makeSafety() {
	if e.Executable == "" || !dir.FileExists(e.Executable) {
		e.Executable = dir.GetExecutablePath()
	}

	if e.WorkDir == "" || !dir.FileExists(e.WorkDir) {
		e.WorkDir = dir.GetExecutableDir()
	}

	if e.TempDir == "" {
		e.TempDir = os.TempDir()
	}
	if e.RunDir == "" {
		e.RunDir = "/var/run"
		if !dir.FileExists(e.RunDir) {
			e.RunDir = os.TempDir()
		}
	}
	if e.LogDir == "" {
		e.LogDir = "/var/log"
		if !dir.FileExists(e.LogDir) {
			e.LogDir = os.TempDir()
		}
	}

	if e.StandardOutPath == "" {
		e.StandardOutPath = fmt.Sprintf("%s/stdout.log", e.LogDir)
	}
	if e.StandardErrorPath == "" {
		e.StandardErrorPath = fmt.Sprintf("%s/stderr.log", e.LogDir)
	}
}
