module github.com/hedzr/cmdr-addons

go 1.14

//replace github.com/hedzr/log => ../10.log

//replace github.com/hedzr/logex => ../15.logex

//replace github.com/hedzr/cmdr => ../50.cmdr

//replace github.com/kardianos/service => ../../kardianos/service

require (
	github.com/ajg/form v1.5.1 // indirect
	github.com/c-bata/go-prompt v0.2.6
	github.com/fasthttp-contrib/websocket v0.0.0-20160511215533-1f3b11f56072 // indirect
	github.com/gin-gonic/gin v1.7.2
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/hedzr/cmdr v1.9.0
	github.com/hedzr/log v1.3.22
	github.com/imkira/go-interpol v1.1.0 // indirect
	github.com/k0kubun/colorstring v0.0.0-20150214042306-9440f1994b88 // indirect
	github.com/kardianos/service v1.2.0
	github.com/kataras/iris/v12 v12.1.8
	github.com/labstack/echo-contrib v0.11.0
	github.com/labstack/echo/v4 v4.3.0
	github.com/labstack/gommon v0.3.0
	github.com/moul/http2curl v1.0.0 // indirect
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/valyala/fasthttp v1.27.0 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/yalp/jsonpath v0.0.0-20180802001716-5cc68e5049a0 // indirect
	github.com/yudai/gojsondiff v1.0.0 // indirect
	github.com/yudai/golcs v0.0.0-20170316035057-ecda9a501e82 // indirect
	github.com/yudai/pp v2.0.1+incompatible // indirect
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	gopkg.in/hedzr/errors.v2 v2.1.5
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)
