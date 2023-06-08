# AI Training Autopilot
This repository is part of the effort described in the IBM Research Challenge #5160.

The goal of this challenge is to enable the OpenShift container platform to become the premier platform to orchestrate the full life cycle of Foundation Model workflows (pre-processing, training, adaptation/distillation, and inference) seamlessly across public, private, and on-prem cloud environments.

From operation perspective, infrastucture stability is always important. We actually saw the various errors and anomaly states in GPU and Network, for instance, so it becomes crucial to provide a tool to detect, avoid, and handle the infrastructure issues while running the AI training job. 

We provide a collection of tools (named Autopilot) to steer and address these infrastructure issues automatically by pre-flight checks, in-flight checks, and also post-flight to learn or improve the issue detection logic. 

Autopilot runs as a DaemonSet on all worker nodes that have GPUs.

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

<!-- - A Mutating Webhook to inject a pre-flight container to jobs before they are executed -->
- The PCIe NVIDIA bandwidth test to check host-to-device connection on each node
<!-- - The memory test is a cuda program performing `daxpy` and `cuda_dgemm` reporting host to device and device to host memory bandwidth measurements, HBM bandwidth, along with other information about temperature, power usage and clock speed -->
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


<!-- ### RBAC, Roles and Service Accounts

For the init containers to run correctly, the Webhook will create a service account along with some RBAC, and the service account will be attached to the workload. 
This is needed because the init container might need to create an HealthCheckReport object (`"create", "get", "list"` verbs)

Those operations are namespaced, that is, the webhook creates a Role, a RoleBinding and a Service Account that are local to the namespace where the workload is running.

These objects will remain in the namespace unless manually deleted or if automation is implemented to delete such objects.

### Health check report objects

In the event that a report is issued, it will be in the namespace where the workload is running because the `create` verb is namespaced. Users cannot delete those objects unless the admin gives them permission to. Also, the relevant node is cordoned and no new workloads will run on it. The node is not flushed, so existing workloads will still be there.

The admin is the only subject that should delete the report object and take actions.
Each object is named after the node and the name is not unique. This means that, if an object exists for `nodeA`, another object for the same node will not be created. This is to avoid generating an unreasonable and not needed amount of API objects. 
Once actions are taken on the relevant node, the admin can proceed with the deletion of the corresponding health check report object.

 -->
## Install autopilot (Admin)
**Installation**: Both projects can be installed through Helm and need admin privileges to create objects like services, serviceaccounts, namespaces and RBAC.

A basic system requirement is that an image pull secret to `icr.io` or `us.icr.io` is available. An image for each component is pushed in each region. 

<!-- Webhook options are in [this page](https://github.ibm.com/hybrid-cloud-infrastructure-research/autopilot-mutating-webhook#customization-available-to-the-admins). 
CRD options are in [this page](https://github.ibm.com/hybrid-cloud-infrastructure-research/healthcheckoperator#customization).
 -->
Helm charts values can be found [here](https://github.ibm.com/hybrid-cloud-infrastructure-research/autopilot/tree/main/autopilot-daemon/helm-charts/autopilot).

By default, it will create a namespace named `autopilot` where to run the components. Users workloads do not run in the autopilot namespace. The creation of the namespace can be disabled in the `autopilot-mutating-webhook/helm-charts/mutating-webhook/values.yaml` file by setting `create` to false  in the namespace block.

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
 
The recommended commands are as follows:

```bash
git clone git@github.ibm.com:ai-foundation/foundation-model-stack.git % or clone this repository and skip the next step
% UPDATE VALUES IN THE HELM CHARTS
make install
% or helm install autopilot-daemon/helm-charts/autopilot
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
make uninstall 
% or helm uninstall autopilot
```

## Run health checks

Health checks can be executed through a utility tool provided with a Helm chart, or by querying the Autopilot service.
Results can be visualized by either checking the logs of the utility tool/service query, or by looking at the data in a Grafana dashboard.
The relevant `json` file can be found [here](https://github.ibm.com/hybrid-cloud-infrastructure-research/autopilot/blob/main/utils/Autopilot-Grafana-Dashboard.json)

### Helm Chart

Users and admins can create a single pod that can run the desired health checks.
Please refer to the [dedicated page](https://github.ibm.com/hybrid-cloud-infrastructure-research/autopilot/tree/main/utils/system-check) for more details and customization.

### Query the Autopilot Service

Autopilot provides a `/status` handler that can be queried to get the entire system status, meaning that it will run all the tests on all the nodes. Autopilot is reachable by service name `autopilot-healthchecks.autopilot.svc` in-cluster only, meaning it can be reached from a pod running in the cluster.

Tests can be tailored by a combination of:

- `host=<hostname>`, to run all tests on a specific node
- `check=<healthcheck>`, to run a single test (`pciebw`, `nic` and `remapped`)

For example:

```bash
curl "http://autopilot-healthchecks.autopilot.svc:3333/status?host=dev-ppv5g-worker-3-with-secondary-jdf7b&check=nic"
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0
Asking to run on remote node dev-ppv5g-worker-3-with-secondary-jdf7bEndpointSlice: autopilot-healthchecks

Endpoint: 10.129.12.39

Node:  dev-ppv5g-worker-3-with-secondary-jdf7b

url:  http://10.129.12.39:3333/status?host=dev-ppv5g-worker-3-with-secondary-jdf7b&check=nic

Response: 
 Checking system status of host dev-ppv5g-worker-3-with-secondary-jdf7b (localhost) 

[[ NETWORK ]] Evaluating reachability of Multi-NIC CNI.
===== Health Status of dev-ppv5g-worker-3-with-secondary-jdf7b =====
Allocatable network devices: 2/2
Connectable network devices: 2/2
Host is OK (all functional and connected).
Reported by multi-nic-cni-health-checker-5cfb794496-57vtw at 2023-06-08T20:12:22Z

[[ NETWORK ]] SUCCESS

dev-ppv5g-worker-3-with-secondary-jdf7b 1 1


Node Status:  Ok

-------------------------------------

Node Summary: 

 dev-ppv5g-worker-3-with-secondary-jdf7b :  Ok
```
###
