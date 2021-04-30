package svr

import (
	"context"
	"github.com/hedzr/cmdr-addons/pkg/svr/tls"
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

// SpecialRun provides the ability how a RouterMux object listen-and-serv to ...
type SpecialRun interface {
	Run(config *tls.CmdrTLSConfig, srv *http.Server, hotReloadListener net.Listener) (err error)
}
