module github.com/hedzr/cmdr-addons

go 1.16

//replace github.com/hedzr/log => ../10.log

//replace github.com/hedzr/logex => ../15.logex

//replace github.com/hedzr/cmdr => ../50.cmdr

//replace github.com/kardianos/service => ../../kardianos/service

require (
	github.com/c-bata/go-prompt v0.2.6
	github.com/gin-gonic/gin v1.7.7
	github.com/gorilla/mux v1.8.0
	github.com/hedzr/cmdr v1.10.13
	github.com/hedzr/log v1.5.9
	github.com/kardianos/service v1.2.1
	github.com/labstack/echo-contrib v0.12.0
	github.com/labstack/echo/v4 v4.6.3
	github.com/labstack/gommon v0.3.1
	golang.org/x/crypto v0.0.0-20220112180741-5e0467b6c7ce
	gopkg.in/hedzr/errors.v2 v2.1.5
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)
