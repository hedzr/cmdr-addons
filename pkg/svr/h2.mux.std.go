package svr

import (
	"io"
	"net"
	"net/http"
)

func newStdMux() *stdImpl {
	d := &stdImpl{}
	d.init()
	return d
}

type stdImpl struct {
	mux *http.ServeMux
}

func (d *stdImpl) init() {
	d.mux = http.NewServeMux()
}

func (d *stdImpl) PreServe() (err error) {
	return
}

func (d *stdImpl) PostServe() (err error) {
	return
}

func (d *stdImpl) Handler() http.Handler {
	return d.mux
}

func (d *stdImpl) Serve(srv *http.Server, listener net.Listener, certFile, keyFile string) (err error) {
	// NOTE that the h2listener have not been reassigned to the exact tlsListener
	return srv.ServeTLS(listener, certFile, keyFile)
}

func (d *stdImpl) BuildRoutes() {
	d.mux.HandleFunc("/hello", helloHandler)
	d.mux.HandleFunc("/", echoHandler)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = io.WriteString(w, "Hello, world!\n")
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = io.WriteString(w, r.URL.Path)
	if r.URL.Path == "/zero" {
		d0(8, 0) // raise a 0-divide panic and it will be recovered by http.Conn.serve(ctx)
	}
}

func d0(a, b int) int {
	return a / b
}
