package utils

import (
	"github.com/prometheus/client_golang/prometheus"
)

type InitConfig struct {
	InitContainerImagePCIeBW string
	InitContainerImageMem    string
	InitContainerImageNet    string
	BWThreshold              string
}

var UserConfig InitConfig

var (
	Requests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "healthcheck_req_total",
			Help: "Number of invocations to Autopilot",
		},
	)

	Hchecks = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "health_report_total",
			Help: "Summary of the health checks measurements on compute nodes.",
		},
		[]string{"health", "node", "deviceid"},
	)
)

func Initmetrics(reg prometheus.Registerer) {
	// Register custom metrics with the global prometheus registry
	reg.MustRegister(Requests, Hchecks)
}
