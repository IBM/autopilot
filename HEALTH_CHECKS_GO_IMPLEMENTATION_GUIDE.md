# Health Checks Go Implementation Guide

## Overview

This guide provides detailed instructions for implementing the Go version of the health check functionality that currently exists as Python scripts. It includes step-by-step instructions, code patterns, and best practices to ensure consistency and quality.

## Prerequisites

1. Go 1.21 or later
2. Access to Kubernetes cluster for testing
3. Container images with required binaries:
   - CUDA `bandwidthTest`
   - CUDA `gpucheck`
   - `dcgmi`
   - `iperf3`
   - `nvidia-smi`
   - `ping`

## Implementation Steps

### 1. Create Package Structure

Create the following directory structure:
```
autopilot-daemon/pkg/healthcheck/
├── pciebw/
├── remapped/
├── gpumem/
├── dcgm/
├── ping/
├── gpupower/
├── pvc/
└── iperf/
```

### 2. Implement PCIe Bandwidth Check

Create `autopilot-daemon/pkg/healthcheck/pciebw/pciebw.go`:

```go
package pciebw

import (
    "fmt"
    "os/exec"
    "strconv"
    "strings"
    
    "github.com/IBM/autopilot/pkg/utils"
    "k8s.io/klog/v2"
)

// RunPCIeBWCheck executes the PCIe bandwidth test with a given threshold
func RunPCIeBWCheck(threshold int) (string, error) {
    // Run briefings check first
    if err := runBriefingsCheck(); err != nil {
        return fmt.Sprintf("[[ PCIEBW ]] ABORT\n%s", err.Error()), nil
    }
    
    klog.Info("[[ PCIEBW ]] Briefings completed. Continue with PCIe Bandwidth evaluation.")
    
    // Execute bandwidth test
    cmd := exec.Command("/home/autopilot/gpu-bw/bandwidthTest", "--htod", "--memory=pinned", "--csv")
    output, err := cmd.CombinedOutput()
    if err != nil {
        klog.Error("Error executing bandwidthTest: ", err.Error())
        return "", err
    }
    
    // Parse output
    bandwidths, parseErr := parseBandwidthOutput(string(output))
    if parseErr != nil {
        return fmt.Sprintf("[[ PCIEBW ]] ABORT\n%s", parseErr.Error()), nil
    }
    
    // Check threshold
    if !checkThreshold(bandwidths, threshold) {
        result := "FAIL\nHost " + utils.NodeName + "\n"
        for _, bw := range bandwidths {
            result += fmt.Sprintf("%.2f ", bw)
        }
        return result, nil
    }
    
    result := "SUCCESS\nHost " + utils.NodeName + "\n"
    for _, bw := range bandwidths {
        result += fmt.Sprintf("%.2f ", bw)
    }
    return result, nil
}

// runBriefingsCheck executes the briefings.sh script to check prerequisites
func runBriefingsCheck() error {
    cmd := exec.Command("bash", "./utils/briefings.sh")
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("briefings check failed: %s", string(output))
    }
    
    if strings.Contains(string(output), "ABORT") {
        return fmt.Errorf("briefings check aborted: %s", string(output))
    }
    
    return nil
}

// parseBandwidthOutput parses the output from bandwidthTest to extract measurements
func parseBandwidthOutput(output string) ([]float64, error) {
    lines := strings.Split(output, "\n")
    var bandwidths []float64
    
    for _, line := range lines {
        if strings.Contains(line, "Bandwidth") {
            parts := strings.Split(line, "= ")
            if len(parts) >= 2 {
                bwStr := strings.TrimSuffix(parts[1], " GB/s")
                bw, err := strconv.ParseFloat(bwStr, 64)
                if err != nil {
                    return nil, fmt.Errorf("failed to parse bandwidth value: %s", bwStr)
                }
                bandwidths = append(bandwidths, bw)
            }
        }
    }
    
    return bandwidths, nil
}

// checkThreshold compares bandwidth measurements against the threshold
func checkThreshold(bandwidths []float64, threshold int) bool {
    for _, bw := range bandwidths {
        if bw < float64(threshold) {
            return false
        }
    }
    return true
}
```

### 3. Implement GPU Remapped Rows Check

Create `autopilot-daemon/pkg/healthcheck/remapped/remapped.go`:

