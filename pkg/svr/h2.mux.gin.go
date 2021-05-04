package svr

import (
	"github.com/gin-gonic/gin"
	"github.com/hedzr/cmdr-addons/pkg/svr/tls"
	"io"
	"net"
	"net/http"
)

func newGin() *ginImpl {
	d := &ginImpl{}
	d.init()
	return d
}

type ginImpl struct {
	router *gin.Engine
}

func (d *ginImpl) init() {
	gin.ForceConsoleColor()
	d.router = gin.New()
	d.router.Use(gin.Logger(), gin.Recovery())
	// d.router.GET("/benchmark", MyBenchLogger(), benchEndpoint)
}

//func (d *ginImpl) Shutdown(ctx context.Context) error {
//	d.router.
//	return nil
//}

func (d *ginImpl) Handler() http.Handler {
	return d.router
}

func (d *ginImpl) App() http.Handler {
	return d.router
}

func (d *ginImpl) PreServe() (err error) {
	return
}

func (d *ginImpl) PostServe() (err error) {
	return
}

func (d *ginImpl) Serve(srv *http.Server, listener net.Listener, certFile, keyFile string) (err error) {
	// note that the h2listener have not been reassigned to the exact tlsListener
	return srv.ServeTLS(listener, certFile, keyFile)
}

func (d *ginImpl) BuildRoutes() {
	// https://github.com/gin-gonic/gin
	// https://github.com/gin-contrib/multitemplate
	// https://github.com/gin-contrib
	//
	// https://www.mindinventory.com/blog/top-web-frameworks-for-development-golang/

	d.router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	d.router.GET("/hello", helloGinHandler)

	d.router.GET("/s/*action", echoGinHandler)
}

func (d *ginImpl) Run(config *tls.CmdrTLSConfig, srv *http.Server, hotReloadListener net.Listener) (err error) {
	if config.Enabled && (config.IsServerCertValid() || srv.TLSConfig.GetCertificate == nil) {
		err = d.router.RunTLS(srv.Addr, config.Cert, config.Key)
	} else if hotReloadListener != nil {
		err = d.router.RunListener(hotReloadListener)
	} else {
		err = d.router.Run(srv.Addr)
	}
	return
}

// func (d *daemonImpl) buildGinRoutes(mux *gin.Engine) (err error) {
// 	// https://github.com/gin-gonic/gin
// 	// https://github.com/gin-contrib/multitemplate
// 	// https://github.com/gin-contrib
// 	//
// 	// https://www.mindinventory.com/blog/top-web-frameworks-for-development-golang/
//
// 	mux.GET("/ping", func(c *gin.Context) {
// 		c.JSON(200, gin.H{
// 			"message": "pong",
// 		})
// 	})
// 	mux.GET("/hello", helloGinHandler)
//
// 	mux.GET("/s/*action", echoGinHandler)
// 	return
// }

func helloGinHandler(c *gin.Context) {
	_, _ = io.WriteString(c.Writer, "Hello, world!\n")
}

func echoGinHandler(c *gin.Context) {
	action := c.Param("action")
	if action == "/zero" {
		d0(8, 0) // raise a 0-divide panic and it will be recovered by http.Conn.serve(ctx)
	}
	_, _ = io.WriteString(c.Writer, action)
}
