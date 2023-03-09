/*
Copyright The Guard Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"time"

	"go.kubeguard.dev/guard/auth/providers/token"
	"go.kubeguard.dev/guard/authz/providers/azure"
	"go.kubeguard.dev/guard/authz/providers/azure/data"
	azureutils "go.kubeguard.dev/guard/util/azure"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"gomodules.xyz/signals"
	"gomodules.xyz/x/ntp"
	v "gomodules.xyz/x/version"
	"k8s.io/klog/v2"
	"kmodules.xyz/client-go/meta"
	"kmodules.xyz/client-go/tools/fsnotify"
)

type Server struct {
	AuthRecommendedOptions  *AuthRecommendedOptions
	AuthzRecommendedOptions *AuthzRecommendedOptions
	TokenAuthenticator      *token.Authenticator
	WriteTimeout            time.Duration
	ReadTimeout             time.Duration
}

func (s *Server) AddFlags(fs *pflag.FlagSet) {
	fs.DurationVar(&s.WriteTimeout, "server-write-timeout", 10*time.Second, "Guard http server write timeout. Default is 10 seconds.")
	fs.DurationVar(&s.ReadTimeout, "server-read-timeout", 5*time.Second, "Guard http server read timeout. Default is 5 seconds.")
	s.AuthRecommendedOptions.AddFlags(fs)
	s.AuthzRecommendedOptions.AddFlags(fs)
}

func (s Server) ListenAndServe() {
	if errs := s.AuthRecommendedOptions.Validate(); errs != nil {
		klog.Fatal(errs)
	}

	if errs := s.AuthzRecommendedOptions.Validate(s.AuthRecommendedOptions); errs != nil {
		klog.Fatal(errs)
	}

	if s.AuthRecommendedOptions.NTP.Enabled() {
		ticker := time.NewTicker(s.AuthRecommendedOptions.NTP.Interval)
		go func() {
			for range ticker.C {
				if err := ntp.CheckSkewFromServer(s.AuthRecommendedOptions.NTP.NTPServer, s.AuthRecommendedOptions.NTP.MaxClodkSkew); err != nil {
					klog.Fatal(err)
				}
			}
		}()
	}

	if s.AuthRecommendedOptions.Token.AuthFile != "" {
		s.TokenAuthenticator = token.New(s.AuthRecommendedOptions.Token)

		err := s.TokenAuthenticator.Configure()
		if err != nil {
			klog.Fatalln(err)
		}
		if meta.PossiblyInCluster() {
			w := fsnotify.Watcher{
				WatchDir: filepath.Dir(s.AuthRecommendedOptions.Token.AuthFile),
				Reload: func() error {
					return s.TokenAuthenticator.Configure()
				},
			}
			stopCh := signals.SetupSignalHandler()
			err = w.Run(stopCh)
			if err != nil {
				klog.Fatal(err)
			}
		}
	}

	// loading file read related data
	if err := s.AuthRecommendedOptions.LDAP.Configure(); err != nil {
		klog.Fatal(err)
	}
	if err := s.AuthRecommendedOptions.Google.Configure(); err != nil {
		klog.Fatal(err)
	}

	/*
		Ref:
		 - http://www.levigross.com/2015/11/21/mutual-tls-authentication-in-go/
		 - https://blog.cloudflare.com/exposing-go-on-the-internet/
		 - http://www.bite-code.com/2015/06/25/tls-mutual-auth-in-golang/
		 - http://www.hydrogen18.com/blog/your-own-pki-tls-golang.html
	*/
	caCert, err := os.ReadFile(s.AuthRecommendedOptions.SecureServing.CACertFile)
	if err != nil {
		klog.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(caCert)
	if !ok {
		klog.Fatal("Failed to add CA cert in CertPool for guard server")
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

	m := chi.NewRouter()
	m.Use(middleware.RealIP)
	m.Use(middleware.Logger)
	m.Use(middleware.Recoverer)

	// Instrument the handlers with all the metrics, injecting the "handler" label by currying.
	// ref:
	// - https://godoc.org/github.com/prometheus/client_golang/prometheus/promhttp#example-InstrumentHandlerDuration
	// - https://github.com/brancz/prometheus-example-app/blob/master/main.go#L44:28
	handler := promhttp.InstrumentHandlerInFlight(inFlightGauge,
		promhttp.InstrumentHandlerDuration(duration.MustCurryWith(prometheus.Labels{"handler": "tokenreviews"}),
			promhttp.InstrumentHandlerCounter(counter,
				promhttp.InstrumentHandlerResponseSize(responseSize.MustCurryWith(prometheus.Labels{"handler": "tokenreviews"}), &s),
			),
		),
	)

	m.Post("/tokenreviews", handler.ServeHTTP)
	m.Get("/metrics", promhttp.Handler().ServeHTTP)
	m.Get("/healthz", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("x-content-type-options", "nosniff")
		err := json.NewEncoder(w).Encode(v.Version)
		if err != nil {
			klog.Fatal(err)
		}
	}))

	authzhandler := Authzhandler{
		AuthRecommendedOptions:  s.AuthRecommendedOptions,
		AuthzRecommendedOptions: s.AuthzRecommendedOptions,
	}

	m.Get("/readyz", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if len(s.AuthzRecommendedOptions.AuthzProvider.Providers) > 0 && s.AuthzRecommendedOptions.AuthzProvider.Has(azure.OrgType) && s.AuthzRecommendedOptions.Azure.DiscoverResources {
			if authzhandler.operationsMap != nil && len(authzhandler.operationsMap) > 0 {
				w.WriteHeader(200)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			w.WriteHeader(200)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("x-content-type-options", "nosniff")
		err := json.NewEncoder(w).Encode(v.Version)
		if err != nil {
			klog.Fatal(err)
		}
	}))

	klog.Infoln("setting up authz providers")
	if len(s.AuthzRecommendedOptions.AuthzProvider.Providers) > 0 {
		authzPromHandler := promhttp.InstrumentHandlerInFlight(inFlightGaugeAuthz,
			promhttp.InstrumentHandlerDuration(duration.MustCurryWith(prometheus.Labels{"handler": "subjectaccessreviews"}),
				promhttp.InstrumentHandlerCounter(counterAuthz,
					promhttp.InstrumentHandlerResponseSize(responseSize.MustCurryWith(prometheus.Labels{"handler": "subjectaccessreview"}), &authzhandler),
				),
			),
		)

		m.Post("/subjectaccessreviews", authzPromHandler.ServeHTTP)

		if s.AuthzRecommendedOptions.AuthzProvider.Has(azure.OrgType) {
			options := data.DefaultOptions
			authzhandler.Store, err = data.NewDataStore(options)
			if authzhandler.Store == nil || err != nil {
				klog.Fatalf("Error in initalizing cache. Error:%s", err.Error())
			}

			if s.AuthzRecommendedOptions.Azure.DiscoverResources {
				clusterType := ""

				switch s.AuthzRecommendedOptions.Azure.AuthzMode {
				case "arc":
					clusterType = azureutils.ConnectedClusters
				case "aks":
					clusterType = azureutils.ManagedClusters
				case "fleet":
					clusterType = azureutils.Fleets
				default:
					klog.Fatalf("Authzmode %s is not supported for fetching list of resources", s.AuthzRecommendedOptions.Azure.AuthzMode)
				}

				settings, err := azureutils.NewDiscoverResourcesSettings(clusterType, s.AuthRecommendedOptions.Azure.Environment, s.AuthzRecommendedOptions.Azure.AKSAuthzTokenURL, s.AuthzRecommendedOptions.Azure.KubeConfigFile, s.AuthRecommendedOptions.Azure.TenantID, s.AuthRecommendedOptions.Azure.ClientID, s.AuthRecommendedOptions.Azure.ClientSecret)
				if err != nil {
					klog.Fatalf("Failed to create settings for discovering resources. Error:%s", err)
				}

				discoverResourcesListStart := time.Now()
				operationsMap, err := azureutils.DiscoverResources(context.Background(), settings)
				discoverResourcesDuration := time.Since(discoverResourcesListStart).Seconds()
				if err != nil {
					azureutils.DiscoverResourcesTotalDuration.Observe(discoverResourcesDuration)
					klog.Fatalf("Failed to create map of data actions. Error:%s", err)
				}

				azureutils.DiscoverResourcesTotalDuration.Observe(discoverResourcesDuration)
				authzhandler.operationsMap = operationsMap
			}
		}
	}

	srv := &http.Server{
		Addr:         s.AuthRecommendedOptions.SecureServing.SecureAddr,
		ReadTimeout:  s.ReadTimeout,
		WriteTimeout: s.WriteTimeout,
		Handler:      m,
		TLSConfig:    tlsConfig,
	}
	klog.Fatalln(srv.ListenAndServeTLS(s.AuthRecommendedOptions.SecureServing.CertFile, s.AuthRecommendedOptions.SecureServing.KeyFile))
}
