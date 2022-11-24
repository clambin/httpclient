package httpclient

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics contains Prometheus metrics to capture during API calls. Each metric is expected to have two labels:
// the first will contain the application issuing the request. The second will contain the endpoint (i.e. Path) of the request.
type Metrics struct {
	Latency *prometheus.SummaryVec // measures latency of an API call
	Errors  *prometheus.CounterVec // measures any errors returned by an API call
}

// NewMetrics creates a standard set of Prometheus metrics to capture during API calls.
func NewMetrics(namespace, subsystem string, r prometheus.Registerer) Metrics {
	m := Metrics{
		Latency: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Name: prometheus.BuildFQName(namespace, subsystem, "api_latency"),
			Help: "Latency of Reporter API calls",
		}, []string{"application", "endpoint", "method"}),
		Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, subsystem, "api_errors_total"),
			Help: "Number of failed Reporter API calls",
		}, []string{"application", "endpoint", "method"}),
	}
	if r == nil {
		r = prometheus.DefaultRegisterer
	}
	r.MustRegister(m.Latency, m.Errors)
	return m
}

// ReportErrors measures any API client call failures:
//
//	err := callAPI(server, endpoint)
//	pm.ReportErrors(err)
func (pm *Metrics) ReportErrors(err error, labelValues ...string) {
	if pm == nil || pm.Errors == nil {
		return
	}

	var value float64
	if err != nil {
		value = 1.0
	}
	pm.Errors.WithLabelValues(labelValues...).Add(value)
}

// MakeLatencyTimer creates a prometheus.Timer to measure the duration (latency) of an API client call
// If no Latency metric was created, timer will be nil:
//
//	timer := pm.MakeLatencyTimer(server, endpoint)
//	callAPI(server, endpoint)
//	if timer != nil {
//		timer.ObserveDuration()
//	}
func (pm *Metrics) MakeLatencyTimer(labelValues ...string) (timer *prometheus.Timer) {
	if pm != nil && pm.Latency != nil {
		timer = prometheus.NewTimer(pm.Latency.WithLabelValues(labelValues...))
	}
	return
}
