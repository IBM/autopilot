package utils

import (
	"os/exec"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog/v2"
)

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
		[]string{"health", "node", "cpumodel", "gpumodel", "deviceid"},
	)
)

func InitMetrics(reg prometheus.Registerer) {
	// Register custom metrics with the global prometheus registry
	reg.MustRegister(HchecksGauge)
}

func InitHardwareMetrics() {
	// Define CPUModel global variable
	cpu := "N/A"

	cmd := "cat /proc/cpuinfo | egrep '^model name' | uniq | awk '{print substr($0, index($0,$4))}'|  sed 's/(//; s/)//'"
	out, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		klog.Info("Error retrieving cpu model info", err.Error())
	} else {
		cpu = strings.TrimSpace(string(out[:]))
	}
	klog.Info("CPU_MODEL: ", cpu)
	CPUModel = cpu

	// Define GPUModel global variable
	gpu := "N/A"

	cmd2 := exec.Command("nvidia-smi", "--query-gpu=gpu_name", "--format=csv,noheader")
	out, err = cmd2.CombinedOutput()
	if err != nil {
		klog.Info("Error retrieving gpu model info", err.Error())
	} else {
		tmp := strings.TrimSpace(string(out[:]))
		gpu = strings.Split(tmp, "\n")[0]
	}
	klog.Info("GPU_MODEL: ", gpu)
	GPUModel = gpu
}
