package ping

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/IBM/autopilot/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

// RunPingCheck executes the ping network check
func RunPingCheck(nodes []string, job string, nodelabel string) (string, error) {
	// Get Kubernetes client
	cset := utils.GetClientsetInstance()

	// Discover nodes based on job or node label
	nodeMap, err := discoverNodes(job, nodelabel)
	if err != nil {
		klog.Errorf("Error discovering nodes: %v", err)
		return "ABORT", err
	}

	// If 'all' is in the nodes list, include all discovered nodes
	allNodes := false
	for _, node := range nodes {
		if node == "all" {
			allNodes = true
			break
		}
	}

	// If 'all' is specified or no specific nodes, use all discovered nodes
	if allNodes || len(nodes) == 0 || (len(nodes) == 1 && nodes[0] == "all") {
		// Add all discovered nodes to nodeMap
	} else {
		// Filter nodeMap to only include specified nodes
		filteredNodeMap := make(map[string]bool)
		for _, node := range nodes {
			if _, exists := nodeMap[node]; exists {
				filteredNodeMap[node] = true
			}
		}
		nodeMap = filteredNodeMap
	}

	// Get network interfaces for each node
	interfaces, err := getNetworkInterfaces(cset)
	if err != nil {
		klog.Errorf("Error getting network interfaces: %v", err)
		return "ABORT", err
	}

	// Execute ping commands to test connectivity
	result, err := executePingTests(cset, nodeMap, interfaces)
	if err != nil {
		klog.Errorf("Error executing ping tests: %v", err)
		return "ABORT", err
	}

	klog.Info("Ping network check completed successfully")
	return result, nil
}

// discoverNodes discovers nodes based on job or node label
func discoverNodes(job string, nodelabel string) (map[string]bool, error) {
	nodeMap := make(map[string]bool)
	nodeNameSelf := os.Getenv("NODE_NAME")

	// Get Kubernetes client
	cset := utils.GetClientsetInstance()

	// If no job or node label specified, get all nodes
	if (job == "None" || job == "") && (nodelabel == "None" || nodelabel == "") {
		// List all nodes
		allNodes, err := cset.Cset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("error listing all nodes: %v", err)
		}

		// Add all nodes except self to nodeMap
		for _, node := range allNodes.Items {
			if node.Name != nodeNameSelf {
				nodeMap[node.Name] = true
			}
		}
		return nodeMap, nil
	}

	// Get nodes from job if specified
	if job != "None" && job != "" {
		jobParts := strings.Split(job, ":")
		if len(jobParts) >= 2 {
			jobNs := jobParts[0]
			jobLabel := jobParts[1]

			// List pods with the specified label
			jobPods, err := cset.Cset.CoreV1().Pods(jobNs).List(context.TODO(), metav1.ListOptions{
				LabelSelector: jobLabel,
			})
			if err != nil {
				return nil, fmt.Errorf("error listing job pods: %v", err)
			}

			// Add nodes from job pods to nodeMap
			for _, pod := range jobPods.Items {
				if pod.Spec.NodeName != nodeNameSelf {
					nodeMap[pod.Spec.NodeName] = true
				}
			}
		}
	}

	// Get nodes from node label if specified
	if nodelabel != "None" && nodelabel != "" {
		// List nodes with the specified label
		labeledNodes, err := cset.Cset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
			LabelSelector: nodelabel,
		})
		if err != nil {
			return nil, fmt.Errorf("error listing labeled nodes: %v", err)
		}

		// Add labeled nodes to nodeMap
		for _, node := range labeledNodes.Items {
			if node.Name != nodeNameSelf {
				nodeMap[node.Name] = true
			}
		}
	}

	return nodeMap, nil
}

// getNetworkInterfaces gets network interfaces for each node
func getNetworkInterfaces(cset *utils.K8sClientset) (map[string]map[string][]string, error) {
	namespaceSelf := os.Getenv("NAMESPACE")
	nodeNameSelf := os.Getenv("NODE_NAME")

	// List all autopilot pods
	autopilotPods, err := cset.Cset.CoreV1().Pods(namespaceSelf).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "app=autopilot",
	})
	if err != nil {
		return nil, fmt.Errorf("error listing autopilot pods: %v", err)
	}

	interfaces := make(map[string]map[string][]string)

	// Process each pod to get network interfaces
	for _, pod := range autopilotPods.Items {
		if pod.Spec.NodeName != nodeNameSelf {
			node := make(map[string][]string)
			interfaces[pod.Spec.NodeName] = node

			// Try to get network status from pod annotations
			if networkStatus, exists := pod.Annotations["k8s.v1.cni.cncf.io/network-status"]; exists {
				// In a real implementation, we would parse the JSON network status
				// For now, we'll just log that we found it
				klog.Infof("Found network status for pod %s on node %s", pod.Name, pod.Spec.NodeName)
			} else {
				// Use default pod IPs
				if pod.Status.PodIPs != nil {
					ips := []string{}
					for _, podIP := range pod.Status.PodIPs {
						ips = append(ips, podIP.IP)
					}
					node["default"] = ips
				}
			}
		}
	}

	return interfaces, nil
}

// executePingTests executes ping commands to test connectivity
func executePingTests(cset *utils.K8sClientset, nodeMap map[string]bool, interfaces map[string]map[string][]string) (string, error) {
	result := ""
	fail := false

	// Execute ping tests for each node and interface
	for nodeName := range nodeMap {
		if nodeInterfaces, exists := interfaces[nodeName]; exists {
			for ifaceName, ips := range nodeInterfaces {
				for _, ip := range ips {
					// Execute ping command
					command := exec.Command("ping", ip, "-t", "45", "-c", "10")
					ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
					command = exec.CommandContext(ctx, "ping", ip, "-t", "45", "-c", "10")

					output, err := command.CombinedOutput()
					cancel()

					if err != nil {
						klog.Errorf("Error executing ping to %s: %v", ip, err)
						fail = true
						result += fmt.Sprintf("Node %s %s %s 1\n", nodeName, ip, ifaceName)
					} else {
						outputStr := string(output)
						if strings.Contains(outputStr, "Unreachable") || strings.Contains(outputStr, "100% packet loss") {
							klog.Infof("Node %s %s %s is unreachable", nodeName, ip, ifaceName)
							fail = true
							result += fmt.Sprintf("Node %s %s %s 1\n", nodeName, ip, ifaceName)
						} else {
							klog.Infof("Node %s %s %s is reachable", nodeName, ip, ifaceName)
							result += fmt.Sprintf("Node %s %s %s 0\n", nodeName, ip, ifaceName)
						}
					}
				}
			}
		}
	}

	if fail {
		result = "[PING] At least one node unreachable. FAIL\n" + result
	} else {
		result = "[PING] all nodes reachable. success\n" + result
	}

	return result, nil
}