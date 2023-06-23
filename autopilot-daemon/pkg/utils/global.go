package utils

import (
	"github.com/prometheus/client_golang/prometheus"
)

type InitConfig struct {
	BWThreshold string
}

var UserConfig InitConfig

var (
	Requests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "autopilot",
			Name:      "health_checks_req_total",
			Help:      "Number of invocations to Autopilot",
		},
	)

	Hchecks = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  "autopilot",
			Name:       "health_report_total",
			Help:       "Summary of the health checks measurements on compute nodes.",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"health", "node", "deviceid"},
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
	reg.MustRegister(Requests, Hchecks, HchecksGauge)
}
