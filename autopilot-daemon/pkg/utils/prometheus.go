package utils

import (
    "os/exec"
    "strings"

    "k8s.io/klog/v2"
    "github.com/prometheus/client_golang/prometheus"
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
		[]string{"health", "node", "cpu", "gpu", "deviceid"},
	)
)

func InitMetrics(reg prometheus.Registerer) {
	// Register custom metrics with the global prometheus registry
	reg.MustRegister(HchecksGauge)
}

func InitHardwareMetrics() {
    // Define CPUModel global variable
    cpu := "N/A"

    cmd1 := "cat /proc/cpuinfo | egrep '^model name' | uniq | awk '{print substr($0, index($0,$4))}'"
    out1, err1 := exec.Command("bash","-c",cmd1).Output()
    if err1 != nil {
        klog.Info("Error retrieving cpu model info", err1.Error())
    } else {
        cpu = string(out1)
    }
    klog.Info("CPU_MODEL: ", cpu)
    CPUModel = cpu

    // Define GPUModel global variable
    gpu := "N/A"

    cmd2 := exec.Command("nvidia-smi", "--query-gpu=gpu_name", "--format=csv,noheader")
    out2, err2 := cmd2.CombinedOutput()
    if err2 != nil {
        klog.Info("Error retrieving gpu model info", err2.Error())
    } else {
        tmp := strings.TrimSpace(string(out2[:]))
        gpu = strings.Split(tmp, "\n")[0]
    }
    klog.Info("GPU_MODEL: ", gpu)
    GPUModel = gpu
}
