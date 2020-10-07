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
	"github.com/hedzr/cmdr/conf"
	"github.com/kataras/iris/v12/core/host"
	"golang.org/x/crypto/acme/autocert"
	"gopkg.in/hedzr/errors.v2"
	"net"
	"net/http"
	"net/url"
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
		cmdr.Logger.Debugf("...autocert enabled")
		d.certManager = &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(d.domains()...), // 测试时使用的域名：example.com
			Cache:      autocert.DirCache(cmdr.GetStringR("server.autocert.dir-cache", "ci/certs")),
		}
		go func() {
			if err := http.ListenAndServe(":80", d.certManager.HTTPHandler(nil)); err != nil {
				cmdr.Logger.Fatalf("autocert tool listening on :80 failed: %v", err)
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
				cmdr.Logger.Debugf("...shutdown going on.")
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
	if sd, ok := d.routerImpl.(GracefulShutdown); ok {
		if err := sd.Shutdown(ctx); err != nil {
			cmdr.Logger.Errorf("   mux Shutdown failed: %v", err)
		} else {
			cmdr.Logger.Debugf("   mux Shutdown ok.")
		}
	} else if err := srv.Shutdown(ctx); err != nil {
		cmdr.Logger.Errorf("   srv Shutdown failed: %v", err)
	} else {
		cmdr.Logger.Debugf("   srv Shutdown ok.")
	}
}

func (d *daemonImpl) enterLoop(prg *dex.Program, stopCh, doneCh chan struct{}, listener net.Listener) (err error) {
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
		cmdr.Logger.Printf("daemon.dex ServeSignals, pid = %v", os.Getpid())
		if err = sig.ServeSignals(); err != nil {
			cmdr.Logger.Errorf("daemon.dex Error: %v", err)
		}
	}

	// if daemonImpl != nil {
	// 	err = daemonImpl.OnStop(cmd, args)
	// }

	if err != nil {
		cmdr.Logger.Fatalf("daemon.dex terminated: %v", err)
	}
	cmdr.Logger.Printf("daemon.dex terminated.")

	return
}

