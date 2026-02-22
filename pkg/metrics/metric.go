package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var MetricRegistry *prometheus.Registry = prometheus.NewRegistry()

func init() {
	MetricRegistry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
}

func MetricHandler(w http.ResponseWriter, r *http.Request) {
	promhttp.HandlerFor(MetricRegistry, promhttp.HandlerOpts{}).
		ServeHTTP(w, r)
}
