package handlers

import (
	"net/http"
	"os"

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

		klog.Info("Batch size ", batch)
		if hosts == "all" {
			klog.Info("Checking status on all nodes")
			w.Write([]byte("Checking status on all nodes\n\n"))
			err, out := runAllTestsRemote("all", checks, batch, jobName)
			if err != nil {
				klog.Error(err.Error())
			}
			w.Write(*out)
		} else {
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
				err, out := runAllTestsRemote(hosts, checks, batch, jobName)
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

func JobNameHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Job's node self discovery test. Use with /status?job=namespace:jobname"))
		jobName := r.URL.Query().Get("job")
		if jobName == "" {
			jobName = "None"
		}
	}
	return http.HandlerFunc(fn)
}

func GPUMemHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("HBM test. NOT IMPLEMENTED YET"))
	}
	return http.HandlerFunc(fn)
}
