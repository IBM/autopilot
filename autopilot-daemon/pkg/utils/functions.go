package utils

import (
	"os"

	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/klog/v2"
	resourcehelper "k8s.io/kubectl/pkg/util/resource"
)

// Returns true if GPUs are not currently requested by any workload
func GPUsAvailability() bool {
	cset := GetClientsetInstance()

	fieldselector, err := fields.ParseSelector("spec.nodeName=" + os.Getenv("NODE_NAME") + ",status.phase!=" + string(corev1.PodSucceeded))
	// + ",status.phase!=" + string(corev1.PodFailed)
	if err != nil {
		klog.Info("Error in creating the field selector", err.Error())
		return false
	}
	pods, err := cset.Cset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: fieldselector.String(),
	})
	if err != nil {
		klog.Info("Cannot list pods:", err.Error())
		return false
	}
	for _, pod := range pods.Items {
		podReqs, podLimits := resourcehelper.PodRequestsAndLimits(&pod)
		gpuReq := podReqs["nvidia.com/gpu"]
		gpuLim := podLimits["nvidia.com/gpu"]
		if gpuReq.Value() > 0 || gpuLim.Value() > 0 {
			klog.Info("Pod ", pod.Name, " with requests ", gpuReq.Value(), " and limits ", gpuLim.Value(), ". Cannot run invasive health checks.")
			return false
		}
	}
	klog.Info("GPUs are free. Can run invasive health checks.")
	return true
}
