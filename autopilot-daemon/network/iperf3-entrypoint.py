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
import requests

parser = argparse.ArgumentParser()
parser.add_argument('--nodes', type=str, default='all', help='Node(s) running autopilot that will be reached out by iperf3. Can be a comma separated list. Default is \"all\". Servers are reached out sequentially')

parser.add_argument('--job', type=str, default='None', help='Workload node discovery w/ given namespace and label. Ex: \"--job=namespace:label-key=label-value\". Default is set to None.')

parser.add_argument('--plane', type=str, default='data', help='Run on either data plane (data) or management plane (mgmt, on eth0). Can be customized with --clients to run multiple clients over a single server on multiple ports. Default is data plane.')

parser.add_argument('--clients', type=str, default='1', help='Number of iperf3 clients to connect to a remote server')

parser.add_argument('--servers', type=str, default='1', help='Number of iperf3 servers per node. If #replicas is less than the number of secondary nics, it will create #replicas server per nic. Otherwise, it will spread #replicas servers as evenly as possible on all interfaces')

parser.add_argument('--source', type=str, default='None', help='Number of iperf3 clients to connect to a remote server')

parser.add_argument('--nodelabel', type=str, default='None', help='Node label to select nodes. Ex: \"label-key=label-value\". Default is set to None.')

args = vars(parser.parse_args())

async def main():
    nodelist = args['nodes'].replace(' ', '').split(',') # list of nodes
    job = args['job']
    nodelabel = args['nodelabel']
    allnodes = False
    nodemap = {}

    config.load_incluster_config()
    if args['source'] != "None" and args['source'] != os.getenv("NODE_NAME"):
        print("[IPERF] Asking to run from a different node. Invoking the test on target node", os.getenv("NODE_NAME"))
        v1 = client.CoreV1Api()
        try:
            endpoints = v1.list_namespaced_endpoints(namespace=os.getenv("NAMESPACE"),field_selector="metadata.name=autopilot-healthchecks")
        except ApiException as e:
            print("Exception when calling CoreV1Api->list_namespaced_endpoints: %s\n" % e)
        for endpointslice in endpoints.items:
            addresses = endpointslice.subsets[0].addresses
            for address in addresses:
                if address.node_name == args['source']:
                    url = 'http://' + address.ip + ':3333/status?check=iperf&host='+args['nodes']+'&job='+args['job']+'&plane='+args['plane']+'&clientsperiface='+args['clients']+'&serverspernode='+args['servers']
                    print("[IPERF] Forward request to", address.ip, "url", url)
                    page = ''
                    while page == '':
                        page = requests.get(url)
                    print((page.text).strip())
                    exit()

    if 'all' in nodelist and job == 'None' and nodelabel == 'None':
        allnodes = True
    else:
        nodemap = get_job_nodes(nodelist)
        
    # print("[IPERF] All? ", allnodes)
    # print("[IPERF] Nodes: ", nodemap.keys())
    # print("[IPERF] Data/mgmt plane:", plane)
    print("[IPERF] Pod running clients: ", os.getenv("POD_NAME"))
    
    address_map= get_addresses(allnodes, nodemap)
    if len(address_map) == 0:
        print("[IPERF] No nodes selected. ABORT.")
        exit() 
    
    print("[IPERF] Starting servers on all other nodes... ")
    interfaces = [iface for iface in netifaces.interfaces() if "net" in iface] ## VERY TEMPORARY
    if len(interfaces)==0:
        print("[IPERF] Cannot launch servers -- secondary nics not found ", os.getenv("POD_NAME"), ". ABORT")
        exit()

    secondary_nics_count = len(interfaces) # quite a lame bet.. excluding eth0 and lo assuming all the other ones are what we want.
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
            exit()

        print('[IPERF] Workload:', ': '.join(job))
        if len(job_pods.items) == 0:
            print ("No pod is labeled with", job_label, " - ABORT.")
            exit()

        for pod in job_pods.items:
            if pod.spec.node_name != node_name_self:
                nodemap[pod.spec.node_name] = True
    
    nodelabel = args['nodelabel']
    if nodelabel != 'None':
        try:
            labeled_nodes = v1.list_node(label_selector=nodelabel)
        except ApiException as e:
            print("Exception when calling CoreV1Api->list_node: %s\n" % e)
            exit()
        if len(labeled_nodes.items) == 0:
            print ("No node is labeled with", nodelabel, " - ABORT.")
            exit()

        for labeled_node in labeled_nodes.items:
            if labeled_node.metadata.name != node_name_self:
                nodemap[labeled_node.metadata.name] = True
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
        exit()

    for pod in autopilot_pods.items:
        if pod.spec.node_name != node_name_self and (allnodes or (pod.spec.node_name in nodemap.keys())):
            entrylist = json.loads(pod.metadata.annotations['k8s.v1.cni.cncf.io/network-status'])
            if len(entrylist) > 0:
                for entry in entrylist:
                    try:
                        iface=entry['interface']
                    except KeyError:
                        print("Interface key not found, assigning default.")
                        iface = "default"
                    if address_map.get(iface) == None:
                        address_map[iface] = []
                    address_map.get(iface).append((entry['ips'],pod.spec.node_name))
    if len(address_map) == 0:
        print("[IPERF] No interfaces found. FAIL.")

    return address_map

async def start_servers(allnodes, nodemap):
    v1 = client.CoreV1Api()
    address_list = []
    try:
        endpoints = v1.list_namespaced_endpoints(namespace=os.getenv("NAMESPACE"),field_selector="metadata.name=autopilot-healthchecks")
    except ApiException as e:
        print("Exception when calling CoreV1Api->list_namespaced_endpoints: %s\n" % e)
    for endpointslice in endpoints.items:
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
    else:
        if len(address_map)>1:
            if client_replicas_per_iface < (len(address_map)-1):
                client_replicas_per_iface = (len(address_map)-1)
            subset = math.ceil(client_replicas_per_iface/(len(address_map)-1)) # remove eth0 from the count
            if subset > maxports:
                print("[IPERF] Mismatch in number of servers. Wants " + str(subset) + ", have " + str(maxports) + " server on each nic. Downsizing number of clients to match the existing servers.")
            else:
                maxports = subset
            print("[IPERF] Number of clients per interface: " + str(subset))
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
                                clients.append((try_connect_popen(command, filename), entry[1], ip))
                            netid+=1
            print("[IPERF] All clients launched from ", os.getenv("POD_NAME"))
        else:
            print("[IPERF] Cannot launch clients -- secondary nics not found ", os.getenv("POD_NAME"), ". ABORT")
            return
        for c in clients:
            try:
                c[0].wait(30)
            except:
               print("Timeout while waiting for", c[2], "on node", c[1])
               continue

    command = ['bash','./network/iperf3-debrief.sh']
    result = subprocess.run(command, text=True, capture_output=True)
    if result.stderr:
        print(result.stderr)
        print("[IPERF] output parse exited with error: " + result.stderr + " FAIL")
    else:
        output = result.stdout
        print(output.strip())
        

def try_connect_popen(command, filename):
    log_file = open(filename, "at")
    p = subprocess.Popen(command, start_new_session=True, text=True, stdout=log_file, stderr=log_file)
    return p

if __name__ == '__main__':
    asyncio.run(main())
