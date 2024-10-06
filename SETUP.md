
# Install Autopilot

Autopilot can be installed through Helm and need enough privileges to create objects like services, serviceaccounts, namespaces and relevant RBAC.

**If on OpenShift**, after completing the installation, manually label the namespace to enable metrics to be scraped by Prometheus with `label ns autopilot "openshift.io/cluster-monitoring"=`

## Requirements

- Install `helm-git` plugin

```bash
helm plugin install https://github.com/aslafy-z/helm-git --version 0.15.1
```

## Helm Chart customization

Helm charts values and how-to for customization can be found [here](https://github.com/IBM/autopilot/autopilot-daemon/helm-charts/autopilot).

## Install

1) Add autopilot repo

```bash
helm repo add autopilot git+https://github.com/IBM/autopilot.git@autopilot-daemon/helm-charts/autopilot?ref=gh-pages
```

2) Install autopilot (idempotent command). The config file is for customizing the helm values. Namespace is where the helm chart will live, not the namespace where Autopilot runs

```bash
helm upgrade autopilot autopilot/autopilot --install --namespace=autopilot --create-namespace -f your-config.yml
```

The controllers should show up in the selected namespace

```bash
kubectl get po -n autopilot
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
 helm uninstall autopilot -n autopilot
```
