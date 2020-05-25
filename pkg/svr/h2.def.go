// Copyright © 2020 Hedzr Yeh.

package svr

const (
	defaultPort = 1379

	typeDefault muxType = iota
	typeGin
	typeIris
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
	// 	routerImpl  routerMux
	// 	// router      *gin.Engine
	// 	// irisApp     *iris.Application
	// }
)
