package server

import (
	v "github.com/appscode/go/version"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	version = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "version",
		Help: "Version information about this binary",
		ConstLabels: map[string]string{
			"version": v.Version.Version,
		},
	})

	inFlightGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "tokenreviews_handler_requests_in_flight",
		Help: "A gauge of requests currently being served by the tokenreviews handler.",
	})

	counter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tokenreviews_hendler_requests_total",
			Help: "A counter for requests to the tokenreviews handler.",
		},
		[]string{"code", "method"},
	)

	// duration is partitioned by the HTTP method and handler. It uses custom
	// buckets based on the expected request duration.
	duration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_duration_seconds",
			Help:    "A histogram of latencies for requests.",
			Buckets: []float64{.25, .5, 1, 2.5, 5, 10},
		},
		[]string{"handler", "method"},
	)

	// responseSize has no labels, making it a zero-dimensional
	// ObserverVec.
	responseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "response_size_bytes",
			Help:    "A histogram of response sizes for requests.",
			Buckets: []float64{100, 200, 400, 1000},
		},
		[]string{"handler"},
	)
)

func init() {
	// Register all of the metrics in the standard registry.
	prometheus.MustRegister(inFlightGauge, counter, duration, responseSize)
}
