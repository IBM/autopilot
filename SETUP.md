
# Install Autopilot

Autopilot can be installed through Helm and need enough privileges to create objects like services, serviceaccounts, namespaces and relevant RBAC.

## Helm Chart customization

Helm charts values and how-to for customization can be found [here](helm-charts/autopilot/README.md).

## Install

1) Add autopilot repo

```bash
helm repo add autopilot https://ibm.github.io/autopilot/
```

2) Install autopilot (idempotent command). The config file is for customizing the helm values. It is not mandatory. If the default values work for you, omit the `-f`. The `--namespace` parameter says where the helm chart will be deployed

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

## Uninstall

```bash
 helm uninstall autopilot -n autopilot
 kubectl delete namespace autopilot
```

## Enabling Prometheus

### Kubernetes Users

The ServiceMonitor object is the one that enables Prometheus to scrape the metrics produced by Autopilot.
In order for Prometheus to find the right objects, the `ServiceMonitor` needs to be annotated with the Prometheus' release name. It is usually `prometheus`, and that's the default added in the Autopilot release.
If that is not the case in your cluster, the correct release label can be found by checking in the `ServiceMonitor` of Prometheus itself, or the name of Prometheus helm chart.
Then, Autopilot's `ServiceMonitor` can be labeled with the following command

```bash
kubectl label servicemonitors.monitoring.coreos.com -n autopilot autopilot-metrics-monitor release=<prometheus-release-name>
```

### OpenShift Users

**If on OpenShift**, after completing the installation, manually label the namespace to enable metrics to be scraped by Prometheus with the following command:
The `ServiceMonitor` labeling is not required.

```bash
kubectl label ns autopilot openshift.io/cluster-monitoring=true
```
