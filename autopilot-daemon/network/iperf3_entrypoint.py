from iperf3_utils import *
from network_workload import NetworkWorkload

parser = argparse.ArgumentParser()

parser.add_argument(
    "--workload",
    type=str,
    default="ring",
    help=('The type of network workload. Supported workload values: "ring"'),
)

parser.add_argument(
    "--pclients",
    type=str,
    default="8",
    help=(
        'The number of clients to run in parallel. Note. This is not using the iperf3 "-P" option. '
        'This spawns "N" number of iperf3 client instances in parallel to a target server. For each client, '
        'a respective port on the target server will be pinned. For instance, if there are 3 "pclients" '
        "specified, then there will be 3 instances of a particular network interface on 3 different ports."
    ),
)

parser.add_argument(
    "--startport",
    type=str,
    default="5200",
    help=(
        'The default port value. In the event that "pclients" is greater than 1, the default port value used '
        "to generate servers will automatically increase to accomdate the clients running in parallel."
    ),
)

parser.add_argument(
    '--cleanup',
    action='store_true',
    help=(
        "When provided, this will kill ALL iperf servers on every node."
    ),
)

args = vars(parser.parse_args())


async def makeconnection(event, address, handle):
    """
    Handles connections to the target autopilot pod on a different worker-node.
    Attempts to ensure synchronization via asyncio events...

    Args:
        address (str): The address of the autopilot pod.
        handle (str): The endpoint handle for the connection.

    """
    # Task waits for the event to be set before starting its work.
    if event != None:
        await event.wait()
    url = f"http://{address}:{AUTOPILOT_PORT}{handle}"
    total_timeout = aiohttp.ClientTimeout(total=60 * 10)
    log.info(f"Initiated connection to: {url}")
    async with aiohttp.ClientSession(timeout=total_timeout) as session:
        async with session.get(url) as resp:
            reply = await resp.text()
            log.info(f"Response from {url}:\n{reply}")


async def iperf_start_servers(node_map, num_servers, port_start):
    """
    Starts iperf3 servers on each node by sending requests to the corresponding endpoints
    derived in the node_map. Each server will be launched from the corresponding autopilot
    pod that the endpoint represents on the worker-node.

    Args:
        node_map (dict): A dictionary mapping worker-nodes to representation data.
        num_servers (str): The number of iperf3 servers to start on each node.
        port_start (str) The port to start launching servers from on each node.
    """
    tasks = [
        makeconnection(
            None,
            node_map[node]["endpoint"],
            f"/iperfservers?numservers={num_servers}&startport={port_start}",
        )
        for node in node_map
    ]
    await asyncio.gather(*tasks)


async def run_workload(workload_type, nodemap, workload, num_clients, port_start):
    """
    Starts network tests according to the specified workload.

    Args:
        workload_type (str): A workload type to run.
        node_map (dict): A dictionary mapping node names to their endpoints, pods, and network interfaces.
        workload (dict): A dictionary specifying the workload and steps for the network tests.
        num_clients (str): The number of parallel clients to test against the server (used to also increase port val.)
        port_start (str): A port associated to the server,
    """
    if SupportedWorkload.RING.value == workload_type:
        event = asyncio.Event()
        # All the nodes "should have" the same amount of interfaces...let's just get the first node and check how many there are...
        netifaces_count = len(nodemap[next(iter(nodemap))]["netifaces"])
        for iface in range(netifaces_count):
            for step in workload:
                log.info(f"\nTime Step {step} for interface: net1-{iface}")
                tasks = []
                for pair in workload[step]:
                    for source, target in pair.items():
                        # source and target are worker node names, which are keys to the dictionary...
                        log.info(f"{source} -> {target}")
                        log.info(
                            f"      From {nodemap[target]['pod']} to {nodemap[target]['netifaces'][iface]}:{port_start}"
                        )
                        task = makeconnection(
                            event,
                            nodemap[source]["endpoint"],
                            f"/iperfclients?dstip={nodemap[target]['netifaces'][iface]}&dstport={port_start}&numclients={num_clients}",
                        )
                        tasks.append(task)
                await asyncio.sleep(1)
                log.info("Starting all connections...")
                # signals all waiting tasks to start concurrently.
                event.set()
                # Runs all the tasks concurrently AND waits for all of them to complete (very important for timesteps).
                await asyncio.gather(*tasks)
    else:
        log.error("Unsupported Workload Attempted")
        sys.exit(1)


async def cleanup_iperf_servers(node_map):
    """
    Removes all started iperf servers across all nodes.

    Args:
    node_map (dict): A dictionary mapping worker-nodes to representation data.
    """
    tasks = [
        makeconnection(
            None,
            node_map[node]["endpoint"],
            f"/iperfstopservers",
        )
        for node in node_map
    ]
    await asyncio.gather(*tasks)


async def main():
    type_of_workload = args["workload"].upper()
    num_parallel_clients = args["pclients"]
    port_start = args["startport"]
    cleanup_iperf = args["cleanup"]

    wl = NetworkWorkload()
    autopilot_node_map = wl.gen_autopilot_node_map_json()
    if type_of_workload in (workload.value for workload in SupportedWorkload):
        if SupportedWorkload.RING.value == type_of_workload:
            wl.print_autopilot_node_map_json(autopilot_node_map)

            ring_workload = wl.generate_ring_topology_json(autopilot_node_map)
            wl.print_ring_workload()

            await iperf_start_servers(
                autopilot_node_map, num_parallel_clients, port_start
            )

            await run_workload(
                type_of_workload,
                autopilot_node_map,
                ring_workload,
                num_parallel_clients,
                port_start,
            )

        else:
            #
            # TODO: Build other workloads...
            #
            log.error("Unsupported Workload Attempted")
            sys.exit(1)
    else:
        log.error("Unsupported Workload Attempted")
        sys.exit(1)

    if cleanup_iperf:
        await cleanup_iperf_servers(autopilot_node_map)

if __name__ == "__main__":
    asyncio.run(main())
