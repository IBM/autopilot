# Helm Chart Customization

## Latest tag

At every PR merge, we automatically build the `latest` tag that can be pulled by using `quay.io/autopilot/autopilot:latest`.

This tag contains the latest changes and it must be considered as a dev image. For stable releases, always refer to the published ones.

## Customize Helm chart

Autopilot is set to run on NVidia GPU nodes. It is possible to run it on heterogeneous nodes (i.e., CPU only and GPU only), GPU only nodes or CPU only nodes.

```yaml
onlyOnGPUNodes: true
```

Running on GPU nodes only, will:

1) add the `nvidia.com/gpu.present: 'true'` label and
2) enable the init container, which checks on the nvidia device plug-in to be setup

Alternatively, `onlyOnGPUNodes` can be set to false and Autopilot will run on all worker nodes, regardless of the accelerators.
Notice that, in this heterogeneous case, the GPU health checks will error out in the non-GPU nodes.

- Autopilot runs tests periodically. The default is set to every hour and 4 hours for regular and deep diagnostics respectively, but these can be customized be changing the following

```yaml
repeat: <hours> # periodic health checks timer (default 1h)
invasive: <hours> # deeper diagnostic timer (default 4h, 0 to disable)
```

- The list of GPU errors considered fatal as a result of a dcgmi run, can be customized through the `DCGM_FATAL_ERRORS` environment variable. This is used to label nodes with extra WARN/EVICT labels. The list defaults to [PCIe,NVLink,ECC,GPU Memory] and refers to https://docs.nvidia.com/datacenter/dcgm/latest/user-guide/feature-overview.html#id3

```yaml
  - name: "DCGM_FATAL_ERRORS"
    value: ""
```

- Invasive jobs (e.g., dcgm level 3), are executed as separate job. The job deletes itself by default after 30s. This parameter can be customized by the env variable below

```yaml
  - name: "INVASIVE_JOB_TTLSEC"
    value: ""
```

- PCIe bandwidth critical value is defaulted to 4GB/s. It is recommended to set a threshold that is 25% or lower of the expected peak PCIe bandwidth capability, which maps to maximum peak from 16 lanes to 4 lanes. For example, for a PCIe Gen4x16, reported peak bandwidth is 63GB/s. A degradation at 25% is 15.75GB/s, which corresponds to PCIe Gen4x4. The measured bandwidth is expected to be at least 80% of the expected peak PCIe generation bandwidth.

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

All these values can be saved in a `config.yaml` file.

## Install

If you have your own configuration file, it can be passed to the `helm` install command with the `-f` parameter. If you want to install the default values, just omit the parameter.

```bash
helm upgrade autopilot autopilot/autopilot --install --namespace=autopilot --create-namespace <-f your-config.yml>
```

For more customization, please refer to `values.yaml`.
