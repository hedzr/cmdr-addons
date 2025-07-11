package service

import (
	"bytes"
	"context"
	"os"
	"strings"

	"github.com/hedzr/cmdr-addons/service/v2/systems"
)

type upstartD struct{}

func (s *upstartD) Choose(ctx context.Context) (ok bool) {
	if systems.HasLinuxBackends {
		ok = hasUpstart(ctx)
	}
	return
}

func hasUpstart(ctx context.Context) bool {
	return false
}

func (s *upstartD) IsValid(ctx context.Context) (valid bool) {
	if systems.HasNTService {
	}
	return
}

func (s *upstartD) Control(ctx context.Context, config *Config, m *mgmtS, cmd Command) (err error) {
	if systems.HasNTService {
	}
	return
}

//

//

type sysvInitD struct{}

func (s *sysvInitD) Choose(ctx context.Context) (ok bool) {
	if systems.HasLinuxBackends {
		ok = hasSysvInitD(ctx)
	}
	return
}

func hasSysvInitD(ctx context.Context) bool {
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
		_ = err
		contents := buf.String()

		if strings.Trim(contents, " \r\n") == "init" {
			return true
		}
	}
	return false
}

func (s *sysvInitD) IsValid(ctx context.Context) (valid bool) {
	if systems.HasNTService {
	}
	return
}

func (s *sysvInitD) Control(ctx context.Context, config *Config, m *mgmtS, cmd Command) (err error) {
	if systems.HasNTService {
	}
	return
}

//

//

//
