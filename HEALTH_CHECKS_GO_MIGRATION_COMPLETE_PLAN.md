# Complete Plan for Migrating Health Checks from Python to Go

## Executive Summary

This document provides a comprehensive plan for migrating the Python-based health check system to native Go implementations. The migration will replace all Python scripts with Go code that integrates seamlessly with the existing Go codebase while maintaining full backward compatibility.

## Current System Analysis

### Python Health Check System
The current implementation consists of:
- Main orchestrator: `autopilot-daemon/utils/runHealthchecks.py`
- Individual health check scripts in Python and shell
- External execution of binaries and scripts
- Kubernetes Python client for cluster interactions
- JSON output parsing and node labeling

### Go Codebase Structure
The existing Go codebase provides:
- HTTP handlers in `autopilot-daemon/pkg/handler/`
- Health check logic in `autopilot-daemon/pkg/healthcheck/`
- Utility functions in `autopilot-daemon/pkg/utils/`
- Kubernetes client integration
- Prometheus metrics collection
- Node labeling functionality

## Migration Strategy

### Approach
1. **Incremental Replacement**: Replace Python scripts with Go implementations one by one
2. **Backward Compatibility**: Maintain identical output format and behavior
3. **Native Integration**: Use Go's native capabilities instead of subprocess calls
4. **Feature Parity**: Ensure all existing functionality is preserved

### Implementation Order
1. PCIe Bandwidth Check (pciebw)
2. GPU Remapped Rows Check (remapped)
3. GPU Memory Check (gpumem)
4. DCGM Diagnostics Check (dcgm)
5. Network Ping Check (ping)
6. GPU Power Check (gpupower)
7. PVC Create/Delete Check (pvc)
8. Network Bandwidth Check (iperf)

## Detailed Implementation Plan

### 1. PCIe Bandwidth Check (pciebw)

**Current Implementation:**
- Python: `gpu-bw/entrypoint.py`
- Shell: `gpu-bw/gpuLocalBandwidthTest.sh`
- Binary: `bandwidthTest`

**Go Implementation:**
```go
// Key functions
func RunPCIeBWCheck(threshold int) (string, error)
func parseBandwidthOutput(output string) ([]float64, error)
func checkThreshold(bandwidths []float64, threshold int) bool
```

**Integration Points:**
- HTTP handler: `handler.PCIeBWHandler()`
- Metrics: `utils.HchecksGauge`
- Status tracking: `healthcheck.HealthCheckStatus[PCIeBW]`

### 2. GPU Remapped Rows Check (remapped)

**Current Implementation:**
- Python: `gpu-remapped/entrypoint.py`
- Shell: `gpu-remapped/remapped-rows.sh`
- Command: `nvidia-smi`

**Go Implementation:**
```go
// Key functions
func RunRemappedRowsCheck() (string, error)
func parseRemappedRowsOutput(output string) (map[int]bool, error)
func detectGPUs() ([]int, error)
```

**Integration Points:**
- HTTP handler: `handler.RemappedRowsHandler()`
- Metrics: `utils.HchecksGauge`
- Status tracking: `healthcheck.HealthCheckStatus[RowRemap]`

### 3. GPU Memory Check (gpumem)

**Current Implementation:**
- Python: `gpu-mem/entrypoint.py`
- Binary: `gpu-mem/gpucheck`

**Go Implementation:**
```go
// Key functions
func RunGPUMemCheck() (string, error)
func parseGPUMemOutput(output string) (bool, error)
```

**Integration Points:**
- HTTP handler: `handler.GpuMemHandler()`
- Metrics: `utils.HchecksGauge`
- Status tracking: `healthcheck.HealthCheckStatus[GPUMem]`

### 4. DCGM Diagnostics Check (dcgm)

**Current Implementation:**
- Python: `gpu-dcgm/entrypoint.py`
- Command: `dcgmi diag`

**Go Implementation:**
```go
// Key functions
func RunDCGMCheck(level string) (string, error)
func parseDCGMJSON(output string) (bool, string, error)
func patchNode(success bool, output string, level string) error
```

**Integration Points:**
- HTTP handler: `handler.DCGMHandler()`
- Metrics: `utils.HchecksGauge`
- Status tracking: `healthcheck.HealthCheckStatus[DCGM]`
- Kubernetes client: `utils.GetClientsetInstance()`

### 5. Network Ping Check (ping)

**Current Implementation:**
- Python: `network/ping-entrypoint.py`
- Command: `ping`

**Go Implementation:**
```go
// Key functions
func RunPingCheck(nodes []string, job string, nodelabel string) (string, error)
func discoverNodes(job string, nodelabel string) (map[string]string, error)
func getNetworkInterfaces() (map[string]map[string][]string, error)
func executePing(ip string) (bool, error)
```

**Integration Points:**
- HTTP handler: `handler.PingHandler()`
- Metrics: `utils.HchecksGauge`
- Status tracking: `healthcheck.HealthCheckStatus[Ping]`
- Kubernetes client: `utils.GetClientsetInstance()`

### 6. GPU Power Check (gpupower)

**Current Implementation:**
- Shell: `gpu-power/power-throttle.sh`
- Command: `nvidia-smi`

**Go Implementation:**
```go
// Key functions
func RunGPUPowerCheck() (string, error)
func parsePowerOutput(output string) ([]float64, error)
func detectGPUs() ([]int, error)
```

