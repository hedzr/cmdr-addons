// Copyright © 2020 Hedzr Yeh.

package svr

import (
	"net"
	"net/http"
)

func getOnGetListener() net.Listener {
	// l := h2listener
	// h2listener = nil
	// return l
	return h2listener
}

var h2listener net.Listener

type routerMux interface {
	Handler() http.Handler
	Serve(srv *http.Server, listener net.Listener, certFile, keyFile string) (err error)
	BuildRoutes()

	PreServe() (err error)
	PostServe() (err error)
}
