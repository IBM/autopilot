package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/IBM/autopilot/pkg/healthcheck"
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
		dcgmR := r.URL.Query().Get("r")
		if dcgmR == "" {
			dcgmR = "1"
		}
		if strings.Contains(checks, string(healthcheck.Iperf)) {
			klog.Info("Running iperf3 on hosts ", hosts, " or job ", jobName)
			w.Write([]byte("Running iperf3 on hosts " + hosts + " or job " + jobName + "\n\n"))
			checks = strings.Trim(checks, "iperf")
			workload := r.URL.Query().Get("workload")
			if workload == "" {
				workload = "ring"
			}
			pclients := r.URL.Query().Get("pclients")
			if pclients == "" {
				pclients = "8"
			}
			startport := r.URL.Query().Get("startport")
			if startport == "" {
				startport = "5200"
			}
			cleanup := ""
			if r.URL.Query().Has("cleanup") {
				cleanup = "--cleanup"
			}
			out, err := healthcheck.RunIperf(workload, pclients, startport, cleanup)
			if err != nil {
				klog.Error(err.Error())
			}
			if out != nil {
				w.Write(*out)
			}
		}
		if checks != "" {
			if hosts == utils.NodeName {
				utils.HealthcheckLock.Lock()
				defer utils.HealthcheckLock.Unlock()
				out, err := healthcheck.RunHealthLocalNode(checks, dcgmR, jobName, nodelabel, r)
				if err != nil {
					klog.Error(err.Error())
				}
				w.Write(*out)
				hasFailures := healthcheck.GetNodeStatus()
				klog.Info("Errors after running local, on demand health checks: ", hasFailures)
				if hasFailures {
					utils.PatchNode(utils.GPUHealthWarnLabel, utils.NodeName, false)
				} else {
					utils.PatchNode(utils.GPUHealthPassLabel, utils.NodeName, false)
				}

			} else {
				klog.Info("Asking to run on remote node(s) ", hosts, " or with node label ", nodelabel)
				w.Write([]byte("Asking to run on remote node(s) " + hosts + " or with node label " + nodelabel + "\n\n"))
				out, err := healthcheck.RunHealthRemoteNodes(hosts, checks, batch, jobName, dcgmR, nodelabel)
				if err != nil {
					klog.Error(err.Error())
				}
				w.Write(*out)
			}
		}

	}
	return http.HandlerFunc(fn)
}

func PCIeBWHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Requesting pcie test with bw: " + strconv.Itoa(utils.UserConfig.BWThreshold) + "\n"))
		out, err := healthcheck.RunPCIeBW()
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
		out, err := healthcheck.RunRemappedRows()
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
		out, err := healthcheck.RunPing(hosts, jobName, nodelabel)
		if err != nil {
			klog.Error(err.Error())
		}
		if out != nil {
			w.Write(*out)
		}
	}
	return http.HandlerFunc(fn)
}

func InvasiveCheckHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Launching invasive health checks. Results will be added to 'autopilot.ibm.com/gpuhealth' and 'autopilot.ibm.com/dcgm.level.3' node labels\n"))
		healthcheck.InvasiveCheck()
	}
	return http.HandlerFunc(fn)
}

func IperfHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		workload := r.URL.Query().Get("workload")
		if workload == "" {
			workload = "ring"
		}
		pclients := r.URL.Query().Get("pclients")
		if pclients == "" {
			pclients = "8"
		}
		startport := r.URL.Query().Get("startport")
		if startport == "" {
			startport = "5200"
		}
		cleanup := ""
		if r.URL.Query().Has("cleanup") {
			cleanup = "--cleanup"
		}
		out, err := healthcheck.RunIperf(workload, pclients, startport, cleanup)
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
		numservers := r.URL.Query().Get("numservers")
		if numservers == "" {
			numservers = "8"
		}
		startport := r.URL.Query().Get("startport")
		if startport == "" {
			startport = "5200"
		}
		out, err := healthcheck.StartIperfServers(numservers, startport)

		if err != nil {
			klog.Error(err.Error())
		}
		if out != nil {
			w.Write(*out)
		}
	}
	return http.HandlerFunc(fn)
}

func StopAllIperfServersHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		out, err := healthcheck.StopAllIperfServers()
		if err != nil {
			klog.Error(err.Error())
		}
		if out != nil {
			w.Write(*out)
		}
	}
	return http.HandlerFunc(fn)
}

func StartIperfClientsHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		dstip := r.URL.Query().Get("dstip")
		dstport := r.URL.Query().Get("dstport")
		numclients := r.URL.Query().Get("numclients")
		out, err := healthcheck.StartIperfClients(dstip, dstport, numclients)
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
		out, err := healthcheck.RunDCGM(dcgmR)
		if err != nil {
			klog.Error(err.Error())
		}
		if out != nil {
			w.Write(*out)
		}
	}
	return http.HandlerFunc(fn)
}

func GpuPowerHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("GPU Power Measurement test"))
		out, err := healthcheck.RunGPUPower()
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
		out, err := healthcheck.RunGPUPower()
		if err != nil {
			klog.Error(err.Error())
		}
		if out != nil {
			w.Write(*out)
		}
	}
	return http.HandlerFunc(fn)
}

func PVCHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("PVC create-delete test\n"))
		out, err := healthcheck.RunCreateDeletePVC()
		if err != nil {
			klog.Error(err.Error())
		}
		if out != nil {
			w.Write(*out)
		}
	}
	return http.HandlerFunc(fn)
}

func ReadinessProbeHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		data := HealthResult{"readinessProbe", "ready"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(data)
	}
	return http.HandlerFunc(fn)
}
