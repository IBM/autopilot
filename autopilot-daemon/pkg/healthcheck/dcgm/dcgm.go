package dcgm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/IBM/autopilot/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

// RunDCGMCheck executes the DCGM diagnostics check
func RunDCGMCheck(level string) (string, error) {
	// Run the dcgmi diag command
	cmd := exec.Command("dcgmi", "diag", "-j", "-r", level)
	output, err := cmd.CombinedOutput()
	if err != nil {
		klog.Errorf("Error running DCGM diagnostics: %v", err)
		return "ABORT", err
	}

	// Parse the JSON output to check for failures
	success, outputStr, err := parseDCGMJSON(string(output))
	if err != nil {
		klog.Errorf("Error parsing DCGM JSON output: %v", err)
		return "ABORT", err
	}

	// Format the output similar to the Python implementation
	result := ""
	if success {
		result = "[[ DCGM ]] SUCCESS\n"
	} else {
		result = "Host " + os.Getenv("NODE_NAME") + "\n"
		result += "[[ DCGM ]] FAIL\n"
		result += outputStr + "\n"
	}

	// Patch the node based on the results
	err = patchNode(success, outputStr, level)
	if err != nil {
		klog.Errorf("Error patching node: %v", err)
		// Don't return error here as the main check was successful
	}

	klog.Info("DCGM diagnostics check completed successfully")
	return result, nil
}

// parseDCGMJSON parses the JSON output from dcgmi
func parseDCGMJSON(output string) (bool, string, error) {
	// Parse the JSON output
	var dcgmData map[string]interface{}
	err := json.Unmarshal([]byte(output), &dcgmData)
	if err != nil {
		return false, "", fmt.Errorf("error parsing JSON: %v", err)
	}

	// Navigate to the test categories
	diagData, ok := dcgmData["DCGM GPU Diagnostic"].(map[string]interface{})
	if !ok {
		return false, "", fmt.Errorf("could not find DCGM GPU Diagnostic data")
	}

	categories, ok := diagData["test_categories"].([]interface{})
	if !ok {
		return false, "", fmt.Errorf("could not find test_categories")
	}

	success := true
	outputStr := ""

	// Process each category and test
	for _, category := range categories {
		categoryMap, ok := category.(map[string]interface{})
		if !ok {
			continue
		}

		tests, ok := categoryMap["tests"].([]interface{})
		if !ok {
			continue
		}

		for _, test := range tests {
			testMap, ok := test.(map[string]interface{})
			if !ok {
				continue
			}

			testName, ok := testMap["name"].(string)
			if !ok {
				continue
			}

			results, ok := testMap["results"].([]interface{})
			if !ok {
				continue
			}

			testFailing := false
			for _, result := range results {
				resultMap, ok := result.(map[string]interface{})
				if !ok {
					continue
				}

				status, ok := resultMap["status"].(string)
				if !ok {
					continue
				}

				if strings.ToLower(status) == "fail" {
					success = false
					if !testFailing {
						if outputStr != "" {
							outputStr += ";"
						}
						outputStr += unifyStringFormat(testName)
						testFailing = true
					}

					// Add GPU ID if available
					if gpuID, hasGPUID := resultMap["gpu_id"]; hasGPUID {
						outputStr += "." + fmt.Sprintf("%v", gpuID)
					} else {
						outputStr += ".NoGPUid"
					}
				}
			}
		}
	}

	return success, outputStr, nil
}

// unifyStringFormat translates key-strings into lowercase and strip spaces
func unifyStringFormat(key string) string {
	// Convert to lowercase and trim spaces
	toLower := strings.ToLower(strings.TrimSpace(key))
	// Replace '/' and spaces with '_'
	re := regexp.MustCompile(`[/\s]+`)
	res := re.ReplaceAllString(toLower, "_")
	return res
}

// patchNode updates node labels based on DCGM results
func patchNode(success bool, output string, level string) error {
	// Get Kubernetes client
	cset := utils.GetClientsetInstance()
	
	// Get current node name
	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		return fmt.Errorf("NODE_NAME environment variable not set")
	}
	
	// Get the current node
	node, err := cset.Cset.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting node: %v", err)
	}
	
	// Create a copy of the node labels
	labels := node.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	
	// Update labels based on success/failure
	if success {
		labels["autopilot.ibm.com/dcgm.level."+level] = "PASS"
	} else {
		labels["autopilot.ibm.com/dcgm.level."+level] = "FAIL"
	}
	
	// Create patch data
	patchData := fmt.Sprintf(`{"metadata":{"labels":%s}}`, func() string {
		data, _ := json.Marshal(labels)
		return string(data)
	}())
	
	// Patch the node
	_, err = cset.Cset.CoreV1().Nodes().Patch(context.TODO(), nodeName, types.MergePatchType, []byte(patchData), metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("error patching node: %v", err)
	}
	
	klog.Infof("Successfully patched node %s with dcgm.level.%s=%s", nodeName, level, func() string {
		if success {
			return "PASS"
		}
		return "FAIL"
	}())
	
	return nil
}