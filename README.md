# AI Training Autopilot
This repository is part of the effort described in the IBM Research Challenge #5160.

The goal of this challenge is to enable the OpenShift container platform to become the premier platform to orchestrate the full life cycle of Foundation Model workflows (pre-processing, training, adaptation/distillation, and inference) seamlessly across public, private, and on-prem cloud environments.

From operation perspective, infrastucture stability is always important. We actually saw the various errors and anomaly states in GPU and Network, for instance, so it becomes crucial to provide a tool to detect, avoid, and handle the infrastructure issues while running the AI training job. 

We provide a collection of tools (named Autopilot) to steer and address these infrastructure issues automatically by pre-flight checks, in-flight checks, and also post-flight to learn or improve the issue detection logic. 

Autopilot runs as a DaemonSet on all worker nodes that have GPUs. All results from health checks are exposed through Prometheus and a Grafana dashboard is available in the `utility-tools` folder.


![image](https://media.github.ibm.com/user/96687/files/0d466863-a19e-459d-a492-e2275377d4b9)


The toolkit currently provides health checks for pre-flight and post-flight phases, while in-flight checks will be enabled in the future. In more details (list subject to change):

- pre-flight checks

  - validate infrastructure before the start of jobs

- in-flight checks

  - workload and system performance is continuously monitored

  - detect anomaly, decide to continue or stop the job

  - issue alert to end users

- post-flight checks

  -  validate infrastructure once the job ends

![image](https://media.github.ibm.com/user/96687/files/4a7c81ba-857a-43d4-bc82-0784ef81b270)

## Health Checks
The current status of Autopilot includes:

- The PCIe NVidia bandwidth test to check host-to-device connection on each node
- GPUs remapped rows evaluation
- NVidia DCGM (Data Center GPU Manager) diagnostics
- Ping to evaluate hosts reachability 
- Iperf3 to evaluate network bandwidth and hosts connectivity 

All test except `iperf3` are executed periodically every hour by default. The time frame can be customized during installation.

### Query the Autopilot Service

Autopilot provides a `/status` handler that can be queried to get the entire system status, meaning that it will run all the tests on all the nodes. Autopilot is reachable by service name `autopilot-healthchecks.autopilot.svc` in-cluster only, meaning it can be reached from a pod running in the cluster, or through port forwarding (see below).

Health check names are `pciebw`, `dcgm`, `remapped`, `ping`, `iperf`.

For example, using port forwarding to localhost and `curl`

```bash
curl "http://localhost:3333/status?check=pciebw&host=nodename1"
```

All tests can be tailored by a combination of:

- `host=<hostname1,hostname2,...>`, to run all tests on a specific node or on a comma separated list of nodes.
- `check=<healthcheck1,healtcheck2,...>`, to run a single test (`pciebw`, `dcgm`, `remapped`, `ping`, `iperf` or `all`) or a list of comma separated tests. When no parameters are specified, `pciebw`, `dcgm`, `remapped`, `ping` tests are run.
- `batch=<#hosts>`, how many hosts to check at a single moment. Requests to the batch are run in parallel asynchronously. Batching is done to avoid running too many requests in parallel when the number of worker nodes increases. Defaults to all nodes.

Some health checks provide further customization.

#### DCGM
This test runs `dcgmi diag`, and we support only `r` as (parameter)[https://docs.nvidia.com/datacenter/dcgm/latest/user-guide/dcgm-diagnostics.html#command-line-options]. 

The default is `1`, but can customize it by `/status?check=dcgm&r=2`.

#### IPERF 
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



## Run Health Checks

Health checks can be executed through a utility tool provided with a Helm chart, or by querying the Autopilot service.
Results can be visualized by either checking the logs of the utility tool/service query, or by looking at the data in a Grafana dashboard.
A very basic `json` for a Grafana Dasnhboard can be found [here](https://github.ibm.com/hybrid-cloud-infrastructure-research/autopilot/blob/main/utility-tools/Autopilot-Grafana-Dashboard.json).

#### Query with Port-Forward

Alternatively, it is possible to port-forward the autopilot healthchecks Service and `curl` from localhost. 

```bash
kubectl port-forward service/autopilot-healthchecks 3333:3333 -n autopilot
```

Will print the following output:
```bash
Forwarding from 127.0.0.1:3333 -> 3333
Forwarding from [::1]:3333 -> 3333
```

Then on another terminal, run the desired curl command. In this example, we target one node and check the pcie bandwidth.
In this scenario, we have a value lower than `4GB/s`, which results in an alert. This error will be exported to the OpenShift web console and on Slack, if that is enabled by admins.

```bash
curl "http://127.0.0.1:3333/status?check=pciebw"
```

The output of the command above, will be similar to the following (edited to save space):
```bash
Checking status on all nodes
Autopilot Endpoint: 10.128.6.187
Node: dev-ppv5g-var-lib-containers-f4v6q
url(s): http://10.128.6.187:3333/status?host=dev-ppv5g-var-lib-containers-f4v6q&check=pciebw
Response:
Checking system status of host dev-ppv5g-var-lib-containers-f4v6q (localhost) 

[[ PCIEBW ]] Briefings completed. Continue with PCIe Bandwidth evaluation.
[[ PCIEBW ]] FAIL
Host  dev-ppv5g-var-lib-containers-f4v6q
12.3 12.3 12.3 12.3 3.3 12.3 12.3 12.3

Node Status: PCIE Failed
-------------------------------------


Autopilot Endpoint: 10.131.4.93
Node: dev-ppv5g-var-lib-containers-pzccw
url(s): http://10.131.4.93:3333/status?host=dev-ppv5g-var-lib-containers-pzccw&check=pciebw
Response:
Checking system status of host dev-ppv5g-var-lib-containers-pzccw (localhost) 

[[ PCIEBW ]] Briefings completed. Continue with PCIe Bandwidth evaluation.
[[ PCIEBW ]] SUCCESS
Host  dev-ppv5g-var-lib-containers-pzccw
12.1 12.0 12.3 12.3 11.9 11.5 12.1 12.1

Node Status: Ok
-------------------------------------

Node Summary:

{'dev-ppv5g-var-lib-containers-f4v6q': ['PCIE Failed'],
 'dev-ppv5g-var-lib-containers-pzccw': ['Ok']}

~~~DEBUGGING BELOW~~~
Processes (randomly ordered) and the nodes they ran (process:[nodes]):
{486: ['dev-ppv5g-var-lib-containers-f4v6q'],
 487: ['dev-ppv5g-var-lib-containers-pzccw']}

runtime: 31.845192193984985 sec
```

```

#### Query from a pod
In the example below, we create a utility `nginx` pod from which we can run `curl` commands against the `autopilot-healthchecks` service.
We run the PCIe bandwidth test on all nodes, and we can see it is failing on one node.

Create a dummy nginx pod:

```bash
kubectl create job curl-pod --image=nginx -- sleep inf
```

Then run an health check:

```bash
kubectl exec jobs/curl-pod -- curl "http://autopilot-healthchecks.autopilot.svc:3333/status?check=pciebw"
```


#### Helm Chart

Users and admins can create a single pod that can run the desired health checks.
Please refer to the [dedicated page](https://github.ibm.com/hybrid-cloud-infrastructure-research/autopilot/tree/main/utility-tools/system-check) for more details and customization.

## Install autopilot (Admin)
**Installation**: Both projects can be installed through Helm and need admin privileges to create objects like services, serviceaccounts, namespaces and RBAC.

A basic system requirement is that an image pull secret to `icr.io` or `us.icr.io` is available. An image for each component is pushed in each region. 


Helm charts values can be found [here](https://github.ibm.com/hybrid-cloud-infrastructure-research/autopilot/tree/main/autopilot-daemon/helm-charts/autopilot).

By default, it will create a namespace named `autopilot` where to run the components. Users workloads do not run in the autopilot namespace. The creation of the namespace can be disabled by setting `create` to false in the namespace block of the `Values.yaml` file.

```yaml
namespace: 
  create: true
  name: autopilot
```

- To pull images from `cil15` registry, the admin needs to add `imagePullSecret` data in one of the helm charts. Both webhook and controller have such entry. It is possible to avoid the creation of the pull secret by setting the value `create` to false in the imagePullSecret block, and by setting the name of the one that will be used (i.e., `all-icr-io`).

```yaml
pullSecrets:
  create: true
  name: autopilot-pull-secret
  imagePullSecretData: 
```

Autopilot needs to run on GPU nodes, but it does not request any GPU in the resource requirements.
To avoid landing on non-GPU nodes (e.g., worker nodes dedicated to infrastructure components), we recommend to have the following in the Helm chart values. It is enabled by default.

```yaml
nodeSelector:
  nvidia.com/gpu.present: 'true'
```  
 
In summary, the recommended commands are as follows. Notice that Helm will install the chart in the current namespace (check with `oc project`):

```bash
git clone https://github.ibm.com/hybrid-cloud-infrastructure-research/autopilot.git
```

Update values for the Helm chart, then:

```bash
helm install autopilot autopilot-daemon/helm-charts/autopilot 
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

### Uninstall

```bash
 helm uninstall autopilot % -n <namespace-where-chart-resides>
```

