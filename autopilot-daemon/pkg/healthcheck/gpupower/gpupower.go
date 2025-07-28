package gpupower

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"k8s.io/klog/v2"
)

// RunGPUPowerCheck executes the GPU power check
func RunGPUPowerCheck() (string, error) {
	// First check if NVIDIA GPUs are available
	gpuIDs, err := detectGPUs()
	if err != nil {
		return "ABORT", err
	}

	if len(gpuIDs) == 0 {
		klog.Info("[GPU POWER] No NVIDIA GPU detected. Skipping the Power Throttle check.")
		return "ABORT", nil
	}

	klog.Infof("[GPU POWER] Detected NVIDIA GPUs: %v", gpuIDs)

	// Check for power throttling on each GPU
	result, err := checkPowerThrottling(gpuIDs)
	if err != nil {
		klog.Errorf("Error checking power throttling: %v", err)
		return "ABORT", err
	}

	// Format the output similar to the shell script
	output := ""
	fail := false
	for _, r := range result {
		if r == 1 {
			fail = true
		}
		output += fmt.Sprintf("%d ", r)
	}
	output = strings.TrimSpace(output)

	if fail {
		output = "[GPU POWER] FAIL\n" + output
	} else {
		output = "[GPU POWER] SUCCESS\n" + output
	}

	klog.Info("GPU Power check completed successfully")
	return output, nil
}

// detectGPUs uses ls command to detect available GPUs
func detectGPUs() ([]int, error) {
	cmd := exec.Command("ls", "-d", "/dev/nvidia*")
	output, err := cmd.Output()
	if err != nil {
		// No NVIDIA GPUs found
		return []int{}, nil
	}

	gpuIDs := []int{}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if strings.Contains(line, "nvidia") {
			deviceID := strings.TrimPrefix(line, "/dev/nvidia")
			if deviceID != "" {
				// Check if it's a number
				if id, err := strconv.Atoi(deviceID); err == nil {
					gpuIDs = append(gpuIDs, id)
				}
			}
		}
	}

	return gpuIDs, nil
}

// checkPowerThrottling uses nvidia-smi to check for GPU power throttling
func checkPowerThrottling(gpuIDs []int) ([]int, error) {
	result := make([]int, len(gpuIDs))

	for i, gpuID := range gpuIDs {
		// Run nvidia-smi to get power throttling information
		cmd := exec.Command("nvidia-smi", "--format=csv", "-i", strconv.Itoa(gpuID), "--query-gpu=clocks_event_reasons.hw_slowdown")
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("error running nvidia-smi for GPU %d: %v", gpuID, err)
		}

		// Parse output to check for power throttling
		scanner := bufio.NewScanner(strings.NewReader(string(output)))
		scanner.Scan() // Skip header line
		if scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "Not Active") {
				result[i] = 0 // No power throttling
			} else {
				result[i] = 1 // Power throttling detected
			}
		}
	}

	return result, nil
}