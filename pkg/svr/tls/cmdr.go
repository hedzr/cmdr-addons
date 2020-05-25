// Copyright © 2020 Hedzr Yeh.

package tls

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/hedzr/cmdr"
	"github.com/sirupsen/logrus"
	"gopkg.in/hedzr/errors.v2"
	"io/ioutil"
	"net"
	"os"
	"path"
)

// NewCmdrTLSConfig builds the *CmdrTLSConfig object from cmdr config file and cmdr command-line arguments
func NewCmdrTLSConfig(appTag, prefixInConfigFile, prefixInCommandline string) *CmdrTLSConfig {
	s := &CmdrTLSConfig{}
	if len(prefixInConfigFile) > 0 {
		s.InitTLSConfigFromConfigFile(appTag, prefixInConfigFile)
	}
	if len(prefixInCommandline) > 0 {
		s.InitTLSConfigFromCommandline(prefixInCommandline)
	}
	return s
}

// CmdrTLSConfig wraps the certificates.
// For server-side, the `Cert` field must be a bundle of server certificates with all root CAs chain.
// For server-side, the `Cacert` is optional for extra client CA's.
type CmdrTLSConfig struct {
	Enabled       bool
	Cacert        string // server-side: optional server's CA;   client-side: client's CA
	ServerCert    string //                                      client-side: the server's cert
	Cert          string // server-side: server's cert bundle;   client-side: client's cert
	Key           string // server-side: server's key;           client-side: client's key
	ClientAuth    bool
	MinTLSVersion uint16
}

// IsServerCertValid checks the server or CA cert are present.
func (s *CmdrTLSConfig) IsServerCertValid() bool {
	return s.ServerCert != "" || s.Cacert != ""
}

// IsCertValid checks the cert and privateKey are present
func (s *CmdrTLSConfig) IsCertValid() bool {
	return s.Cert != "" && s.Key != ""
}

// IsClientAuthEnabled checks if the client-side authentication is enabled
func (s *CmdrTLSConfig) IsClientAuthEnabled() bool {
	return s.ClientAuth && s.Cert != "" && s.Key != ""
}

// InitTLSConfigFromCommandline loads the parsed command-line arguments to *CmdrTLSConfig
func (s *CmdrTLSConfig) InitTLSConfigFromCommandline(prefix string) {
	var b bool
	var sz string
	b = cmdr.GetBoolRP(prefix, "client-auth")
	if b {
		s.ClientAuth = b
	}
	sz = cmdr.GetStringRP(prefix, "cacert")
	if sz != "" {
		s.Cacert = sz
	}
	sz = cmdr.GetStringRP(prefix, "cert")
	if sz != "" {
		s.Cert = sz
	}
	sz = cmdr.GetStringRP(prefix, "key")
	if sz != "" {
		s.Key = sz
	}

	for _, loc := range cmdr.GetStringSliceRP(prefix, "locations") {
		if s.Cacert != "" && cmdr.FileExists(path.Join(loc, s.Cacert)) {
			s.Cacert = path.Join(loc, s.Cacert)
		} else if s.Cacert != "" {
			continue
		}
		if s.Cert != "" && cmdr.FileExists(path.Join(loc, s.Cert)) {
			s.Cert = path.Join(loc, s.Cert)
		} else if s.Cert != "" {
			continue
		}
		if s.Key != "" && cmdr.FileExists(path.Join(loc, s.Key)) {
			s.Key = path.Join(loc, s.Key)
		} else if s.Key != "" {
			continue
		}
	}

	switch cmdr.GetIntRP(prefix, "tls-version", 2) {
	case 0:
		s.MinTLSVersion = tls.VersionTLS10
	case 1:
		s.MinTLSVersion = tls.VersionTLS11
	case 3:
		s.MinTLSVersion = tls.VersionTLS13
	default:
		s.MinTLSVersion = tls.VersionTLS12
	}
}

// InitTLSConfigFromConfigFile loads CmdrTLSConfig members from cmdr config file.
//
// The entries in config file looks like:
//
//     prefix := "my-service.server.tls"
//     tls:
//       enabled: true
//       cacert: root.pem
//       cert: cert.pem
//       key: cert.key
//       locations:
//     	   - ./ci/certs
//     	   - $CFG_DIR/certs
func (s *CmdrTLSConfig) InitTLSConfigFromConfigFile(appTag, prefix string) {
	if len(appTag) > 0 {
		prefix = appTag + "." + prefix
	}
	enabled := cmdr.GetBoolRP(prefix, "enabled")
	if enabled {
		s.ClientAuth = cmdr.GetBoolRP(prefix, "client-auth")
		s.Cacert = cmdr.GetStringRP(prefix, "cacert")
		s.Cert = cmdr.GetStringRP(prefix, "cert")
		s.Key = cmdr.GetStringRP(prefix, "key")

		for _, loc := range cmdr.GetStringSliceRP(prefix, "locations") {
			logrus.Debugf("> tls - testing loc: %v", loc)
			if s.Cacert != "" && cmdr.FileExists(path.Join(loc, s.Cacert)) {
				s.Cacert = path.Join(loc, s.Cacert)
			} else if s.Cacert != "" {
				continue
			}
			if s.Cert != "" && cmdr.FileExists(path.Join(loc, s.Cert)) {
				s.Cert = path.Join(loc, s.Cert)
			} else if s.Cert != "" {
				continue
			}
			if s.Key != "" && cmdr.FileExists(path.Join(loc, s.Key)) {
				s.Key = path.Join(loc, s.Key)
			} else if s.Key != "" {
				continue
			}
		}

		switch cmdr.GetIntRP(prefix, "tls-version", int(s.MinTLSVersion-tls.VersionTLS10)) {
		case 0:
			s.MinTLSVersion = tls.VersionTLS10
		case 1:
			s.MinTLSVersion = tls.VersionTLS11
		case 3:
			s.MinTLSVersion = tls.VersionTLS13
		default:
			s.MinTLSVersion = tls.VersionTLS12
		}

		logrus.Debugf("> cfg_dir: %v / %v", path.Dir(cmdr.GetUsedConfigFile()), os.Getenv("CFG_DIR"))
		logrus.Debugf("> tls matched: %+v", s)
	}
}

