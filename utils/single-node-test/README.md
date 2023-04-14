# Single Node PCIe Bandwidth Test

This Helm chart can be used to launch a Pod on a targeted node, and run the pcie-bw test on it.

The Helm chart can be configured by updating the following values in `values.yaml`:

- `namespace` where to run the Pod. It is set to `default`. The namespace needs to have a valid `ImagePullSecret` to get images from `us.icr.io` or `icr.io`
- `imagePullSecret` defaulted to `all-icr-io`
- `nodename` where to schedule the Pod
- `bw` minimum bandwidth value that is acceptable. Anything lower than that, will create a Health Check Report. It is set to `4`

To run the Pod:

```bash
helm install pcie-test utils/single-node-test/charts
```

To uninstall:

```bash
helm uninstall pcie-test
```
