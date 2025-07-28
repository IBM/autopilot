# Go Health Check Implementation Plan

## Overview

This document provides detailed implementation plans for each health check that needs to be reimplemented in Go. Each section describes the specific implementation approach, key functions, and integration points.

## 1. PCIe Bandwidth Check (pciebw)

### Implementation Approach
Create a Go package that directly executes the CUDA bandwidthTest binary and parses its output.

### Key Functions
```go
// RunPCIeBWCheck executes the PCIe bandwidth test with a given threshold
func RunPCIeBWCheck(threshold int) (string, error)

// parseBandwidthOutput parses the output from bandwidthTest to extract measurements
func parseBandwidthOutput(output string) ([]float64, error)

// checkThreshold compares bandwidth measurements against the threshold
func checkThreshold(bandwidths []float64, threshold int) bool
```

### Integration Points
- Called from `handler.PCIeBWHandler()` in `autopilot-daemon/pkg/handler/handler.go`
- Updates `healthcheck.HealthCheckStatus[PCIeBW]` in `autopilot-daemon/pkg/healthcheck/global.go`
- Exports metrics via `utils.HchecksGauge` in `autopilot-daemon/pkg/utils/prometheus.go`

### Dependencies
- CUDA `bandwidthTest` binary must be available at `/home/autopilot/gpu-bw/bandwidthTest`
- `nvidia-smi` for GPU detection

## 2. GPU Remapped Rows Check (remapped)

### Implementation Approach
Create a Go package that uses `nvidia-smi` to check for GPU memory remapped rows.

### Key Functions
```go
// RunRemappedRowsCheck executes the remapped rows check
func RunRemappedRowsCheck() (string, error)

// parseRemappedRowsOutput parses nvidia-smi output to check for remapped rows
func parseRemappedRowsOutput(output string) (map[int]bool, error)

// detectGPUs uses nvidia-smi to detect available GPUs
func detectGPUs() ([]int, error)
```

### Integration Points
- Called from `handler.RemappedRowsHandler()` in `autopilot-daemon/pkg/handler/handler.go`
- Updates `healthcheck.HealthCheckStatus[RowRemap]` in `autopilot-daemon/pkg/healthcheck/global.go`
- Exports metrics via `utils.HchecksGauge` in `autopilot-daemon/pkg/utils/prometheus.go`

### Dependencies
- `nvidia-smi` must be available in the container

## 3. GPU Memory Check (gpumem)

### Implementation Approach
Create a Go package that executes the gpucheck binary and parses its output.

### Key Functions
```go
// RunGPUMemCheck executes the GPU memory check
func RunGPUMemCheck() (string, error)

// parseGPUMemOutput parses the output from gpucheck to determine pass/fail
func parseGPUMemOutput(output string) (bool, error)
```

### Integration Points
- Called from `handler.GpuMemHandler()` in `autopilot-daemon/pkg/handler/handler.go`
- Updates `healthcheck.HealthCheckStatus[GPUMem]` in `autopilot-daemon/pkg/healthcheck/global.go`
- Exports metrics via `utils.HchecksGauge` in `autopilot-daemon/pkg/utils/prometheus.go`

### Dependencies
- CUDA `gpucheck` binary must be available at `/home/autopilot/gpu-mem/gpucheck`

## 4. DCGM Diagnostics Check (dcgm)

### Implementation Approach
Create a Go package that uses `dcgmi` CLI tool to run diagnostics and parse JSON output.

### Key Functions
```go
// RunDCGMCheck executes the DCGM diagnostics check
func RunDCGMCheck(level string) (string, error)

// parseDCGMJSON parses the JSON output from dcgmi
func parseDCGMJSON(output string) (bool, string, error)

// patchNode updates node labels based on DCGM results
func patchNode(success bool, output string, level string) error
```

### Integration Points
- Called from `handler.DCGMHandler()` in `autopilot-daemon/pkg/handler/handler.go`
- Updates `healthcheck.HealthCheckStatus[DCGM]` in `autopilot-daemon/pkg/healthcheck/global.go`
- Exports metrics via `utils.HchecksGauge` in `autopilot-daemon/pkg/utils/prometheus.go`
- Uses Kubernetes client from `utils.GetClientsetInstance()` in `autopilot-daemon/pkg/utils/functions.go`

### Dependencies
- `dcgmi` must be available in the container
- Kubernetes client already available

## 5. Network Ping Check (ping)

### Implementation Approach
Create a Go package that uses the Kubernetes API to discover nodes and network interfaces, then executes ping commands.

### Key Functions
```go
// RunPingCheck executes the ping network check
func RunPingCheck(nodes []string, job string, nodelabel string) (string, error)

// discoverNodes discovers nodes based on job or node label
func discoverNodes(job string, nodelabel string) (map[string]string, error)

// getNetworkInterfaces gets network interfaces for each node
func getNetworkInterfaces() (map[string]map[string][]string, error)

// executePing executes ping commands to test connectivity
func executePing(ip string) (bool, error)
```

