package pciebw

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/IBM/autopilot/pkg/utils"
	"k8s.io/klog/v2"
)

const (
	bandwidthTestPath = "/home/autopilot/gpu-bw/bandwidthTest"
)

// RunPCIeBWCheck executes the PCIe bandwidth test with a given threshold
func RunPCIeBWCheck(threshold int) (string, error) {
	// First check if NVIDIA GPUs are available
	gpuIDs, err := detectGPUs()
	if err != nil {
		return "ABORT", err
	}

	if len(gpuIDs) == 0 {
		klog.Info("No NVIDIA GPU detected. Skipping the bandwidth test.")
		return "SKIP", nil
	}

	klog.Infof("Detected NVIDIA GPUs: %v", gpuIDs)

	// Run the bandwidth test
	output, err := runBandwidthTest(gpuIDs)
	if err != nil {
		klog.Errorf("Error running bandwidth test: %v", err)
		return "ABORT", err
	}

	// Parse the output to extract bandwidth measurements
	bandwidths, err := parseBandwidthOutput(output)
	if err != nil {
		klog.Errorf("Error parsing bandwidth output: %v", err)
		return "ABORT", err
	}

	// Check if any bandwidth is below threshold
	failed := checkThreshold(bandwidths, threshold)

	// Format the output similar to the Python implementation
	result := "SUCCESS\n"
	result += fmt.Sprintf("Host %s\n", os.Getenv("NODE_NAME"))
	
	bwStr := ""
	for _, bw := range bandwidths {
		bwStr += fmt.Sprintf("%.2f ", bw)
	}
	result += strings.TrimSpace(bwStr) + "\n"

	if failed {
		klog.Info("PCIe BW test failed - below threshold")
		return result, fmt.Errorf("bandwidth below threshold")
	}

	klog.Info("PCIe BW test completed successfully")
	return result, nil
}

// detectGPUs uses nvidia-smi to detect available GPUs
func detectGPUs() ([]int, error) {
	cmd := exec.Command("ls", "-d", "/dev/nvidia*")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	gpuIDs := []int{}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	numre := "^[0-9]+$"

	for _, line := range lines {
		if strings.Contains(line, "nvidia") {
			deviceID := strings.TrimPrefix(line, "/dev/nvidia")
			if matched, _ := strconv.ParseBool(fmt.Sprintf("%s", deviceID)); matched || deviceID != "" {
				// Check if it's a number
				if _, err := strconv.Atoi(deviceID); err == nil {
					id, _ := strconv.Atoi(deviceID)
					gpuIDs = append(gpuIDs, id)
				}
			}
		}
	}

	return gpuIDs, nil
}

// runBandwidthTest executes the CUDA bandwidthTest binary for each GPU
func runBandwidthTest(gpuIDs []int) (string, error) {
	var fullOutput string

	for _, gpuID := range gpuIDs {
		cmd := exec.Command(bandwidthTestPath, "--htod", "--memory=pinned", fmt.Sprintf("--device=%d", gpuID), "--csv")
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Check if it's a critical error
			outputStr := string(output)
			if strings.Contains(strings.ToLower(outputStr), "error") || strings.Contains(outputStr, "802") {
				return "", fmt.Errorf("critical error with GPU %d: %s", gpuID, outputStr)
			}
		}
		fullOutput += string(output) + "\n"
	}

	return fullOutput, nil
}

// parseBandwidthOutput parses the output from bandwidthTest to extract measurements
func parseBandwidthOutput(output string) ([]float64, error) {
	bandwidths := []float64{}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Bandwidth =") {
			parts := strings.Split(line, "= ")
			if len(parts) >= 2 {
				bwPart := strings.Split(parts[1], " GB/s")
				if len(bwPart) >= 1 {
					bw, err := strconv.ParseFloat(strings.TrimSpace(bwPart[0]), 64)
					if err != nil {
						return nil, fmt.Errorf("error parsing bandwidth value: %v", err)
					}
					bandwidths = append(bandwidths, bw)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning output: %v", err)
	}

	return bandwidths, nil
}

// checkThreshold compares bandwidth measurements against the threshold
func checkThreshold(bandwidths []float64, threshold int) bool {
	for _, bw := range bandwidths {
		if bw < float64(threshold) {
			return true
		}
	}
	return false
}