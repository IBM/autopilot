package remapped

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"k8s.io/klog/v2"
)

// RunRemappedRowsCheck executes the remapped rows check
func RunRemappedRowsCheck() (string, error) {
	// First check if NVIDIA GPUs are available
	gpuIDs, err := detectGPUs()
	if err != nil {
		return "ABORT", err
	}

	if len(gpuIDs) == 0 {
		klog.Info("No NVIDIA GPU detected. Skipping the Remapped Rows check.")
		return "SKIP", nil
	}

	klog.Infof("Detected NVIDIA GPUs: %v", gpuIDs)

	// Check for remapped rows on each GPU
	result, err := checkRemappedRows(gpuIDs)
	if err != nil {
		klog.Errorf("Error checking remapped rows: %v", err)
		return "ABORT", err
	}

	// Format the output similar to the Python implementation
	// Return just the result without the host name, as the main healthcheck
	// function handles the host name formatting
	klog.Info("Remapped Rows check completed successfully")
	return result, nil
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

// checkRemappedRows uses nvidia-smi to check for GPU memory remapped rows
func checkRemappedRows(gpuIDs []int) (string, error) {
	result := ""
	fail := false

	for _, gpuID := range gpuIDs {
		// Run nvidia-smi to get remapped rows information
		cmd := exec.Command("nvidia-smi", "-q", "-i", strconv.Itoa(gpuID))
		output, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("error running nvidia-smi for GPU %d: %v", gpuID, err)
		}

		// Parse output to check for remapped rows
		hasRemappedRows, err := parseRemappedRowsOutput(string(output))
		if err != nil {
			return "", fmt.Errorf("error parsing remapped rows output for GPU %d: %v", gpuID, err)
		}

		if hasRemappedRows {
			result += "1 "
			fail = true
		} else {
			result += "0 "
		}
	}

	// Trim trailing space
	result = strings.TrimSpace(result)

	if fail {
		return "FAIL\n" + result, nil
	}

	return result, nil
}

// parseRemappedRowsOutput parses nvidia-smi output to check for remapped rows
func parseRemappedRowsOutput(output string) (bool, error) {
	// Look for "Remapped Rows" section in the output
	lines := strings.Split(output, "\n")
	inRemappedRowsSection := false

	for _, line := range lines {
		// Check if we're entering the Remapped Rows section
		if strings.Contains(line, "Remapped Rows") {
			inRemappedRowsSection = true
			continue
		}

		// If we're in the Remapped Rows section, look for "Pending : Yes"
		if inRemappedRowsSection {
			if strings.Contains(line, "Pending") {
				// Extract the value after "Pending :"
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					value := strings.TrimSpace(parts[1])
					if value == "Yes" {
						return true, nil
					}
				}
			}

			// If we've moved past the Remapped Rows section, stop checking
			if strings.Contains(line, "Temperature") || strings.Contains(line, "Performance State") {
				inRemappedRowsSection = false
			}
		}
	}

	return false, nil
}