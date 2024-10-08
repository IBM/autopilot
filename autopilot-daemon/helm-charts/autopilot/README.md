# Helm Chart Customization

Autopilot is set to run on NVidia GPU nodes. It is possible to run it on heterogeneous nodes (i.e., CPU only and GPU only), GPU only nodes or CPU only nodes.

```yaml
onlyOnGPUNodes: true
```

Running on GPU nodes only, will:

1) add the `nvidia.com/gpu.present: 'true'` label and 
2) enable the init container, which checks on the nvidia device plug-in to be setup

Alternatively, `onlyOnGPUNodes` can be set to false and Autopilot will run on all worker nodes, regardless of the accelerators.
Notice that, in this heterogeneous case, the GPU health checks will error out in the non-GPU nodes.

If you do not want to create a new namespace and use an existing one, then set `create: false` and specify the namespace name.
On OpenShift, please ntice that you **must** label the namespace `oc label ns <namespace> openshift.io/cluster-monitoring=true` to have Prometheus scrape metrics from Autopilot.

- To pull the image from a private registry, the admin needs to add `imagePullSecret` data in one of the helm charts. It is possible to avoid the creation of the pull secret by setting the value `create` to false in the imagePullSecret block, and by setting the name of the one that will be used (i.e., `autopilot-pull-secret`).

```yaml
pullSecrets:
  create: true
  name: autopilot-pull-secret
  imagePullSecretData: <base64 encoded-key>
```

- Autopilot runs tests periodically. The default is set to every hour and 4 hours for regular and deep diagnostics respectively, but these can be customized be changing the following

```yaml
repeat: <hours> # periodic health checks timer (default 1h)
invasive: <hours> # deeper diagnostic timer (default 4h, 0 to disable)
```

- PCIe bandwidth critical value is defaulted to 4GB/s. It can be customized by changing the following

```yaml
PCIeBW: <val>
```

- If secondary nics are available by, for instance, Multus or Multi-Nic-Operator, those can be enabled in autopilot by setting the following

```yaml
annotations:
  k8s.v1.cni.cncf.io/networks: <network-config-name>
```

- The list of periodic health checks can be customized through an environment variable. In the example below, we select all health checks and specify the storage class for the `pvc` test

If running on CPU nodes only, `pciebw,remapped,dcgm and gpupower` can be removed

```yaml
env:
  - name: "PERIODIC_CHECKS"
    value: "pciebw,remapped,dcgm,ping,gpupower,pvc"
  - name: "PVC_TEST_STORAGE_CLASS"
    value: "example-storage-class"
```

All these values can be saved in a `config.yaml` file, which can be passed to the `helm` install command

```bash
helm upgrade autopilot autopilot/autopilot --install --namespace=autopilot --create-namespace -f your-config.yml
```