// onRunHttp2Server NOTE
// listener: a copy from parent linux process, just for live reload.
func (d *daemonImpl) onRunHttp2Server(prg *dex.Program, stopCh, doneCh chan struct{}, hotReloadListener net.Listener) (err error) {
	d.appTag = prg.Command.GetRoot().AppName
	if conf.AppName != d.appTag && d.appTag != "" {
		conf.AppName = d.appTag
		conf.Version = prg.Command.GetRoot().Version
		conf.ServerTag = d.appTag
		conf.ServerID = d.appTag
	}

	cmdr.Logger.Debugf("%q daemon OnRun, pid = %v, ppid = %v", d.appTag, os.Getpid(), os.Getppid())

	// Tweak configuration values here.
	var (
		port      = cmdr.GetIntRP(d.appTag, "server.port")
		portNoTls = 0
		config    = tls2.NewCmdrTLSConfig(d.appTag, "server.tls", "server.start")
		tlsConfig = d.checkAndEnableAutoCert(config)
	)

	cmdr.Logger.Tracef("used config file: %v", cmdr.GetUsedConfigFile())
	cmdr.Logger.Tracef("logger level: %v", cmdr.GetLoggerLevel())

	if config.IsServerCertValid() || tlsConfig.GetCertificate == nil {
		port = cmdr.GetIntRP(d.appTag, "server.ports.tls")
	}
	portNoTls = cmdr.GetIntRP(d.appTag, "server.ports.default", 0)

	if port == 0 {
		cmdr.Logger.Fatalf("port not defined.")
	}
	addr := fmt.Sprintf(":%d", port) // ":3300"

	serverType := cmdr.GetStringRP(d.appTag, "server.type", "")
	serverType = cmdr.GetStringR("server.Mux", serverType)
	switch serverType {
	case "iris":
		d.Type = typeIris
	case "echo":
		d.Type = typeEcho
	case "gin":
		d.Type = typeGin
	case "gorilla":
		d.Type = typeGorilla
	case "default":
		d.Type = typeDefault
	default:
		d.Type = typeGin
	}

	if d.routerImpl == nil {
		switch d.Type {
		case typeIris:
			d.routerImpl = newIris()
		case typeEcho:
			d.routerImpl = newEcho()
		case typeGin:
			d.routerImpl = newGin()
		case typeGorilla:
			d.routerImpl = newGorilla()
		case typeDefault:
			d.routerImpl = newStdMux()
		default:
			d.routerImpl = newGin()
		}
		cmdr.Logger.Printf("serverType got: %v, %v", serverType, d.Type)
	} else {
		cmdr.Logger.Printf("serverType is: %v, %v | routerImpl was preset.", serverType, d.Type)
	}

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

		if pps, ok := d.routerImpl.(ForLoggerInitializing); ok {
			pps.PrePreServe()
		}
		if err = d.routerImpl.PreServe(); err != nil {
			cmdr.Logger.Fatalf("%+v", err)
		}
		defer func() {
			if err = d.routerImpl.PostServe(); err != nil {
				cmdr.Logger.Fatalf("%+v", err)
			}
		}()

		// Start the server with TLS, since we are running HTTP/2 it must be
		// run with TLS.
		// Exactly how you would run an HTTP/1.1 server with TLS connection.
		if config.IsServerCertValid() || srv.TLSConfig.GetCertificate == nil {
			cmdr.Logger.Printf("> Serving on %v with HTTPS...", addr)
			if d.Type == typeIris && portNoTls > 0 {
				// 转发 80 到 https
				target, _ := url.Parse("https://" + addr)
				source := fmt.Sprintf("%s:%v", target.Hostname(), portNoTls)
				go func() {
					cmdr.Logger.Printf("  > Proxy from %v to HTTPS...", source)
					err := host.NewProxy(source, target).ListenAndServe()
					if err != nil {
						cmdr.Logger.Fatalf("proxy at %v failed: %v", source, err)
					}
				}()
			}
			// if cmdr.FileExists("ci/certs/server.cert") && cmdr.FileExists("ci/certs/server.key") {
			if err = d.serve(prg, srv, hotReloadListener, config.Cert, config.Key); err != http.ErrServerClosed && err != nil {
				if dex.IsErrorAddressAlreadyInUse(err) {
					if present, process := dex.FindDaemonProcess(); present {
						cmdr.Logger.Fatalf("cannot serve, last pid=%v, error is: %+v", process.Pid, err)
					}
				}
				cmdr.Logger.Fatalf("listen at port %v failed: %v", port, err)
			}
			// if err = d.serve(srv, hotReloadListener, "ci/certs/server.cert", "ci/certs/server.key"); err != http.ErrServerClosed {
			// 	cmdr.Logger.Fatal(err)
			// }
			cmdr.Logger.Printf("   end")
			// 		} else {
			// 			cmdr.Logger.Fatalf(`ci/certs/server.{cert,key} NOT FOUND under '%s'. You might generate its at command line:
			//
			// [ -d ci/certs ] || mkdir -p ci/certs
			// openssl genrsa -out ci/certs/server.key 2048
			// openssl req -new -x509 -key ci/certs/server.key -out ci/certs/server.cert -days 3650 -subj /CN=localhost
			//
			// 			`, cmdr.GetCurrentDir())
			// 		}
		} else {
			cmdr.Logger.Printf("Serving on %v with HTTP...", addr)
			if err = d.serve(prg, srv, hotReloadListener, "", ""); err != http.ErrServerClosed && err != nil {
				cmdr.Logger.Fatalf("%+v", err)
			}
			cmdr.Logger.Printf("   end")
		}
	}()

	// go worker(stopCh, doneCh)
	return
}

func (d *daemonImpl) serve(prg *dex.Program, srv *http.Server, listener net.Listener, certFile, keyFile string) (err error) {
	// if srv.shuttingDown() {
	// 	return http.ErrServerClosed
	// }

	addr := srv.Addr
	if addr == "" {
		addr = ":https"
	}

	if listener == nil {
		if cmdr.GetBoolR("server.start.socket") {
			sf := prg.SocketFileName()
			if cmdr.GetBoolR("server.start.reset-socket-file") && cmdr.FileExists(sf) {
				err = os.Remove(sf)
				if err != nil {
					return err
				}
			}
			cmdr.Logger.Infof("listening on unix sock file: %v", sf)
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
		cmdr.Logger.Printf("   h2listener closed, pid=%v", os.Getpid())
	}()

	h2listener = listener
	return d.routerImpl.Serve(srv, h2listener, certFile, keyFile)
}

func (d *daemonImpl) handle(w http.ResponseWriter, r *http.Request) {
	// Log the request protocol
	cmdr.Logger.Printf("Got connection: %s", r.Proto)
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
