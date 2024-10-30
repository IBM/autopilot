# Network Validation Tests

Autopilot provides two network validation tests:

- Reachability: runs `ping` against all network interfaces available in all the Autopilot pods
- Bandwidth: runs `iperf3` to validate the network bandwidth available.

## Iperf

This test, in it's current form, is primarily for running `TCP` `data plane` `port-to-port` network workloads to gather key performance statistics. This performs a `Ring Traversal` (or as we call it, a `ring workload`) through all network interfaces (net1-X interfaces) at varying intensity (number of simultaneous client & servers per interface). In future versions of Autopilot, more workloads and customization to the workloads may be provided.

### Ring workload
A "Ring Workload", in our case is similar the commonly known "Ring Topology" such that the execution calls flow sequentially in a   particular _direction_ that forms a "ring" like pattern. _Most importantly, none of the the compute infrastructure is actually configured in a ring, we merely develop workloads that resemble a ring pattern._ The motivation for these workloads is to achieve full line rate throughput on a port-by-port (in our case network interfaces net1-X) basis for a single logical cluster.

Assume we have the following set of nodes `[A,B,C]`.  We can create a `ring` starting from node `A` that flows to the direction of `C`:

```console
A -> B
B -> C
C -> A
```

In our case, a "Ring Workload" will exhaust all starting pointings. We call these iterations, `timesteps`. In a compute infrastructure with `n` number of nodes, we can say there will be `n-1` total timesteps. Said differently, there's `n-1` possible starting points that form a ring such that no node flows to itself.  Each of the pairs of execution in a given timestep will execute in parallel.

```console
Timestep 1:
------------
A -> B
B -> C
C -> A

Timestep 2:
------------
A -> C
B -> A
C -> B
```

As part of this workload, Autopilot will generate the Ring Workload and then start `iperf3 servers` on each interface on each Autopilot pod based on the configuration options provided by the user.  Only after the `iperf3 servers` are started, Autopilot will begin executing the workload by starting `iperf3 clients` based on the configuration options provided by the user. All results are logged back to the user.

For each network interface on each node, an `iperf3 server` is started. The number of `iperf3 servers` is dependent on the `number of clients` intended on being run. For example, if the  `number of clients` is `8`, then there will be `8` `iperf3 servers` started per interface on a unique `port`.

For each timestep, all `pairs` are executed simultaneously. For each pair some `number of clients` are started in parallel and will run for `5 seconds` using `zero-copies` against a respective `iperf3 server`

Metrics such `minimum`, `maximum`, `mean`, `aggregate` bitrates and transfers are tracked for both the `sender` and the `receiver` for each `client -> server` execution. The results are stored both as `JSON` in the respective `pod` as well as summarized and dumped into the `pod logs`.

Invocation from the exposed Autopilot API is as follows below:

```bash
    # Invoked via the `status` handle:
curl "http://autopilot-healthchecks-autopilot.<domain>/status?check=iperf&workload=ring&pclients=<NUMBER_OF_IPERF3_CLIENTS>&startport=<STARTING_IPERF3_SERVER_PORT>"

    # Invoked via the `iperf` handle directly:
curl "http://autopilot-healthchecks-autopilot.<domain>/iperf?workload=ring&pclients=<NUMBER_OF_IPERF3_CLIENTS>&startport=<STARTING_IPERF3_SERVER_PORT>"
```
