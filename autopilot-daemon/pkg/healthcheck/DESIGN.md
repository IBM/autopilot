# Go Health Check Implementation Design

## Overview

This document outlines the design for reimplementing the Python health check functionality in Go. The goal is to replace the Python scripts with native Go implementations that provide the same functionality while integrating seamlessly with the existing Go codebase.

## Health Check Components

### 1. PCIe Bandwidth Check (pciebw)

**Current Implementation:**
- Python script: `autopilot-daemon/gpu-bw/entrypoint.py`
- Shell script: `autopilot-daemon/gpu-bw/gpuLocalBandwidthTest.sh`
- CUDA binary: `bandwidthTest`

**Go Implementation Plan:**
- Create a Go package `pciebw` in `autopilot-daemon/pkg/healthcheck/pciebw`
- Implement a function `RunPCIeBWCheck(threshold int) (string, error)` that:
  - Executes the `bandwidthTest` binary directly using `os/exec`
  - Parses the output to extract bandwidth measurements
  - Compares against the threshold to determine pass/fail
  - Returns results in the same format as the Python implementation

**Dependencies:**
- CUDA `bandwidthTest` binary must be available in the container

### 2. GPU Remapped Rows Check (remapped)

**Current Implementation:**
- Python script: `autopilot-daemon/gpu-remapped/entrypoint.py`
- Shell script: `autopilot-daemon/gpu-remapped/remapped-rows.sh`

**Go Implementation Plan:**
- Create a Go package `remapped` in `autopilot-daemon/pkg/healthcheck/remapped`
- Implement a function `RunRemappedRowsCheck() (string, error)` that:
  - Uses `os/exec` to run `nvidia-smi` commands
  - Parses the output to check for remapped rows
  - Returns results in the same format as the Python implementation

**Dependencies:**
- `nvidia-smi` must be available in the container

### 3. GPU Memory Check (gpumem)

**Current Implementation:**
- Python script: `autopilot-daemon/gpu-mem/entrypoint.py`
- CUDA binary: `gpucheck`

**Go Implementation Plan:**
- Create a Go package `gpumem` in `autopilot-daemon/pkg/healthcheck/gpumem`
- Implement a function `RunGPUMemCheck() (string, error)` that:
  - Executes the `gpucheck` binary directly using `os/exec`
  - Parses the output to determine pass/fail
  - Returns results in the same format as the Python implementation

**Dependencies:**
- CUDA `gpucheck` binary must be available in the container

### 4. DCGM Diagnostics Check (dcgm)

**Current Implementation:**
- Python script: `autopilot-daemon/gpu-dcgm/entrypoint.py`
- Uses `dcgmi` CLI tool

**Go Implementation Plan:**
- Create a Go package `dcgm` in `autopilot-daemon/pkg/healthcheck/dcgm`
- Implement a function `RunDCGMCheck(level string) (string, error)` that:
  - Uses `os/exec` to run `dcgmi diag` commands
  - Parses JSON output to check for failures
  - Implements the same result parsing logic as the Python version
  - Updates node labels using Kubernetes API (already available in utils)
  - Returns results in the same format as the Python implementation

**Dependencies:**
- `dcgmi` must be available in the container
- Kubernetes client already available

### 5. Network Ping Check (ping)

**Current Implementation:**
- Python script: `autopilot-daemon/network/ping-entrypoint.py`
- Uses Kubernetes API to discover nodes and network interfaces
- Uses `ping` command to test connectivity

**Go Implementation Plan:**
- Create a Go package `ping` in `autopilot-daemon/pkg/healthcheck/ping`
- Implement a function `RunPingCheck(nodes []string, job string, nodelabel string) (string, error)` that:
  - Uses existing Kubernetes client from utils package
  - Discovers nodes and network interfaces
  - Executes `ping` commands to test connectivity
  - Parses output to determine pass/fail
  - Returns results in the same format as the Python implementation

