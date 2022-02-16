// Copyright © 2020 Hedzr Yeh.

package svr

import (
	"github.com/hedzr/cmdr"
	"github.com/hedzr/cmdr-addons/pkg/plugins/dex"
	"github.com/hedzr/cmdr-addons/pkg/svr/tls"
	"time"
)

func (d *daemonImpl) OnCmdrPrepare(prog *dex.Program, root *cmdr.RootCommand) (err error) {
	serverCmd := root.FindSubCommand("server")
	serverStartCmd := serverCmd.FindSubCommand("start")

	opt := cmdr.NewCmdFrom(serverCmd)

	if flg := serverCmd.FindFlag("port"); flg != nil {
		flg.DefaultValue = defaultPort

	} else {
		cmdr.NewInt(defaultPort).Titles("port", "p").
			Description("the port to listen.", "").
			Group("").
			Placeholder("PORT").
			AttachTo(opt)
	}

	ox := cmdr.NewCmdFrom(serverStartCmd)

	cmdr.NewBool(false).
		Titles("socket", "se").
		Description("enable the listening on unix sock", "").
		Group("Unix Sock").
		AttachTo(ox)
	cmdr.NewString("").
		Titles("socket-file", "sf").
		Description("the listening unix sock file (/var/run/app/app.sock)", "").
		Placeholder("FILE").
		Group("Unix Sock").
		AttachTo(ox)
	cmdr.NewBool(false).
		Titles("reset-socket-file", "").
		Description("unlink/delete the residual unix sock file at first (for the abnormal terminated)", "").
		Group("Unix Sock").
		AttachTo(ox)

	// Server-Type radio group

	cmdr.NewBool(true).
		Titles("h2-server", "h2", "").
		Description("start as a HTTP/2 server", "").
		ToggleGroup("Server-Type").
		AttachTo(ox)
	cmdr.NewBool(false).
		Titles("cmdexec-loop", "lp").
		Description("start a worker and loop for cmd exec", "").
		ToggleGroup("Server-Type").
		AttachTo(ox)

	//

	certOptCmd := cmdr.NewSubCmd().Titles("certs", "ca").
		Description("certificates operations...", "").
		Group("CA").
		AttachTo(opt)
	certCreateCmd := cmdr.NewSubCmd().Titles("create", "c").
		Description("create CA, server and client certificates").
		Action(tls.CertCreate).
		AttachTo(certOptCmd)
	cmdr.NewStringSlice().Titles("host", "h").
		Description("Comma-separated hostnames and IPs to generate a certificate for").
		Placeholder("HOSTNAMES,...").
		AttachTo(certCreateCmd)
	cmdr.NewString().Titles("start-date", "f", "from", "valid-from").
		Description("Creation date formatted as Jan 1 15:04:05 2011 (default now)").
		Placeholder("DATETIME").
		AttachTo(certCreateCmd)
	cmdr.NewDuration(365*10*24*time.Hour).Titles("valid-for", "duration", "d").
		Description("Duration that certificate is valid for").
		AttachTo(certCreateCmd)
	cmdr.NewString().Titles("CN", "cn", "common-name").
		Description("common name string").
		Placeholder("CN").
		AttachTo(certCreateCmd)

	// caCmd := certOptCmd.NewSubCommand("ca").
	// 	Description("certification tool (such as create-ca, create-cert, ...)", "certification tool (such as create-ca, create-cert, ...          )\nverbose long descriptions here.").
	// 	Group("CA")
	//
	// caCreateCmd := caCmd.NewSubCommand("create", "c").
	// 	Description("create NEW CA certification", "").
	// 	Group("Tool").
	// 	Action(tls.CaCreate)

	cmdr.NewBool(false).
		Titles("iris", "iris").
		Description("use Iris engine", "").
		ToggleGroup("Mux").
		AttachTo(opt)
	cmdr.NewBool(true).
		Titles("gin", "gin").
		Description("use Gin engine", "").
		ToggleGroup("Mux").
		AttachTo(opt)
	cmdr.NewBool(false).
		Titles("gorilla", "gorilla").
		Description("use Gorilla engine", "").
		ToggleGroup("Mux").
		AttachTo(opt)
	cmdr.NewBool(false).
		Titles("echo", "echo").
		Description("use Echo engine", "").
		ToggleGroup("Mux").
		AttachTo(opt)
	cmdr.NewBool(false).
		Titles("std", "std").
		Description("use stdlib http.ServerMux engine", "").
		ToggleGroup("Mux").
		AttachTo(opt)

	// http 2 client

	cmdr.NewSubCmd().
		Titles("h2-test", "h2").
		Description("test http 2 client", "test http 2 client,\nverbose long descriptions here.").
		Group("Test").
		Action(func(cmd *cmdr.Command, args []string) (err error) {
			// runClient()
			return
		}).
		AttachToRoot(root)
	return
}

func (d *daemonImpl) BeforeServiceStart(prog *dex.Program, root *cmdr.Command) (err error) {
	return
}

func (d *daemonImpl) AfterServiceStop(prog *dex.Program, root *cmdr.Command) (err error) {
	return
}
