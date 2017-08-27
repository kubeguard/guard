package hostfacts

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/appscode/log"
	"github.com/appscode/pat"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	WebAddress string
	CACertFile string
	CertFile   string
	KeyFile    string

	OpsAddress      string
	EnableAnalytics bool
}

func (s Server) ListenAndServe() {
	// Run Monitoring Server with both /metric and /debug
	go func() {
		if s.OpsAddress != "" {
			http.Handle("/metrics", promhttp.Handler())
			log.Errorln("Failed to start monitoring server, cause", http.ListenAndServe(s.OpsAddress, nil))
		}
	}()

	m := pat.New()
	srv := &http.Server{
		Addr:         s.WebAddress,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      m,
	}
	if s.CACertFile == "" && s.CertFile == "" && s.KeyFile == "" {
		log.Fatalln(srv.ListenAndServe())
	} else {
		/*
			Ref:
			 - https://blog.cloudflare.com/exposing-go-on-the-internet/
			 - http://www.bite-code.com/2015/06/25/tls-mutual-auth-in-golang/
			 - http://www.hydrogen18.com/blog/your-own-pki-tls-golang.html
		*/
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
			ClientAuth: tls.VerifyClientCertIfGiven,
			NextProtos: []string{"h2", "http/1.1"},
		}
		if s.CACertFile != "" {
			caCert, err := ioutil.ReadFile(s.CACertFile)
			if err != nil {
				log.Fatal(err)
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig.ClientCAs = caCertPool
		}
		tlsConfig.BuildNameToCertificate()

		srv.TLSConfig = tlsConfig
		log.Fatalln(srv.ListenAndServeTLS(s.CertFile, s.KeyFile))
	}
}
