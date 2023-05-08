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
	Mutations = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "healthcheck_req_total",
			Help: "Number of invocations to Autopilot",
		},
	)
)

func Initmetrics(reg prometheus.Registerer) {
	// Register custom metrics with the global prometheus registry
	reg.MustRegister(Mutations)
}
