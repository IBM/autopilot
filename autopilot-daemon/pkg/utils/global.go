package utils

import (
	"os"
	"sync"

	"k8s.io/client-go/kubernetes"
)

type InitConfig struct {
	BWThreshold int
}

var UserConfig InitConfig

type K8sClientset struct {
	Cset *kubernetes.Clientset
}

var k8sClientset *K8sClientset
var csetLock sync.Mutex

var HealthcheckLock sync.Mutex

var CPUModel string
var GPUModel string

var NodeName string = os.Getenv("NODE_NAME")
var Namespace string = os.Getenv("NAMESPACE")
var PodName string = os.Getenv("POD_NAME")