**Integration Points:**
- HTTP handler: `handler.GpuPowerHandler()`
- Metrics: `utils.HchecksGauge`
- Status tracking: `healthcheck.HealthCheckStatus[GPUPower]`

### 7. PVC Create/Delete Check (pvc)

**Current Implementation:**
- Partially in Go: `healthcheck/functions.go`
- Kubernetes API calls

**Go Implementation:**
```go
// Key functions
func RunPVCCheck() (string, error)
func createPVC() error
func deletePVC(name string) error
func listPVC(name string) (string, error)
```

**Integration Points:**
- HTTP handler: `handler.PVCHandler()`
- Metrics: `utils.HchecksGauge`
- Status tracking: `healthcheck.HealthCheckStatus[PVC]`
- Kubernetes client: `utils.GetClientsetInstance()`

### 8. Network Bandwidth Check (iperf)

**Current Implementation:**
- Multiple Python scripts in `network/`
- Command: `iperf3`

**Go Implementation:**
```go
// Key functions
func RunIperfCheck(workload string, pclients string, startport string) (string, error)
func StartIperfServers(numservers string, startport string) (string, error)
func StopAllIperfServers() (string, error)
func StartIperfClients(dstip string, dstport string, numclients string) (string, error)
func generateRingWorkload(nodeMap map[string]NodeInfo) map[int][]map[string]string
func executeRingWorkload(workload map[int][]map[string]string, nodeMap map[string]NodeInfo, numClients string, startPort string) (string, error)
```

**Integration Points:**
- HTTP handler: `handler.IperfHandler()`
- Kubernetes client: `utils.GetClientsetInstance()`

## Package Structure

```
autopilot-daemon/pkg/healthcheck/
├── DESIGN.md
├── IMPLEMENTATION_PLAN.md
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

## Integration with Existing Codebase

### HTTP Handlers
Update `autopilot-daemon/pkg/handler/handler.go` to call Go functions instead of executing Python scripts:

```go
// Before
tmp, err = exec.Command("python3", "./gpu-bw/entrypoint.py", "-t", strconv.Itoa(utils.UserConfig.BWThreshold)).CombinedOutput()

// After
tmp, err = pciebw.RunPCIeBWCheck(utils.UserConfig.BWThreshold)
```

### Metrics Collection
Reuse existing Prometheus metrics:
```go
utils.HchecksGauge.WithLabelValues(string(PCIeBW), utils.NodeName, utils.CPUModel, utils.GPUModel, strconv.Itoa(gpuid)).Set(bw)
```

### Node Labeling
Reuse existing node labeling functions:
```go
utils.PatchNode(utils.GPUHealthWarnLabel, utils.NodeName, false)
```

## Testing Strategy

### Unit Testing
Each package will include comprehensive unit tests:
- Input validation
- Output parsing
- Error conditions
- Edge cases

### Integration Testing
- Verify integration with existing HTTP handlers
- Validate metrics collection
- Confirm node labeling functionality
- Test Kubernetes API interactions

### Regression Testing
- Compare output format with Python implementations
- Validate monitoring and alerting system compatibility
- Ensure log parsing scripts continue to work
- Confirm API compatibility

## Deployment Plan

### Phase 1: Development (Weeks 1-2)
1. Create package structure
2. Implement PCIe bandwidth, GPU remapped rows, GPU memory checks
3. Write unit tests

### Phase 2: Advanced Checks (Weeks 3-4)
1. Implement DCGM diagnostics, network ping, GPU power checks
2. Enhance PVC implementation
3. Write integration tests

### Phase 3: Network Bandwidth (Weeks 5-6)
1. Implement network bandwidth (iperf) check
2. Complete integration testing
3. Performance optimization

### Phase 4: Validation (Week 7)
1. Comprehensive testing against Python implementations
2. Performance benchmarking
3. Documentation updates

### Phase 5: Deployment (Week 8)
1. Gradual rollout with feature flags
2. Monitor for issues
3. Full production deployment

## Risk Mitigation

### Compatibility Risks
- **Mitigation:** Maintain identical output format
- **Validation:** Extensive comparison testing

### Performance Risks
- **Mitigation:** Profile against Python implementations
- **Optimization:** Use Go's concurrency features

### Deployment Risks
- **Mitigation:** Feature flags for gradual rollout
- **Rollback:** Maintain Python scripts as fallback

## Success Criteria

1. All health checks reimplemented in Go with identical functionality
2. No performance degradation compared to Python implementations
3. Full compatibility with existing monitoring, alerting, and logging systems
4. Successful integration with existing HTTP handlers and metrics collection
5. Comprehensive test coverage (minimum 80% code coverage)
6. Updated documentation reflecting the new implementation

## Benefits

### Performance
- Faster execution without subprocess overhead
- Better resource utilization
- Native concurrency support

### Reliability
- Stronger typing reduces runtime errors
- Better error handling
- Native Kubernetes client integration

### Maintainability
- Consistent language across the codebase
- Better integration with existing Go tooling
- Improved debugging capabilities

### Deployment
- Single binary deployment
- No Python dependencies
- Reduced container image size

## Conclusion

This migration plan provides a comprehensive approach to replacing the Python health check system with native Go implementations. By following the incremental approach and maintaining backward compatibility, the migration can be completed with minimal risk while providing significant benefits in performance, reliability, and maintainability.