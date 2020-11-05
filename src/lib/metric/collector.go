package metric

import (
	"os"

	"github.com/prometheus/client_golang/prometheus"
)

// RegisterCollectors register all the common static collector
func RegisterCollectors() {
	prometheus.MustRegister([]prometheus.Collector{
		TotalInFlightGauge,
		TotalReqCnt,
		TotalReqDurSummary,
	}...)
}

var (
	// TotalInFlightGauge used to collect total in flight number
	TotalInFlightGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: os.Getenv(NamespaceEnvKey),
			Subsystem: os.Getenv(SubsystemEnvKey),
			Name:      "http_request_inflight",
			Help:      "The total number of requests",
		},
		[]string{"url"},
	)

	// TotalReqCnt used to collect total request counter
	TotalReqCnt = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: os.Getenv(NamespaceEnvKey),
			Subsystem: os.Getenv(SubsystemEnvKey),
			Name:      "http_request",
			Help:      "The total number of requests",
		},
		[]string{"method", "code", "url"},
	)

	// TotalReqDurSummary used to collect total request duration summaries
	TotalReqDurSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  os.Getenv(NamespaceEnvKey),
			Subsystem:  os.Getenv(SubsystemEnvKey),
			Name:       "http_request_duration_seconds",
			Help:       "The time duration of the requests",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"method", "url"})
)
