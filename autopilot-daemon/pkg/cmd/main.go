package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/IBM/autopilot/pkg/handler"
	"github.com/IBM/autopilot/pkg/healthcheck"
	"github.com/IBM/autopilot/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog/v2"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:  "run-healthchecks",
				Usage: "Run health checks on all nodes or a specific node(s)",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "service",
						Value: "autopilot-healthchecks",
						Usage: "Autopilot healthchecks service name",
					},
					&cli.StringFlag{
						Name:  "namespace",
						Value: "autopilot",
						Usage: "Namespace where autopilot DaemonSet is deployed",
					},
					&cli.StringFlag{
						Name:  "nodes",
						Value: "all",
						Usage: "Node(s) that will run a healthcheck. Can be a comma separated list",
					},
					&cli.StringFlag{
						Name:  "check",
						Value: "all",
						Usage: "The specific test(s) that will run: \"all\", \"pciebw\", \"dcgm\", \"remapped\", \"ping\", \"gpumem\", \"pvc\" or \"gpupower\". Can be a comma separated list",
					},
					&cli.IntFlag{
						Name:  "batchSize",
						Value: 0,
						Usage: "Number of nodes to check in parallel",
					},
					&cli.StringFlag{
						Name:  "wkload",
						Value: "None",
						Usage: "Workload node discovery w/ given namespace and label. Ex: \"--wkload=namespace:label-key=label-value\"",
					},
					&cli.StringFlag{
						Name:  "dcgmR",
						Value: "1",
						Usage: "Run a diagnostic in dcgmi",
					},
					&cli.StringFlag{
						Name:  "nodelabel",
						Value: "None",
						Usage: "Node label to select nodes. Ex: \"label-key=label-value\"",
					},
				},
				Action: func(c *cli.Context) error {
					return healthcheck.RunHealthChecks(c)
				},
			},
			{
				Name:  "daemon",
				Usage: "Run the autopilot daemon",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "port",
						Value: "3333",
						Usage: "Port for the webhook to listen to",
					},
					&cli.IntFlag{
						Name:  "bw",
						Value: 4,
						Usage: "Sets bandwidth threshold for the init container",
					},
					&cli.StringFlag{
						Name:  "logfile",
						Usage: "File where to save all the events",
					},
					&cli.StringFlag{
						Name:  "loglevel",
						Value: "2",
						Usage: "Log level",
					},
					&cli.IntFlag{
						Name:  "w",
						Value: 24,
						Usage: "Run all tests periodically on each node. Time set in hours",
					},
					&cli.IntFlag{
						Name:  "invasive-check-timer",
						Value: 4,
						Usage: "Run invasive checks (e.g., dcgmi level 3) on each node when GPUs are free. Time set in hours",
					},
				},
				Action: func(c *cli.Context) error {
					port := c.String("port")
					bwThreshold := c.Int("bw")
					logFile := c.String("logfile")
					v := c.String("loglevel")
					repeat := c.Int("w")
					invasive := c.Int("invasive-check-timer")

					klog.InitFlags(nil)
					flag.Set("alsologtostderr", "true")
					if logFile != "" {
						flag.Set("log_file", logFile)
					}
					flag.Set("v", v)
					flag.Set("logtostderr", "false")
					klog.OsExit = func(exitCode int) {
						fmt.Printf("os.Exit(%d)\n", exitCode)
					}

					utils.UserConfig = utils.InitConfig{
						BWThreshold: bwThreshold,
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
						Addr:         ":" + port,
						Handler:      hcMux,
						ReadTimeout:  30 * time.Minute,
						WriteTimeout: 30 * time.Minute,
						IdleTimeout:  30 * time.Minute,
					}

					go func() {
						klog.Info("Serving Health Checks on port :", port)
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

					// Run the health checks at startup, then start the timer
					healthcheck.PeriodicCheck()

					periodicChecksTicker := time.NewTicker(time.Duration(repeat) * time.Hour)
					defer periodicChecksTicker.Stop()
					invasiveChecksTicker := time.NewTicker(time.Duration(invasive) * time.Hour)
					defer invasiveChecksTicker.Stop()
					for {
						select {
						case <-periodicChecksTicker.C:
							healthcheck.PeriodicCheck()
						case <-invasiveChecksTicker.C:
							if invasive > 0 {
								healthcheck.InvasiveCheck()
							}
						}
					}
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		klog.Fatal(err)
	}
}
