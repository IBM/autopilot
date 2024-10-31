from iperf3_utils import *


#
# TODO: Make this an abstract class...
#
#


class NetworkWorkload:
    def __init__(self, namespace=None, workload_name="Ring Topology"):
        self.namespace = namespace or os.getenv("NAMESPACE")
        self.workload = workload_name
        self.log = logging.getLogger(__name__)
        logging.basicConfig(
            format="[NETWORK] - [WORKLOAD-GEN] - [%(levelname)s] : %(message)s",
            level=logging.INFO,
        )

        try:
            config.load_incluster_config()
            self.v1 = client.CoreV1Api()
        except:
            self.log.error("Failed to load Kubernetes CoreV1API.")
            exit(1)

    def get_all_ifaces(self):
        address_map = {}

        try:
            autopilot_pods = self.v1.list_namespaced_pod(
                namespace=self.namespace, label_selector="app=autopilot"
            )
        except ApiException as e:
            self.log.error(
                "Exception when calling CoreV1Api->list_namespaced_pod: %s\n" % e
            )
            exit(1)
        entrylist = json.loads('{}')
        for pod in autopilot_pods.items:
            try:
                entrylist = json.loads(
                    pod.metadata.annotations["k8s.v1.cni.cncf.io/network-status"]
                )
            except KeyError:
                log.info(
                    f'Key k8s.v1.cni.cncf.io/network-status not found on pod "{CURR_POD_NAME}" on "{CURR_WORKER_NODE_NAME}"')
            if len(entrylist) > 0:
                for entry in entrylist:
                    try:
                        iface = entry["interface"]
                    except KeyError:
                        self.log.info("Interface key name not found, assigning 'k8s-pod-network'.")
                        iface = "k8s-pod-network"
                    if address_map.get(iface) == None:
                        address_map[iface] = []
                    address_map.get(iface).append((pod.spec.node_name, entry["ips"]))
            else:
                pod_ips = pod.status.pod_i_ps
                if pod_ips != None:
                    iface = "default"
                    if address_map.get(iface) == None:
                        address_map[iface] = []
                    ips = []
                    for pod_ip in pod_ips:
                        ips.append(pod_ip.ip)
                    address_map.get(iface).append((pod.spec.node_name, ips))

        if len(address_map) == 0:
            self.log.error("No interfaces found. FAIL.")
        return address_map

    def gen_autopilot_node_map_json(self):
        #
        # TODO: This is bad because it gets all endpoints, but what happens if
        # we have a failing worker that doesn't have any pods?
        #
        # Well we skip it...this bad...why? Well, the user won't know...
        #
        # Proposal, warn the user at least that NOT ALL work nodes will be tested...
        #
        try:
            endpoints = self.v1.list_namespaced_endpoints(
                self.namespace,
                field_selector="metadata.name=autopilot-healthchecks",
            )
        except ApiException as e:
            self.log.error(
                "Exception when calling Kubernetes CoreV1Api->list_namespaced_endpoints: %s\n"
                % e
            )
            exit(1)

        autopilot_node_map = {}
        for endpointslice in endpoints.items:
            addresses = endpointslice.subsets[0].addresses
            for item in addresses:
                node_name = item.node_name
                if node_name not in autopilot_node_map:
                    pod_name = item.target_ref.name
                    ip_address = item.ip
                    autopilot_node_map[node_name] = {
                        "pod": pod_name,
                        "endpoint": ip_address,
                    }

        addresses = self.get_all_ifaces()
        for add in addresses:
            if add != "eth0":
                for entry in addresses.get(add):
                    worker_node_name = entry[0]
                    net_interfaces = entry[1]
                    if worker_node_name in autopilot_node_map:
                        autopilot_node_map[worker_node_name][
                            "netifaces"
                        ] = net_interfaces

                return autopilot_node_map

    def generate_ring_topology_json(self, worker_nodes_map):
        pair_links = {}
        node_count = len(worker_nodes_map)
        if node_count > 1:
            worker_nodes = list(worker_nodes_map.keys())
            for t in range(1, node_count):
                step_pairs = []
                for i in range(node_count):
                    source = worker_nodes[i]
                    target = worker_nodes[(i + t) % node_count]
                    step_pairs.append({source: target})
                pair_links[t] = step_pairs
        return pair_links

    def print_autopilot_node_map_json(self, worker_node_map):
        self.log.info(f"\n{json.dumps(worker_node_map, indent=4)}")

    def print_ring_topology_json(self, ring_workload):
        output = ""
        for step in ring_workload:
            output += f"Time Step {step}:\n"
            for pair in ring_workload[step]:
                for source, target in pair.items():
                    output += f"    {source} -> {target}\n"
        self.log.info(f"\n{output}")

    def print_ring_workload(self):
        autopilot_node_map_json = self.gen_autopilot_node_map_json()
        ring_workload_pairs_json = self.generate_ring_topology_json(
            autopilot_node_map_json
        )
        output = ""
        for step in ring_workload_pairs_json:
            output += f"Time Step {step}\n"
            for pair in ring_workload_pairs_json[step]:
                for source, dest in pair.items():
                    output += (
                        f"    Pod-to-Pod: {autopilot_node_map_json[source]['pod']} "
                        f"-> {autopilot_node_map_json[dest]['pod']}\n"
                        f"        Endpoint-to-Endpoint: {autopilot_node_map_json[source]['endpoint']} -> "
                        f"{autopilot_node_map_json[dest]['endpoint']}\n"
                    )
            output += f"\n"
        self.log.info(f"\n{output}")
