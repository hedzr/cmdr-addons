package service

import (
	"context"
	"os"
	"testing"

	"github.com/hedzr/is"
	"gopkg.in/hedzr/errors.v3"
)

var demoConfig = &Config{
	Name:           "demo",
	DisplayName:    "demo service",
	Description:    "demo",
	WorkDir:        "",
	Executable:     "",
	ArgsForInstall: nil,
	Env:            nil,
	RunAs:          "",
	Dependencies:   nil,
}

func TestInstall(t *testing.T) {
	ctx := context.Background()
	t.Run("uninstall at first", func(t *testing.T) {
		svc := New(ctx)
		err := svc.Control(ctx, demoConfig, Uninstall)
		if err != nil {
			if errors.Is(err, errors.Unavailable) {
				return
			}
			t.Fatal(err)
		}
	})
	t.Run("install", func(t *testing.T) {
		svc := New(ctx)
		err := svc.Control(ctx, demoConfig, Install)
		if err != nil {
			if errors.Is(err, errors.Unavailable) {
				return
			}
			t.Fatal(err)
		}
	})
	t.Run("run", func(t *testing.T) {
		svc := New(ctx)
		inCI := is.ToBool(os.Getenv("CI_RUNNING"))
		_ = inCI
		err := svc.Run(ctx, demoConfig)
		if err != nil {
			if errors.Iss(err, errors.Unavailable, ErrServiceIsRunning) {
				return
			}
			if demoConfig.RetCode == 1 && inCI {
				// sometimes this is not a real error if service was dead or some resource unavailable.
				return
			}
			t.Fatal(err)
		}
	})
	t.Run("uninstall", func(t *testing.T) {
		svc := New(ctx)
		err := svc.Control(ctx, demoConfig, Uninstall)
		if err != nil {
			if errors.Is(err, errors.Unavailable) || errors.Is(err, ErrServiceIsNotEnabled) {
				return
			}
			t.Fatal(err)
		}
	})
}

// func TestNew(t *testing.T) {
// 	ctx := context.Background()
// }
//
// func TestUninstall(t *testing.T) {
// 	ctx := context.Background()
// }

func TestCommand_MarshalText(t *testing.T) {
	t.Log(Install)
	println(Install)
}
