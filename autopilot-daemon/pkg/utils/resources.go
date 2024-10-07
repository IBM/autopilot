package utils

import (
	"os"

	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	resourcehelper "k8s.io/kubectl/pkg/util/resource"
)

var gpu_type corev1.ResourceName = "nvidia/gpu"
var resources = []corev1.ResourceName{"cpu", "memory", "nvidia.com/gpu", "nvidia.com/roce_gdr", "ephemeral-storage"} //storage, resourcequotas

func getNodeResourceUsage() (map[corev1.ResourceName]resource.Quantity, map[corev1.ResourceName]resource.Quantity) {
	cset := GetClientsetInstance()
	fieldselector, err := fields.ParseSelector("spec.nodeName=" +
		os.Getenv("NODE_NAME") + ",status.phase!=" + string(corev1.PodSucceeded))
	if err != nil {
		panic(err)
	}
	pods, err := cset.Cset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: fieldselector.String(),
	})
	if err != nil {
		panic(err)
	}
	requests := make(map[corev1.ResourceName]resource.Quantity)
	limits := make(map[corev1.ResourceName]resource.Quantity)

	for _, r := range resources {
		requests[r] = resource.Quantity{}
		limits[r] = resource.Quantity{}
	}
	for _, pod := range pods.Items {
		reqs, lims := resourcehelper.PodRequestsAndLimits(&pod)
		for _, r := range resources {
			if _, ok := reqs[r]; ok {
				tmp := requests[r]
				tmp.Add(reqs[r])
				requests[r] = tmp
			}
			if _, ok := lims[r]; ok {
				tmp := limits[r]
				tmp.Add(lims[r])
				limits[r] = tmp
			}
		}
	}
	return requests, limits
}

func confirmResourceType(res corev1.ResourceName) bool {
	for _, val := range resources {
		if val == res {
			return true
		}
	}
	return false
}

func CheckLimitCapacity(new_amount string, res corev1.ResourceName) bool {
	// convert new amount into ResourceQuantity
	if !confirmResourceType(res) {
		panic("Unknown ResourceName: " + res)
	}
	// compare
	new_res, err := resource.ParseQuantity(new_amount)
	if err != nil {
		panic(err)
	}

	_, node_lim := getNodeResourceUsage() // Hmm.. we may not want to call this each time...
	node_cap := getNodeResourceCapacity()

	res_cap := node_cap[res]
	res_lim := node_lim[res]
	res_lim.Add(new_res)
	if res_cap.Cmp(res_lim) < 0 {
		return false
	}

	return true
}

func getNodeResourceCapacity() map[corev1.ResourceName]resource.Quantity {
	// Get node information and capacity
	cset := GetClientsetInstance()
	node, err := cset.Cset.CoreV1().Nodes().Get(context.Background(), os.Getenv("NODE_NAME"), metav1.GetOptions{})
	if err != nil {
		panic(err)
	}
	return node.Status.Capacity

}

func PrintResourceUsageHeader() string {
	line1 := "Resource\t Requests\t Limits\t Capacity\n"
	line2 := "________\t ________\t ______\t ________\n"
	return line1 + line2
}

func PrintResourceUsage() string {
	reqs, lims := getNodeResourceUsage()
	cap := getNodeResourceCapacity()
	output := ""
	for _, r := range resources {
		output += r.String() + "\t"
		if _, ok := reqs[r]; ok {
			tmp := reqs[r]
			output += tmp.String() + "\t"
		}
		if _, ok := lims[r]; ok {
			tmp := lims[r]
			output += tmp.String() + "\t"
		}
		if _, ok := cap[r]; ok {
			tmp := cap[r]
			output += tmp.String()
		}
		output += "\n"
	}
	return output
}

// Returns true if GPUs are not currently requested by any workload
// TODO: CONFIRM THAT THIS CAN REPLACE WHATS IN functions.go
//func GPUsAvailability() bool {
//	reqs, lims := getNodeResourceUsage()
//	gpu_req := reqs[gpu_type]
//	gpu_lim := lims[gpu_type]
//	if gpu_req.Value() > 0 || gpu_lim.Value() > 0 {
//		return false
//	}
//	return true
//}
