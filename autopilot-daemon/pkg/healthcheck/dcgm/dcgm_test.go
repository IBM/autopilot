package dcgm

import (
	"testing"
)

// TestUnifyStringFormat tests the unifyStringFormat function
func TestUnifyStringFormat(t *testing.T) {
	// Test case 1: Normal string
	input := "GPU Memory"
	expected := "gpu_memory"
	result := unifyStringFormat(input)
	if result != expected {
		t.Errorf("Expected unifyStringFormat(%s) to be %s, got %s", input, expected, result)
	}

	// Test case 2: String with spaces and slashes
	input = "PCIe / NVLink"
	expected = "pcie_/_nvlink"
	result = unifyStringFormat(input)
	if result != expected {
		t.Errorf("Expected unifyStringFormat(%s) to be %s, got %s", input, expected, result)
	}

	// Test case 3: String with mixed case
	input = "Hardware Test"
	expected = "hardware_test"
	result = unifyStringFormat(input)
	if result != expected {
		t.Errorf("Expected unifyStringFormat(%s) to be %s, got %s", input, expected, result)
	}
}

// TestParseDCGMJSON tests the parseDCGMJSON function
func TestParseDCGMJSON(t *testing.T) {
	// Test case 1: Successful test with no failures
	jsonOutput := `{
		"DCGM GPU Diagnostic": {
			"test_categories": [
				{
					"category": "Hardware",
					"tests": [
						{
							"name": "GPU Memory",
							"results": [
								{
									"status": "Pass",
									"gpu_id": 0
								}
							]
						}
					]
				}
			]
		}
	}`

	success, output, err := parseDCGMJSON(jsonOutput)
	if err != nil {
		t.Errorf("parseDCGMJSON returned error: %v", err)
	}

	if !success {
		t.Errorf("Expected parseDCGMJSON to return success=true, got %t", success)
	}

	if output != "" {
		t.Errorf("Expected parseDCGMJSON to return empty output, got %s", output)
	}

	// Test case 2: Failed test
	jsonOutput = `{
		"DCGM GPU Diagnostic": {
			"test_categories": [
				{
					"category": "Hardware",
					"tests": [
						{
							"name": "GPU Memory",
							"results": [
								{
									"status": "Fail",
									"gpu_id": 0
								}
							]
						}
					]
				}
			]
		}
	}`

	success, output, err = parseDCGMJSON(jsonOutput)
	if err != nil {
		t.Errorf("parseDCGMJSON returned error: %v", err)
	}

	if success {
		t.Errorf("Expected parseDCGMJSON to return success=false, got %t", success)
	}

	if output != "gpu_memory.0" {
		t.Errorf("Expected parseDCGMJSON to return 'gpu_memory.0', got %s", output)
	}

	// Test case 3: Multiple failed tests
	jsonOutput = `{
		"DCGM GPU Diagnostic": {
			"test_categories": [
				{
					"category": "Hardware",
					"tests": [
						{
							"name": "GPU Memory",
							"results": [
								{
									"status": "Fail",
									"gpu_id": 0
								}
							]
						},
						{
							"name": "PCIe",
							"results": [
								{
									"status": "Fail",
									"gpu_id": 1
								}
							]
						}
					]
				}
			]
		}
	}`

	success, output, err = parseDCGMJSON(jsonOutput)
	if err != nil {
		t.Errorf("parseDCGMJSON returned error: %v", err)
	}

	if success {
		t.Errorf("Expected parseDCGMJSON to return success=false, got %t", success)
	}

	expectedOutput := "gpu_memory.0;pcie.1"
	if output != expectedOutput {
		t.Errorf("Expected parseDCGMJSON to return '%s', got %s", expectedOutput, output)
	}
}