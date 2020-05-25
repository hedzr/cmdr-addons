// Copyright © 2020 Hedzr Yeh.

package svr

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/hedzr/cmdr"
	"github.com/hedzr/cmdr-addons/pkg/plugins/dex"
	"github.com/hedzr/cmdr-addons/pkg/plugins/dex/sig"
	tls2 "github.com/hedzr/cmdr-addons/pkg/svr/tls"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/acme/autocert"
	"gopkg.in/hedzr/errors.v2"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"
)

func (d *daemonImpl) domains() (domainList []string) {
	for _, top := range cmdr.GetStringSliceR("server.autocert.domains", "example.com") {
		domainList = append(domainList, top)
		for _, s := range cmdr.GetStringSliceR("server.autocert.second-level-domains", "aurora", "api", "home", "res") {
			domainList = append(domainList, fmt.Sprintf("%s.%s", s, top))
		}
	}
	return
}

func (d *daemonImpl) checkAndEnableAutoCert(config *tls2.CmdrTLSConfig) (tlsConfig *tls.Config) {
	tlsConfig = &tls.Config{}

	if config.IsServerCertValid() {
		tlsConfig = config.ToServerTLSConfig()
	}

	if cmdr.GetBoolR("server.autocert.enabled") {
		logrus.Debugf("...autocert enabled")
		d.certManager = &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(d.domains()...), // 测试时使用的域名：example.com
			Cache:      autocert.DirCache(cmdr.GetStringR("server.autocert.dir-cache", "ci/certs")),
		}
		go func() {
			if err := http.ListenAndServe(":80", d.certManager.HTTPHandler(nil)); err != nil {
				logrus.Fatal("autocert tool listening on :80 failed.", err)
			}
		}()
		tlsConfig.GetCertificate = d.certManager.GetCertificate
	}

	return
}

func (d *daemonImpl) enableGracefulShutdown(srv *http.Server, stopCh, doneCh chan struct{}) {

	go func() {
		for {
			select {
			case <-stopCh:
				logrus.Debugf("...shutdown going on.")
				d.shutdown(srv)
				<-doneCh
				return
			}
		}
	}()

}

func (d *daemonImpl) shutdown(srv *http.Server) {
	ctx, cancelFunc := context.WithTimeout(context.TODO(), 8*time.Second)
	defer cancelFunc()
	if err := srv.Shutdown(ctx); err != nil {
		logrus.Error("Shutdown failed: ", err)
	} else {
		logrus.Debugf("Shutdown ok.")
	}
}

func (d *daemonImpl) enterLoop(prog *dex.Program, stopCh, doneCh chan struct{}, listener net.Listener) (err error) {
	switch runtime.GOOS {
	case "windows":
	LOOP:
		// Developer raises stopCh signal to make the daemon shutdown itself gracefully.
		// The doneCh will be triggered once shutdown completely.
		for {
			select {
			case <-doneCh:
				break LOOP
			}
		}
		break

	default:
		log.Printf("daemon.dex ServeSignals, pid = %v", os.Getpid())
		if err = sig.ServeSignals(); err != nil {
			log.Println("daemon.dex Error:", err)
		}
	}

	// if daemonImpl != nil {
	// 	err = daemonImpl.OnStop(cmd, args)
	// }

	if err != nil {
		log.Fatal("daemon.dex terminated.", err)
	}
	log.Println("daemon.dex terminated.")

	return
}

