# Run Health Checks on All Nodes

This Helm chart can be used to launch a DaemonSet on a the cluster, and run the pcie-bw and memory-bw test on all nodes. Notice that, if some nodes don't have all GPUs available, the checks are skipped.
The `gpu-mem` container will check run a longer test, which may take a few minutes.

Since a DaemonSet runs on all nodes, if some nodes do not have all GPUs available, the pods will stay pending.

Thus it is important to keep an eye on the DaemonSet and delete it once the targeted nodes have been tested, so that all GPUs occupied by the DaemonSet can be released.

The Helm chart can be configured by updating the following values in `values.yaml`:

- `namespace` where to run the Pod. It is set to `default`. The namespace needs to have a valid `ImagePullSecret` to get images from `us.icr.io` or `icr.io`
- `imagePullSecret` defaulted to `all-icr-io`

To run the Pod:

```bash
helm install all-nodes-test utils/all-nodes-test/charts
```

Once the pods complete all the tests, a successful run looks like this

```bash
claudias-air:autopilot cmisale$ oc get po
NAME                            READY   STATUS    RESTARTS   AGE
autopilot-allnodes-test-42hp9   1/1     Running   0          6m53s
autopilot-allnodes-test-hnx25   1/1     Running   0          6m55s
autopilot-allnodes-test-vfn42   1/1     Running   0          6m57s
autopilot-allnodes-test-wwl2s   1/1     Running   0          6m56s

claudias-air:autopilot cmisale$ oc logs logs autopilot-allnodes-test-42hp9 -c completion
Thu Apr 13 19:59:01 UTC 2023
All tests completed successfully
```

To uninstall:

```bash
helm uninstall all-nodes-test
```
