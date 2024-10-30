from iperf3_utils import *

parser = argparse.ArgumentParser()
parser.add_argument(
    "--numservers",
    type=int,
    default=1,
    help=(
        'The number of servers (on different ports) to have running on a single IP. Note. For "numservers" values greater than 1 '
        'the "startport" value will be adjusted for each subsequently started server by a factor of 1.'
    ),
)

parser.add_argument(
    "--startport",
    type=int,
    default=5200,
    help=(
        'The default port value. In the event that "numservers" is greater than 1, the default port value used '
        "to generate servers will automatically increase to accomdate the clients running in parallel."
    ),
)
args = vars(parser.parse_args())


def main():
    num_server = args["numservers"]
    port = args["startport"]
    interfaces = []
    entrylist = json.loads('{}')

    try:
        config.load_incluster_config()
        v1 = client.CoreV1Api()
    except:
        log.error("Failed to load Kubernetes CoreV1API.")
        exit(1)
    try:
        autopilot_pods = v1.list_namespaced_pod(
                namespace=AUTOPILOT_NAMESPACE, field_selector="metadata.name="+CURR_POD_NAME
                )
    except ApiException as e:
        log.error(
            "Exception when calling CoreV1Api->list_namespaced_pod: %s\n" % e
        )
        exit(1)

    pod = autopilot_pods.items[0]
    try:
        entrylist = json.loads(
                pod.metadata.annotations["k8s.v1.cni.cncf.io/network-status"]
        )
    except KeyError:
        log.info(
            f'Key k8s.v1.cni.cncf.io/network-status not found on pod "{CURR_POD_NAME}" on "{CURR_WORKER_NODE_NAME}"')
    if len(entrylist) > 0:
        interfaces = [
            iface
            for iface in netifaces.interfaces()
            if "net" in iface and iface not in ("lo", "eth0", "tunl0")
        ]
    else:
        interfaces = [
            iface
            for iface in netifaces.interfaces()
            if iface not in ("lo", "tunl0")
        ]

    
    if not interfaces:
        log.error(
            f'Secondary nics not found for "{CURR_POD_NAME}" on "{CURR_WORKER_NODE_NAME}".'
        )
        sys.exit(1)

    for iface in interfaces:
        for i in range(num_server):
            try:
                address = netifaces.ifaddresses(iface)
                ip = address[netifaces.AF_INET][0]["addr"]
                command = ["iperf3", "-s", "-B", ip, "-p", str(port + i), "-D"]
                log.info(
                    f"Starting iperf3 server {ip}:{port + i} using {iface} in {CURR_POD_NAME} on {CURR_WORKER_NODE_NAME}..."
                )
                subprocess.run(command, text=True, capture_output=True, check=True)
            except subprocess.CalledProcessError as e:
                log.error(
                    f"Server failed to start on {ip}:{port + i} using {iface} in {CURR_POD_NAME} on {CURR_WORKER_NODE_NAME}.\n "
                    f"Exited with error: {e.stderr}"
                )
                sys.exit(1)
            except KeyError:
                log.error(
                    f"No AF_INET (IPv4) address found for interface {iface} in {CURR_POD_NAME} on {CURR_WORKER_NODE_NAME}."
                )
                sys.exit(1)


if __name__ == "__main__":
    main()
