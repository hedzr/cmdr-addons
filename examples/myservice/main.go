package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hedzr/cmdr-addons/v2/service"
	"github.com/hedzr/cmdr-addons/v2/tool/dbglog"
	"github.com/hedzr/is"
	"github.com/hedzr/is/basics"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("expected 'foo' or 'bar' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "foo":
		fooCmd.Parse(os.Args[2:])
		fmt.Println("subcommand 'foo'")
		fmt.Println("  enable:", *fooEnable)
		fmt.Println("  name:", *fooName)
		fmt.Println("  tail:", fooCmd.Args())
		return
	case "bar":
		barCmd.Parse(os.Args[2:])
		fmt.Println("subcommand 'bar'")
		fmt.Println("  level:", *barLevel)
		fmt.Println("  tail:", barCmd.Args())
		return

	default:
		ctx := context.Background()
		err := serverOper(ctx, os.Args[1], os.Args[2:])
		if err != nil {
			log.Printf("Application Error: %+v\n", err)
			os.Exit(1)
		}
	}
}

func serverOper(ctx context.Context, cmd string, args []string) (err error) {
	defer basics.Close()
	switch cmd {
	case "start", "s":
		startCmd.Parse(args)
		if debug {
			is.SetDebugMode(true)
		}

		ctx1, cancel := context.WithCancel(ctx)
		defer cancel()

		cfg := newConfig()
		svc := newService(ctx1)
		svc.SetServiceMode(true)
		// err = svc.Control(ctx1, cfg, service.Start)

		// closeChan := make(chan struct{}, 8)
		// defer func() { close(closeChan) }()
		dbglog.Info("starting service...")
		err = svc.Control(ctx1, cfg, service.Start)

		// c := is.Signals().Catch()
		// c.WithOnSignalCaught(func(ctx context.Context, sig os.Signal, wgShutdown *sync.WaitGroup) {
		// 	// closeChan <- struct{}{}
		// 	println()
		// 	fmt.Printf("signal %q caught\n", sig)
		// 	cancel()
		// }).WaitFor(ctx1, func(ctx context.Context, closer func()) {
		// 	// service.Start won't return till ctx1 cancelled.
		// 	err = svc.Control(ctx1, cfg, service.Start)
		// 	closer()
		// })

		// is.PressAnyKeyToContinue(os.Stdin)

	case "run":
		err = serviceOper(ctx, startCmd, args, service.Start, true, false)

	case "stop", "halt", "shutdown", "quit":
		err = serviceOper(ctx, stopCmd, args, service.Stop, false, false)

	case "restart", "re":
		err = serviceOper(ctx, restartCmd, args, service.Restart, false, false)

	case "install", "in":
		err = serviceOper(ctx, installCmd, args, service.Install, false, false)

	case "uninstall", "remove", "rm":
		err = serviceOper(ctx, uninstallCmd, args, service.Uninstall, false, false)

	default:
		err = errors.New("expected 'start' or 'stop' subcommands")
	}
	return
}

func serviceOper(ctx context.Context, c *flag.FlagSet, args []string, cmd service.Command, foreMode, serviceMode bool) (err error) {
	c.Parse(args)

	if debug {
		is.SetDebugMode(true)
	}

	cfg := newConfig()
	svc := newService(ctx)
	if force {
		svc.SetForceMode(force)
	}
	svc.SetForegroundMode(foreMode)
	svc.SetServiceMode(serviceMode)
	err = svc.Control(ctx, cfg, cmd)
	return
}

func newConfig() (cfg *service.Config) {
	cfg = &service.Config{
		Name:        "myservice",
		DisplayName: "myservice service",
		Description: "myservice service desc here",
		// WorkDir:        "",
		// Executable:     "",
		// ArgsForInstall: nil,
		// Env:            nil,
		// RunAs:          "",
		// Dependencies:   nil,
	}
	return
}

func newService(ctx context.Context) (svc service.Manager) {
	svc = service.New(ctx)
	return
}

var (
	fooCmd *flag.FlagSet
	barCmd *flag.FlagSet

	startCmd     *flag.FlagSet
	stopCmd      *flag.FlagSet
	restartCmd   *flag.FlagSet
	installCmd   *flag.FlagSet
	uninstallCmd *flag.FlagSet

	fooEnable *bool
	fooName   *string
	barLevel  *int

	debug bool
	force bool
)

func init() {
	fooCmd = flag.NewFlagSet("foo", flag.ExitOnError)
	fooEnable = fooCmd.Bool("enable", false, "enable")
	fooName = fooCmd.String("name", "", "name")

	barCmd = flag.NewFlagSet("bar", flag.ExitOnError)
	barLevel = barCmd.Int("level", 0, "level")

	startCmd = flag.NewFlagSet("start", flag.ExitOnError)
	stopCmd = flag.NewFlagSet("stop", flag.ExitOnError)
	restartCmd = flag.NewFlagSet("restart", flag.ExitOnError)
	installCmd = flag.NewFlagSet("install", flag.ExitOnError)
	uninstallCmd = flag.NewFlagSet("uninstall", flag.ExitOnError)

	startCmd.BoolVar(&debug, "debug", false, "debug mode")
	stopCmd.BoolVar(&debug, "debug", false, "debug mode")
	restartCmd.BoolVar(&debug, "debug", false, "debug mode")
	installCmd.BoolVar(&debug, "debug", false, "debug mode")
	uninstallCmd.BoolVar(&debug, "debug", false, "debug mode")

	installCmd.BoolVar(&force, "force", false, "force reinstall")
	uninstallCmd.BoolVar(&force, "force", false, "force uninstall")
}
