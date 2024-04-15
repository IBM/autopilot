package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/IBM/autopilot/pkg/handlers"
	"github.com/IBM/autopilot/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	toolswatch "k8s.io/client-go/tools/watch"
	"k8s.io/klog/v2"
)

func watchNode() {

	watchFunc := func(options metav1.ListOptions) (watch.Interface, error) {
		timeout := int64(60)
		fieldselector, err := fields.ParseSelector("metadata.name=" + os.Getenv("NODE_NAME"))
		if err != nil {
			klog.Info("Error in creating the field selector", err.Error())
			return nil, err
		}
		return utils.GetClientsetInstance().Cset.CoreV1().Nodes().Watch(context.Background(), metav1.ListOptions{TimeoutSeconds: &timeout, FieldSelector: fieldselector.String()})
	}

	watcher, _ := toolswatch.NewRetryWatcher("1", &cache.ListWatch{WatchFunc: watchFunc})

	for event := range watcher.ResultChan() {
		item := event.Object.(*corev1.Node)

		switch event.Type {
		case watch.Modified:
			{
				key := "autopilot/dcgm.level.3"
				labels := item.GetLabels()
				if val, found := labels[key]; found {
					var res float64
					res = 0
					if strings.Contains(val, "PASS") {
						klog.Info("[DCGM level 3] Update observation: ", os.Getenv("NODE_NAME"), " Pass")
					} else {
						res = 1
						klog.Info("[DCGM level 3] Update observation: ", os.Getenv("NODE_NAME"), " Error found")
					}
					utils.HchecksGauge.WithLabelValues("dcgm", os.Getenv("NODE_NAME"), "").Set(res)
				}
			}
		}
	}

}

func main() {
	port := flag.String("port", "3333", "Port for the webhook to listen to. Defaulted to 3333")
	bwThreshold := flag.String("bw", "4", "Sets bandwidth threshold for the init container")
	logFile := flag.String("logfile", "", "File where to save all the events")
	v := flag.String("loglevel", "2", "Log level")
	repeat := flag.Int("w", 24, "Run all tests periodically on each node. Time set in hours. Defaults to 24h")
	intrusive := flag.Int("intrusive-check-timer", 4, "Run intrusive checks (e.g., dcgmi level 3) on each node when GPUs are free. Time set in hours. Defaults to 4h. Set to 0 to avoid intrusive checks")

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

	// Create a Watcher over nodes. Needed to export metrics from data created by external jobs (i.e., dcgm Jobs or PytorchJob NCCL tests)
	go watchNode()
	// Run the health checks at startup, then start the timer
	handlers.PeriodicCheckTimer()

	periodicChecksTicker := time.NewTicker(time.Duration(*repeat) * time.Hour)
	defer periodicChecksTicker.Stop()
	intrusiveChecksTicker := time.NewTicker(time.Duration(*intrusive) * time.Hour)
	defer periodicChecksTicker.Stop()
	for {
		select {
		case <-periodicChecksTicker.C:
			handlers.PeriodicCheckTimer()
		case <-intrusiveChecksTicker.C:
			if *intrusive > 0 {
				handlers.IntrusiveCheckTimer()
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
