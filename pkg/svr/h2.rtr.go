package svr

import (
	"net"
	"net/http"
)

type RouterMux interface {
	BuildRoutes()

	Handler() http.Handler
	App() http.Handler

	PreServe() (err error)
	Serve(srv *http.Server, listener net.Listener, certFile, keyFile string) (err error)
	PostServe() (err error)
}

// ForLoggerInitializing can be used for your logger initializing inside a router, such as iris.Use(logger.New()).
type ForLoggerInitializing interface {
	PrePreServe()
}
