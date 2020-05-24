module github.com/hedzr/cmdr-addons

go 1.14

// replace github.com/hedzr/logex => ../logex

replace github.com/hedzr/cmdr => ../cmdr

replace github.com/kardianos/service => ../../kardianos/service

require (
	github.com/hedzr/cmdr v1.6.39
	github.com/kardianos/service v1.0.0
	github.com/sirupsen/logrus v1.6.0
	gopkg.in/hedzr/errors.v2 v2.0.12
	gopkg.in/yaml.v3 v3.0.0-20200504145624-a81ed60d5f3a
)
