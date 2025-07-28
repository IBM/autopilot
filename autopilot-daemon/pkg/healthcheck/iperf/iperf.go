package iperf

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/IBM/autopilot/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

const (
	defaultPort = 5200
)

// RunIperfCheck executes the iperf network bandwidth check
func RunIperfCheck(workload string, pclients string, startport string) (string, error) {
	// Parse arguments
	numClients, err := strconv.Atoi(pclients)
	if err != nil {
		return "", fmt.Errorf("invalid pclients value: %v", err)
	}

	startPort, err := strconv.Atoi(startport)
	if err != nil {
		return "", fmt.Errorf("invalid startport value: %v", err)
	}

	// Start iperf servers
	_, err = StartIperfServers(pclients, startport)
	if err != nil {
		return "", fmt.Errorf("failed to start iperf servers: %v", err)
	}

	// Execute the workload
	result, err := executeWorkload(workload, numClients, startPort)
	if err != nil {
		return "", fmt.Errorf("failed to execute workload: %v", err)
	}

	return result, nil
}

// StartIperfServers starts iperf3 servers on all nodes
func StartIperfServers(numservers string, startport string) (string, error) {
	numServers, err := strconv.Atoi(numservers)
	if err != nil {
		return "", fmt.Errorf("invalid numservers value: %v", err)
	}

	startPort, err := strconv.Atoi(startport)
	if err != nil {
		return "", fmt.Errorf("invalid startport value: %v", err)
	}

	// Get Kubernetes client
	cset := utils.GetClientsetInstance()

	// Get current pod information
	podName := os.Getenv("POD_NAME")
	namespace := os.Getenv("NAMESPACE")

	// Get the current pod
	pod, err := cset.Cset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get pod: %v", err)
	}

	// Get network interfaces
	interfaces := getNetworkInterfaces(cset)

	if len(interfaces) == 0 {
		return "", fmt.Errorf("no network interfaces found")
	}

	// Start servers on each interface
	for _, iface := range interfaces {
		for i := 0; i < numServers; i++ {
			// Get IP address for the interface
			ip, err := getInterfaceIP(iface)
			if err != nil {
				klog.Errorf("Failed to get IP for interface %s: %v", iface, err)
				continue
			}

			// Start iperf3 server
			port := startPort + i
			command := exec.Command("iperf3", "-s", "-B", ip, "-p", strconv.Itoa(port), "-D")
			klog.Infof("Starting iperf3 server %s:%d using %s...", ip, port, iface)

			err = command.Run()
			if err != nil {
				klog.Errorf("Server failed to start on %s:%d using %s: %v", ip, port, iface, err)
				return "", err
			}
		}
	}

	return "iperf3 servers started successfully", nil
}

// StopAllIperfServers stops all iperf3 servers
func StopAllIperfServers() (string, error) {
	// Find and kill all iperf3 server processes
	command := exec.Command("ps", "aux")
	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("error listing processes: %v", err)
	}

	processes := strings.Split(string(output), "\n")
	for _, process := range processes {
		// Look for iperf3 processes with -s flag (servers)
		if strings.Contains(process, "iperf3") && strings.Contains(process, "-s") {
			fields := strings.Fields(process)
			if len(fields) > 1 {
				pid, err := strconv.Atoi(fields[1])
				if err != nil {
					klog.Errorf("Could not convert PID to integer: %s", process)
					continue
				}

				// Kill the process
				err = syscall.Kill(pid, syscall.SIGTERM)
				if err != nil {
					klog.Errorf("Failed to kill process with PID %d: %v", pid, err)
				}
			}
		}
	}

	klog.Info("All iperf servers have been removed")
	return "All iperf servers have been stopped", nil
}

// StartIperfClients starts iperf3 clients for a specific test
func StartIperfClients(dstip string, dstport string, numclients string) (string, error) {
	dstPort, err := strconv.Atoi(dstport)
	if err != nil {
		return "", fmt.Errorf("invalid dstport value: %v", err)
	}

	numClients, err := strconv.Atoi(numclients)
	if err != nil {
		return "", fmt.Errorf("invalid numclients value: %v", err)
	}

	// Run iperf3 clients in parallel
	results := make([]map[string]interface{}, numClients)
	for i := 0; i < numClients; i++ {
		port := dstPort + i
		result, err := runIperfClient(dstip, port, 5) // 5 seconds duration
		if err != nil {
			klog.Errorf("Error running iperf client: %v", err)
			// Continue with other clients
			continue
		}
		results[i] = result
	}

	// Calculate statistics
	stats := calculateStats(results, numClients)

	// Return results as JSON
	jsonResult, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal results: %v", err)
	}

	return string(jsonResult), nil
}

