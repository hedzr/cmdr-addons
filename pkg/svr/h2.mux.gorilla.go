package svr

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net"
	"net/http"
	"strings"
)

// https://github.com/gorilla/mux

func newGorilla() *gorillaImpl {
	d := &gorillaImpl{}
	d.init()
	return d
}

type gorillaImpl struct {
	router *mux.Router
}

func (d *gorillaImpl) PreServe() (err error) {
	return
}

func (d *gorillaImpl) PostServe() (err error) {
	return
}

func (d *gorillaImpl) init() {
	d.router = mux.NewRouter()

	// r := d.Handler().(*mux.Router)
	// // Only matches if domain is "www.example.com".
	// r.Host("www.example.com")
	// // Matches a dynamic subdomain.
	// r.Host("{subdomain:[a-z]+}.example.com")
}

func (d *gorillaImpl) Handler() http.Handler {
	return d.router
}

func (d *gorillaImpl) Serve(srv *http.Server, listener net.Listener, certFile, keyFile string) (err error) {
	d.walkRoutes()

	http.Handle("/", d.router)

	// NOTE that the h2listener have not been reassigned to the exact tlsListener
	return srv.ServeTLS(listener, certFile, keyFile)
	// return
}

func (d *gorillaImpl) walkRoutes() {
	r := d.Handler().(*mux.Router)
	err := r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			fmt.Println("ROUTE:", pathTemplate)
		}
		pathRegexp, err := route.GetPathRegexp()
		if err == nil {
			fmt.Println("Path regexp:", pathRegexp)
		}
		queriesTemplates, err := route.GetQueriesTemplates()
		if err == nil {
			fmt.Println("Queries templates:", strings.Join(queriesTemplates, ","))
		}
		queriesRegexps, err := route.GetQueriesRegexp()
		if err == nil {
			fmt.Println("Queries regexps:", strings.Join(queriesRegexps, ","))
		}
		methods, err := route.GetMethods()
		if err == nil {
			fmt.Println("Methods:", strings.Join(methods, ","))
		}
		fmt.Println()
		return nil
	})

	if err != nil {
		fmt.Println(err)
	}
}

func (d *gorillaImpl) BuildRoutes() {
	r := d.Handler().(*mux.Router)

	dir := "./public"
	// This will serve files under http://localhost:8000/static/<filename>
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(dir))))

	r.HandleFunc("/", muxHomeHandler)
	r.HandleFunc("/products", muxProductsHandler)
	r.HandleFunc("/articles", muxArticlesHandler)
	r.HandleFunc("/products/{key}", muxProductHandler)
	r.HandleFunc("/articles/{category}/", muxArticlesCategoryHandler)
	r.HandleFunc("/articles/{category}/{id:[0-9]+}", muxArticleHandler)

	r.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		// an example API handler
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})
}

func muxHomeHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("Gorilla!\n"))
}

func muxProductsHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("Gorilla!\n"))
}

func muxArticlesHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("Gorilla!\n"))
}

func muxProductHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("Gorilla!\n"))
}

func muxArticlesCategoryHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "Category: %v\n", vars["category"])
}

func muxArticleHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("Gorilla!\n"))
}
