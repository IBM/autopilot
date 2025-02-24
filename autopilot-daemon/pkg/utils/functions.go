package utils

import (
	"errors"
	"os"
	"strconv"

	"context"

	"github.com/thanhpk/randstr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	resourcehelper "k8s.io/kubectl/pkg/util/resource"
)

func GetClientsetInstance() *K8sClientset {
	csetLock.Lock()
	if k8sClientset == nil {
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
	csetLock.Unlock()
	return k8sClientset
}

func GetNode(nodename string) (*corev1.Node, error) {
	cset := GetClientsetInstance()
	fieldselector, err := fields.ParseSelector("metadata.name=" + nodename)
	if err != nil {
		klog.Info("Error in creating the field selector ", err.Error())
		return nil, err
	}
	instance, err := cset.Cset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{FieldSelector: fieldselector.String()})
	if err != nil {
		klog.Info("Error in creating the watcher ", err.Error())
		return nil, err
	}
	return &instance.Items[0], nil
}

// Returns true if GPUs are not currently requested by any workload
func GPUsAvailability() bool {
	node, _ := GetNode(NodeName)
	nodelabels := node.Labels
	if _, found := nodelabels["nvidia.com/gpu.present"]; !found {
		klog.Info("At least one GPU busy on node ", NodeName, ". Cannot run invasive health checks.")
		return false
	}
	// Once cleared, list pods using gpus and abort the check if gpus are in use
	fieldselector, err := fields.ParseSelector("spec.nodeName=" + NodeName + ",status.phase!=" + string(corev1.PodSucceeded))
	if err != nil {
		klog.Info("Error in creating the field selector ", err.Error())
		return false
	}
	cset := GetClientsetInstance()
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
	klog.Info("GPUs are free. Will run invasive health checks.")
	return true
}

func CreateJob(healthcheck string) error {
	var args []string
	var cmd []string
	switch healthcheck {
	case "dcgm":
		cmd = []string{"python3"}
		args = []string{"gpu-dcgm/entrypoint.py", "-r", "3", "-l", "-v"}
	}
	cset := GetClientsetInstance()

	fieldselector, err := fields.ParseSelector("metadata.name=" + PodName)
	if err != nil {
		klog.Info("Error in creating the field selector", err.Error())
		return err
	}
	pods, err := cset.Cset.CoreV1().Pods("autopilot").List(context.TODO(), metav1.ListOptions{
		FieldSelector: fieldselector.String(),
	})
	if err != nil {
		klog.Info("Cannot get pod:", err.Error())
		return err
	}
	autopilotPod := pods.Items[0]
	// setting TTL to 30 sec, but looking for used defined value
	ttlsec := int32(30)
	if os.Getenv("INVASIVE_JOB_TTLSEC") != "" {
		val, _ := strconv.Atoi(os.Getenv("INVASIVE_JOB_TTLSEC"))
		ttlsec = int32(val)
	}

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
					NodeName:           NodeName,
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
									Value: NodeName,
								},
							},
						},
					},
				},
			},
		},
	}
	klog.Info("Try create Job")
	_, err = cset.Cset.BatchV1().Jobs(Namespace).Create(context.TODO(), job,
		metav1.CreateOptions{})
	if err != nil {
		klog.Info("Couldn't create Job ", err.Error())
		return err
	}
	klog.Info("Created")
	return nil
}

func PatchNode(label string, nodename string, force bool) error {
	cset := GetClientsetInstance()

	// Should not patch the gpuhealth label if it's currently in TESTING or EVICT
	node, err := cset.Cset.CoreV1().Nodes().Get(context.TODO(), nodename, v1.GetOptions{})
	if err != nil {
		klog.Info("[Node Patch] Failed read node ", err.Error())
		return err
	}
	labels := node.GetLabels()
	if current, found := labels["autopilot.ibm.com/gpuhealth"]; found {
		klog.Info("Node ", nodename, " label found ", current)
		if current == "TESTING" || current == "EVICT" {
			if !force {
				klog.Info("Cannot patch node's label, value found: ", current)
				return errors.New("Node status " + current)
			} else {
				klog.Info("Force patch for completed testing")
			}
		}
	} else {
		klog.Info("No label found, will go ahead patching the node")
	}
	_, err = cset.Cset.CoreV1().Nodes().Patch(context.TODO(), nodename, types.StrategicMergePatchType, []byte(label), v1.PatchOptions{})
	if err != nil {
		klog.Info("[Node Patch] Failed. ", err.Error())
		return err
	}
	klog.Info("Node patched with label ", label)
	return nil
}
