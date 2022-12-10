# Auto-pilot
Collection of tool for the Auto-pilot project.
This repo contains submodules. You can pull them by running 
```
make submodule-init
```


To build the GPU bandwidth test:
```
make gpu-bw-image 
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
