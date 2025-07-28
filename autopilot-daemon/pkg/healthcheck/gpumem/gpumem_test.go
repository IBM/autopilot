package gpumem

import (
	"testing"
)

// TestParseGPUMemOutput tests the parseGPUMemOutput function
func TestParseGPUMemOutput(t *testing.T) {
	// Test case 1: No errors (contains "NONE")
	output := ` GPU H2D(p)  H2D   D2H(p)  D2H   daxpy  dgemm   temp     power     smMHz
 0   6.25   6.25   6.25   6.25    800.00   50.00      0        0        0
Summary of GPU errors: NONE `
	result, err := parseGPUMemOutput(output)
	if err != nil {
		t.Errorf("parseGPUMemOutput returned error: %v", err)
	}

	if result != true {
		t.Errorf("Expected parseGPUMemOutput to return true, got %t", result)
	}

	// Test case 2: Has errors (does not contain "NONE" but contains "GPU errors")
	output = ` GPU H2D(p)  H2D   D2H(p)  D2H   daxpy  dgemm   temp     power     smMHz
 0   6.25   6.25   6.25   6.25    800.00   50.00      0        0        0
Summary of GPU errors:GPU 0 -- H2D(p): 6.25; daxpy: 800.00; dgemm: 50.00`
	result, err = parseGPUMemOutput(output)
	if err != nil {
		t.Errorf("parseGPUMemOutput returned error: %v", err)
	}

	if result != false {
		t.Errorf("Expected parseGPUMemOutput to return false, got %t", result)
	}

	// Test case 3: Empty output (should return error)
	output = ""
	result, err = parseGPUMemOutput(output)
	if err == nil {
		t.Errorf("Expected parseGPUMemOutput to return error for empty output, but got no error")
	}

	if result != false {
		t.Errorf("Expected parseGPUMemOutput to return false for empty output, got %t", result)
	}
}