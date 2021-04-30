/*
 * Copyright © 2019 Hedzr Yeh.
 */

package dex

import "github.com/hedzr/cmdr"

var (
	// DaemonServerCommand defines a group of sub-commands for daemon operations.
	DaemonServerCommand = &cmdr.Command{
		BaseOpt: cmdr.BaseOpt{
			// Name:        "server",
			Short:       "s",
			Full:        "server",
			Aliases:     []string{"svr", "daemon"},
			Description: "server ops: for linux daemon.",
			Group:       "Daemonization",
		},
		Flags: []*cmdr.Flag{
			{
				BaseOpt: cmdr.BaseOpt{
					Short:       "p",
					Full:        "port",
					Description: "main port (RESTful).",
				},
				DefaultValue: 3000,
			},
		},
		SubCommands: []*cmdr.Command{
			{
				BaseOpt: cmdr.BaseOpt{
					Short:       "s",
					Full:        "start",
					Aliases:     []string{"run", "startup"},
					Description: "startup this system service/daemon.",
					Action:      daemonStart,
					LongDescription: `**start** command make Program running as a daemon background.
**run** command make Program running in current tty foreground.
`,
					Examples: `
$ {{.AppName}} start
					make Program running as a daemon background.
$ {{.AppName}} start --foreground
					make Program running in current tty foreground.
$ {{.AppName}} run
					make Program running in current tty foreground.
$ {{.AppName}} stop
					stop daemonized Program.
$ {{.AppName}} reload
					send signal to trigger Program reload its configurations.
$ {{.AppName}} hot-reload [TODO]
					send signal to make Program restart itself without broken any connections.
$ {{.AppName}} status
					display the demonized Program running status.
$ {{.AppName}} install
					install Program as a daemon service (win, macOS, linux systemd/upstart/init).
$ {{.AppName}} uninstall
					remove the installed daemon service.
`,
				},
				Flags: []*cmdr.Flag{
					{
						BaseOpt: cmdr.BaseOpt{
							Short:       "f",
							Full:        "foreground",
							Aliases:     []string{"fg"},
							Description: "run on foreground, instead of demonized.",
						},
						DefaultValue: false,
					},
					{
						BaseOpt: cmdr.BaseOpt{
							Short:       "hr",
							Full:        "in-hot-reload",
							Aliases:     []string{"hot-restart"},
							Description: "app is been running in hot reload mode.",
							Hidden:      true,
						},
						DefaultValue: false,
					},
					{
						BaseOpt: cmdr.BaseOpt{
							Short:       "",
							Full:        "in-daemon",
							Description: "app is been running in daemon mode (special for windows service).",
							Hidden:      true,
						},
						DefaultValue: false,
					},
				},
			},
			{
				BaseOpt: cmdr.BaseOpt{
					Short:       "t",
					Full:        "stop",
					Aliases:     []string{"halt", "pause", "shutdown"},
					Description: "stop this system service/daemon.",
					Action:      daemonStop,
				},
				Flags: []*cmdr.Flag{
					{
						BaseOpt: cmdr.BaseOpt{
							Short:       "1",
							Full:        "hup",
							Description: "send SIGHUP - to reload service configurations",
						},
						DefaultValue: false,
					},
					{
						BaseOpt: cmdr.BaseOpt{
							Short:       "3",
							Full:        "quit",
							Description: "send SIGQUIT - to quit service gracefully",
						},
						DefaultValue: false,
					},
					{
						BaseOpt: cmdr.BaseOpt{
							Short:       "9",
							Full:        "kill",
							Description: "send SIGKILL - to quit service unconditionally",
						},
						DefaultValue: false,
					},
					{
						BaseOpt: cmdr.BaseOpt{
							Short:       "15",
							Full:        "term",
							Description: "send SIGTERM - to quit service gracefully",
						},
						DefaultValue: false,
					},
					{
						BaseOpt: cmdr.BaseOpt{
							Short:       "31",
							Full:        "usr2",
							Description: "send SIGUSR2 - to hot-restart service gracefully",
						},
						DefaultValue: false,
					},
				},
			},
			{
				BaseOpt: cmdr.BaseOpt{
					Short:       "re",
					Full:        "restart",
					Aliases:     []string{"reload"},
					Description: "reload configurations for this system service/daemon.",
					Action:      daemonRestart,
				},
			},
			{
				BaseOpt: cmdr.BaseOpt{
					Short:       "hr",
					Full:        "hot-reload",
					Aliases:     []string{"hot-restart", "live-reload"},
					Description: "hot-reload this system service/daemon.",
					LongDescription: `hot-restart/hot-reload/live-reload: 

This action will start a new child process and transfer all 
living connections to the child, and shutdown itself 
gracefully.
With this action, the service will keep serving without broken.
`,
					Action: daemonHotReload,
				},
			},
			{
				BaseOpt: cmdr.BaseOpt{
					Short:       "ss",
					Full:        "status",
					Aliases:     []string{"st"},
					Description: "display its running status as a system service/daemon.",
					Action:      daemonStatus,
				},
			},
			{
				BaseOpt: cmdr.BaseOpt{
					Short:       "i",
					Full:        "install",
					Aliases:     []string{"setup"},
					Description: "install as a system service/daemon.",
					Group:       "Config",
					Action:      daemonInstall,
				},
				Flags: []*cmdr.Flag{
					{
						BaseOpt: cmdr.BaseOpt{
							Short:       "s",
							Full:        "systemd",
							Aliases:     []string{"sys"},
							Description: "install as a systemd service.",
						},
						DefaultValue: true,
						ToggleGroup:  "service-type",
					},
				},
			},
			{
				BaseOpt: cmdr.BaseOpt{
					Short:       "u",
					Full:        "uninstall",
					Aliases:     []string{"remove"},
					Description: "remove from a system service/daemon.",
					Group:       "Config",
					Action:      daemonUninstall,
				},
				Flags: []*cmdr.Flag{
					{
						BaseOpt: cmdr.BaseOpt{
							Short:       "s",
							Full:        "systemd",
							Aliases:     []string{"sys"},
							Description: "uninstall the systemd service.",
						},
						DefaultValue: true,
						ToggleGroup:  "service-type",
					},
				},
			},
		},
	}
)
