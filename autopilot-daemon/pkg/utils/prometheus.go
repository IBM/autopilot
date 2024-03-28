package utils

import "github.com/prometheus/client_golang/prometheus"

var (
	Requests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "autopilot",
			Name:      "health_checks_req_total",
			Help:      "Number of invocations to Autopilot",
		},
	)

	HchecksGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "autopilot",
			Name:      "health_checks",
			Help:      "Summary of the health checks measurements on compute nodes. Gauge Vector version",
		},
		[]string{"health", "node", "deviceid"},
	)
)

func Initmetrics(reg prometheus.Registerer) {
	// Register custom metrics with the global prometheus registry
	reg.MustRegister(HchecksGauge)
}