### Integration Points
- Called from `handler.PingHandler()` in `autopilot-daemon/pkg/handler/handler.go`
- Updates `healthcheck.HealthCheckStatus[Ping]` in `autopilot-daemon/pkg/healthcheck/global.go`
- Exports metrics via `utils.HchecksGauge` in `autopilot-daemon/pkg/utils/prometheus.go`
- Uses Kubernetes client from `utils.GetClientsetInstance()` in `autopilot-daemon/pkg/utils/functions.go`

### Dependencies
- `ping` command must be available in the container
- Kubernetes client already available

## 6. GPU Power Check (gpupower)

### Implementation Approach
Create a Go package that uses `nvidia-smi` to check for GPU power throttling.

### Key Functions
```go
// RunGPUPowerCheck executes the GPU power check
func RunGPUPowerCheck() (string, error)

// parsePowerOutput parses nvidia-smi output to check for power throttling
func parsePowerOutput(output string) ([]float64, error)

// detectGPUs uses nvidia-smi to detect available GPUs
func detectGPUs() ([]int, error)
```

### Integration Points
- Called from `handler.GpuPowerHandler()` in `autopilot-daemon/pkg/handler/handler.go`
- Updates `healthcheck.HealthCheckStatus[GPUPower]` in `autopilot-daemon/pkg/healthcheck/global.go`
- Exports metrics via `utils.HchecksGauge` in `autopilot-daemon/pkg/utils/prometheus.go`

### Dependencies
- `nvidia-smi` must be available in the container

## 7. PVC Create/Delete Check (pvc)

### Implementation Approach
Enhance the existing implementation in the healthcheck package to use the Kubernetes API directly.

### Key Functions
```go
// RunPVCCheck executes the PVC create/delete check
func RunPVCCheck() (string, error)

// createPVC creates a PVC using the Kubernetes API
func createPVC() error

// deletePVC deletes a PVC using the Kubernetes API
func deletePVC(name string) error

// listPVC checks the status of a PVC
func listPVC(name string) (string, error)
```

### Integration Points
- Called from `handler.PVCHandler()` in `autopilot-daemon/pkg/handler/handler.go`
- Updates `healthcheck.HealthCheckStatus[PVC]` in `autopilot-daemon/pkg/healthcheck/global.go`
- Exports metrics via `utils.HchecksGauge` in `autopilot-daemon/pkg/utils/prometheus.go`
- Uses Kubernetes client from `utils.GetClientsetInstance()` in `autopilot-daemon/pkg/utils/functions.go`

### Dependencies
- Kubernetes client already available

## 8. Network Bandwidth Check (iperf)

### Implementation Approach
Create a Go package that uses `iperf3` to test network bandwidth and the Kubernetes API to coordinate tests.

### Key Functions
```go
// RunIperfCheck executes the iperf network bandwidth check
func RunIperfCheck(workload string, pclients string, startport string) (string, error)

// StartIperfServers starts iperf3 servers on all nodes
func StartIperfServers(numservers string, startport string) (string, error)

// StopAllIperfServers stops all iperf3 servers
func StopAllIperfServers() (string, error)

// StartIperfClients starts iperf3 clients for a specific test
func StartIperfClients(dstip string, dstport string, numclients string) (string, error)

// generateRingWorkload generates a ring workload topology
func generateRingWorkload(nodeMap map[string]NodeInfo) map[int][]map[string]string

// executeRingWorkload executes the ring workload
func executeRingWorkload(workload map[int][]map[string]string, nodeMap map[string]NodeInfo, numClients string, startPort string) (string, error)
```

### Integration Points
- Called from `handler.IperfHandler()` in `autopilot-daemon/pkg/handler/handler.go`
- Uses existing Kubernetes client from `utils.GetClientsetInstance()` in `autopilot-daemon/pkg/utils/functions.go`

### Dependencies
- `iperf3` must be available in the container
- Kubernetes client already available

## Implementation Order

1. PCIe Bandwidth Check (pciebw)
2. GPU Remapped Rows Check (remapped)
3. GPU Memory Check (gpumem)
4. DCGM Diagnostics Check (dcgm)
5. Network Ping Check (ping)
6. GPU Power Check (gpupower)
7. PVC Create/Delete Check (pvc)
8. Network Bandwidth Check (iperf)

## Testing Strategy

Each health check package will include:
1. Unit tests for parsing functions
2. Mock tests for external command execution
3. Integration tests with the existing metrics and labeling systems
4. Validation that output matches the Python implementation format

## Backward Compatibility

All implementations will produce identical output to the Python scripts to ensure:
1. Existing monitoring and alerting systems continue to work
2. Log parsing scripts continue to function
3. API consumers receive the same response format