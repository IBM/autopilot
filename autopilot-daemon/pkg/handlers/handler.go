package handlers

import (
	"net/http"
	"os"
	"strings"

	"k8s.io/klog/v2"
)

func SystemStatusHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
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
			batch = "1"
		}
		jobName := r.URL.Query().Get("job")
		if jobName == "" {
			jobName = "None"
		}
		replicas := r.URL.Query().Get("replicas")
		if replicas == "" {
			replicas = "1"
		}

		klog.Info("Batch size ", batch)
		// if hosts == "all" {
		// 	klog.Info("Checking status on all nodes")
		// 	w.Write([]byte("Checking status on all nodes\n\n"))
		// 	err, out := runAllTestsRemote("all", checks, batch, jobName)
		// 	if err != nil {
		// 		klog.Error(err.Error())
		// 	}
		// 	w.Write(*out)
		// } else {
		if strings.Contains(checks, "iperf") {
			klog.Info("Running iperf3 on hosts ", hosts, " or job ", jobName)
			w.Write([]byte("Running iperf3 on hosts " + hosts + " or job " + jobName + "\n\n"))
			checks = strings.Trim(checks, "iperf")
			plane := r.URL.Query().Get("plane")
			if plane == "" {
				plane = "data"
			}
			err, out := runIperf(hosts, jobName, plane, replicas)
			if err != nil {
				klog.Error(err.Error())
			}
			w.Write(*out)
		}
		if checks != "" {
			if hosts == os.Getenv("NODE_NAME") {
				klog.Info("Checking system status of host " + hosts + " (localhost)")
				w.Write([]byte("Checking system status of host " + hosts + " (localhost) \n\n"))
				err, out := runAllTestsLocal(checks)
				if err != nil {
					klog.Error(err.Error())
				}
				w.Write(*out)
			} else {
				klog.Info("Asking to run on remote node(s) ", hosts)
				w.Write([]byte("Asking to run on remote node(s) " + hosts + "\n\n"))
				err, out := runAllTestsRemote(hosts, checks, batch, jobName, replicas)
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
		err, out := runPCIeBw()
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
		err, out := runRemappedRows()
		if err != nil {
			klog.Error(err.Error())
		}
		if out != nil {
			w.Write(*out)
		}

	}
	return http.HandlerFunc(fn)
}

func NetReachHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Requesting secondary nics reachability test\n"))
		err, out := netReachability()
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
		iface := r.URL.Query().Get("iface")
		if iface == "" {
			iface = "eth0"
		}
		replicas := r.URL.Query().Get("replicas")
		if replicas == "" {
			replicas = "1"
		}
		plane := r.URL.Query().Get("plane")
		if plane == "" {
			plane = "data"
		}
		err, out := runIperf(hosts, jobName, plane, replicas)
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
		w.Write([]byte("Starting Iperf3 Servers"))
		replicas := r.URL.Query().Get("replicas")
		if replicas == "" {
			replicas = "1"
		}
		err, out := startIperfServers(replicas)
		if err != nil {
			klog.Error(err.Error())
		}
		if out != nil {
			w.Write(*out)
		}
	}
	return http.HandlerFunc(fn)
}
