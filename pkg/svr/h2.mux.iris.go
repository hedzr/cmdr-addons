package svr

import (
	"context"
	"github.com/hedzr/cmdr"
	"github.com/hedzr/cmdr/conf"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/host"
	"github.com/kataras/iris/v12/middleware/accesslog"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"
	"net"
	"net/http"
	"time"
)

func newIris() *irisImpl {
	d := &irisImpl{
		BaseIrisImpl: NewRouterBaseIrisImpl(),
	}
	return d
}

type irisImpl struct {
	BaseIrisImpl
}

func NewRouterBaseIrisImpl() BaseIrisImpl {
	d := BaseIrisImpl{}
	d.init()
	return d
}

type BaseIrisImpl struct {
	irisApp *iris.Application
	ac      *accesslog.AccessLog
}

func (d *BaseIrisImpl) init() {
	app := iris.New()
	d.irisApp = app

	app.UseRouter(recover.New())
}

func (d *BaseIrisImpl) Shutdown(ctx context.Context) error {
	return d.irisApp.Shutdown(ctx)
}

func (d *BaseIrisImpl) PrePreServe() {
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
	}
	customLogger := logger.New(logger.Config{
		// Status displays status code
		Status: true,
		// IP displays request's remote address
		IP: true,
		// Method displays the http method
		Method: true,
		// Path displays the request path
		Path: true,
		// Query appends the url query to the Path.
		Query: true,
		// Columns: false,

		// if !empty then its contents derives from `ctx.Values().Get("logger_message")
		// will be added to the logs.
		MessageContextKeys: []string{"logger_message"},

		// if !empty then its contents derives from `ctx.GetHeader("User-Agent")
		MessageHeaderKeys: []string{"User-Agent"},

		LogFunc: d.print,
	})

	app := d.irisApp

	app.Use(customLogger)
	app.Logger().SetLevel(n)
	//app.OnAnyErrorCode(customLogger, func(ctx iris.Context) {
	//	// this should be added to the logs, at the end because of the `logger.Config#MessageContextKey`
	//	ctx.Values().Set("logger_message",
	//		"a dynamic message passed to the logs")
	//	ctx.Writef("My Custom error page")
	//})

	// Note, it's buffered, so make sure it's closed so it can flush any buffered contents.
	d.ac = accesslog.File("./access.log")
	// defer ac.Close()
	app.UseRouter(d.ac.Handler)

	if cmdr.GetBoolRP(conf.AppName, "server.statics.enabled") {
		// Serve our front-end and its assets.
		app.HandleDir(cmdr.GetStringRP(conf.AppName, "server.statics.url"), iris.Dir(cmdr.GetStringRP(conf.AppName, "server.statics.path")))
	}

	//loc := cmdr.GetStringRP(conf.AppName, "server.statics.path")
	//yaag.Init(&yaag.Config{ // <- IMPORTANT, init the middleware.
	//	On:       true,
	//	DocTitle: cmdr.AppName + " via Iris",
	//	DocPath:  path.Join(loc, "apidoc.html"),
	//	BaseUrls: map[string]string{"Production": "", "Staging": ""},
	//})
	//app.UseRouter(irisyaag.New()) // <- IMPORTANT, register the middleware.
}

func (d *BaseIrisImpl) print(endTime time.Time, latency time.Duration, status, ip, method, path string, message interface{}, headerMessage interface{}) {
	cmdr.Logger.Infof("%v %v %v %v %v %v %v %v", endTime, latency, status, ip, method, path, message, headerMessage)
}

func (d *BaseIrisImpl) PreServe() (err error) {
	return
}

func (d *BaseIrisImpl) PostServe() (err error) {
	return
}

func (d *BaseIrisImpl) Handler() http.Handler { return d.irisApp }
func (d *BaseIrisImpl) App() http.Handler     { return d.irisApp }

func (d *BaseIrisImpl) Serve(srv *http.Server, listener net.Listener, certFile, keyFile string) (err error) {
	// return d.irisApp.Run(iris.Raw(func() error {
	// 	su := d.irisApp.NewHost(srv)
	// 	if netutil.IsTLS(su.Server) {
	// 		h2listener = tls.NewListener(listener, su.Server.TLSConfig)
	// 		// cmdr.Logger.Debugf("new h2listener: %v", su.Server.TLSConfig)
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
		// v1 - http 1.1 with tls:
		//h2listener = tls.NewListener(listener, srv.TLSConfig)
		//return d.irisApp.Run(iris.Listener(h2listener), iris.WithoutServerError(iris.ErrServerClosed))

		// v2 - http 2 with iris tls, but h2listener not ready
		_ = listener.Close()
		return d.irisApp.Run(d.irisTLSServer(srv),
			// iris.WithoutVersionChecker,
			//iris.WithOptimizations,
			iris.WithoutServerError(iris.ErrServerClosed))
	}

	return d.irisApp.Run(iris.Server(srv), iris.WithoutServerError(iris.ErrServerClosed))
}

//type super struct {
//	*iris.Supervisor
//}
//
//func irisTLSListener(l net.Listener, srv *http.Server, hostConfigs ...host.Configurator) iris.Runner {
//	return func(app *iris.Application) error {
//		//app.config.vhost = netutil.ResolveVHost(l.Addr().String())
//		s:= &super{Supervisor: app.NewHost(srv),}
//		return s.Configure(hostConfigs...).
//			ServeTLS(l)
//	}
//}

func (d *BaseIrisImpl) irisTLSServer(srv *http.Server, hostConfigs ...host.Configurator) iris.Runner {
	return func(app *iris.Application) error {
		host := app.NewHost(srv)
		host.RegisterOnShutdown(func() {
			d.ac.Close()
		})
		return host.Configure(hostConfigs...).
			ListenAndServeTLS("", "")
	}
}

func (d *BaseIrisImpl) BuildRoutes() {
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
