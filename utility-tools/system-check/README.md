# Complete Health Check Assessment

This Helm chart can be used to launch a Pod that runs the health checks on all the nodes.

The health checks are:

- PCIe Bandwidth on all GPUs, host to device
- Check on existing GPUs remapped rows
- Check on correct availability of Secondary Nics

This workload is a Python program listing the Autopilot endpoints belonging to the Autopilot Kubernetes Service that exposes the health checks API.

This program needs to deploy a ClusterRole and ClusterRoleBinding required to list the endpoints, therefore this workload can be deployed only if the necessary privileges are granted (i.e., cluster admins).

## Installation

The Helm chart can be configured by updating the following values in `values.yaml`:

- `namespace` where to run the Pod. The namespace needs to have a valid `ImagePullSecret` to get images from `us.icr.io` or `icr.io`
- `imagePullSecret` defaulted to `all-icr-io`
- `autopilotService` is the name of the Service that exposes the health checks endpoints. It is defaulted to `autopilot-healthchecks`
- `autopilotNamespace` is the namespace where the Autopilot daemons are running. It is defaulted to `autopilot`
- `targetNode` to run the test(s) on a specific node only, rather than on the entire system
- `testType` is the type of test that will run i.e. pciebw, nic, remapped, or all
- `batchSize` is the number of nodes running a health check per processor. It is defaulted to `1` node per processor

To deploy the Pod:

```bash
helm install system-check utility-tools/system-check/charts/
```

Logs can be streamed with

```bash
kubectl logs -f system-check 
```

All the health checks expose metrics that can be plotted through Grafana dashboards. A `json` file for a set of predefined dashboards in Autopilot, can be found in the `utiliy-tools` directory in this repository.

## Uninstall

To uninstall:

```bash
helm uninstall system-check 
```
