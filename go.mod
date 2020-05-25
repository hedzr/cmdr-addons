module github.com/hedzr/cmdr-addons

go 1.14

// replace github.com/hedzr/logex => ../logex

// replace github.com/hedzr/cmdr => ../cmdr

// replace github.com/kardianos/service => ../../kardianos/service

require (
	github.com/gin-gonic/gin v1.6.3
	github.com/gorilla/mux v1.7.4
	github.com/hedzr/cmdr v1.6.39
	github.com/kardianos/service v1.0.0
	github.com/kataras/iris/v12 v12.1.8
	github.com/labstack/echo-contrib v0.9.0
	github.com/labstack/echo/v4 v4.1.16
	github.com/labstack/gommon v0.3.0
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/sirupsen/logrus v1.6.0
	golang.org/x/crypto v0.0.0-20200510223506-06a226fb4e37
	gopkg.in/hedzr/errors.v2 v2.0.12
	gopkg.in/yaml.v3 v3.0.0-20200504145624-a81ed60d5f3a
)
