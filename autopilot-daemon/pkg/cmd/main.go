package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.ibm.com/hybrid-cloud-infrastructure-research/autopilot-daemon/pkg/handlers"
	"github.ibm.com/hybrid-cloud-infrastructure-research/autopilot-daemon/pkg/utils"
	"k8s.io/klog/v2"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	port := flag.String("port", "3333", "Port for the webhook to listen to. Defaulted to 3333")
	// InitContainerImagePCIeBW := flag.String("pciebw", "", "Init container for PCIe bandwidth test")
	// InitContainerImageMem := flag.String("gpumem", "", "Init container for gpu memory test")
	// InitContainerImageNet := flag.String("netreach", "", "Init container for secondary nic reachability test")
	bwThreshold := flag.String("bw", "4", "Sets bandwidth threshold for the init container")
	logFile := flag.String("logfile", "report.log", "File where requests counter and info is being stored")
	v := flag.String("loglevel", "2", "Log level")
	flag.Parse()

	// var fs flag.FlagSet
	klog.InitFlags(nil)
	flag.Set("alsologtostderr", "true")
	flag.Set("log_file", *logFile)
	flag.Set("v", *v)
	flag.Set("logtostderr", "false")
	klog.OsExit = func(exitCode int) {
		fmt.Printf("os.Exit(%d)\n", exitCode)
	}

	utils.UserConfig = utils.InitConfig{
		// InitContainerImagePCIeBW: *InitContainerImagePCIeBW,
		// InitContainerImageMem:    *InitContainerImageMem,
		// InitContainerImageNet:    *InitContainerImageNet,
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
	hcMux.Handle("/pciebw", handlers.PCIeBWHandler("4"))
	hcMux.Handle("/nic", handlers.NetReachHandler())
	// hcMux.Handle("/gpumem", handlers.GPUMemHandler())
	hcMux.Handle("/remapped", handlers.RemappedRowsHandler())
	hcMux.Handle("/status", handlers.SystemStatusHandler())

	klog.Info("Serving Health Checks on port :", *port)
	err := http.ListenAndServe(":"+*port, hcMux)
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
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
