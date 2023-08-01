# AI Training Autopilot
This repository is part of the effort described in the IBM Research Challenge #5160.

The goal of this challenge is to enable the OpenShift container platform to become the premier platform to orchestrate the full life cycle of Foundation Model workflows (pre-processing, training, adaptation/distillation, and inference) seamlessly across public, private, and on-prem cloud environments.

From operation perspective, infrastucture stability is always important. We actually saw the various errors and anomaly states in GPU and Network, for instance, so it becomes crucial to provide a tool to detect, avoid, and handle the infrastructure issues while running the AI training job. 

We provide a collection of tools (named Autopilot) to steer and address these infrastructure issues automatically by pre-flight checks, in-flight checks, and also post-flight to learn or improve the issue detection logic. 

Autopilot runs as a DaemonSet on all worker nodes that have GPUs. All results from health checks are exposed through Prometheus and a Grafana dashboard is available in the `utility-tools` folder.



![autopilot-pod](https://media.github.ibm.com/user/96687/files/3f513944-2b23-4ce1-92ce-5cbbf5a40f10)


The toolkit will provide pre-flight, in-flight and post-flight checks. In more details (list subject to changes):

- pre-flight checks

  - validate infrastructure before the start of jobs

- in-flight checks

  - workload and system performance is continuously monitored

  - detect anomaly, decide to continue or stop the job

  - issue alert to end users

- post-flight learning

  - improve anomaly detection based on infrastructure validation data

### Health Checks
The current status of the Autopilot includes:

- The PCIe NVIDIA bandwidth test to check host-to-device connection on each node
- A check on GPUs remapped rows
- A check on the multi-nic CNI availability
<!-- - A HealthCheckReport Custom Resource Definition (CRD) and a controller that takes action based on the bandwidth test result -->

<!-- The Mutating Webhook and HealthCheckReport Operator are linked in this repository as submodules.
Please follow the links to get more information about each sub-project.

The image below shows the current execution flow of a pre-flight check 

![execflow-autopilot](https://media.github.ibm.com/user/96687/files/8fa9e470-7007-4d5a-af7a-fb66d7da5429)

At a high level, the flow is the following (omitting the MCAD part for simplification):

- A job is created by the user, containing the label `autopilot:""`.
- The mutating webhook will check if the pods are also requesting GPUs. If so, it will inject the init container with the PCIe bandwidth test.
- At execution time, each pod will first run the health check container. If the test will succeed, then the pod will keep running normally.
<!-- - If the test fails, the init container will create a HealthCheckReport CRD indicating the result of the test and the node involved. Also, the pod will label itself with `deschedule` so that it can be removed from the faulty node. -->


## Run Health Checks

Health checks can be executed through a utility tool provided with a Helm chart, or by querying the Autopilot service.
Results can be visualized by either checking the logs of the utility tool/service query, or by looking at the data in a Grafana dashboard.
The relevant `json` file can be found [here](https://github.ibm.com/hybrid-cloud-infrastructure-research/autopilot/blob/main/utility-tools/Autopilot-Grafana-Dashboard.json)

### Helm Chart

Users and admins can create a single pod that can run the desired health checks.
Please refer to the [dedicated page](https://github.ibm.com/hybrid-cloud-infrastructure-research/autopilot/tree/main/utility-tools/system-check) for more details and customization.

### Query the Autopilot Service

Autopilot provides a `/status` handler that can be queried to get the entire system status, meaning that it will run all the tests on all the nodes. Autopilot is reachable by service name `autopilot-healthchecks.autopilot.svc` in-cluster only, meaning it can be reached from a pod running in the cluster.

Tests can be tailored by a combination of:

- `host=<hostname1,hostname2,...>`, to run all tests on a specific node or on a comma separated list of nodes.
- `check=<healthcheck1,healtcheck2,...>`, to run a single test (`pciebw`, `nic` and `remapped`, or `all`) or a list of comma separated tests. When no parameters are specified, all tests are run.
- `batch=<#hosts>`, how many hosts to check at a single moment. Requests to the batch are run in parallel. Batching is done to avoid running too many requests in parallel when the number of worker nodes increases. Default to 1.

#### Query from a pod
In the example below, we create a utility `nginx` pod from which we can run `curl` commands against the `autopilot-healthchecks` service.
We run the PCIe bandwidth test on all nodes, and we can see it is failing on one node.

```bash
$ kubectl create job curl-pod --image=nginx -- sleep inf
$ kubectl exec jobs/curl-pod -- curl "http://autopilot-healthchecks.autopilot.svc:3333/status?check=pciebw"
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
  0     0    0     0    0     0      0      0 --:--:--  0:00:32 --:--:--     0Checking status on all nodes


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

#### Query with Port-Forward

Alternatively, it is possible to port-forward the autopilot healthchecks Service and `curl` from localhost. 

```bash
$ kubectl port-forward service/autopilot-healthchecks 3333:3333
Forwarding from 127.0.0.1:3333 -> 3333
Forwarding from [::1]:3333 -> 3333
```

Then on another terminal, run the desired curl command. In this example, we target one node and check if the secondary nics are reachable.

```bash
$ curl "http://127.0.0.1:3333/status?host=dev-ppv5g-worker-3-with-secondary-h5vb6&check=nic"
Checking system status of host dev-ppv5g-worker-3-with-secondary-h5vb6 (localhost) 

[[ NETWORK ]] Evaluating reachability of Multi-NIC CNI.
===== Health Status of dev-ppv5g-worker-3-with-secondary-h5vb6 =====
Allocatable network devices: 2/2
Connectable network devices: 2/2
Host is OK (all functional and connected).
Reported by multi-nic-cni-health-checker-5cfb794496-57vtw at 2023-06-09T01:43:15Z

[[ NETWORK ]] SUCCESS

dev-ppv5g-worker-3-with-secondary-h5vb6 1 1

```


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
git clone git@github.ibm.com:ai-foundation/foundation-model-stack.git % or clone this repository and skip the next step
% UPDATE VALUES IN THE HELM CHARTS
helm install autopilot autopilot-daemon/helm-charts/autopilot 
% or make install
```

The controllers should show up in the selected namespace

```bash
$ oc get po -n autopilot
NAME                               READY   STATUS    RESTARTS   AGE
autopilot-daemon-autopilot-g7j6h   1/1     Running   0          70m
autopilot-daemon-autopilot-g822n   1/1     Running   0          70m
autopilot-daemon-autopilot-x6h8d   1/1     Running   0          70m
autopilot-daemon-autopilot-xhntv   1/1     Running   0          70m
```

### Uninstall

```bash
 helm uninstall autopilot % -n <namespace-where-chart-resides>
% or make uninstall (must be in the chart's namespace)
```

