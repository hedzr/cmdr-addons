// Copyright © 2020 Hedzr Yeh.

package svr

const (
	defaultPort = 1379

	typeDefault muxType = iota
	typeGin
	typeIrisDisabled
	typeGorilla
	typeEcho
)

type (
	muxType int

	// daemonImpl struct {
	// 	appTag      string
	// 	certManager *autocert.Manager
	// 	Type        muxType
	// 	mux         *http.ServeMux
	// 	routerImpl  RouterMux
	// 	// router      *gin.Engine
	// 	// irisApp     *iris.Application
	// }
)
