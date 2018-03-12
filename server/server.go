package server

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"path/filepath"
	"time"

	"github.com/appscode/go/log"
	"github.com/appscode/go/ntp"
	"github.com/appscode/go/signals"
	v "github.com/appscode/go/version"
	"github.com/appscode/guard/token"
	"github.com/appscode/kutil/meta"
	"github.com/appscode/kutil/tools/fsnotify"
	"github.com/appscode/pat"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
)

type Server struct {
	RecommendedOptions *RecommendedOptions
	TokenAuthenticator *token.Authenticator
}

func (s *Server) AddFlags(fs *pflag.FlagSet) {
	s.RecommendedOptions.AddFlags(fs)
}

func (s Server) ListenAndServe() {
	if s.RecommendedOptions.NTP.Enabled() {
		ticker := time.NewTicker(s.RecommendedOptions.NTP.Interval)
		go func() {
			for range ticker.C {
				if err := ntp.CheckSkew(s.RecommendedOptions.NTP.MaxClodkSkew); err != nil {
					log.Fatal(err)
				}
			}
		}()
	}

	if s.RecommendedOptions.Token.AuthFile != "" {
		s.TokenAuthenticator = token.New(s.RecommendedOptions.Token)

		err := s.TokenAuthenticator.Configure()
		if err != nil {
			log.Fatalln(err)
		}
		if meta.PossiblyInCluster() {
			w := fsnotify.Watcher{
				WatchDir: filepath.Dir(s.RecommendedOptions.Token.AuthFile),
				Reload: func() error {
					return s.TokenAuthenticator.Configure()
				},
			}
			stopCh := signals.SetupSignalHandler()
			w.Run(stopCh)
		}
	}

	// caCertPool for self signed LDAP sever certificate
	if s.RecommendedOptions.LDAP.CaCertFile != "" {
		caCert, err := ioutil.ReadFile(s.RecommendedOptions.LDAP.CaCertFile)
		if err != nil {
			log.Fatal(err)
		}
		s.RecommendedOptions.LDAP.CaCertPool = x509.NewCertPool()
		s.RecommendedOptions.LDAP.CaCertPool.AppendCertsFromPEM(caCert)
		ok := s.RecommendedOptions.LDAP.CaCertPool.AppendCertsFromPEM(caCert)
		if !ok {
			log.Fatal("Failed to add CA cert in CertPool for LDAP")
		}
	}

	/*
		Ref:
		 - http://www.levigross.com/2015/11/21/mutual-tls-authentication-in-go/
		 - https://blog.cloudflare.com/exposing-go-on-the-internet/
		 - http://www.bite-code.com/2015/06/25/tls-mutual-auth-in-golang/
		 - http://www.hydrogen18.com/blog/your-own-pki-tls-golang.html
	*/
	caCert, err := ioutil.ReadFile(s.RecommendedOptions.SecureServing.CACertFile)
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(caCert)
	if !ok {
		log.Fatal("Failed to add CA cert in CertPool for guard server")
	}

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
		// ClientAuth: tls.VerifyClientCertIfGiven needed to pass healthz check
		ClientAuth: tls.VerifyClientCertIfGiven,
		ClientCAs:  caCertPool,
		NextProtos: []string{"h2", "http/1.1"},
	}
	tlsConfig.BuildNameToCertificate()

	m := pat.New()

	// Instrument the handlers with all the metrics, injecting the "handler"
	// label by currying.
	handler := promhttp.InstrumentHandlerInFlight(inFlightGauge,
		promhttp.InstrumentHandlerDuration(duration.MustCurryWith(prometheus.Labels{"handler": "tokenreviews"}),
			promhttp.InstrumentHandlerCounter(counter,
				promhttp.InstrumentHandlerResponseSize(responseSize.MustCurryWith(prometheus.Labels{"handler": "tokenreviews"}), s),
			),
		),
	)
	m.Post("/tokenreviews", handler)
	m.Get("/metrics", promhttp.Handler())
	m.Get("/healthz", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("x-content-type-options", "nosniff")
		json.NewEncoder(w).Encode(v.Version)
	}))
	srv := &http.Server{
		Addr:         s.RecommendedOptions.SecureServing.SecureAddr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      m,
		TLSConfig:    tlsConfig,
	}
	log.Fatalln(srv.ListenAndServeTLS(s.RecommendedOptions.SecureServing.CertFile, s.RecommendedOptions.SecureServing.KeyFile))
}
