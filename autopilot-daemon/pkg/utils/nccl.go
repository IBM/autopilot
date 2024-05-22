package utils

import (
	"context"
	"github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	clientv1 "github.com/kubeflow/training-operator/pkg/client/clientset/versioned/typed/kubeflow.org/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"os"
	"sync"
)

type KubeflowClientset struct {
	Cset *clientv1.KubeflowV1Client
}

var kubeflowClientset *KubeflowClientset
var kubeflowLock sync.Mutex

func getKubeflowClientsetInstance() *KubeflowClientset {
	if kubeflowClientset == nil {
		kubeflowLock.Lock()
		defer kubeflowLock.Unlock()
		if kubeflowClientset == nil {
			kubeflowClientset = &KubeflowClientset{}
			config, err := rest.InClusterConfig()
			if err != nil {
				panic(err.Error())
			}
			kubeflowClientset.Cset, err = clientv1.NewForConfig(config)
			if err != nil {
				panic(err.Error())
			}
		}
	}
	return kubeflowClientset
}

func ConfirmNCCLSupport() bool {

	_, exists := os.LookupEnv("NCCL_POD_IMAGE_URI")
	if !exists {
		klog.Info("NCCL Test Error: NCCL_POD_IMAGE_URI UNDEFINED")
		return false
	}
	nccl_cpu, exists := os.LookupEnv("NCCL_POD_RESOURCE_LIMIT_CPU")
	if !exists {
		klog.Info("NCCL Test Error: NCCL_POD_RESOURCE_LIMIT_CPU UNDEFINED")
		return false
	}
	nccl_gpu, exists := os.LookupEnv("NCCL_POD_RESOURCE_LIMIT_GPU")
	if !exists {
		klog.Info("NCCL Test Error: NCCL_POD_RESOURCE_LIMIT_GPU UNDEFINED")
		return false
	}
	nccl_mem, exists := os.LookupEnv("NCCL_POD_RESOURCE_LIMIT_MEMORY")
	if !exists {
		klog.Info("NCCL Test Error: NCCL_POD_RESOURCE_LIMIT_MEMORY UNDEFINED")
		return false
	}
	nccl_rdma, exists := os.LookupEnv("NCCL_POD_RESOURCE_LIMIT_RDMA_DEVICES")
	if !exists {
		klog.Info("NCCL Test Error: NCCL_POD_RESOURCE_LIMIT_RDMA_DEVICES UNDEFINED")
		return false
	}

	klog.Info("NCCL test requires the following resources: CPU:", nccl_cpu, " GPU:",
		nccl_gpu, " MEM:", nccl_mem, " RDMA DEVS:", nccl_rdma)

	if !CheckLimitCapacity(nccl_gpu, "nvidia.com/gpu") {
		klog.Info("NCCL Test Error: Inefficient GPUs")
		return false
	}
	if !CheckLimitCapacity(nccl_rdma, "nvidia.com/roce_gdr") {
		klog.Info("NCCL Test Error: Inefficient RMDA Devices")
		return false
	}
	if !CheckLimitCapacity(nccl_cpu, "cpu") {
		klog.Info("NCCL Test Error: Inefficient CPU")
		return false
	}
	if !CheckLimitCapacity(nccl_mem, "memory") {
		klog.Info("NCCL Test Error: Inefficient Memory")
		return false
	}

	// TODO: check that kubeflow APIs are available

	return true
}

