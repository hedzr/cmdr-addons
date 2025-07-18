package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hedzr/is"
	"github.com/hedzr/is/term"
	"github.com/hedzr/is/term/color"
)

func main() {
	var cancelled int32
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	is.RegisterStateGetter("custom", func() bool { return is.InVscodeTerminal() })

	println("state.InTesting:   ", is.InTesting())
	println("state.in-testing:  ", is.State("in-testing"))
	println("state.custom:      ", is.State("custom")) // detects a state with custom detector
	println("env.GetDebugLevel: ", is.Env().GetDebugLevel())
	if is.InDebugMode() {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug})))
	}

	fmt.Printf("\n%v", color.GetCPT().Translate(`<code>code</code> | <kbd>CTRL</kbd>
		<b>bold / strong / em</b>
		<i>italic / cite</i>
		<u>underline</u>
		<mark>inverse mark</mark>
		<del>strike / del </del>
		<font color="green">green text</font>
`, color.FgDefault))

	println("term.IsTerminal:               ", term.IsTerminal(int(os.Stdout.Fd())))
	println("term.IsAnsiEscaped:            ", term.IsAnsiEscaped(color.GetCPT().Translate(`<code>code</code>`, color.FgDefault)))
	println("term.IsCharDevice(stdout):     ", term.IsCharDevice(os.Stdout))
	rows, cols, err := term.GetFdSize(os.Stdout.Fd())
	println("term.GetFdSize(stdout):        ", rows, cols, err)
	rows, cols, err = term.GetTtySizeByFd(os.Stdout.Fd())
	println("term.GetTtySizeByFd(stdout):   ", rows, cols, err)
	rows, cols, err = term.GetTtySizeByFile(os.Stdout)
	println("term.GetTtySizeByFile(stdout): ", rows, cols, err)
	println("term.IsStartupByDoubleClick:   ", term.IsStartupByDoubleClick())

	catcher := is.Signals().Catch()
	catcher.
		WithPrompt("Press CTRL-C to quit...").
		// WithOnLoop(dbStarter, cacheStarter, mqStarter).
		WithOnSignalCaught(func(ctx context.Context, sig os.Signal, wg *sync.WaitGroup) {
			println()
			slog.Info("signal caught", "sig", sig)
			cancel() // cancel user's loop, see Wait(...)
		}).
		WaitFor(ctx, func(ctx context.Context, closer func()) {
			slog.Debug("entering looper's loop...")
			go func() {
				// to terminate this app after a while automatically:
				time.Sleep(10 * time.Second)
				// closer will send a os.SigInt to cause `WithOnSignalCaught` calling `cancel()`
				closer()
			}()

			<-ctx.Done() // waiting until any os signal caught

			if atomic.CompareAndSwapInt32(&cancelled, 0, 1) {
				is.PressAnyKeyToContinue(os.Stdin)
			}
		})
}

type dbMgr struct{}

func (*dbMgr) Close()                           {}         // before app terminatine
func (*dbMgr) Open(context.Context) (err error) { return } // ran before catcher.WaitFor()
