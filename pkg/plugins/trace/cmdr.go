// Copyright © 2020 Hedzr Yeh.

package trace

import (
	"github.com/hedzr/cmdr"
)

// WithTraceEnable enables a minimal `trace` option at cmdr Root Command Level.
func WithTraceEnable(enabled bool) cmdr.ExecOption {
	return func(w *cmdr.ExecWorker) {
		// daemonImpl = daemonImplX

		w.AddOnBeforeXrefBuilding(func(root *cmdr.RootCommand, args []string) {

			if enabled {
				// attaches `--trace` to root command
				cmdr.NewBool(false).
					Titles("trace", "tr").
					Description("enable trace mode for tcp/mqtt send/recv data dump").
					Group(cmdr.SysMgmtGroup).
					EnvKeys("TRACE").
					OnSet(func(keyPath string, value interface{}) {
						// fmt.Printf("trace: %v\n", value)
						b := cmdr.ToBool(value)
						if b {
							_ = Start()
							root.AppendPostActions(func(cmd *cmdr.Command, args []string) {
								Stop()
							})
						}
					}).
					AttachToRoot(root)
			}

		})
	}
}
