package client

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	namespace = "devnull"
	subsystem = "http_client"
)

// NewInstrumented returns a new instrumented HTTP client.
func NewInstrumented(r prometheus.Registerer, timeout time.Duration) *http.Client {
	inFlightRequests := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "in_flight_requests",
		Help:      "A gauge of in-flight requests for the HTTP client.",
		Namespace: namespace,
		Subsystem: subsystem,
	})
	totalRequests := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:      "requests_total",
			Help:      "A counter for requests from the HTTP client.",
			Namespace: namespace,
			Subsystem: subsystem,
		},
		[]string{"code", "method"},
	)
	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:      "request_duration_seconds",
			Help:      "A histogram of request latencies.",
			Buckets:   prometheus.DefBuckets,
			Namespace: namespace,
			Subsystem: subsystem,
		},
		[]string{},
	)

	r.MustRegister(totalRequests, requestDuration, inFlightRequests)

	c := &http.Client{}
	if timeout > 0 {
		c.Timeout = timeout
	}

	// Wrap the default RoundTripper with middleware.
	c.Transport = promhttp.InstrumentRoundTripperInFlight(inFlightRequests,
		promhttp.InstrumentRoundTripperCounter(totalRequests,
			promhttp.InstrumentRoundTripperDuration(requestDuration, http.DefaultTransport),
		),
	)
	return c
}