**Dependencies:**
- `ping` command must be available in the container
- Kubernetes client already available

### 6. GPU Power Check (gpupower)

**Current Implementation:**
- Shell script: `autopilot-daemon/gpu-power/power-throttle.sh`

**Go Implementation Plan:**
- Create a Go package `gpupower` in `autopilot-daemon/pkg/healthcheck/gpupower`
- Implement a function `RunGPUPowerCheck() (string, error)` that:
  - Uses `os/exec` to run `nvidia-smi` commands
  - Parses output to check for power throttling
  - Returns results in the same format as the shell script

**Dependencies:**
- `nvidia-smi` must be available in the container

### 7. PVC Create/Delete Check (pvc)

**Current Implementation:**
- Implemented in `autopilot-daemon/pkg/healthcheck/healthcheck.go` and `functions.go`
- Already partially implemented in Go using Kubernetes API

**Go Implementation Plan:**
- Enhance existing implementation in `autopilot-daemon/pkg/healthcheck/functions.go`
- Implement a function `RunPVCCheck() (string, error)` that:
  - Uses Kubernetes API to create and delete PVCs
  - Follows the same logic as the current implementation
  - Returns results in the same format as the Python implementation

**Dependencies:**
- Kubernetes client already available

### 8. Network Bandwidth Check (iperf)

**Current Implementation:**
- Python scripts in `autopilot-daemon/network/`
- Uses `iperf3` to test network bandwidth

**Go Implementation Plan:**
- Create a Go package `iperf` in `autopilot-daemon/pkg/healthcheck/iperf`
- Implement functions:
  - `RunIperfCheck(workload string, pclients string, startport string) (string, error)`
  - `StartIperfServers(numservers string, startport string) (string, error)`
  - `StopAllIperfServers() (string, error)`
  - `StartIperfClients(dstip string, dstport string, numclients string) (string, error)`
- Use existing Kubernetes client from utils package
- Execute `iperf3` commands using `os/exec`
- Parse JSON output to extract bandwidth metrics
- Return results in the same format as the Python implementation

**Dependencies:**
- `iperf3` must be available in the container
- Kubernetes client already available

## Integration with Existing Codebase

### HTTP Handlers
The existing HTTP handlers in `autopilot-daemon/pkg/handler/handler.go` will be updated to call the new Go implementations instead of executing Python scripts.

### Health Check Status Tracking
The existing `HealthCheckStatus` map in `autopilot-daemon/pkg/healthcheck/global.go` will be used to track the status of each health check.

### Metrics Collection
The existing Prometheus metrics in `autopilot-daemon/pkg/utils/prometheus.go` will be used to export health check results.

### Node Labeling
The existing node labeling functions in `autopilot-daemon/pkg/utils/functions.go` will be used to update node labels based on health check results.

## Package Structure

```
autopilot-daemon/pkg/healthcheck/
├── DESIGN.md (this file)
├── global.go (existing)
├── functions.go (existing)
├── healthcheck.go (existing)
├── pciebw/
│   ├── pciebw.go
│   └── pciebw_test.go
├── remapped/
│   ├── remapped.go
│   └── remapped_test.go
├── gpumem/
│   ├── gpumem.go
│   └── gpumem_test.go
├── dcgm/
│   ├── dcgm.go
│   └── dcgm_test.go
├── ping/
│   ├── ping.go
│   └── ping_test.go
├── gpupower/
│   ├── gpupower.go
│   └── gpupower_test.go
├── pvc/
│   ├── pvc.go
│   └── pvc_test.go
└── iperf/
    ├── iperf.go
    └── iperf_test.go
```

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

Each health check will have unit tests that:
- Mock external command execution where possible
- Verify output format matches Python implementation
- Test error conditions
- Validate integration with existing metrics and labeling systems

## Backward Compatibility

The new Go implementations will produce identical output to the Python scripts to ensure backward compatibility with:
- Existing monitoring and alerting systems
- Log parsing scripts
- API consumers