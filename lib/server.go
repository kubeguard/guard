package lib

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"path/filepath"
	"time"

	"github.com/appscode/go/log"
	"github.com/appscode/go/signals"
	v "github.com/appscode/go/version"
	"github.com/appscode/kutil/meta"
	"github.com/appscode/kutil/tools/fsnotify"
	"github.com/appscode/pat"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
)

type Server struct {
	WebAddress    string
	CACertFile    string
	CertFile      string
	KeyFile       string
	OpsAddress    string
	TokenAuthFile string
	Google        GoogleOptions
	Azure         AzureOptions
	LDAP          LDAPOptions
}

func (s *Server) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.WebAddress, "web-address", s.WebAddress, "Http server address")
	fs.StringVar(&s.CACertFile, "ca-cert-file", s.CACertFile, "File containing CA certificate")
	fs.StringVar(&s.CertFile, "cert-file", s.CertFile, "File container server TLS certificate")
	fs.StringVar(&s.KeyFile, "key-file", s.KeyFile, "File containing server TLS private key")
	fs.StringVar(&s.OpsAddress, "ops-addr", s.OpsAddress, "Address to listen on for web interface and telemetry.")

	fs.StringVar(&s.TokenAuthFile, "token-auth-file", "", "To enable static token authentication")
	s.Google.AddFlags(fs)
	s.Azure.AddFlags(fs)
	s.LDAP.AddFlags(fs)
}

func (s Server) UseTLS() bool {
	return s.CACertFile != "" && s.CertFile != "" && s.KeyFile != ""
}

func (s Server) ListenAndServe() {
	// Run Monitoring Server with both /metric and /debug
	go func() {
		if s.OpsAddress != "" {
			http.Handle("/metrics", promhttp.Handler())
			http.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(200)
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("x-content-type-options", "nosniff")
				json.NewEncoder(w).Encode(v.Version)
			})
			log.Errorln("Failed to start monitoring server, cause", http.ListenAndServe(s.OpsAddress, nil))
		}
	}()

	if s.TokenAuthFile != "" {
		var err error
		tokenMap, err = LoadTokenFile(s.TokenAuthFile)
		if err != nil {
			log.Fatalln(err)
		}
		if meta.PossiblyInCluster() {
			w := fsnotify.Watcher{
				WatchDir: filepath.Dir(s.TokenAuthFile),
				Reload: func() error {
					lock.Lock()
					defer lock.Unlock()

					data, err := LoadTokenFile(s.TokenAuthFile)
					if err != nil {
						return err
					}
					tokenMap = data
					return nil
				},
			}
			stopCh := signals.SetupSignalHandler()
			w.Run(stopCh)
		}
	}

	/*
		Ref:
		 - http://www.levigross.com/2015/11/21/mutual-tls-authentication-in-go/
		 - https://blog.cloudflare.com/exposing-go-on-the-internet/
		 - http://www.bite-code.com/2015/06/25/tls-mutual-auth-in-golang/
		 - http://www.hydrogen18.com/blog/your-own-pki-tls-golang.html
	*/
	caCert, err := ioutil.ReadFile(s.CACertFile)
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		MinVersion:               tls.VersionTLS12,
		SessionTicketsDisabled:   true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305, // Go 1.8 only
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,   // Go 1.8 only
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:  caCertPool,
		NextProtos: []string{"h2", "http/1.1"},
	}
	tlsConfig.BuildNameToCertificate()

	m := pat.New()
	m.Post("/apis/authentication.k8s.io/v1beta1/tokenreviews", s)
	srv := &http.Server{
		Addr:         s.WebAddress,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      m,
		TLSConfig:    tlsConfig,
	}
	log.Fatalln(srv.ListenAndServeTLS(s.CertFile, s.KeyFile))
}
