package httpclient_test

import (
	"errors"
	"github.com/clambin/httpclient"
	"github.com/prometheus/client_golang/prometheus"
	pcg "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

func TestClientMetrics_MakeLatencyTimer(t *testing.T) {
	cfg := httpclient.Metrics{}

	// MakeLatencyTimer returns nil if no Latency metric is set
	assert.Nil(t, cfg.MakeLatencyTimer())

	cfg = httpclient.NewMetrics("foo", "")

	// collect metrics
	timer := cfg.MakeLatencyTimer("foo", "/bar", http.MethodGet)
	require.NotNil(t, timer)
	time.Sleep(10 * time.Millisecond)
	timer.ObserveDuration()

	// one measurement should be collected
	m, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)
	var found bool
	for _, entry := range m {
		if *entry.Name == "foo_api_latency" {
			require.Equal(t, pcg.MetricType_SUMMARY, *entry.Type)
			require.NotZero(t, entry.Metric)
			assert.NotZero(t, entry.Metric[0].Summary.GetSampleCount())
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestClientMetrics_ReportErrors(t *testing.T) {
	cfg := httpclient.Metrics{}

	// ReportErrors doesn't crash when no Errors metric is set
	cfg.ReportErrors(nil)

	cfg = httpclient.NewMetrics("bar", "")

	// collect metrics
	cfg.ReportErrors(nil, "foo", "/bar", http.MethodGet)

	// do a measurement
	count := getErrorMetrics(t, prometheus.DefaultGatherer, "bar_")
	assert.Equal(t, map[string]float64{"/bar": 0}, count)

	// record an error
	cfg.ReportErrors(errors.New("some error"), "foo", "/bar", http.MethodGet)

	// counter should now be 1
	count = getErrorMetrics(t, prometheus.DefaultGatherer, "bar_")
	assert.Equal(t, map[string]float64{"/bar": 1}, count)
}

func TestClientMetrics_Nil(t *testing.T) {
	cfg := httpclient.Metrics{}

	timer := cfg.MakeLatencyTimer("snafu")
	assert.Nil(t, timer)
	cfg.ReportErrors(nil, "foo")
}

func getErrorMetrics(t *testing.T, g prometheus.Gatherer, prefix string) map[string]float64 {
	t.Helper()

	counters := make(map[string]float64)
	m, err := g.Gather()
	require.NoError(t, err)
	for _, entry := range m {
		if *entry.Name == prefix+"api_errors_total" {
			require.Equal(t, pcg.MetricType_COUNTER, *entry.Type)
			for _, metric := range entry.Metric {
				counters[*metric.Label[1].Value] = *metric.Counter.Value
			}
		}
	}
	return counters
}

func getLatencyCounters(t *testing.T, g prometheus.Gatherer, prefix string) map[string]uint64 {
	t.Helper()

	counters := make(map[string]uint64)
	m, err := g.Gather()
	require.NoError(t, err)
	for _, entry := range m {
		if *entry.Name == prefix+"api_latency" {
			require.Equal(t, pcg.MetricType_SUMMARY, *entry.Type)
			for _, metric := range entry.Metric {
				counters[*metric.Label[1].Value] = *metric.Summary.SampleCount
			}
		}
	}
	return counters
}
