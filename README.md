# AI Training Autopilot
This repository is part of the effort described in the IBM Research Challenge #5160.

The goal of this challenge is to enable the OpenShift container platform to become the premier platform to orchestrate the full life cycle of Foundation Model workflows (pre-processing, training, adaptation/distillation, and inference) seamlessly across public, private, and on-prem cloud environments.

From operation perspective, infrastucture stability is always important. We actually saw the various errors and anomaly states in GPU and Network, for instance, so it becomes crucial to provide a tool to detect, avoid, and handle the infrastructure issues while running the AI training job. 

We will provide a collection of tools (named Autopilot) to steer and address these infrastructure issues automatically by pre-flight checks, in-flight checks, and also post-flight to learn or improve the issue detection logic. 

The toolkit will provide pre-flight, in-flight and post-flight checks. In more details (list subject to changes):

- pre-flight checks

  - validate infrastructure before the start of jobs

  - swaps any sub-optimal components

- in-flight checks

  - workload and system performance is continuously monitored

  - detect anomaly, decide to continue or stop the job

  - issue alert to end users

- post-flight learning

  - improve anomaly detection based on infrastructure validation data

### Pre-Flight check
The current status of the [Autopilot toolkit](https://github.ibm.com/hybrid-cloud-infrastructure-research/autopilot-mutating-webhook#autopilot-mutating-webhook) includes:

- A Mutating Webhook to inject a pre-flight container to jobs before they are executed
- The PCIe NVIDIA bandwidth test to check host-to-device connection locally to each node, distributed via Docker container
- The memory test is a cuda program performing `daxpy` and `cuda_dgemm` reporting host to device and device to host memory bandwidth measurements, HBM bandwidth, along with other information about temperature, power usage and clock speed
- A HealthCheckReport Custom Resource Definition (CRD) and a controller that takes action based on the bandwidth test result

The Mutating Webhook and HealthCheckReport Operator are linked in this repository as submodules.
Please follow the links to get more information about each sub-project.

The image below shows the current execution flow of a pre-flight check 

![execflow-autopilot](https://media.github.ibm.com/user/96687/files/8fa9e470-7007-4d5a-af7a-fb66d7da5429)

At a high level, the flow is the following (omitting the MCAD part for simplification):

- A job is created by the user, containing the label `autopilot:""`.
- The mutating webhook will check if the pods are also requesting GPUs. If so, it will inject the init container with the PCIe bandwidth test.
- At execution time, each pod will first run the health check container. If the test will succeed, then the pod will keep running normally.
<!-- - If the test fails, the init container will create a HealthCheckReport CRD indicating the result of the test and the node involved. Also, the pod will label itself with `deschedule` so that it can be removed from the faulty node. -->
- The HealthCheckReport controller will get the report created in the previous step. It will then cordon the node if this option is selected at installation time. Eventually, the scheduler will try to place the pod somewhere else until a good node is found. If all the nodes are bad, the job will never run and the bad nodes are all marked as unschedulable.

### RBAC, Roles and Service Accounts

For the init containers to run correctly, the Webhook will create a service account along with some RBAC, and the service account will be attached to the workload. 
This is needed because the init container might need to create an HealthCheckReport object (`"create", "get", "list"` verbs)

Those operations are namespaced, that is, the webhook creates a Role, a RoleBinding and a Service Account that are local to the namespace where the workload is running.

These objects will remain in the namespace unless manually deleted or if automation is implemented to delete such objects.

### Health check report objects

In the event that a report is issued, it will be in the namespace where the workload is running because the `create` verb is namespaced. Users cannot delete those objects unless the admin gives them permission to. Also, the relevant node is cordoned and no new workloads will run on it. The node is not flushed, so existing workloads will still be there.

The admin is the only subject that should delete the report object and take actions.
Each object is named after the node and the name is not unique. This means that, if an object exists for `nodeA`, another object for the same node will not be created. This is to avoid generating an unreasonable and not needed amount of API objects. 
Once actions are taken on the relevant node, the admin can proceed with the deletion of the corresponding health check report object.


## Install autopilot (Admin)
**Installation**: Both projects can be installed through Helm and need admin privileges to create objects like services, serviceaccounts, namespaces and RBAC.

A basic system requirement is that an image pull secret to `icr.io` or `us.icr.io` is available. An image for each component is pushed in each region. 

Webhook options are in [this page](https://github.ibm.com/hybrid-cloud-infrastructure-research/autopilot-mutating-webhook#customization-available-to-the-admins). 
CRD options are in [this page](https://github.ibm.com/hybrid-cloud-infrastructure-research/healthcheckoperator#customization).

Helm charts values can be found in these links: [webhook](https://github.ibm.com/hybrid-cloud-infrastructure-research/autopilot-mutating-webhook/tree/main/helm-charts/mutating-webhook/templates) and [controller](https://github.ibm.com/hybrid-cloud-infrastructure-research/healthcheckoperator/tree/main/helm-charts/healthcheckoperator/templates).



### Clone main repository with submodules

The Autopilot components are in this repository, but we incorporated it as a submodule in the [Foundation Model Stack](https://github.ibm.com/ai-foundation/foundation-model-stack).

- By default, it will create a namespace named `autopilot` where to run the components. Users workloads do not run in the autopilot namespace. The creation of the namespace can be disabled in the `autopilot-mutating-webhook/helm-charts/mutating-webhook/values.yaml` file by setting `create` to false  in the namespace block.

```yaml
namespace: 
  create: true
  name: autopilot
```

- To pull images from `cil15` registry, the admin needs to add `imagePullSecret` data in one of the helm charts. Both webhook and controller have such entry. It is possible to avoid the creation of the pull secret by setting the value `create` to false in the imagePullSecret block, and by setting the name of the one that will be used (i.e., `all-icr-io`).

```yaml
pullSecrets:
  create: true
  imagePullSecrets: [name: "mutating-webhook-pull"]
  imagePullSecretData: 
```

The recommended commands are as follows:

```bash
git clone git@github.ibm.com:ai-foundation/foundation-model-stack.git % or clone this repository and skip the next step
git submodule update --init --recursive
cd autopilot
git submodule update --remote
% UPDATE VALUES IN THE HELM CHARTS
make install
```

The controllers should show up in the selected namespace

```bash
$ oc get po -n autopilot
NAME                                                              READY   STATUS      RESTARTS   AGE
autopilot-webhook-webhook-v0-c956bd6c9-9jxzd                      1/1     Running     0          22s
healthcheckreport-controller-manager-7c757d848d-d25mp             2/2     Running     0          89s
```

#### Alternative: install each repository 

```bash
# Mutating Webhook
helm install mw-v0 autopilot-mutating-webhook/helm-charts/mutating-webhook

# Health check report CRD and Controller
helm install hcr-v0 healthcheckoperator/helm-charts/healthcheckoperator
```

### Uninstall

```bash
make uninstall 
% or helm uninstall mw-v0 && helm uninstall hcr-v0
```

### Run a basic example
This example creates a pod with an existing init container and one main container. The webhook will prepend the init container to the list of existing init containers. For instance, in case of PyTorch jobs, the KubeFlow operator will inject some other container. We want the pre-flight check to run before any other container. This test simulates such scenario.

The complete instructions and expected output can be found (here)[https://github.ibm.com/hybrid-cloud-infrastructure-research/autopilot-mutating-webhook#run-the-most-basic-example]. Edit the yaml as needed.
To quickly run it:

```bash
oc create -f autopilot-mutating-webhook/manifests/incomplete-pod.yaml
```


## User

### Automation through Helm charts
Users can enable autopilot through the [Helm charts based automation](https://github.ibm.com/ai-foundation/foundation-model-stack/tree/main/tools/scripts/appwrapper-pytorchjob) in their PytorchJobs.

```yaml
# Autopilot health checks
autopilotHealthChecks: # <optional> List of labels enabling one or more system health pre-flight checks. 
For the time being, we only provide the host-to-device PCIe bandwidth test, which is checking that the expected bandwidth is above 4Gb/s and that all GPUs are accessible through nvidia-smi. 
The test runs in an init container of the PytorchJob. 
The pod will be deleted if the health check fails, and MCAD will try to reschedule it. 
The init container is added only if all GPUs in the node are requested, in order to guarantee the correctness of the result (i.e., no other jobs are using the GPUs on the node). 
Other health check tests will be added in the future. 
NOTE: if autopilot is not installed by the admin, those labels have no effect.
  # - gpu-pcie-bw
```

### Manual enablement
Users can use the health checks by adding a few labels in their job yaml files.

1. **Enable all existing health checks**
PCI-e and memory tests are both enabled by default when adding the `autopilot: ""` label in the job's metadata. 

```yaml
labels:
  autopilot: ""
```

2. **Enable a subset of health checks**
Users can decide to get only one of the two tests by adding an extra label in the metadata along with the `autopilot:""` one, which is mandatory to enable the injection of the init containers.

For example, to enable admission webhook and PCI-e BW test only:

```yaml
labels:
  autopilot: ""
  gpu-pcie-bw: ""
```

Otherwise, to enable the memory test only:

```yaml
labels:
  autopilot: ""
  gpu-mem: ""
```
