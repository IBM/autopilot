# Health Checks Implementation Plan: Python to Go Migration

## Executive Summary

This document outlines the comprehensive plan for migrating the Python-based health checks to native Go implementations. The goal is to replace all Python scripts with Go code that provides identical functionality while better integrating with the existing Go codebase.

## Current State Analysis

### Python Health Check Script
The current implementation uses `autopilot-daemon/utils/runHealthchecks.py` as the main entry point, which:
- Uses Kubernetes Python client to interact with the cluster
- Executes various Python and shell scripts for different health checks
- Runs health checks either locally or remotely on other nodes
- Collects and reports results

### Existing Go Codebase
The Go implementation already has:
- HTTP handlers for each health check in `autopilot-daemon/pkg/handler/handler.go`
- Health check execution functions in `autopilot-daemon/pkg/healthcheck/healthcheck.go`
- Kubernetes client integration in `autopilot-daemon/pkg/utils/`
- Prometheus metrics collection
- Node labeling functionality

### Integration Approach
The Python script is currently called from the Go codebase using `exec.Command` in `RunHealthRemoteNodes` function. The plan is to replace these external calls with native Go implementations.

## Health Checks to Migrate

### 1. PCIe Bandwidth Check (pciebw)
**Current:** Python script calls `gpu-bw/entrypoint.py` which executes `gpu-bw/gpuLocalBandwidthTest.sh`
**Migration:** Implement native Go function that executes `bandwidthTest` binary directly

### 2. GPU Remapped Rows Check (remapped)
**Current:** Python script calls `gpu-remapped/entrypoint.py` which executes `gpu-remapped/remapped-rows.sh`
**Migration:** Implement native Go function that uses `nvidia-smi` directly

### 3. GPU Memory Check (gpumem)
**Current:** Python script calls `gpu-mem/entrypoint.py` which executes `gpu-mem/gpucheck`
**Migration:** Implement native Go function that executes `gpucheck` binary directly

### 4. DCGM Diagnostics Check (dcgm)
**Current:** Python script calls `gpu-dcgm/entrypoint.py` which uses `dcgmi` CLI
**Migration:** Implement native Go function that uses `dcgmi` CLI directly with JSON parsing

### 5. Network Ping Check (ping)
**Current:** Python script calls `network/ping-entrypoint.py` which uses Kubernetes API and `ping` command
**Migration:** Implement native Go function using Kubernetes Go client and `ping` command

### 6. GPU Power Check (gpupower)
**Current:** Shell script `gpu-power/power-throttle.sh` which uses `nvidia-smi`
**Migration:** Implement native Go function that uses `nvidia-smi` directly

### 7. PVC Create/Delete Check (pvc)
**Current:** Partially implemented in Go in `healthcheck/functions.go`
**Migration:** Enhance existing implementation to be fully native

### 8. Network Bandwidth Check (iperf)
**Current:** Python scripts in `network/` directory that use `iperf3`
**Migration:** Implement native Go functions that use `iperf3` and Kubernetes Go client

## Implementation Architecture

### Package Structure
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

### Interface Design
Each health check package will implement a consistent interface:
```go
type HealthCheck interface {
    Run() (string, error)
    Name() string
    Type() HealthCheckType
}
```

### Integration with Existing Code
The new implementations will integrate with:
1. HTTP handlers in `pkg/handler/handler.go`
2. Health check status tracking in `pkg/healthcheck/global.go`
3. Prometheus metrics in `pkg/utils/prometheus.go`
4. Node labeling in `pkg/utils/functions.go`

## Implementation Steps

### Phase 1: Foundation (Week 1)
1. Create package structure
2. Implement PCIe bandwidth check
3. Implement GPU remapped rows check
4. Implement GPU memory check

### Phase 2: Advanced Checks (Week 2)
1. Implement DCGM diagnostics check
2. Implement network ping check
3. Implement GPU power check
4. Enhance PVC check

### Phase 3: Network Checks (Week 3)
1. Implement network bandwidth (iperf) check
2. Complete integration testing
3. Performance optimization

### Phase 4: Testing and Validation (Week 4)
1. Unit testing for all components
2. Integration testing with existing codebase
3. Validation against Python implementation outputs
4. Documentation updates

## Testing Strategy

### Unit Testing
Each package will have comprehensive unit tests covering:
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

## Risk Mitigation

### Compatibility Risks
- **Mitigation:** Maintain identical output format
- **Validation:** Extensive comparison testing with Python implementations

### Performance Risks
- **Mitigation:** Profile Go implementations against Python counterparts
- **Optimization:** Use Go's concurrency features where appropriate

### Deployment Risks
- **Mitigation:** Implement feature flags for gradual rollout
- **Rollback:** Maintain Python scripts as fallback option

## Success Criteria

1. All health checks reimplemented in Go with identical functionality
2. No degradation in performance compared to Python implementations
3. Full compatibility with existing monitoring, alerting, and logging systems
4. Successful integration with existing HTTP handlers and metrics collection
5. Comprehensive test coverage (minimum 80% code coverage)
6. Updated documentation reflecting the new implementation

## Timeline

| Week | Activities |
|------|------------|
| 1 | Foundation implementation (PCIe, Remapped, GPU Memory) |
| 2 | Advanced checks implementation (DCGM, Ping, Power, PVC) |
| 3 | Network bandwidth implementation and integration |
| 4 | Testing, validation, and documentation |

## Resources Required

1. Development team with Go and Kubernetes experience
2. Access to test environment with NVIDIA GPUs
3. Container images with required binaries (bandwidthTest, gpucheck, dcgmi, iperf3)
4. Monitoring and alerting system for validation

## Next Steps

1. Review and approve this implementation plan
2. Create feature branch for development
3. Begin Phase 1 implementation
4. Schedule weekly progress reviews