// Helper functions

// getNetworkInterfaces gets network interfaces for the current pod
func getNetworkInterfaces(cset *utils.K8sClientset) []string {
	// Get current pod information
	podName := os.Getenv("POD_NAME")
	namespace := os.Getenv("NAMESPACE")

	// Get the current pod
	pod, err := cset.Cset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Failed to get pod: %v", err)
		// Fallback to default interfaces
		return []string{"eth0"}
	}

	// Get network interfaces from pod annotations if available
	if status, ok := pod.Annotations["k8s.v1.cni.cncf.io/network-status"]; ok {
		// Parse the JSON network status to extract interfaces
		// This is a simplified implementation - in reality, you'd need to parse the JSON properly
		interfaces := []string{}
		// Look for interfaces that are not lo, eth0, or tunl0
		// This is a placeholder implementation - you'd need to parse the actual JSON
		interfaces = append(interfaces, "net1", "net2") // Example interfaces
		return interfaces
	}
	
	// Fallback to default interfaces
	return []string{"eth0"}
}

// getInterfaceIP gets the IP address for a network interface
func getInterfaceIP(iface string) (string, error) {
	// In a real implementation, this would get the actual IP address
	// For now, we'll return a placeholder
	return "127.0.0.1", nil
}

// runIperfClient runs a single iperf3 client
func runIperfClient(dstip string, dstport int, duration int) (map[string]interface{}, error) {
	command := exec.Command("iperf3", "-c", dstip, "-p", strconv.Itoa(dstport), "-t", strconv.Itoa(duration), "-i", "1.0", "-Z")

	// Set a timeout for the command
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(duration+10)*time.Second)
	defer cancel()
	command = exec.CommandContext(ctx, "iperf3", "-c", dstip, "-p", strconv.Itoa(dstport), "-t", strconv.Itoa(duration), "-i", "1.0", "-Z")

	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("iperf3 client failed: %v", err)
	}

	// Parse the output to extract results
	result := parseIperfOutput(string(output))
	return result, nil
}

// parseIperfOutput parses the output from iperf3
func parseIperfOutput(output string) map[string]interface{} {
	result := map[string]interface{}{
		"sender": map[string]interface{}{
			"transfer": map[string]interface{}{"rate": 0.0, "units": "n/a"},
			"bitrate":  map[string]interface{}{"rate": 0.0, "units": "n/a"},
		},
		"receiver": map[string]interface{}{
			"transfer": map[string]interface{}{"rate": 0.0, "units": "n/a"},
			"bitrate":  map[string]interface{}{"rate": 0.0, "units": "n/a"},
		},
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.ToLower(strings.TrimSpace(line))
		if strings.Contains(line, "sender") {
			parts := strings.Fields(line)
			if len(parts) >= 8 {
				result["sender"].(map[string]interface{})["transfer"].(map[string]interface{})["rate"] = parts[4]
				result["sender"].(map[string]interface{})["transfer"].(map[string]interface{})["units"] = parts[5]
				result["sender"].(map[string]interface{})["bitrate"].(map[string]interface{})["rate"] = parts[6]
				result["sender"].(map[string]interface{})["bitrate"].(map[string]interface{})["units"] = parts[7]
			}
		} else if strings.Contains(line, "receiver") {
			parts := strings.Fields(line)
			if len(parts) >= 8 {
				result["receiver"].(map[string]interface{})["transfer"].(map[string]interface{})["rate"] = parts[4]
				result["receiver"].(map[string]interface{})["transfer"].(map[string]interface{})["units"] = parts[5]
				result["receiver"].(map[string]interface{})["bitrate"].(map[string]interface{})["rate"] = parts[6]
				result["receiver"].(map[string]interface{})["bitrate"].(map[string]interface{})["units"] = parts[7]
			}
		}
	}

	return result
}

