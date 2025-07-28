package gpumem

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"k8s.io/klog/v2"
)

const (
	gpucheckPath = "/home/autopilot/gpu-mem/gpucheck"
)

// RunGPUMemCheck executes the GPU memory check
func RunGPUMemCheck() (string, error) {
	// Check if the gpucheck binary exists
	if _, err := os.Stat(gpucheckPath); os.IsNotExist(err) {
		klog.Info("GPU Memory check binary not found. Skipping the memory check.")
		return "SKIP", nil
	}

	// Run the gpucheck binary
	cmd := exec.Command(gpucheckPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		klog.Errorf("Error running GPU memory check: %v", err)
		return "ABORT", err
	}

	// Parse the output to determine pass/fail
	result, err := parseGPUMemOutput(string(output))
	if err != nil {
		klog.Errorf("Error parsing GPU memory check output: %v", err)
		return "ABORT", err
	}

	// Format the output similar to the Python implementation
	formattedOutput := string(output)
	if result {
		// Success case - no errors
		formattedOutput = "[[ GPU-MEM ]] Health Check successful\n" + formattedOutput
	} else {
		// Failure case - errors found
		formattedOutput = "[[ GPU-MEM ]] Health Check unsuccessful. FAIL.\n" + formattedOutput
	}

	klog.Info("GPU Memory check completed successfully")
	return formattedOutput, nil
}

// parseGPUMemOutput parses the output from gpucheck to determine pass/fail
func parseGPUMemOutput(output string) (bool, error) {
	// Check if "NONE" is in the output, which indicates no errors
	if strings.Contains(output, "NONE") {
		return true, nil // Pass
	}

	// Check if there are any GPU errors reported
	if strings.Contains(output, "GPU errors") && !strings.Contains(output, "NONE") {
		return false, nil // Fail
	}

	// If we can't determine the result from the output, return an error
	return false, fmt.Errorf("unable to determine GPU memory check result from output")
}