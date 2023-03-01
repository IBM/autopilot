# Slack Notifications for cordoned nodes with Alert Manager

This folder contains the files needed to set up Slack notifications for cordoned nodes using Prometheus and AlertManager on OpenShift. 

There are 3 main steps to set it up:
1) Create a Prometheus `AlertingRule`
2) Create a Slack webhook application
3) Create an `AlertManager` Receiver

These steps are explained in more detail below.

## Create an `AlertingRule` for Prometheus
```console
kubectl create -f alertingrule.yaml
```

Note the following lines in the alerting rule below:

- The `PrometheusRule` is created in the `openshift-monitoring` namespace - this is the namespace where Prometheus and Alert manager is deployed on the OpenShift cluster.
- A `cordon: autopilot` label is added - this is important to match the alert with an Alert Manager receiver that we will create in the last step. This is how Prometheus knows which Alert Manager receiver to send the alert to. 
- `expr: sum (kube_node_spec_unschedulable) by (node) > 0` is the expression used to count how many nodes are unschedulable (or cordoned) and if it's more than 0
```
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: cordon-node-alert
  namespace: openshift-monitoring
  labels:
    cordon: autopilot
spec:
  groups:
  - name: example
    rules:
    - alert: CordonAlert
      annotations:
        description: A node has been cordoned - check health of nodes
        summary: A node has been cordoned - check health of nodes
      expr: sum (kube_node_spec_unschedulable) by (node) > 0
      for: 1m
      labels:
        severity: critical
        cordon: autopilot
```
## Create a Slack webhook application
- Create a Slack workspace using your personal Slack account (not your IBM Slack)
- Go to https://slack.com/apps and select your workspace from the dropdown menu in the top right of the page
- Click on `Get Essential Apps` and search the App Directory for `Incoming WebHooks`

You should see a page like this:
![Slack Webhook](images/slack.png)

- Click on `Add to Slack` and choose which Slack channel to post messages to from the dropdown menu or create a new channel.

- Click on `Add Incoming Webhooks Integration`

- Copy and paste the `WebhookURL`. We will use this when we configure the `AlertManager` Receiver in the next step.
It should look something like this:
```
https://hooks.slack.com/services/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
```

## Create an `AlertManager` receiver using the OpenShift UI

