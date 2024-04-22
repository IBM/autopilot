package handlers

import (
	"net/http"
	"os"
	"strings"

	"github.com/IBM/autopilot/pkg/utils"
	"k8s.io/klog/v2"
)

func SystemStatusHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		nodelabel := r.URL.Query().Get("nodelabel")
		if nodelabel == "" {
			nodelabel = "None"
		}
		hosts := r.URL.Query().Get("host")
		if hosts == "" {
			hosts = "all"
		}
		checks := r.URL.Query().Get("check")
		if checks == "" {
			checks = "all"
		}
		batch := r.URL.Query().Get("batch")
		if batch == "" {
			batch = "0"
		}
		jobName := r.URL.Query().Get("job")
		if jobName == "" {
			jobName = "None"
		}
		iperfclients := r.URL.Query().Get("clientsperiface")
		if iperfclients == "" {
			iperfclients = "1"
		}
		iperfservers := r.URL.Query().Get("serverspernode")
		if iperfservers == "" {
			iperfservers = "1"
		}
		dcgmR := r.URL.Query().Get("r")
		if dcgmR == "" {
			dcgmR = "1"
		}
		if strings.Contains(checks, "iperf") {
			klog.Info("Running iperf3 on hosts ", hosts, " or job ", jobName)
			w.Write([]byte("Running iperf3 on hosts " + hosts + " or job " + jobName + "\n\n"))
			checks = strings.Trim(checks, "iperf")
			plane := r.URL.Query().Get("plane")
			if plane == "" {
				plane = "data"
			}
			sourceNode := r.URL.Query().Get("source")
			if sourceNode == "" {
				sourceNode = "None"
			}
			out, err := runIperf(hosts, jobName, plane, iperfclients, iperfservers, sourceNode, nodelabel)
			if err != nil {
				klog.Error(err.Error())
			}
			w.Write(*out)
		}
		if checks != "" {
			if hosts == os.Getenv("NODE_NAME") {
				klog.Info("Checking system status of host " + hosts + " (localhost)")
				w.Write([]byte("Checking system status of host " + hosts + " (localhost) \n\n"))
				utils.HealthcheckLock.Lock()
				defer utils.HealthcheckLock.Unlock()
				out, err := runAllTestsLocal(hosts, checks, dcgmR, jobName, nodelabel, r)
				if err != nil {
					klog.Error(err.Error())
				}
				w.Write(*out)
			} else {
				klog.Info("Asking to run on remote node(s) ", hosts, " or with node label ", nodelabel)
				w.Write([]byte("Asking to run on remote node(s) " + hosts + " or with node label " + nodelabel + "\n\n"))
				out, err := runAllTestsRemote(hosts, checks, batch, jobName, dcgmR, nodelabel, "status")
				if err != nil {
					klog.Error(err.Error())
				}
				w.Write(*out)
			}
		}

	}
	return http.HandlerFunc(fn)
}

func PCIeBWHandler(pciebw string) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Requesting pcie test with bw: " + pciebw + "\n"))
		out, err := runPCIeBw()
		if err != nil {
			klog.Error(err.Error())
		}
		if out != nil {
			w.Write(*out)
		}

	}
	return http.HandlerFunc(fn)
}

func RemappedRowsHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Requesting Remapped Rows check on all GPUs\n"))
		out, err := runRemappedRows()
		if err != nil {
			klog.Error(err.Error())
		}
		if out != nil {
			w.Write(*out)
		}

	}
	return http.HandlerFunc(fn)
}

func PingHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Ping test"))
		hosts := r.URL.Query().Get("host")
		if hosts == "" {
			hosts = "all"
		}
		jobName := r.URL.Query().Get("job")
		if jobName == "" {
			jobName = "None"
		}
		nodelabel := r.URL.Query().Get("nodelabel")
		if nodelabel == "" {
			nodelabel = "None"
		}
		out, err := runPing(hosts, jobName, nodelabel)
		if err != nil {
			klog.Error(err.Error())
		}
		if out != nil {
			w.Write(*out)
		}
	}
	return http.HandlerFunc(fn)
}

func IperfHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Iperf3 test"))
		hosts := r.URL.Query().Get("host")
		if hosts == "" {
			hosts = "all"
		}
		jobName := r.URL.Query().Get("job")
		if jobName == "" {
			jobName = "None"
		}
		sourceNode := r.URL.Query().Get("source")
		if sourceNode == "" {
			sourceNode = "None"
		}
		iperfclients := r.URL.Query().Get("clientsperiface")
		if iperfclients == "" {
			iperfclients = "1"
		}
		iperfservers := r.URL.Query().Get("serverspernode")
		if iperfservers == "" {
			iperfservers = "1"
		}
		plane := r.URL.Query().Get("plane")
		if plane == "" {
			plane = "data"
		}
		nodelabel := r.URL.Query().Get("nodelabel")
		if nodelabel == "" {
			nodelabel = "None"
		}
		out, err := runIperf(hosts, jobName, plane, iperfclients, iperfservers, sourceNode, nodelabel)
		if err != nil {
			klog.Error(err.Error())
		}
		if out != nil {
			w.Write(*out)
		}
	}
	return http.HandlerFunc(fn)
}

func StartIperfServersHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		replicas := r.URL.Query().Get("replicas")
		if replicas == "" {
			replicas = "1"
		}
		out, err := startIperfServers(replicas)

		if err != nil {
			klog.Error(err.Error())
		}
		if out != nil {
			w.Write(*out)
		}
	}
	return http.HandlerFunc(fn)
}

func DCGMHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("DCGM test"))
		dcgmR := r.URL.Query().Get("r")
		if dcgmR == "" {
			dcgmR = "1"
		}
		out, err := runDCGM(dcgmR)
		if err != nil {
			klog.Error(err.Error())
		}
		if out != nil {
			w.Write(*out)
		}
	}
	return http.HandlerFunc(fn)
}

func CoordinationHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Start Coordination Handler\n"))
		nodelabel := r.URL.Query().Get("nodelabel")
		if nodelabel == "" {
			nodelabel = "None"
		}
		hosts := r.URL.Query().Get("host")
		if hosts == "" {
			hosts = "all"
		}
		checks := r.URL.Query().Get("check")
		if checks == "" {
			checks = "all"
		}
		batch := r.URL.Query().Get("batch")
		if batch == "" {
			batch = "0"
		}
		jobName := r.URL.Query().Get("job")
		if jobName == "" {
			jobName = "None"
		}
		if checks == "resources" {
			if hosts == os.Getenv("NODE_NAME") {
				klog.Info("Checking resources of node " + hosts + " (localhost)")
				w.Write([]byte("Checking resources of node " + hosts + " (localhost) \n\n"))
				utils.HealthcheckLock.Lock()
				defer utils.HealthcheckLock.Unlock()

				output := "Allocated resources:\n"
				output += utils.PrintResourceUsageHeader()
				output += utils.PrintResourceUsage()

				output += "\nNCCL TEST: "
				if utils.ConfirmNCCLSupport() {
					output += "AVAILABLE \n"
				} else {
					output += "UNAVAILABLE \n"
				}

				return_str := []byte(output)
				w.Write(return_str)

			} else {
				klog.Info("Asking to run on remote node(s) ", hosts, " or with node label ", nodelabel)
				w.Write([]byte("Asking to run on remote node(s) " + hosts + " or with node label " + nodelabel + "\n\n"))
				out, err := runAllTestsRemote(hosts, checks, batch, jobName, "", nodelabel, "coordinate")
				if err != nil {
					klog.Error(err.Error())
				}
				w.Write(*out)
			}
		}
	}
	return http.HandlerFunc(fn)
}

func GpuPowerHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("GPU Power Measurement test"))
		out, err := runGPUPower()
		if err != nil {
			klog.Error(err.Error())
		}
		if out != nil {
			w.Write(*out)
		}
	}
	return http.HandlerFunc(fn)
}

func GpuMemHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("GPU Memory DGEMM+DAXPY test"))
		out, err := runGPUPower()
		if err != nil {
			klog.Error(err.Error())
		}
		if out != nil {
			w.Write(*out)
		}
	}
	return http.HandlerFunc(fn)
}
