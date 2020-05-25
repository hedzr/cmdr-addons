package svr

import (
	"crypto/tls"
	"github.com/hedzr/cmdr"
	"github.com/labstack/echo-contrib/jaegertracing"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// https://echo.labstack.com/
// https://github.com/labstack/echo

func newEcho() *echoImpl {
	d := &echoImpl{}
	d.init()
	return d
}

type echoImpl struct {
	e             *echo.Echo
	jaegerTracing io.Closer
}

func (d *echoImpl) init() {
	d.e = echo.New()

	// https://echo.labstack.com/middleware/logger
	l := cmdr.GetLoggerLevel()
	n := log.DEBUG
	switch l {
	case cmdr.OffLevel:
		n = log.OFF
	case cmdr.FatalLevel, cmdr.PanicLevel:
		n = log.ERROR
	case cmdr.ErrorLevel:
		n = log.ERROR
	case cmdr.WarnLevel:
		n = log.WARN
	case cmdr.InfoLevel:
		n = log.INFO
	default:
		n = log.DEBUG
		d.e.Debug = true
	}
	d.e.Logger.SetLevel(n)

	d.e.Use(middleware.Recover())
	d.e.Use(middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {}))
	d.e.Use(middleware.BodyLimit("32M"))
	d.e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 5}))

	// d.e.Logger.Fatal(d.e.Start(":1323"))
}

// // DefaultSkipper returns false which processes the middleware.
// func DefaultSkipper(echo.Context) bool {
// 	return false
// }

// urlSkipper ignores metrics route on some middleware
func urlSkipper(c echo.Context) bool {
	if strings.HasPrefix(c.Path(), "/testurl") {
		return true
	}
	return false
}

func (d *echoImpl) PreServe() (err error) {
	// https://echo.labstack.com/middleware/jaegertracing
	//
	// Usage (in bash):
	// $ JAEGER_AGENT_HOST=192.168.1.10 JAEGER_AGENT_PORT=6831 ./myserver
	//
	d.jaegerTracing = jaegertracing.New(d.e, urlSkipper)
	return
}

func (d *echoImpl) PostServe() (err error) {
	if d.jaegerTracing != nil {
		err = d.jaegerTracing.Close()
	}
	return
}

func (d *echoImpl) Handler() http.Handler {
	// panic("implement me")
	return d.e
}

func (d *echoImpl) Serve(srv *http.Server, listener net.Listener, certFile, keyFile string) (err error) {
	// panic("implement me")
	// d.e.Logger.Fatal(d.e.Start(":1323"))

	if listener != nil {
		h2listener = tls.NewListener(listener, srv.TLSConfig)
		// d.e.Listener = h2listener
		d.e.TLSListener = h2listener
	}

	err = d.e.StartServer(srv)
	return
}

func (d *echoImpl) BuildRoutes() {
	// panic("implement me")

	d.e.GET("/", func(c echo.Context) error {
		// Wrap slowFunc on a new span to trace it's execution passing the function arguments
		jaegertracing.TraceFunction(c, slowFunc, "Test String")

		return c.String(http.StatusOK, "Hello, echo World!")
	})

	d.e.GET("/child-span", func(c echo.Context) error {
		// Do something before creating the child span
		time.Sleep(40 * time.Millisecond)
		sp := jaegertracing.CreateChildSpan(c, "Child span for additional processing")
		defer sp.Finish()
		sp.LogEvent("Test log")
		sp.SetBaggageItem("Test baggage", "baggage")
		sp.SetTag("Test tag", "New Tag")
		time.Sleep(100 * time.Millisecond)
		return c.String(http.StatusOK, "Hello, echo World, child span!")
	})
}

// A function to be wrapped. No need to change it's arguments due to tracing
func slowFunc(s string) {
	time.Sleep(200 * time.Millisecond)
	return
}