// onRunHttp2Server NOTE
// listener: a copy from parent linux process, just for live reload.
func (d *daemonImpl) onRunHttp2Server(prog *dex.Program, stopCh, doneCh chan struct{}, hotReloadListener net.Listener) (err error) {
	d.appTag = prog.Command.GetRoot().AppName
	logrus.Debugf("%q daemon OnRun, pid = %v, ppid = %v", d.appTag, os.Getpid(), os.Getppid())

	// Tweak configuration values here.
	var (
		port      = cmdr.GetIntRP(d.appTag, "server.port")
		config    = tls2.NewCmdrTLSConfig(d.appTag, "server.tls", "server.start")
		tlsConfig = d.checkAndEnableAutoCert(config)
	)

	logrus.Tracef("used config file: %v", cmdr.GetUsedConfigFile())
	logrus.Tracef("logger level: %v / %v", logrus.GetLevel(), cmdr.GetLoggerLevel())

	if config.IsServerCertValid() || tlsConfig.GetCertificate == nil {
		port = cmdr.GetIntRP(d.appTag, "server.ports.tls")
	}

	if port == 0 {
		logrus.Fatal("port not defined")
	}
	addr := fmt.Sprintf(":%d", port) // ":3300"

	serverType := cmdr.GetStringR("server.Mux")
	switch serverType {
	case "iris":
		d.Type = typeIris
	case "echo":
		d.Type = typeEcho
	case "gin":
		d.Type = typeGin
	case "gorilla":
		d.Type = typeGorilla
	default:
		d.Type = typeDefault
	}

	switch d.Type {
	case typeIris:
		d.routerImpl = newIris()
	case typeEcho:
		d.routerImpl = newEcho()
	case typeGin:
		d.routerImpl = newGin()
	case typeGorilla:
		d.routerImpl = newGorilla()
	default:
		d.routerImpl = newStdMux()
	}
	logrus.Printf("serverType: %v, %v", serverType, d.Type)

	d.routerImpl.BuildRoutes()

	// Create a server on port 8000
	// Exactly how you would run an HTTP/1.1 server
	srv := &http.Server{
		Addr:              addr,
		Handler:           d.routerImpl.Handler(), // d.mux, // http.HandlerFunc(d.handle),
		TLSConfig:         tlsConfig,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
	}

	d.enableGracefulShutdown(srv, stopCh, doneCh)

	// TODO server push, ...
	// https://posener.github.io/http2/

	go func() {

		// this routine will be terminated safely via golang http shutdown gracefully.

		if err = d.routerImpl.PreServe(); err != nil {
			logrus.Fatalf("%+v", err)
		}
		defer func() {
			if err = d.routerImpl.PostServe(); err != nil {
				logrus.Fatalf("%+v", err)
			}
		}()

		// Start the server with TLS, since we are running HTTP/2 it must be
		// run with TLS.
		// Exactly how you would run an HTTP/1.1 server with TLS connection.
		if config.IsServerCertValid() || srv.TLSConfig.GetCertificate == nil {
			logrus.Printf("Serving on %v with HTTPS...", addr)
			// if cmdr.FileExists("ci/certs/server.cert") && cmdr.FileExists("ci/certs/server.key") {
			if err = d.serve(prog, srv, hotReloadListener, config.Cert, config.Key); err != http.ErrServerClosed && err != nil {
				if dex.IsErrorAddressAlreadyInUse(err) {
					if present, process := dex.FindDaemonProcess(); present {
						logrus.Fatalf("cannot serve, last pid=%v, error is: %+v", process.Pid, err)
					}
				}
				logrus.Fatalf("listen at port %v failed: %v", port, err)
			}
			// if err = d.serve(srv, hotReloadListener, "ci/certs/server.cert", "ci/certs/server.key"); err != http.ErrServerClosed {
			// 	logrus.Fatal(err)
			// }
			logrus.Println("end")
			// 		} else {
			// 			logrus.Fatalf(`ci/certs/server.{cert,key} NOT FOUND under '%s'. You might generate its at command line:
			//
			// [ -d ci/certs ] || mkdir -p ci/certs
			// openssl genrsa -out ci/certs/server.key 2048
			// openssl req -new -x509 -key ci/certs/server.key -out ci/certs/server.cert -days 3650 -subj /CN=localhost
			//
			// 			`, cmdr.GetCurrentDir())
			// 		}
		} else {
			logrus.Printf("Serving on %v with HTTP...", addr)
			if err = d.serve(prog, srv, hotReloadListener, "", ""); err != http.ErrServerClosed && err != nil {
				logrus.Fatalf("%+v", err)
			}
			logrus.Println("end")
		}
	}()

	// go worker(stopCh, doneCh)
	return
}

func (d *daemonImpl) serve(prog *dex.Program, srv *http.Server, listener net.Listener, certFile, keyFile string) (err error) {
	// if srv.shuttingDown() {
	// 	return http.ErrServerClosed
	// }

	addr := srv.Addr
	if addr == "" {
		addr = ":https"
	}

	if listener == nil {
		if cmdr.GetBoolR("server.start.socket") {
			sf := prog.SocketFileName()
			if cmdr.GetBoolR("server.start.reset-socket-file") && cmdr.FileExists(sf) {
				err = os.Remove(sf)
				if err != nil {
					return err
				}
			}
			logrus.Infof("listening on unix sock file: %v", sf)
			listener, err = net.Listen("unix", sf)
			if err != nil {
				err = errors.New("Cannot bind to unix sock %q", sf).Attach(err)
				return err
			}
		} else {
			listener, err = net.Listen("tcp", addr)
			if err != nil {
				err = errors.New("Cannot bind to address %v", addr).Attach(err)
				return err
			}
		}
	}

	defer func() {
		if h2listener != nil {
			h2listener.Close()
		}
		logrus.Printf("h2listener closed, pid=%v", os.Getpid())
	}()

	h2listener = listener
	return d.routerImpl.Serve(srv, h2listener, certFile, keyFile)
}

func (d *daemonImpl) handle(w http.ResponseWriter, r *http.Request) {
	// Log the request protocol
	log.Printf("Got connection: %s", r.Proto)
	// Send a message back to the client
	_, _ = w.Write([]byte("Hello"))
}

const (
	// for http client
	activeTimeout       = 10 * time.Minute
	maxIdleConns        = 1000
	maxIdleConnsPerHost = 100

	// for http server
	idleTimeout       = 5 * time.Minute
	readHeaderTimeout = 1 * time.Second
	writeTimeout      = 10 * time.Second
	maxHeaderBytes    = http.DefaultMaxHeaderBytes
	shutdownTimeout   = 30 * time.Second
)
