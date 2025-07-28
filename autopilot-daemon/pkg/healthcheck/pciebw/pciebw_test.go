package pciebw

import (
	"testing"
)

// TestParseBandwidthOutput tests the parseBandwidthOutput function
func TestParseBandwidthOutput(t *testing.T) {
	// Test case 1: Valid output with bandwidth values
	output := `Device 0: NVIDIA A100-SXM4-40GB
 HostToDevice bandwidth (GB/s): 
 Pinned memory : 25.00
 HostToDevice bandwidth (GB/s): 
 Pinned memory : 25.00
 Bandwidth = 25.00 GB/s
Device 1: NVIDIA A100-SXM4-40GB
 HostToDevice bandwidth (GB/s): 
 Pinned memory : 24.50
 HostToDevice bandwidth (GB/s): 
 Pinned memory : 24.50
 Bandwidth = 24.50 GB/s`

	bandwidths, err := parseBandwidthOutput(output)
	if err != nil {
		t.Errorf("parseBandwidthOutput returned error: %v", err)
	}

	if len(bandwidths) != 2 {
		t.Errorf("Expected 2 bandwidth values, got %d", len(bandwidths))
	}

	if bandwidths[0] != 25.00 {
		t.Errorf("Expected first bandwidth to be 25.00, got %f", bandwidths[0])
	}

	if bandwidths[1] != 24.50 {
		t.Errorf("Expected second bandwidth to be 24.50, got %f", bandwidths[1])
	}

	// Test case 2: Empty output
	output = ""
	bandwidths, err = parseBandwidthOutput(output)
	if err != nil {
		t.Errorf("parseBandwidthOutput returned error for empty output: %v", err)
	}

	if len(bandwidths) != 0 {
		t.Errorf("Expected 0 bandwidth values for empty output, got %d", len(bandwidths))
	}
}

// TestCheckThreshold tests the checkThreshold function
func TestCheckThreshold(t *testing.T) {
	// Test case 1: All values above threshold
	bandwidths := []float64{25.0, 24.5, 26.0}
	threshold := 20
	result := checkThreshold(bandwidths, threshold)
	if result != false {
		t.Errorf("Expected checkThreshold to return false, got %t", result)
	}

	// Test case 2: One value below threshold
	bandwidths = []float64{25.0, 15.0, 26.0}
	threshold = 20
	result = checkThreshold(bandwidths, threshold)
	if result != true {
		t.Errorf("Expected checkThreshold to return true, got %t", result)
	}

	// Test case 3: All values below threshold
	bandwidths = []float64{15.0, 14.5, 16.0}
	threshold = 20
	result = checkThreshold(bandwidths, threshold)
	if result != true {
		t.Errorf("Expected checkThreshold to return true, got %t", result)
	}

	// Test case 4: Empty slice
	bandwidths = []float64{}
	threshold = 20
	result = checkThreshold(bandwidths, threshold)
	if result != false {
		t.Errorf("Expected checkThreshold to return false for empty slice, got %t", result)
	}
}