```go
package remapped

import (
    "fmt"
    "os/exec"
    "strconv"
    "strings"
    
    "github.com/IBM/autopilot/pkg/utils"
    "k8s.io/klog/v2"
)

// RunRemappedRowsCheck executes the remapped rows check
func RunRemappedRowsCheck() (string, error) {
    // Run briefings check first
    if err := runBriefingsCheck(); err != nil {
        return fmt.Sprintf("[[ REMAPPED ROWS ]] ABORT\n%s", err.Error()), nil
    }
    
    klog.Info("[[ REMAPPED ROWS ]] Briefings completed. Continue with remapped rows evaluation.")
    
    // Execute remapped rows check
    cmd := exec.Command("bash", "./gpu-remapped/remapped-rows.sh")
    output, err := cmd.CombinedOutput()
    if err != nil {
        klog.Error("Error executing remapped-rows.sh: ", err.Error())
        return "", err
    }
    
    resultStr := string(output)
    if strings.Contains(resultStr, "FAIL") {
        klog.Info("[[ REMAPPED ROWS ]] FAIL")
        klog.Info("Host ", utils.NodeName)
        klog.Info(strings.TrimSpace(resultStr))
        return fmt.Sprintf("[[ REMAPPED ROWS ]] FAIL\nHost %s\n%s", utils.NodeName, strings.TrimSpace(resultStr)), nil
    } else if strings.Contains(resultStr, "SKIP") {
        klog.Info("[[ REMAPPED ROWS ]] SKIP")
        return fmt.Sprintf("[[ REMAPPED ROWS ]] SKIP\nHost %s\n%s", utils.NodeName, strings.TrimSpace(resultStr)), nil
    } else {
        klog.Info("[[ REMAPPED ROWS ]] SUCCESS")
        klog.Info("Host ", utils.NodeName)
        klog.Info(strings.TrimSpace(resultStr))
        return fmt.Sprintf("[[ REMAPPED ROWS ]] SUCCESS\nHost %s\n%s", utils.NodeName, strings.TrimSpace(resultStr)), nil
    }
}

// runBriefingsCheck executes the briefings.sh script to check prerequisites
func runBriefingsCheck() error {
    cmd := exec.Command("bash", "./utils/briefings.sh")
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("briefings check failed: %s", string(output))
    }
    
    if strings.Contains(string(output), "ABORT") {
        return fmt.Errorf("briefings check aborted: %s", string(output))
    }
    
    return nil
}
```

### 4. Common Implementation Patterns

#### Command Execution Pattern
```go
cmd := exec.Command("nvidia-smi", "--query-gpu=name", "--format=csv")
output, err := cmd.CombinedOutput()
if err != nil {
    klog.Error("Error executing command: ", err.Error())
    return "", err
}
```

#### Output Parsing Pattern
```go
func parseJSONOutput(output string) (map[string]interface{}, error) {
    var result map[string]interface{}
    err := json.Unmarshal([]byte(output), &result)
    if err != nil {
        return nil, fmt.Errorf("failed to parse JSON output: %s", err.Error())
    }
    return result, nil
}
```

#### Error Handling Pattern
```go
if err != nil {
    klog.Error("Error description: ", err.Error())
    // Return appropriate error or continue with fallback
    return "", err
}
```

#### Metrics Collection Pattern
```go
utils.HchecksGauge.WithLabelValues("pciebw", utils.NodeName, utils.CPUModel, utils.GPUModel, "0").Set(bw)
```

#### Node Labeling Pattern
```go
err := utils.PatchNode(utils.GPUHealthWarnLabel, utils.NodeName, false)
if err != nil {
    klog.Error("Failed to patch node: ", err.Error())
}
```

## Testing Guidelines

### Unit Testing Template
```go
func TestRunPCIeBWCheck(t *testing.T) {
    // Setup mock environment
    
    // Execute function
    result, err := RunPCIeBWCheck(4)
    
    // Verify results
    assert.NoError(t, err)
    assert.Contains(t, result, "SUCCESS")
    
    // Cleanup
}
```

### Integration Testing
1. Test each health check individually
2. Verify output format matches Python implementation
3. Confirm metrics are collected correctly
4. Validate node labeling works as expected

## Code Review Checklist

- [ ] Follows existing code style and patterns
- [ ] Proper error handling with logging
- [ ] Correct metrics collection
- [ ] Node labeling integration
- [ ] Output format matches Python implementation
- [ ] Unit tests cover main functionality
- [ ] Error cases handled appropriately
- [ ] No hardcoded paths or values
- [ ] Proper use of Kubernetes client where needed

## Deployment Considerations

### Feature Flags
Use environment variables to enable/disable new implementations:
```go
if os.Getenv("USE_GO_HEALTHCHECKS") == "true" {
    // Use Go implementation
} else {
    // Fall back to Python
}
```

### Rollback Strategy
Maintain Python scripts as fallback option during transition period.

## Monitoring and Validation

### Key Metrics to Monitor
1. Execution time compared to Python implementation
2. Error rates and failure patterns
3. Resource utilization (CPU, memory)
4. Compatibility with existing monitoring systems

### Validation Process
1. Compare output format with Python implementations
2. Verify metrics collection matches existing patterns
3. Confirm node labeling behavior is identical
4. Test in staging environment before production rollout

## Troubleshooting Guide

### Common Issues
1. **Binary Not Found**: Ensure required binaries are in the container
2. **Permission Denied**: Check container user permissions
3. **Kubernetes API Errors**: Verify RBAC permissions
4. **Output Format Mismatches**: Compare with Python implementation

### Debugging Tips
1. Enable verbose logging with `--v=4` flag
2. Use `klog.Info` statements to trace execution flow
3. Test commands manually in container
4. Compare with Python script behavior

## Next Steps

1. Implement the health checks following this guide
2. Write comprehensive unit tests
3. Test in development environment
4. Validate against Python implementation outputs
5. Prepare for production deployment