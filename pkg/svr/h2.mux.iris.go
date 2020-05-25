package svr

import (
	"crypto/tls"
	"github.com/hedzr/cmdr"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"
	"net"
	"net/http"
)

func newIris() *irisImpl {
	d := &irisImpl{}
	d.init()
	return d
}

type irisImpl struct {
	irisApp *iris.Application
}

func (d *irisImpl) init() {
	d.irisApp = iris.New()

	l := cmdr.GetLoggerLevel()
	n := "debug"
	switch l {
	case cmdr.OffLevel:
		n = "disable"
	case cmdr.FatalLevel, cmdr.PanicLevel:
		n = "fatal"
	case cmdr.ErrorLevel:
		n = "error"
	case cmdr.WarnLevel:
		n = "warn"
	case cmdr.InfoLevel:
		n = "info"
	default:
		n = "debug"
	}

	d.irisApp.Use(recover.New())
	d.irisApp.Use(logger.New())
	d.irisApp.Logger().SetLevel(n)
}

func (d *irisImpl) PreServe() (err error) {
	return
}

func (d *irisImpl) PostServe() (err error) {
	return
}

func (d *irisImpl) Handler() http.Handler {
	return d.irisApp
}

func (d *irisImpl) Serve(srv *http.Server, listener net.Listener, certFile, keyFile string) (err error) {
	// return d.irisApp.Run(iris.Raw(func() error {
	// 	su := d.irisApp.NewHost(srv)
	// 	if netutil.IsTLS(su.Server) {
	// 		h2listener = tls.NewListener(listener, su.Server.TLSConfig)
	// 		// logrus.Debugf("new h2listener: %v", su.Server.TLSConfig)
	// 		su.Configure(func(su *host.Supervisor) {
	// 			rs := reflect.ValueOf(su).Elem()
	// 			// rf := rs.FieldByName("manuallyTLS")
	// 			rf := rs.Field(2)
	// 			// rf can't be read or set.
	// 			rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
	// 			// Now rf can be read and set.
	//
	// 			// su.manuallyTLS = true
	// 			i := true
	// 			ri := reflect.ValueOf(&i).Elem() // i, but writeable
	// 			rf.Set(ri)
	// 		})
	// 	}
	// 	err = su.Serve(listener)
	// 	return err
	// }), iris.WithoutServerError(iris.ErrServerClosed))

	if listener != nil {
		h2listener = tls.NewListener(listener, srv.TLSConfig)
		return d.irisApp.Run(iris.Listener(h2listener), iris.WithoutServerError(iris.ErrServerClosed))
	}

	return d.irisApp.Run(iris.Server(srv), iris.WithoutServerError(iris.ErrServerClosed))
}

func (d *irisImpl) BuildRoutes() {
	// https://iris-go.com/start/
	// https://github.com/kataras/iris
	//
	// https://www.slant.co/topics/1412/~best-web-frameworks-for-go

	d.irisApp.Get("/", func(c iris.Context) {
		_, _ = c.JSON(iris.Map{"message": "Hello Iris!"})
	})
	d.irisApp.Get("/ping", func(ctx iris.Context) {
		_, _ = ctx.WriteString("pong")
	})
	// Resource: http://localhost:1380
	d.irisApp.Handle("GET", "/welcome", func(ctx iris.Context) {
		_, _ = ctx.HTML("<h1>Welcome</h1>")
	})

	// app.Run(iris.Addr(":8080"), iris.WithoutServerError(iris.ErrServerClosed))

	d.irisApp.Get("/s/:path", d.echoIrisHandler)

	//
	// d.irisApp.Get("/users/{id:uint64}", func(ctx iris.Context){
	// 	id := ctx.Params().GetUint64Default("id", 0)
	// 	// [...]
	// })
	// d.irisApp.Get("/profile/{name:alphabetical max(255)}", func(ctx iris.Context){
	// 	name := ctx.Params().Get("name")
	// 	// len(name) <=255 otherwise this route will fire 404 Not Found
	// 	// and this handler will not be executed at all.
	// })
	//
	// d.irisApp.Get("/someGet", getting)
	// d.irisApp.Post("/somePost", posting)
	// d.irisApp.Put("/somePut", putting)
	// d.irisApp.Delete("/someDelete", deleting)
	// d.irisApp.Patch("/somePatch", patching)
	// d.irisApp.Head("/someHead", head)
	// d.irisApp.Options("/someOptions", options)

	// user.Register(d.irisApp)
}

func (d *irisImpl) echoIrisHandler(ctx iris.Context) {
	p := ctx.Params().GetString("path")
	if p == "zero" {
		d0(8, 0) // raise a 0-divide panic and it will be recovered by http.Conn.serve(ctx)
	}
	_, _ = ctx.WriteString(p)
}
