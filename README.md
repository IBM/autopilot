# AI Training Auto-Pilot
This repository is part of the effort described in the IBM Research Challenge #5160.

The goal of this challenge is to enable the OpenShift container platform to become the premier platform to orchestrate the full life cycle of Foundation Model workflows (pre-processing, training, adaptation/distillation, and inference) seamlessly across public, private, and on-prem cloud environments.

From operation perspective, infrastucture stability is always important. We actually saw the various errors and anomaly states in GPU and Network, for instance, so it becomes crucial to provide a tool to detect, avoid, and handle the infrastructure issues while running the AI training job. 

We will provide a collection of tools (named Auto-Pilot) to steer and address these infrastructure issues automatically by pre-flight checks, in-flight checks, and also post-flight to learn or improve the issue detection logic. 

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
The current status of the [Auto-Pilot toolkit](https://github.ibm.com/hybrid-cloud-infrastructure-research/autopilot-mutating-webhook#autopilot-mutating-webhook) includes:

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
- If the test fails, the init container will create a HealthCheckReport CRD indicating the result of the test and the node involved. Also, the pod will label itself with `deschedule` so that it can be removed from the faulty node.
- The HealthCheckReport controller will get the report created in the previous step. It will then `delete` the pod marked with `deschedule` label and cordon the node. This way, the scheduler will try to place the pod somewhere else until a good node is found. If all the nodes are bad, the job will never run and the bad nodes are all marked as unschedulable.

## Initialize this repository
Collection of tool for the Auto-pilot project.
This repo contains submodules. You can pull them by running 
```
make submodule-init
```


To build the GPU bandwidth test:
```
make gpu-bw-image 
```

To build the GPU memory test
```
make gpu-mem-image
```

# Install the Auto-pilot components
For detailed instructions and information about the Mutating Webhook, HealthCheckReport CRD and Controller, please refer to the relevant submodules' README.

The init container injected by the mutating webhook, needs the HealthCheckReport CRD being available. Therefore, we install the CRD first.
For a quick deploy, run

```
helm install hrr-operator-v0 healthcheckoperator/helm-charts/healthcheckoperator/
helm install webhook-v0 autopilot-mutating-webhook/helm-charts/mutating-webhook/
```

The controllers should show up in the default namespace
```
claudias-air:autopilot cmisale$ oc get po
NAME                                                              READY   STATUS      RESTARTS   AGE
autopilot-webhook-webhook-v0-c956bd6c9-9jxzd                      1/1     Running     0          22s
healthcheckreport-controller-manager-7c757d848d-d25mp             2/2     Running     0          89s
```

## Run a basic example
This example creates a pod with an existing init container and one main container. The webhook will prepend the init container to the list of existing init containers. For instance, in case of PyTorch jobs, the KubeFlow operator will inject some other container. We want the pre-flight check to run before any other container. This test simulates such scenario.

The complete instructions and expected output can be found (here)[https://github.ibm.com/hybrid-cloud-infrastructure-research/autopilot-mutating-webhook#run-the-most-basic-example].
To quickly run it:
```
oc create -f autopilot-mutating-webhook/manifests/incomplete-pod.yaml
```

### Important side effect! Do not skip this paragraph!

It is **VERY IMPORTANT** to remember that the entire pipeline of webhook+operator may disable one or more nodes by marking them as `unschedulable`. 

Depending on the result of the `incomplete-pod` test, this may happen. 

Make sure to check the nodes with `oc get nodes` and `oc uncordon <nodename>` the ones affected by the test.


# Uninstall 
To remove the charts, just run the `uninstall` command with the chosen release names for the webhook and the controller.

```
helm uninstall hrr-operator-v0
helm uninstall webhook-v0
```
