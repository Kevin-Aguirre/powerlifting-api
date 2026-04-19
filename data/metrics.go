package data

import "github.com/prometheus/client_golang/prometheus"

var (
	DataLoadDurationSeconds = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "data_load_duration_seconds",
		Help: "Duration of the most recent full data load in seconds.",
	})
	DataLoadsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "data_loads_total",
		Help: "Total number of completed data loads.",
	})
)

func init() {
	prometheus.MustRegister(DataLoadDurationSeconds, DataLoadsTotal)
}