// ToServerTLSConfig builds an tls.Config object for server.Serve
func (s *CmdrTLSConfig) ToServerTLSConfig() (config *tls.Config) {
	var err error
	config, err = s.newTLSConfig()
	if err == nil {
		if s.Cacert != "" {
			var rootPEM []byte
			rootPEM, err = ioutil.ReadFile(s.Cacert)
			if err != nil || rootPEM == nil {
				return
			}
			pool := x509.NewCertPool()
			ok := pool.AppendCertsFromPEM([]byte(rootPEM))
			if ok {
				config.ClientCAs = pool
			}
		}
	} else {
		logrus.Errorf("%+v", err)
	}
	return config
}

// ToTLSConfig converts to *tls.Config
func (s *CmdrTLSConfig) ToTLSConfig() (config *tls.Config) {
	config, _ = s.newTLSConfig()
	return config
}

func (s *CmdrTLSConfig) newTLSConfig() (config *tls.Config, err error) {
	var cert tls.Certificate
	cert, err = tls.LoadX509KeyPair(s.Cert, s.Key)
	if err != nil {
		err = errors.New("error parsing X509 certificate/key pair, pwd=%q, cert=%q", cmdr.GetCurrentDir(), s.Cert).Attach(err)
		return
	}
	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		err = errors.New("error parsing certificate").Attach(err)
		return
	}

	// Create TLSConfig
	// We will determine the cipher suites that we prefer.
	config = &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   s.MinTLSVersion,
	}

	// Require client certificates as needed
	if s.IsClientAuthEnabled() {
		config.ClientAuth = tls.RequireAndVerifyClientCert
	}

	// Add in CAs if applicable.
	if s.ClientAuth {
		if s.Cacert != "" {
			var rootPEM []byte
			rootPEM, err = ioutil.ReadFile(s.Cacert)
			if err != nil || rootPEM == nil {
				return nil, err
			}
			pool := x509.NewCertPool()
			ok := pool.AppendCertsFromPEM([]byte(rootPEM))
			if !ok {
				err = errors.New("failed to parse root ca certificate")
			}
			config.ClientCAs = pool
		}

		config.ClientAuth = tls.RequireAndVerifyClientCert
	}

	if err != nil {
		config = nil
	}
	return
}

// NewTLSListener builds net.Listener for tls mode or not
func (s *CmdrTLSConfig) NewTLSListener(l net.Listener) (listener net.Listener, err error) {
	if s != nil && s.IsCertValid() {
		var config *tls.Config
		config, err = s.newTLSConfig()
		if err != nil {
			logrus.Fatal(err)
		}
		listener = tls.NewListener(l, config)
	}
	return
}

// Dial connects to the given network address using net.Dial
// and then initiates a TLS handshake, returning the resulting
// TLS connection.
//
// Dial interprets a nil configuration as equivalent to
// the zero configuration; see the documentation of Config
// for the defaults.
func (s *CmdrTLSConfig) Dial(network, addr string) (conn net.Conn, err error) {
	if s != nil && s.IsServerCertValid() {
		roots := x509.NewCertPool()

		err = s.addCert(roots, s.ServerCert)
		if err != nil {
			return
		}
		err = s.addCert(roots, s.Cacert)
		if err != nil {
			return
		}

		cfg := &tls.Config{
			RootCAs: roots,
		}

		if s.IsClientAuthEnabled() {
			var cert tls.Certificate
			cert, err = tls.LoadX509KeyPair(s.Cert, s.Key)
			if err != nil {
				return
			}
			cfg.Certificates = []tls.Certificate{cert}
			cfg.InsecureSkipVerify = true
		}

		logrus.Printf("Connecting to %s over TLS...\n", addr)
		// Use the tls.Config here in http.Transport.TLSClientConfig
		conn, err = tls.Dial(network, addr, cfg)
	} else {
		logrus.Printf("Connecting to %s...\n", addr)
		conn, err = net.Dial(network, addr)
	}
	return
}

func (s *CmdrTLSConfig) addCert(roots *x509.CertPool, certPath string) (err error) {
	if certPath != "" {
		var rootPEM []byte
		rootPEM, err = ioutil.ReadFile(certPath)
		if err != nil {
			return
		}

		ok := roots.AppendCertsFromPEM(rootPEM)
		if !ok {
			// panic("failed to parse root certificate")
			err = errors.New("failed to parse root certificate")
			return
		}
	}
	return
}
