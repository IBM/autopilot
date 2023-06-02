package handlers

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.ibm.com/hybrid-cloud-infrastructure-research/autopilot-daemon/pkg/utils"
	"k8s.io/klog/v2"
)

func PCIeBWHandler(pciebw string) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Requesting pcie test with bw: " + pciebw + "\n"))
		out, err := exec.Command("python3", "./gpubw/entrypoint.py").Output()
		if err != nil {
			klog.Error(err.Error())
		} else {
			klog.Info("GPU PCIe BW test completed:")
			w.Write(out)
			output := strings.TrimSuffix(string(out[:]), "\n")

			split := strings.Split(output, "\n")
			bws := split[len(split)-1]
			final := strings.Split(bws, " ")

			for gpuid, v := range final {
				bw, err := strconv.ParseFloat(v, 64)
				if err != nil {
					klog.Error(err.Error())
				} else {
					klog.Info("Observation: ", os.Getenv("NODE_NAME"), " ", strconv.Itoa(gpuid), " ", bw)
					utils.Hchecks.WithLabelValues("pciebw", os.Getenv("NODE_NAME"), strconv.Itoa(gpuid)).Observe(bw)
					utils.HchecksGauge.WithLabelValues("pciebw", os.Getenv("NODE_NAME"), strconv.Itoa(gpuid)).Set(bw)
				}
			}
		}

	}
	return http.HandlerFunc(fn)
}

func RemappedRowsHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Requesting Remapped Rows check on all GPUs\n"))
		out, err := exec.Command("python3", "./gpu-remapped/entrypoint.py").Output()
		if err != nil {
			klog.Error(err.Error())
		} else {
			klog.Info("Remapped Rows check test completed:")
			// output := string(out[:])
			// fmt.Println(output)
			output := strings.TrimSuffix(string(out[:]), "\n")

			split := strings.Split(output, "\n")
			rmr := split[len(split)-1]
			final := strings.Split(rmr, " ")

			for gpuid, v := range final {
				rm, err := strconv.ParseFloat(v, 64)
				if err != nil {
					klog.Error(err.Error())
				} else {
					klog.Info("Observation: ", os.Getenv("NODE_NAME"), " ", strconv.Itoa(gpuid), " ", rm)
					utils.Hchecks.WithLabelValues("remapped", os.Getenv("NODE_NAME"), strconv.Itoa(gpuid)).Observe(rm)
					utils.HchecksGauge.WithLabelValues("remapped", os.Getenv("NODE_NAME"), strconv.Itoa(gpuid)).Set(rm)
				}
			}
		}
		w.Write(out)

	}
	return http.HandlerFunc(fn)
}

func GPUMemHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Requesting HBM test "))
	}
	return http.HandlerFunc(fn)
}

func NetReachHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Requesting secondary nics reachability test\n"))
		// utils.Hchecks.WithLabelValues("netdevice", "worker-2", "1").Observe(1.0)
		// utils.Hchecks.WithLabelValues("netdevice", "worker-2", "0").Observe(1.0)
		out, err := exec.Command("python3", "./network/metrics-entrypoint.py").Output()
		if err != nil {
			klog.Error(err.Error())
		} else {
			klog.Info("Secondary NIC health check test completed:")
			output := string(out[:])
			fmt.Println(output)
		}
		w.Write(out)
	}
	return http.HandlerFunc(fn)
}
