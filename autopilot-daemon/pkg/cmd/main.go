package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/IBM/autopilot/pkg/handler"
	"github.com/IBM/autopilot/pkg/healthcheck"
	"github.com/IBM/autopilot/pkg/utils"
	"github.com/IBM/autopilot/pkg/worker"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog/v2"
)

func main() {
	port := flag.String("port", "3333", "Port for the webhook to listen to. Defaulted to 3333")
	bwThreshold := flag.Int("bw", 4, "Sets bandwidth threshold for the init container")
	logFile := flag.String("logfile", "", "File where to save all the events")
	v := flag.String("loglevel", "2", "Log level")
	repeat := flag.String("w", "24h", "Run all tests periodically on each node. Time set in interval format. Defaults to 24h")
	invasive := flag.String("invasive-check-timer", "4h", "Run invasive checks (e.g., dcgmi level 3) on each node when GPUs are free. Time set in interval format. Defaults to 4h. Set to 0 to avoid invasive checks")
	poolSizeInput := flag.Int("workers", 0, "Number of workers to use for concurrent health checks. Defaults to 0 which uses 2*number_of_logical_CPU_cores")

	flag.Parse()

	klog.InitFlags(nil)
	flag.Set("alsologtostderr", "true")
	if *logFile != "" {
		flag.Set("log_file", *logFile)
	}
	flag.Set("v", *v)
	flag.Set("logtostderr", "false")
	klog.OsExit = func(exitCode int) {
		fmt.Printf("os.Exit(%d)\n", exitCode)
	}

	utils.UserConfig = utils.InitConfig{
		BWThreshold: *bwThreshold,
	}

	reg := prometheus.NewRegistry()
	utils.InitMetrics(reg)

	utils.InitHardwareMetrics()

	// Init the node status map
	healthcheck.InitNodeStatusMap()

	pMux := http.NewServeMux()
	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	pMux.Handle("/metrics", promHandler)

	go func() {
		klog.Info("Serving metrics on :8081")
		err := http.ListenAndServe(":8081", pMux)
		if err != nil {
			klog.Error(err.Error())
			os.Exit(1)
		}
	}()

	readinessMux := http.NewServeMux()
	readinessMux.Handle("/readinessprobe", handler.ReadinessProbeHandler())

	go func() {
		klog.Info("Serving Readiness Probe on :8080")
		err := http.ListenAndServe(":8080", readinessMux)
		if err != nil {
			klog.Error(err.Error())
			os.Exit(1)
		}
	}()

	hcMux := http.NewServeMux()

	hcMux.Handle("/dcgm", handler.DCGMHandler())
	hcMux.Handle("/gpumem", handler.GpuMemHandler())
	hcMux.Handle("/gpupower", handler.GpuPowerHandler())
	hcMux.Handle("/iperf", handler.IperfHandler())
	hcMux.Handle("/iperfservers", handler.StartIperfServersHandler())
	hcMux.Handle("/iperfstopservers", handler.StopAllIperfServersHandler())
	hcMux.Handle("/iperfclients", handler.StartIperfClientsHandler())
	hcMux.Handle("/invasive", handler.InvasiveCheckHandler())
	hcMux.Handle("/pciebw", handler.PCIeBWHandler())
	hcMux.Handle("/ping", handler.PingHandler())
	hcMux.Handle("/pvc", handler.PVCHandler())
	hcMux.Handle("/remapped", handler.RemappedRowsHandler())
	hcMux.Handle("/status", handler.SystemStatusHandler())

	s := &http.Server{
		Addr:         ":" + *port,
		Handler:      hcMux,
		ReadTimeout:  30 * time.Minute,
		WriteTimeout: 30 * time.Minute,
		IdleTimeout:  30 * time.Minute,
	}

	go func() {
		klog.Info("Serving Health Checks on port :", *port)
		err := s.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			klog.Info("Server Closed")
		} else if errors.Is(err, http.ErrAbortHandler) {
			klog.Info("Server Aborted")
		} else if errors.Is(err, http.ErrContentLength) {
			klog.Info("Response size too large")
		} else if errors.Is(err, http.ErrBodyReadAfterClose) {
			klog.Info("Read after close")
		} else if errors.Is(err, http.ErrHandlerTimeout) {
			klog.Info("Handler timed out")
		}
		if err != nil {
			klog.Info("EXITING")
			klog.Error(err.Error())
			os.Exit(1)
		}
	}()

	// Create a Watcher over nodes. Needed to export metrics from data created by external jobs (i.e., dcgm Jobs)
	go utils.WatchNode()

	// Set the pool size based on the number of CPU cores
	poolSize := runtime.NumCPU() * 2 // use 2 workers per CPU core
	if *poolSizeInput > 0 {
		// if user has set a limit, use it
		poolSize = *poolSizeInput
	}

	// Create a WorkerPool to handle tasks concurrently
	workerPool := worker.CreateWorkerPool(poolSize)
	klog.Infof("Starting WorkerPool with %d workers", poolSize)

	// Run the health checks at startup, then start the timer
	workerPool.Submit(worker.TaskPeriodicCheck)

	// Parse the repeat and invasive intervals to durations
	repeatDuration, err := utils.ParseInterval(*repeat)
	if err != nil {
		klog.Error("Error parsing repeat interval: ", err)
		os.Exit(1)
	}
	invasiveDuration, err := utils.ParseInterval(*invasive)
	if err != nil {
		klog.Error("Error parsing invasive check interval: ", err)
		os.Exit(1)
	}

	periodicChecksTicker := time.NewTicker(repeatDuration)
	defer periodicChecksTicker.Stop()
	invasiveChecksTicker := time.NewTicker(invasiveDuration)
	defer invasiveChecksTicker.Stop()
	for {
		select {
		case <-periodicChecksTicker.C:
			workerPool.Submit(worker.TaskPeriodicCheck)
		case <-invasiveChecksTicker.C:
			if invasiveDuration > 0 {
				workerPool.Submit(worker.TaskInvasiveCheck)
			}
		}
	}

	// cert := "/etc/admission-webhook/tls/tls.crt"
	// key := "/etc/admission-webhook/tls/tls.key"

	// err := http.ListenAndServeTLS(":"+*port, cert, key, mux)
	// if errors.Is(err, http.ErrServerClosed) {
	// 	klog.Error("Server closed")
	// } else if err != nil {
	// 	klog.Error("error starting server: %s\n", err)
	// 	os.Exit(1)
	// }
}