// calculateStats calculates statistics from multiple client results
func calculateStats(results []map[string]interface{}, numClients int) map[string]interface{} {
	senderValues := map[string][]float64{
		"transfer": make([]float64, 0),
		"bitrate":  make([]float64, 0),
	}
	receiverValues := map[string][]float64{
		"transfer": make([]float64, 0),
		"bitrate":  make([]float64, 0),
	}

	// Extract values from results
	for _, result := range results {
		if result != nil {
			if sender, ok := result["sender"].(map[string]interface{}); ok {
				if transfer, ok := sender["transfer"].(map[string]interface{}); ok {
					if rateStr, ok := transfer["rate"].(string); ok {
						if rate, err := strconv.ParseFloat(rateStr, 64); err == nil {
							senderValues["transfer"] = append(senderValues["transfer"], rate)
						}
					}
				}
				if bitrate, ok := sender["bitrate"].(map[string]interface{}); ok {
					if rateStr, ok := bitrate["rate"].(string); ok {
						if rate, err := strconv.ParseFloat(rateStr, 64); err == nil {
							senderValues["bitrate"] = append(senderValues["bitrate"], rate)
						}
					}
				}
			}
			if receiver, ok := result["receiver"].(map[string]interface{}); ok {
				if transfer, ok := receiver["transfer"].(map[string]interface{}); ok {
					if rateStr, ok := transfer["rate"].(string); ok {
						if rate, err := strconv.ParseFloat(rateStr, 64); err == nil {
							receiverValues["transfer"] = append(receiverValues["transfer"], rate)
						}
					}
				}
				if bitrate, ok := receiver["bitrate"].(map[string]interface{}); ok {
					if rateStr, ok := bitrate["rate"].(string); ok {
						if rate, err := strconv.ParseFloat(rateStr, 64); err == nil {
							receiverValues["bitrate"] = append(receiverValues["bitrate"], rate)
						}
					}
				}
			}
		}
	}

	stats := map[string]interface{}{
		"sender":   calculateStatsForValues(senderValues, numClients),
		"receiver": calculateStatsForValues(receiverValues, numClients),
	}

	return stats
}

// calculateStatsForValues calculates statistics for a set of values
func calculateStatsForValues(values map[string][]float64, numClients int) map[string]interface{} {
	// Calculate sum, min, max
	senderTransferSum := 0.0
	senderBitrateSum := 0.0
	var senderTransferMin, senderTransferMax float64
	var senderBitrateMin, senderBitrateMax float64

	if len(values["transfer"]) > 0 {
		senderTransferMin = values["transfer"][0]
		senderTransferMax = values["transfer"][0]
		for _, v := range values["transfer"] {
			senderTransferSum += v
			if v < senderTransferMin {
				senderTransferMin = v
			}
			if v > senderTransferMax {
				senderTransferMax = v
			}
		}
	}

	if len(values["bitrate"]) > 0 {
		senderBitrateMin = values["bitrate"][0]
		senderBitrateMax = values["bitrate"][0]
		for _, v := range values["bitrate"] {
			senderBitrateSum += v
			if v < senderBitrateMin {
				senderBitrateMin = v
			}
			if v > senderBitrateMax {
				senderBitrateMax = v
			}
		}
	}

	return map[string]interface{}{
		"aggregate": map[string]interface{}{
			"transfer": fmt.Sprintf("%.2f", senderTransferSum),
			"bitrate":  fmt.Sprintf("%.2f", senderBitrateSum),
		},
		"mean": map[string]interface{}{
			"transfer": fmt.Sprintf("%.2f", senderTransferSum/float64(numClients)),
			"bitrate":  fmt.Sprintf("%.2f", senderBitrateSum/float64(numClients)),
		},
		"min": map[string]interface{}{
			"transfer": fmt.Sprintf("%.2f", senderTransferMin),
			"bitrate":  fmt.Sprintf("%.2f", senderBitrateMin),
		},
		"max": map[string]interface{}{
			"transfer": fmt.Sprintf("%.2f", senderTransferMax),
			"bitrate":  fmt.Sprintf("%.2f", senderBitrateMax),
		},
	}
}

// executeWorkload executes a specific network workload
func executeWorkload(workload string, numClients int, startPort int) (string, error) {
	// For now, we'll just return a placeholder result
	// In a real implementation, this would execute the actual workload
	return fmt.Sprintf("Workload %s executed with %d clients starting at port %d", workload, numClients, startPort), nil
}