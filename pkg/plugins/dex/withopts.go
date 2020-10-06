// Copyright © 2020 Hedzr Yeh.

package dex

import (
	"github.com/hedzr/cmdr"
	"github.com/hedzr/log"
	"github.com/kardianos/service"
)

// Opt is functional option type
type Opt func()

// WithLogger sets the user-customizable logger
func WithLogger(logger log.Logger) Opt {
	return func() {
		pd.log = logger
	}
}

// WithServiceConfig appends a customized *service.Config object
func WithServiceConfig(config *service.Config) Opt {
	return func() {
		pd.Config = config
	}
}

// WithLoggerForward forwards all logging outputs to the out and
// err files.
// Typically the files could be found at `/var/log/<appname>/`.
func WithLoggerForward(force bool) Opt {
	return func() {
		pd.ForwardLogToFile = force
	}
}

// WithCommandsModifier could specify a modifier, which allows the tune can
// be applied onto daemon-server-command.
func WithCommandsModifier(modifier func(daemonServerCommand *cmdr.Command) *cmdr.Command) Opt {
	return func() {
		pd.modifier = modifier
	}
}

// WithPreAction appends a pre-action handler to the pre-actions chain.
func WithPreAction(action func(cmd *cmdr.Command, args []string) (err error)) Opt {
	return func() {
		pd.preActions = append(pd.preActions, action)
	}
}

// WithPostAction appends a post-action handler to the post-actions chain.
func WithPostAction(action func(cmd *cmdr.Command, args []string)) Opt {
	return func() {
		pd.postActions = append(pd.postActions, action)
	}
}

// // WithOnGetListener returns tcp/http listener for daemon hot-restarting
// func WithOnGetListener(fn func() net.Listener) Opt {
// 	return func() {
// 		impl.SetOnGetListener(fn)
// 	}
// }
