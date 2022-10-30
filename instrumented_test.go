package httpclient_test

import (
	"encoding/json"
	"fmt"
	"github.com/clambin/httpclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Do(t *testing.T) {
	metrics := httpclient.NewMetrics("foo", "bar")
	s := httptest.NewServer(http.HandlerFunc(handler))
	c := &httpclient.InstrumentedClient{
		Options:     httpclient.Options{PrometheusMetrics: metrics},
		Application: "foo",
	}

	response, err := doCall(c, s.URL+"/foo")
	require.NoError(t, err)
	assert.Equal(t, "bar", response.Name)
	assert.Equal(t, 42, response.Age)

	response, err = doCall(c, s.URL+"/bar")
	require.Error(t, err)

	s.Close()
	response, err = doCall(c, s.URL+"/foo")
	require.Error(t, err)

	assert.Equal(t, map[string]uint64{
		"/foo": 2,
		"/bar": 1,
	}, getLatencyCounters(t, prometheus.DefaultGatherer, "foo_bar_"))

	assert.Equal(t, map[string]float64{
		"/foo": 1,
		"/bar": 0,
	}, getErrorMetrics(t, prometheus.DefaultGatherer, "foo_bar_"))
}

type testStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func doCall(c httpclient.Caller, url string) (response testStruct, err error) {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	var resp *http.Response
	if resp, err = c.Do(req); err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		return response, fmt.Errorf("call failed: %s", resp.Status)
	}
	defer func() { _ = resp.Body.Close() }()

	err = json.NewDecoder(resp.Body).Decode(&response)
	return
}

func handler(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/foo" {
		http.Error(w, "invalid endpoint", http.StatusNotFound)
		return
	}

	_ = json.NewEncoder(w).Encode(testStruct{
		Name: "bar",
		Age:  42,
	})
}
