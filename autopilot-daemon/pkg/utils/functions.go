package utils

import (
	"os"

	"context"

	"github.com/thanhpk/randstr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	resourcehelper "k8s.io/kubectl/pkg/util/resource"
)

func GetClientsetInstance() *K8sClientset {
	if k8sClientset == nil {
		csetLock.Lock()
		defer csetLock.Unlock()
		if k8sClientset == nil {
			k8sClientset = &K8sClientset{}
			config, err := rest.InClusterConfig()
			if err != nil {
				panic(err.Error())
			}
			k8sClientset.Cset, err = kubernetes.NewForConfig(config)
			if err != nil {
				panic(err.Error())
			}
		}

	}
	return k8sClientset
}

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

func CreateJob(healthcheck string) {
	var args []string
	var cmd []string
	switch healthcheck {
	case "dcgm":
		// cmd = []string{"dcgmi"}
		// args = []string{"diag", "-r", "1"}
		cmd = []string{"python3"}
		args = []string{"gpu-dcgm/entrypoint.py", "-r", "3", "-l"}
	}
	cset := GetClientsetInstance()

	fieldselector, err := fields.ParseSelector("metadata.name=" + os.Getenv("POD_NAME"))
	if err != nil {
		klog.Info("Error in creating the field selector", err.Error())
	}
	pods, err := cset.Cset.CoreV1().Pods("autopilot").List(context.TODO(), metav1.ListOptions{
		FieldSelector: fieldselector.String(),
	})
	if err != nil {
		klog.Info("Cannot get pod:", err.Error())
	}
	autopilotPod := pods.Items[0]
	ttlsec := int32(4 * 60 * 60) // setting TTL to 4 hours
	backofflimits := int32(0)
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      healthcheck + "-" + randstr.Hex(6),
			Namespace: autopilotPod.Namespace,
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &ttlsec,
			BackoffLimit:            &backofflimits,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy:      "Never",
					ServiceAccountName: "autopilot",
					NodeName:           os.Getenv("NODE_NAME"),
					InitContainers: []corev1.Container{
						{
							Name:            "init",
							Image:           autopilotPod.Spec.InitContainers[0].DeepCopy().Image,
							ImagePullPolicy: "IfNotPresent",
							Command:         autopilotPod.Spec.InitContainers[0].DeepCopy().Command,
							Args:            autopilotPod.Spec.InitContainers[0].DeepCopy().Args,
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "main",
							Image:           autopilotPod.Spec.Containers[0].DeepCopy().Image,
							ImagePullPolicy: "IfNotPresent",
							Command:         cmd,
							Args:            args,
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("8"),
								},
								Requests: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("8"),
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "NODE_NAME",
									Value: os.Getenv("NODE_NAME"),
								},
							},
						},
					},
				},
			},
		},
	}
	klog.Info("Try create Job")
	_, err = cset.Cset.BatchV1().Jobs("autopilot").Create(context.TODO(), job,
		metav1.CreateOptions{})
	if err != nil {
		klog.Info("Couldn't create Job ", err.Error())
	}
}
