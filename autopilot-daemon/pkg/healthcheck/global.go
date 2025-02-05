package healthcheck

import (
	"os"
	"strings"

	"k8s.io/klog/v2"
)

type HealthCheck string

// Holding each test current status to facilitate node labeling
var HealthCheckStatus map[HealthCheck]bool
var defaultPeriodicChecks string = "pciebw,remapped,dcgm,ping,gpupower"

const (
	Undefined HealthCheck = ""
	DCGM      HealthCheck = "dcgm"
	GPUMem    HealthCheck = "gpumem"
	GPUPower  HealthCheck = "gpupower"
	Iperf     HealthCheck = "iperf"
	PCIeBW    HealthCheck = "pciebw"
	Ping      HealthCheck = "ping"
	PVC       HealthCheck = "pvc"
	RowRemap  HealthCheck = "remapped"
)

func GetPeriodicChecks() string {
	checks, exists := os.LookupEnv("PERIODIC_CHECKS")
	if !exists {
		klog.Info("Run all periodic health checks\n")
		return defaultPeriodicChecks
	}
	return checks
}

func InitNodeStatusMap() {
	HealthCheckStatus = make(map[HealthCheck]bool)
	checklist := GetPeriodicChecks()
	for _, v := range strings.Split(checklist, ",") {
		klog.Info("Init entry map ", v)
		HealthCheckStatus[HealthCheck(v)] = false
	}
}

func GetNodeStatus() bool {
	hasFailures := false
	for v := range HealthCheckStatus {
		hasFailures = hasFailures || HealthCheckStatus[v]
	}
	return hasFailures
}
