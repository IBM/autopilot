package utils

import (
	"context"
	"os"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	toolswatch "k8s.io/client-go/tools/watch"
	"k8s.io/klog/v2"
)

func WatchNode() {

	watchFunc := func(options metav1.ListOptions) (watch.Interface, error) {
		timeout := int64(60)
		fieldselector, err := fields.ParseSelector("metadata.name=" + os.Getenv("NODE_NAME"))
		if err != nil {
			klog.Info("Error in creating the field selector", err.Error())
			return nil, err
		}
		instance, err := GetClientsetInstance().Cset.CoreV1().Nodes().Watch(context.Background(), metav1.ListOptions{TimeoutSeconds: &timeout, FieldSelector: fieldselector.String()})
		if err != nil {
			klog.Info("Error in creating the watcher", err.Error())
			return nil, err
		}
		return instance, err
	}

	watcher, _ := toolswatch.NewRetryWatcher("1", &cache.ListWatch{WatchFunc: watchFunc})

	for event := range watcher.ResultChan() {
		item := event.Object.(*corev1.Node)

		switch event.Type {
		case watch.Modified:
			{
				key := "autopilot.ibm.com/dcgm.level.3"
				labels := item.GetLabels()
				if val, found := labels[key]; found {
					var res float64
					res = 0
					if strings.Contains(val, "ERR") {
						res = 1
						klog.Info("[DCGM level 3] Update observation: ", os.Getenv("NODE_NAME"), " Error found")
					}
					HchecksGauge.WithLabelValues("dcgm", os.Getenv("NODE_NAME"), CPUModel, GPUModel, "").Set(res)
				}
			}
		}
	}

}

func ListPVC() (string, error) {
	pvc, err := GetClientsetInstance().Cset.CoreV1().PersistentVolumeClaims(os.Getenv("NAMESPACE")).Get(context.Background(), os.Getenv("POD_NAME"), metav1.GetOptions{})
	if err != nil {
		klog.Error("Error in creating the lister", err.Error())
		return "ABORT", err
	}
	switch pvc.Status.Phase {
	case "Bound":
		{
			klog.Info("[PVC Create-Delete] PVC Bound: SUCCESS")
			klog.Info("Observation: ", os.Getenv("NODE_NAME"), " 0")
			HchecksGauge.WithLabelValues("pvc", os.Getenv("NODE_NAME"), CPUModel, GPUModel, "").Set(0)
		}
	case "Pending":
		{
			waitonpvc := time.NewTicker(time.Minute)
			defer waitonpvc.Stop()
			<-waitonpvc.C
			pvc, err := GetClientsetInstance().Cset.CoreV1().PersistentVolumeClaims(os.Getenv("NAMESPACE")).Get(context.Background(), os.Getenv("POD_NAME"), metav1.GetOptions{})
			if err != nil {
				klog.Error("[PVC Create-Delete] Error in creating the lister: ", err.Error())
				return "[PVC Create-Delete] PVC not found. ABORT ", err
			}
			phase := pvc.Status.Phase
			if pvc.Status.Phase == "Pending" {
				klog.Info("[PVC Create-Delete] Timer is up with PVC Pending. Force delete. FAIL")
				klog.Info("Observation: ", os.Getenv("NODE_NAME"), " 1")
				HchecksGauge.WithLabelValues("pvc", os.Getenv("NODE_NAME"), CPUModel, GPUModel, "").Set(1)
				err := DeletePVC(os.Getenv("POD_NAME"))
				if err != nil {
					return "[PVC Create-Delete] Error in deleting the PVC. ABORT ", err
				}
				return "[PVC Create-Delete] FAIL", nil
			}
			if phase == "Bound" {
				klog.Info("[PVC Create-Delete] PVC Bound: SUCCESS")
				klog.Info("Observation: ", os.Getenv("NODE_NAME"), " 0")
				HchecksGauge.WithLabelValues("pvc", os.Getenv("NODE_NAME"), CPUModel, GPUModel, "").Set(0)
			}
		}
	}
	err = DeletePVC(os.Getenv("POD_NAME"))
	if err != nil {
		return "Error in deleting the PVC. ABORT ", err
	}
	return "[PVC Create-Delete] PVC SUCCESS", nil
}
