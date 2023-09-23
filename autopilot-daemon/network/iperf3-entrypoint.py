from kubernetes import client, config
from kubernetes.client.rest import ApiException
import os
import json
import subprocess
import argparse
import math
import asyncio
import aiohttp
import netifaces

parser = argparse.ArgumentParser()
parser.add_argument('--nodes', type=str, default='all', help='Node(s) running autopilot that will be reached out by iperf3. Can be a comma separated list. Default is \"all\". Servers are reached out sequentially')

parser.add_argument('--job', type=str, default='None', help='Workload node discovery w/ given namespace and label. Ex: \"--job=namespace:label-key=label-value\". Default is set to None.')

parser.add_argument('--plane', type=str, default='data', help='Run on either data plane (data) or management plane (mgmt, on eth0). Can be customized with --clients to run multiple clients over a single server on multiple ports. Default is data plane.')

parser.add_argument('--clients', type=str, default='1', help='Number of iperf3 clients to connect to a remote server')

parser.add_argument('--servers', type=str, default='1', help='Number of iperf3 servers per node. If #replicas is less than the number of secondary nics, it will create #replicas server per nic. Otherwise, it will spread #replicas servers as evenly as possible on all interfaces')

args = vars(parser.parse_args())

async def main():
    nodelist = args['nodes'].replace(' ', '').split(',') # list of nodes
    plane = args['plane']
    job = args['job']
    allnodes = False
    nodemap = {}

    config.load_incluster_config()

    if 'all' in nodelist and job == 'None':
        allnodes = True
    else:
        nodemap = get_job_nodes(nodelist)
        

    print("[IPERF] All? ", allnodes)
    print("[IPERF] Nodes: ", nodemap.keys())
    print("[IPERF] Data/mgmt plane:", plane)
    print("[IPERF] Pod running clients: ", os.getenv("POD_NAME"))
    
    address_map= get_addresses(allnodes, nodemap)
    if len(address_map) == 0:
        return 
    
    print("[IPERF] Starting servers on all other nodes... ")
    interfaces = netifaces.interfaces()
    if len(interfaces)<3:
        print("[IPERF] Cannot launch servers -- secondary nics not found ", os.getenv("POD_NAME"), ". ABORT")
        return

    secondary_nics_count = (len(interfaces)-2)# quite a lame bet.. excluding eth0 and lo assuming all the other ones are what we want.
    server_replicas = int(args['servers'])
    if server_replicas > secondary_nics_count:
        maxports = int(server_replicas/secondary_nics_count) 
    else:
        maxports = server_replicas
    await start_servers(allnodes, nodemap)

    run_clients(address_map, maxports)


def get_job_nodes(nodelist):
    v1 = client.CoreV1Api()
    # get nodes from job is specified
    nodemap = {}
    node_name_self = os.getenv("NODE_NAME")
    job = args['job']
    if job != 'None':
        job = args['job'].split(':') 
        job_ns = job[0] # ex: "default"
        job_label = job[1] # ex: "job-name=my-job" or "app=my-app"
        try:
            job_pods = v1.list_namespaced_pod(namespace=job_ns, label_selector=job_label)
        except ApiException as e:
            print("[IPERF] Exception when calling CoreV1Api->list_namespaced_pod: %s\n" % e)

        print('[IPERF] Workload:', ': '.join(job))
        for pod in job_pods.items:
            if pod.spec.node_name != node_name_self:
                nodemap[pod.spec.node_name] = True
    # get nodes from input list, if any
    if 'all' not in nodelist:
        for i in nodelist:
            if i != node_name_self:
                nodemap[i] = True
    return nodemap

def get_addresses(allnodes, nodemap):
    v1 = client.CoreV1Api()
    address_map = {}
    namespace_self = os.getenv("NAMESPACE")
    node_name_self = os.getenv("NODE_NAME")

    try:
        autopilot_pods = v1.list_namespaced_pod(namespace=namespace_self, label_selector="app=autopilot")
    except ApiException as e:
        print("[IPERF] Exception when calling CoreV1Api->list_namespaced_pod: %s\n" % e)

    for pod in autopilot_pods.items:
        if pod.spec.node_name != node_name_self and (allnodes or (pod.spec.node_name in nodemap.keys())):
            entrylist = json.loads(pod.metadata.annotations['k8s.v1.cni.cncf.io/network-status'])
            if len(entrylist) > 0:
                for entry in entrylist:
                    if address_map.get(entry['interface']) == None:
                        address_map[entry['interface']] = []
                    address_map.get(entry['interface']).append((entry['ips'],pod.spec.node_name))
    if len(address_map) == 0:
        print("[IPERF] No interfaces found. FAIL.")

    return address_map

