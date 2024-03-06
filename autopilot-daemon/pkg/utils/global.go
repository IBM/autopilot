package utils

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type InitConfig struct {
	BWThreshold string
}

var UserConfig InitConfig

type K8sClientset struct {
	Cset *kubernetes.Clientset
}

var k8sClientset *K8sClientset

func GetClientsetInstance() *K8sClientset {
	var lock = &sync.Mutex{}
	if k8sClientset == nil {
		// creates the in-cluster config
		lock.Lock()
		defer lock.Unlock()
		if k8sClientset == nil {
			k8sClientset = &K8sClientset{}
			config, err := rest.InClusterConfig()
			if err != nil {
				panic(err.Error())
			}
			// creates the clientset
			k8sClientset.Cset, err = kubernetes.NewForConfig(config)
			if err != nil {
				panic(err.Error())
			}
		}

	}
	return k8sClientset
}

var (
	Requests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "autopilot",
			Name:      "health_checks_req_total",
			Help:      "Number of invocations to Autopilot",
		},
	)

	HchecksGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "autopilot",
			Name:      "health_checks",
			Help:      "Summary of the health checks measurements on compute nodes. Gauge Vector version",
		},
		[]string{"health", "node", "deviceid"},
	)
)

func Initmetrics(reg prometheus.Registerer) {
	// Register custom metrics with the global prometheus registry
	reg.MustRegister(HchecksGauge)
}
