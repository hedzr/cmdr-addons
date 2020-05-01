module github.com/hedzr/cmdr-addons

go 1.14

// replace github.com/hedzr/logex => ../logex

replace github.com/hedzr/cmdr => ../cmdr

replace github.com/kardianos/service => ../../kardianos/service

require (
	github.com/hedzr/cmdr v1.6.35
	github.com/kardianos/service v1.0.0
	github.com/sirupsen/logrus v1.5.0
)
