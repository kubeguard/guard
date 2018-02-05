package lib

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/appscode/go/log"
	v "github.com/appscode/go/version"
	"github.com/appscode/pat"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	WebAddress    string
	CACertFile    string
	CertFile      string
	KeyFile       string
	OpsAddress    string
	TokenAuthFile string
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

	/*
		Ref:
		 - http://www.levigross.com/2015/11/21/mutual-tls-authentication-in-go/
		 - https://blog.cloudflare.com/exposing-go-on-the-internet/
		 - http://www.bite-code.com/2015/06/25/tls-mutual-auth-in-golang/
		 - http://www.hydrogen18.com/blog/your-own-pki-tls-golang.html
	*/
	tokenAuthCsvFile = s.TokenAuthFile

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
	m.Post("/apis/authentication.k8s.io/v1beta1/tokenreviews", http.HandlerFunc(Authenticate))
	srv := &http.Server{
		Addr:         s.WebAddress,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      m,
		TLSConfig:    tlsConfig,
	}
	log.Fatalln(srv.ListenAndServeTLS(s.CertFile, s.KeyFile))
}
