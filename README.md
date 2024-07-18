# AI Training Autopilot
Autopilot is a Kubernetes-native daemon that continuously monitors and evaluates GPUs, network and storage health, designed to detect and report infrastructure-level issues during the lifetime of AI workloads. It is an open-source project developed by IBM Research.

In AI training jobs, which may run for weeks or months, anomalies in the GPUs and network can happen anytime and often go undetected. In this case, performance degrades suddenly and a deep diagnostic is needed to identify the root cause, delaying or deleting the current job. Similarly, hardware anomalies can greatly disrupt the throughput and latency of an AI inference server.

The role of Autopilot is to detect and report any problems that are detected during the lifetime of the job and the existence of a cluster.

It implements a set of health checks evaluating the status of the system. These health checks focus mainly on subtle/software issues (i.e., row-remapping or PCIe link degradation), but also run connectivity tests (i.e., ping, iperf) to verify that secondary NICs are reachable. It can also verify that persistent volume claims (PVC) creation is functional for a given storage class.


![image](https://media.github.ibm.com/user/96687/files/0d466863-a19e-459d-a492-e2275377d4b9)


Autopilot is deployed as a Kubernetes DaemonSet on all worker nodes that have GPUs. Each pod exposes a Service that can be accessed through RESTful API to request the execution of health checks. Therefore, each health check has its own entry point, but also a generic “status” entry point is provided.

The DaemonSet does not run as privileged and requires access to GPUs without requesting them as resources. Therefore, the GPUs are seen as available by the scheduler.

The main code is written in Go, while health checks are written in a combination of Python, Go, bash and CUDA. Each Autopilot pod runs health checks only on the node it resides. A pod can request other pods to run health checks on their nodes, and in that case, results are gathered and showed by the requestor pod. 

If Autopilot requires full access to GPUs to run more invasive workloads, it will spawn a separate job with resources requests and limits set.

![image](https://media.github.ibm.com/user/96687/files/4a7c81ba-857a-43d4-bc82-0784ef81b270)

The toolkit currently provides health checks for pre-flight and post-flight phases, while in-flight checks will be enabled in the future. In more details (list subject to change):

- pre-flight checks

  - validate infrastructure before the start of jobs

- in-flight checks

  - workload and system performance is continuously monitored

  - detect anomaly, and issue notification

  - controllers can take actions if errors are found

- post-flight checks

  - validate infrastructure once the job ends

# Health Checks
The current status of Autopilot includes:

- **GPU PCIe Link Bandwidth**: The PCIe NVidia bandwidth test to check host-to-device connection on each node
- **GPU Memory**: GPUs remapped rows evaluation through `nvidia-smi`
- **GPU Memory Bandwidth Performance**: GPUs memory bandwidth evaluation through DAXPY and DGEMM
- **GPU Diagnostics**: NVidia DCGM (Data Center GPU Manager) diagnostics through `dcgmi diag`
- **GPU Power Slowdown**: verify if power throttle is active through `nvidia-smi`
- **Network Reachability**: `ping` to evaluate hosts reachability
- **Network Bandwidth**: `iperf3` to evaluate network bandwidth and hosts connectivity
- **PVC Create/Delete**: given a storageclass, test the ability to successfully provision a Persistent Volume Claim
- **DCGM level 3**: deep diagnostics through NVidia DCGM tool. This test runs as a separate Job that reserves all the GPUs in the node if they are free

A subset of the tests is enabled by default, and they run by default every hour. Both the the list of health checks and the timer can be customized at initialization time.

By default, the periodic checks list contains PCIe, rows remapping, GPUs power, DCGM level 1 and ping.
Health checks can an also run them manually if needed. 

Results from health checks are exported as Prometheus Gauges, so that users and admins can easily check the status of the system on Grafana.

## Deep Diagnostics and Node Labeling
Autopilot runs health checks periodically and labels the nodes with `autopilot.ibm.com/gpuhealth: ERR` is any of the GPU health checks returns an error.

Also, more extensive tests, namely DCGM diagnostics level 3, are also executed automatically only on nodes that have free GPUs. This deeper analysis is needed to reveal problems in the GPUs that can be found only after running level 3 DCGM diagnostic.
This type of diagnostics can help deciding if the worker node should be used for running workloads or not. To facilitate this task, Autopilot will label nodes with key `autopilot/dcgm.level.3`. 

If errors are found, the label `autopilot/dcgm.level.3` will contain the value `ERR`, a timestamp, the test(s) that failed and the GPU id(s) if available. Otherwise, the value is set to `PASS_timestamp`.

## Logs and Metrics

All health checks results are exported through Prometheus, but they can be also found in each pod's logs.

All metrics are accessible through Prometheus and Grafana dashboards. The gauge exposed is `autopilot_health_checks` and can be customized with the following filters:

- `check`, select one or more specific health checks
- `node`, filter by node name
- `cpumodel` and `gpumodel`, for heterogeneous clusters
- `deviceid` to select specific GPUs, when available

# Install Autopilot 
Autopilot can be installed through Helm and need admin privileges to create objects like services, serviceaccounts, namespaces and relevant RBAC.

## Requirements

- Need to install `helm-git` plugin on all hosts

```bash
helm plugin install https://github.com/aslafy-z/helm-git --version 0.15.1
```

## Helm Chart customization

Helm charts values and how-to for customization can be found [here](https://github.com/IBM/autopilot/tree/main/autopilot-daemon/helm-charts/autopilot).


## Install

1) Add autopilot repo

```bash
helm repo add autopilot git+https://github.com/IBM/autopilot.git@autopilot-daemon/helm-charts/autopilot?ref=gh-pages
```

2) Install autopilot (idempotent command). The config file is for customizing the helm values. Namespace is where the helm chart will live, not the namespace where Autopilot runs

```bash
helm upgrade autopilot autopilot/autopilot-daemon --install --namespace=<default> -f your-config.yml
```

The controllers should show up in the selected namespace

```bash
oc get po -n autopilot
```

```bash
NAME                               READY   STATUS    RESTARTS   AGE
autopilot-daemon-autopilot-g7j6h   1/1     Running   0          70m
autopilot-daemon-autopilot-g822n   1/1     Running   0          70m
autopilot-daemon-autopilot-x6h8d   1/1     Running   0          70m
autopilot-daemon-autopilot-xhntv   1/1     Running   0          70m
```

## Uninstall

```bash
 helm uninstall autopilot # -n <default>
```


# Manually Query the Autopilot Service

Autopilot provides a `/status` handler that can be queried to get the entire system status, meaning that it will run all the tests on all the nodes. Autopilot is reachable by service name `autopilot-healthchecks.autopilot.svc` in-cluster only, meaning it can be reached from a pod running in the cluster, or through port forwarding (see below).

Health check names are `pciebw`, `dcgm`, `remapped`, `ping`, `iperf`, `pvc`.

For example, using port forwarding to localhost or by exposing the service

```bash
kubectl port-forward service/autopilot-healthchecks 3333:3333 -n autopilot
# or kubectl expose service autopilot-healthchecks -n autopilot
```

If using port forward, then launch `curl` on another terminal

```bash
curl "http://localhost:3333/status?check=pciebw&host=nodename1"
```

Alternatively, retrieve the route with `kubectl get routes autopilot-healthchecks -n autopilot`

```bash
curl "<route-name>/status?check=pciebw&host=nodename1"
```


All tests can be tailored by a combination of:

- `host=<hostname1,hostname2,...>`, to run all tests on a specific node or on a comma separated list of nodes.
- `check=<healthcheck1,healtcheck2,...>`, to run a single test (`pciebw`, `dcgm`, `remapped`, `gpumem`, `ping`, `iperf` or `all`) or a list of comma separated tests. When no parameters are specified, only `pciebw`, `dcgm`, `remapped`, `ping` tests are run.
- `batch=<#hosts>`, how many hosts to check at a single moment. Requests to the batch are run in parallel asynchronously. Batching is done to avoid running too many requests in parallel when the number of worker nodes increases. Defaults to all nodes.

Some health checks provide further customization.

### DCGM
This test runs `dcgmi diag`, and we support only `r` as [parameter](https://docs.nvidia.com/datacenter/dcgm/latest/user-guide/dcgm-diagnostics.html#command-line-options). 

The default is `1`, but can customize it by `/status?check=dcgm&r=2`.

### IPERF 
This tests runs from a client node, which

- Issues several RPCs to start remote `iperf3` servers
- Launches a certain number of clients towards each of those servers

Both can be customized.
- `serverspernode` can be used to create a certain number of servers on each remote node.
  - if the value is lower than the number of secondary network interfaces, it will create the minimum number of `1` server per interface (excludes `eth0` and `lo`). Each server runs on a separate port.
  - otherwise, it will divide that value by the number of network interfaces existing in the cluster.
- `clientsperiface` can be used to launch a desired number of clients against a single remote server.

Another possible customization is to decide which network plane to test. By default is `data` plane, that is, what runs on secondary interfaces.

To test connection on `eth0`, that is, the management plane (`mgmt`), can use the `plane` parameter as follows `/status?check=iperf&plane=mgmt`.
It will create only one client and there is a single server per node.

### Example

In this example, we target one node and check the pcie bandwidth and use the port-forwarding method.
In this scenario, we have a value lower than `8GB/s`, which results in an alert. This error will be exported to the OpenShift web console and on Slack, if that is enabled by admins.

```bash
curl "http://127.0.0.1:3333/status?check=pciebw"
```

The output of the command above, will be similar to the following (edited to save space):
```bash
Checking status on all nodes
Autopilot Endpoint: 10.128.6.187
Node: hostname
url(s): http://10.128.6.187:3333/status?host=hostname&check=pciebw
Response:
Checking system status of host hostname (localhost) 

[[ PCIEBW ]] Briefings completed. Continue with PCIe Bandwidth evaluation.
[[ PCIEBW ]] FAIL
Host  hostname
12.3 12.3 12.3 12.3 5.3 12.3 12.3 12.3

Node Status: PCIE Failed
-------------------------------------


Autopilot Endpoint: 10.131.4.93
Node: hostname2
url(s): http://10.131.4.93:3333/status?host=hostname2&check=pciebw
Response:
Checking system status of host hostname2 (localhost) 

[[ PCIEBW ]] Briefings completed. Continue with PCIe Bandwidth evaluation.
[[ PCIEBW ]] SUCCESS
Host  hostname2
12.1 12.0 12.3 12.3 11.9 11.5 12.1 12.1

Node Status: Ok
-------------------------------------

Node Summary:

{'hostname': ['PCIE Failed'],
 'hostname2': ['Ok']}

runtime: 31.845192193984985 sec
```
