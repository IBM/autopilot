# Manually Query the Autopilot Service

Autopilot provides a `/status` handler that can be queried to get the entire system status, meaning that it will run all the tests on all the nodes. Autopilot is reachable by service name `autopilot-healthchecks.autopilot.svc` in-cluster only, meaning it can be reached from a pod running in the cluster, or through port forwarding (see below).

Health check names are `pciebw`, `dcgm`, `remapped`, `ping`, `iperf`, `pvc`, `gpumem`.

For example, using port forwarding to localhost or by exposing the service

```bash
kubectl port-forward service/autopilot-healthchecks 3333:3333 -n autopilot
# or oc expose service autopilot-healthchecks -n autopilot in OpenShift
```

If using port forward, then launch `curl` on another terminal

```bash
curl "http://localhost:3333/status?check=pciebw&host=nodename1"
```

Alternatively, retrieve the route with `kubectl get routes autopilot-healthchecks -n autopilot`
When using routes, it is recommended to [increase the timeout](https://docs.openshift.com/container-platform/4.10/networking/routes/route-configuration.html#nw-configuring-route-timeouts_route-configuration) with the following command

```bash
oc annotate route autopilot-healthchecks -n autopilot --overwrite haproxy.router.openshift.io/timeout=30m 
```

Then:

```bash
curl "http://<route-name>/status?check=pciebw&host=nodename1"
```

All tests can be tailored by a combination of:

- `host=<hostname1,hostname2,...>`, to run all tests on a specific node or on a comma separated list of nodes.
- `check=<healthcheck1,healtcheck2,...>`, to run a single test (`pciebw`, `dcgm`, `remapped`, `gpumem`, `ping`, `iperf` or `all`) or a list of comma separated tests. When no parameters are specified, only `pciebw`, `dcgm`, `remapped`, `ping` tests are run.
- `job=<namespace:key=value>`, run tests on nodes running a job labeled with `key=value` in a specific namespace.
- `nodelabel=<key=value>`, run tests on nodes having the `key=value` label.
- `batch=<#hosts>`, how many hosts to check at a single moment. Requests to the batch are run in parallel asynchronously. Batching is done to avoid running too many requests in parallel when the number of worker nodes increases. Defaults to all nodes.

Some health checks provide further customization. More details on all the tests can be found [here](https://github.com/IBM/autopilot/autopilot-daemon/HEALTH_CHECKS.md)
Note that if multiple node selection parameters (`host`, `job`, `nodelabel`) are provided together, Autopilot will run tests on nodes that match _any_  of the specified parameters (set union). For example, the following command will run the `pciebw` test on all nodes that either have the label `label1` OR are running the job `jobKey=job2` because both `nodelabel` and `job` parameters are provided in the input:
`curl "http://<route-name>/status?check=pciebw&nodelabel=label1&job=default:jobKey=job2"`

## DCGM

This test runs `dcgmi diag`, and we support only `r` as [parameter](https://docs.nvidia.com/datacenter/dcgm/latest/user-guide/dcgm-diagnostics.html#command-line-options).

The default is `1`, but can customize it by `/status?check=dcgm&r=2`.

## Network Bandwidth Validation with IPERF

As part of this workload, Autopilot will generate the Ring Workload and then start `iperf3 servers` on each interface on each Autopilot pod based on the configuration options provided by the user.  Only after the `iperf3 servers` are started, Autopilot will begin executing the workload by starting `iperf3 clients` based on the configuration options provided by the user. All results are logged back to the user.

- For each network interface on each node, an `iperf3 server` is started. The number of `iperf3 servers` is dependent on the `number of clients` intended on being run. For example, if the  `number of clients` is `8`, then there will be `8` `iperf3 servers` started per interface on a unique `port`.

- Invocation from the exposed Autopilot API is as follows below:

```bash
# Invoked via the `status` handle:
curl "http://autopilot-healthchecks-autopilot.<domain>/status?check=iperf&workload=ring&pclients=<NUMBER_OF_IPERF3_CLIENTS>&startport=<STARTING_IPERF3_SERVER_PORT>"

# Invoked via the `iperf` handle directly:
curl "http://autopilot-healthchecks-autopilot.<domain>/iperf?workload=ring&pclients=<NUMBER_OF_IPERF3_CLIENTS>&startport=<STARTING_IPERF3_SERVER_PORT>"
```

## Concrete Example

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
