package utils

import (
	"context"
	"strings"

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
		fieldselector, err := fields.ParseSelector("metadata.name=" + NodeName)
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
					if strings.Contains(val, "EVICT") {
						res = 1
						klog.Info("[DCGM level 3] Update observation: ", NodeName, " Fatal error found")
					}
					HchecksGauge.WithLabelValues("dcgm", NodeName, CPUModel, GPUModel, "").Set(res)
				}
			}
		}
	}
}