async def start_servers(allnodes, nodemap):
    v1 = client.CoreV1Api()
    address_list = []
    try:
        endpoints = v1.list_namespaced_endpoints(namespace=os.getenv("NAMESPACE"))
    except ApiException as e:
        print("Exception when calling CoreV1Api->list_namespaced_endpoints: %s\n" % e)
    for endpointslice in endpoints.items:
        if endpointslice.metadata.name == "autopilot-healthchecks":
            addresses = endpointslice.subsets[0].addresses
            for address in addresses:
                if address.node_name != os.getenv("NODE_NAME") and (allnodes or (address.node_name in nodemap.keys())):
                    address_list.append(address.ip)
    res = await asyncio.gather(*(makeconnection(addr) for addr in address_list))
    return res

async def makeconnection(address):
    server_replicas = args['servers']
    url = 'http://' + address + ':3333/iperfservers?replicas=' + server_replicas

    print(f"Initiated connection to {url}.")
    total_timeout=aiohttp.ClientTimeout(total=60*10)
    async with aiohttp.ClientSession(timeout=total_timeout) as session:
        async with session.get(url) as resp:
            reply = await resp.text()
    print(reply)


def run_clients(address_map, maxports):
    client_replicas_per_iface = int(args['clients'])
    plane = args['plane']
    clients = []
    print("[IPERF] Trying to start " + str(client_replicas_per_iface) + " clients per iface")
    command = ['iperf3', '-c', '', '-p', '', '-t', '10']
    # if simple iperf
    if plane == "mgmt":
        command[4] = "6310"
        # iterate over eth0 entries only
        addresses = address_map.get("eth0")
        for ip in addresses:
            # print("[IPERF] Connect to " + ip[0] + " on " + ip[1])
            command[2] = ip[0]
            filename="out-"+ip[0]+"-"+command[4]
            clients.append(try_connect_popen(command, filename))
            # try_connect(command, ip)
    # else if stresstest, run_clients(addresses)
    else:
        if len(address_map)>1:
            if client_replicas_per_iface < (len(address_map)-1):
                client_replicas_per_iface = (len(address_map)-1)
            subset = math.ceil(client_replicas_per_iface/(len(address_map)-1)) # remove eth0 from the count
            if subset > maxports:
                print("[IPERF] Mismatch in number of servers. Wants " + str(subset) + ", have " + str(maxports) + " server on each nic. Downsizing number of clients to match the existing servers.")
            else:
                maxports = subset
            print("[IPERF] Number of servers per interface: " + str(subset))
            # create the port and iterate over address map
            for ifacegroup in address_map: 
                if ifacegroup != "eth0":
                    for entry in address_map.get(ifacegroup):
                        netid = 0
                        for ip in entry[0]:
                            for r in range(maxports):
                                if r > 9:
                                    port = '51'+str(r)
                                else:
                                    port = '510'+str(r)
                                # print("[IPERF] Connect to " + ip + " on " + entry[1] + " :" + port)
                                command[2] = ip
                                command[4] = port
                                filename="out:"+entry[1]+":"+ip+"_net-"+str(netid)
                                clients.append(try_connect_popen(command, filename))
                            netid+=1
            print("[IPERF] Clients launched from ", os.getenv("POD_NAME"))
        else:
            print("[IPERF] Cannot launch clients -- secondary nics not found ", os.getenv("POD_NAME"), ". ABORT")
            return
        [c.wait() for c in clients]
    command = ['bash','./network/iperf3-debrief.sh']
    result = subprocess.run(command, text=True, capture_output=True)
    if result.stderr:
        print(result.stderr)
        print("[IPERF] output parse exited with error: " + result.stderr + " FAIL")
    else:
        output = result.stdout
        print(output)
        

def try_connect_popen(command, filename):
    log_file = open(filename, "at")
    p = subprocess.Popen(command, start_new_session=True, text=True, stdout=log_file, stderr=log_file)
    return p


def try_connect(command, ip):
    timeout_s = 60
    try:
        result = subprocess.run(command, text=True, capture_output=True, timeout=timeout_s)
    except subprocess.TimeoutExpired:
        print("[IPERF] server " + ip[0] + " on " + ip[1] + " unreachable - timeout error. FAIL")
        return
    if result.stderr:
        print(result.stderr)
        print("[IPERF] server " + ip[0] + " on " + ip[1] + " exited with error: " + result.stderr + " FAIL")
    else:
        output = result.stdout
        print(output)

if __name__ == '__main__':
    asyncio.run(main())