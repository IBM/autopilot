package utils

import (
	"sync"

	"k8s.io/client-go/kubernetes"
)

type InitConfig struct {
	BWThreshold string
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
