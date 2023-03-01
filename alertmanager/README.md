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

## Create an `AlertManager` receiver using the OpenShift Web UI
- Log into the OpenShift WebUI as an admin
- Click on Administration -> Cluster Settings -> Configuration -> Alertmanager

You should see this page:
![Alert Manager](images/alertmanager.png)

Click on `Create Receiver`

- Choose a Receiver name and set the Receiver type as Slack
- Click on `Create` and fill out the following fields:

  - Paste the Slack Webhook URL you copied in the previous step into the `Slack API URL` field
  - Write the Slack channel name to send notifications to in the `Channel field`
  - Click on `Show advanced configuration`
    - If you want to set a custom Slack username for the AlertManager such as `Cordoned Node Alert` and a custom icon such as a Slack emoji (I used `:warning:`), you can set those fields there.
    - The title can remain as default title.
    - The required field to set here is the Text field - this is what the Slack message for the notification will be. I set mine to `Node is cordoned - check node health`
  - In the Routing Labels section, provide the label that we provided in the Prometheus `AlertingRule` in the first step.
    - Set Name to `cordon`
    - Set Value to `autopilot`

    This ensures that Prometheus will route the `AlertingRule` we created to this specific `AlertManager` receiver. 

- Click on `Save`

This will generate a yaml file like `alertmanager.yaml` in this folder and will update the `AlertManager` pod configuration to add your new receiver. Now we will start receiving alerts from the Prometheus `AlertingRule` we created. Note that in `alertmanager.yaml`, I added a `default` receiver to catch all alerts that Prometheus fires and route them to my `#random` channel in my Slack workspace. This is optional.

You can check the status of the `AlertManager` pod with this command:
```console
$kubectl logs alertmanager-main-0 -c alertmanager -n openshift-monitoring
```

Once the AlertManager is updated, you can test the notifications by cordoning a node:
```console
kubectl get nodes
kubectl cordon a100-huge-m25p7-worker-2-2tfzd
```
You should see a Slack message in the channel you specified in your personal Slack workspace like this:
![Slack notif](images/notif.png)

Now if you uncordon that node and there are no cordoned nodes left, you should see a `Resolved` message like this:
![Slack resolved](images/resolved.png)

That's it! Now you can get notifications in Slack everytime a node is cordoned on your OpenShift cluster. If there is something else you wish to get notification for, you simply need to create a new `PrometheusRule` with a new `expr` and label, and create a new `AlertManager` Slack receiver with a matching label. 





