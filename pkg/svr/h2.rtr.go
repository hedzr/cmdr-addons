package svr

import (
	"context"
	"net"
	"net/http"
)

// RouterMux wrap a generic mux object
type RouterMux interface {
	BuildRoutes()

	Handler() http.Handler
	App() http.Handler

	PreServe() (err error)
	Serve(srv *http.Server, listener net.Listener, certFile, keyFile string) (err error)
	PostServe() (err error)
}

// GracefulShutdown expose a Shutdown method to parent
type GracefulShutdown interface {
	Shutdown(ctx context.Context) error
}

// ForLoggerInitializing can be used for your logger initializing inside a router, such as iris.Use(logger.New()).
type ForLoggerInitializing interface {
	PrePreServe()
}