func ncclPodSpec() (*corev1.PodSpec, error) {
	var args []string
	var cmd []string
	img := os.Getenv("NCCL_POD_IMAGE_URI")

	// command & args
	pytorchjobcmd := "torchrun --nnodes=${WORLD_SIZE} --node_rank=${RANK} --nproc_per_node=" + os.Getenv("NCCL_POD_RESOURCE_LIMIT_GPU") + " --master_addr=${MASTER_ADDR} --master_port=${MASTER_PORT} /apps/allreduce-loop.py"
	cmd = []string{"/bin/bash"}
	args = []string{"-c", "source /apps/nccl_envs.sh; " + pytorchjobcmd}

	// parse resource quantities
	cpus, err := resource.ParseQuantity(os.Getenv("NCCL_POD_RESOURCE_LIMIT_CPU"))
	if err != nil {
		return nil, err
	}
	mem, err := resource.ParseQuantity(os.Getenv("NCCL_POD_RESOURCE_LIMIT_MEMORY"))
	if err != nil {
		return nil, err
	}
	gpus, err := resource.ParseQuantity(os.Getenv("NCCL_POD_RESOURCE_LIMIT_GPU"))
	if err != nil {
		return nil, err
	}
	rdma, err := resource.ParseQuantity(os.Getenv("NCCL_POD_RESOURCE_LIMIT_RDMA_DEVICES"))
	if err != nil {
		return nil, err
	}

	// Environment Settings
	// Defaults set in /utility-tools/nccl/nccl_envs.sh

	// TODO: CUDA_VISABLE_DEVICES != 8
	// TODO: NCCL_ALGO = Tree vs Ring?
	// TODO: NCCL_NSOCKS_PERTHREAD when RoCE/RDR is disabled

	return &corev1.PodSpec{
		RestartPolicy:      "Never",
		ServiceAccountName: "default", // TODO: make this configurable ??
		//NodeName:           os.Getenv("NODE_NAME"),
		//NodeSelector: map[string]string{ // keeping this for now
		//	"cpu-model.id": "85",
		//},
		Containers: []corev1.Container{
			{
				Name:            "pytorch",
				ImagePullPolicy: "Always",
				Image:           img,
				Command:         cmd,
				Args:            args,
				SecurityContext: &corev1.SecurityContext{
					Capabilities: &corev1.Capabilities{
						Add: []corev1.Capability{
							"IPC_LOCK",
						},
					},
				},
				Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						"cpu":                 cpus,
						"memory":              mem,
						"nvidia.com/gpu":      gpus,
						"nvidia.com/roce_gdr": rdma,
					},
					Requests: corev1.ResourceList{
						"cpu":                 cpus,
						"memory":              mem,
						"nvidia.com/gpu":      gpus,
						"nvidia.com/roce_gdr": rdma,
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "dshm",
						MountPath: "/dev/shm",
					},
				},
				Env: []corev1.EnvVar{ // Leaving this here for now..
					{
						Name:  "NODE_NAME",
						Value: os.Getenv("NODE_NAME"),
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "dshm",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{
						Medium: corev1.StorageMediumMemory,
					},
				},
			},
		},
	}, nil
}

func CreateNCCLPyTorchJob() (string, error) {

	// Create the PyTorchJob client
	pytorchJobClient := getKubeflowClientsetInstance()

	// Define the PodSpec
	podSpec, err := ncclPodSpec()
	if err != nil {
		klog.Info("NCCL Test Error: Cannot get podSpec:", err.Error())
		return "", nil
	}

	replicas := int32(1)

	// Define the PyTorchJob
	pytorchJob := &v1.PyTorchJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PyTorchJob",
			APIVersion: "kubeflow.org/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pytorchjob-example",
			Namespace: "default",
		},
		Spec: v1.PyTorchJobSpec{
			PyTorchReplicaSpecs: map[v1.ReplicaType]*v1.ReplicaSpec{
				v1.PyTorchJobReplicaTypeMaster: {
					Replicas:      &replicas,
					RestartPolicy: v1.RestartPolicyOnFailure,
					Template: corev1.PodTemplateSpec{
						Spec: *podSpec,
					},
				},
				v1.PyTorchJobReplicaTypeWorker: {
					Replicas:      &replicas,
					RestartPolicy: v1.RestartPolicyOnFailure,
					Template: corev1.PodTemplateSpec{
						Spec: *podSpec,
					},
				},
			},
		},
	}

	// Create the PyTorchJob
	newJob, err := pytorchJobClient.Cset.PyTorchJobs("default").Create(context.TODO(), pytorchJob, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}
	return newJob.Name, nil
}
