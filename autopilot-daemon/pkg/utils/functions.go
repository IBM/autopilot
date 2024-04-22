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
	//resourcehelper "k8s.io/kubectl/pkg/util/resource"
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

// Returns the daemonset of this pod
func GetDaemonset() string {
	if value, ok := os.LookupEnv("DAEMONSET"); ok {
		return value
	} else {
		cset := GetClientsetInstance()
		pod, err := cset.Cset.CoreV1().Pods(os.Getenv("NAMESPACE")).Get(context.TODO(), os.Getenv("POD_NAME"), metav1.GetOptions{})
		if err == nil {
			for _, ownerRef := range pod.OwnerReferences {
				if ownerRef.Kind == "DaemonSet" {
					os.Setenv("DAEMONSET", ownerRef.Name)
					return ownerRef.Name
				}
			}
		}
		return ""
	}
}

// Returns the service account of this pod
func GetServiceAccount() string {
	if value, ok := os.LookupEnv("SERVICE_ACCOUNT"); ok {
		return value
	} else {
		cset := GetClientsetInstance()
		pod, err := cset.Cset.CoreV1().Pods(os.Getenv("NAMESPACE")).Get(context.TODO(), os.Getenv("POD_NAME"), metav1.GetOptions{})
		if err != nil {
			return ""
		} else {
			os.Setenv("SERVICE_ACCOUNT", pod.Spec.ServiceAccountName)
			return pod.Spec.ServiceAccountName
		}
	}
}

// Creates a new job on this node, returns job name
func CreateJob(healthcheck string) string {
	var args []string
	var cmd []string
	switch healthcheck {
	case "dcgm":
		cmd = []string{"python3"}
		args = []string{"gpu-dcgm/entrypoint.py", "-r", "3", "-l"}
	}
	cset := GetClientsetInstance()

	fieldselector, err := fields.ParseSelector("metadata.name=" + os.Getenv("POD_NAME"))
	if err != nil {
		klog.Info("Error in creating the field selector", err.Error())
	}
	pods, err := cset.Cset.CoreV1().Pods(os.Getenv("NAMESPACE")).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fieldselector.String(),
	})
	if err != nil {
		klog.Info("Cannot get pod:", err.Error())
		return ""
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
					ServiceAccountName: GetServiceAccount(),
					NodeName:           os.Getenv("NODE_NAME"), // TODO: Make a parameter???
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
	new_job, err := cset.Cset.BatchV1().Jobs(os.Getenv("NAMESPACE")).Create(context.TODO(), job,
		metav1.CreateOptions{})
	if err != nil {
		klog.Info("Couldn't create Job ", err.Error())
		return ""
	}
	return new_job.Name
}
