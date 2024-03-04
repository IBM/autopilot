package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/IBM/autopilot/pkg/handlers"
	"github.com/IBM/autopilot/pkg/utils"
	"k8s.io/klog/v2"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	port := flag.String("port", "3333", "Port for the webhook to listen to. Defaulted to 3333")
	bwThreshold := flag.String("bw", "4", "Sets bandwidth threshold for the init container")
	logFile := flag.String("logfile", "", "File where to save all the events")
	devmode := flag.Bool("dev", false, "Dev mode disables the execution of health checks at pod startup. Default set to True, therefore health checks are executed at pod startup first, and then periodically.")
	v := flag.String("loglevel", "2", "Log level")
	repeat := flag.Int("w", 24, "Run all tests periodically on each node. Time set in hours. Defaults to 24h")

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
	utils.Initmetrics(reg)

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

	hcMux := http.NewServeMux()

	hcMux.Handle("/pciebw", handlers.PCIeBWHandler(utils.UserConfig.BWThreshold))
	hcMux.Handle("/remapped", handlers.RemappedRowsHandler())
	hcMux.Handle("/status", handlers.SystemStatusHandler())
	hcMux.Handle("/iperf", handlers.IperfHandler())
	hcMux.Handle("/iperfservers", handlers.StartIperfServersHandler())
	hcMux.Handle("/dcgm", handlers.DCGMHandler())
	hcMux.Handle("/ping", handlers.PingHandler())
	hcMux.Handle("/gpupower", handlers.GpuPowerHandler())
	hcMux.Handle("/gpumem", handlers.GpuMemHandler())

	s := &http.Server{
		Addr:         ":" + *port,
		Handler:      hcMux,
		ReadTimeout:  30 * time.Minute,
		WriteTimeout: 30 * time.Minute,
		IdleTimeout:  30 * time.Minute,
	}

	go func() {
		klog.Info("Serving Health Checks on port :", *port)
		// err := http.ListenAndServe(":"+*port, hcMux)
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

	if !*devmode {
		handlers.TimerRun()
	}

	testsTicker := time.NewTicker(time.Duration(*repeat) * time.Hour)
	defer testsTicker.Stop()
	for {
		select {
		case <-testsTicker.C:
			handlers.TimerRun()
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
