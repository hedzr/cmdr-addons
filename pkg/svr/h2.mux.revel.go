package svr

import (
	"net"
	"net/http"
)

// https://revel.github.io/
// https://github.com/revel/revel

func newRevel() *revelImpl {
	d := &revelImpl{}
	d.init()
	return d
}

type revelImpl struct {
}

func (d *revelImpl) init() {
}

func (d *revelImpl) PreServe() (err error) {
	return
}

func (d *revelImpl) PostServe() (err error) {
	return
}

func (d *revelImpl) Handler() http.Handler {
	panic("implement me")
}

func (d *revelImpl) Serve(srv *http.Server, listener net.Listener, certFile, keyFile string) (err error) {
	panic("implement me")
}

func (d *revelImpl) BuildRoutes() {
	panic("implement me")
}
