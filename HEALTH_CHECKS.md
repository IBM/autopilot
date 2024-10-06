# Health Checks

Here is a breakdown of the existing health checks:

1. PCIe Bandwidth Check (pciebw):
    - Description  : Host-to-device connection speeds, one measurement per GPU. Codebase in tag [v12.4.1](https://github.com/NVIDIA/cuda-samples/tree/master/Samples/1_Utilities/bandwidthTest)
    - Outputs: Pass/fail results based on PCIe bandwidth thresholds.
    - Implementation: Compares bandwidth results to a threshold (e.g., 8 GB/s). If the measured bandwidth falls below the threshold, it triggers a failure.
2. GPU Memory Check (remapped):
    - Description: Information from nvidia-smi regarding GPU memory remapped rows.
    - Outputs: Reports the state of GPU memory (normal/faulty).
    - Implementation: Analyzes remapped rows information to assess potential GPU memory issues.
3. GPU Memory Bandwidth Performance (gpumem):
    - Description: Memory bandwidth measurements using DAXPY and DGEMM.
    - Outputs: Performance metrics (eg., TFlops, power).
    - Implementation: CUDA code that valuates memory bandwidth and flags deviations from expected performance values.
4. GPU Diagnostics (dcgm):
    - Description: Runs NVidia DCGM diagnostics using dcgmi diag.
    - Outputs: Diagnostic results (pass/fail).
    - Implementation: Analyzes GPU health, including memory, power, and thermal performance.
5. PVC Create/Delete (pvc):
    - Description: Given a storage class, tests if a PVC can be created and deleted.
    - Output: pass/fail depending on the success or failure of creation and deletion of a PVC. If either operation fail, the result is a failure.
    - Implementation: creation of a PVC through K8s APIs.
6. Network Reachability Check (ping):
    - Description: Pings between nodes to assess connectivity.
    - Outputs: Pass/fail based on ping success.
    - Implementation: all-to-all reachability test.
7. Network Bandwidth Check (iperf):
    - Description: Tests network bandwidth by launching clients and servers on multiple interfaces through iperf3. Results are aggregated per interface results from network tests. Further details can be found in [the dedicated page](https://github.com/IBM/autopilot/tree/main/autopilot-daemon/network).
    - Outputs: Aggregate bandwidth on each interface, per node (in Gb/s).
    - Implementation: Tests network bandwidth by launching clients and servers on multiple interfaces and by running a ring topology on all network interfaces found on the pod that are exposed by network controllers like multi-nic CNI, which exposes fast network interfaces in the pods requesting them. Does not run on `eth0`.

These checks are configured to run periodically (e.g., hourly), and results are accessible via Prometheus, direct API queries or labels on the worker nodes.
