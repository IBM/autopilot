from kubernetes import client, config
from kubernetes.client.rest import ApiException
import os
import json
import subprocess
import argparse
import math

parser = argparse.ArgumentParser()
parser.add_argument('--nodes', type=str, default='all', help='Node(s) running autopilot that will be reached out by iperf3. Can be a comma separated list. Default is \"all\". Servers are reached out sequentially')

parser.add_argument('--job', type=str, default='None', help='Workload node discovery w/ given namespace and label. Ex: \"--job=namespace:label-key=label-value\". Default is set to None.')

parser.add_argument('--iface', type=str, default='eth0', help='Name of the interface to test with iperf3. Will spawn a single client to connect to a single server. The client will connect to IP on the selected interface. Can be management plane (eth0) or data plane (e.g., net1). Only one value allowed. Default is set to eth0.')

parser.add_argument('--plane', type=str, default='data', help='Run on either data plane (data) or management plane (mgmt, on eth0). Can be customized with --replicas to run multiple clients over a single server on multiple ports. Default is data plane.')

parser.add_argument('--replicas', type=str, default='1', help='Number of iperf3 clients to connect to a remote server')

args = vars(parser.parse_args())

def main():
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
    
    run_clients(address_map)


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
                    address_map.get(entry['interface']).append((entry['ips'][0],pod.spec.node_name))
    if len(address_map) == 0:
        print("[IPERF] No interfaces found. FAIL.")

    return address_map


def run_clients(address_map):
    client_replicas_per_iface = args['replicas']
    plane = args['plane']
    clients = []
    print("[IPERF] Starting " + client_replicas_per_iface + " clients per iface")
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
            clients.append(try_connect_popen(command, ip, filename))
            # try_connect(command, ip)
    # else if stresstest, run_clients(addresses)
    else:
        subset = math.ceil(int(client_replicas_per_iface)/(len(address_map)-1)) # remove eth0 from the count
        print("Number of servers per interface: " + str(subset))
        # create the port and iterate over address map
        
        for ifacegroup in address_map: 
            if ifacegroup != "eth0":
                for ip in address_map.get(ifacegroup):
                    for r in range(subset):
                        if r > 9:
                            port = '51'+str(r)
                        else:
                            port = '510'+str(r)
                        # print("[IPERF] Connect to " + ip[0] + " on " + ip[1] + " :" + port)
                        command[2] = ip[0]
                        command[4] = port
                        filename="out-"+ip[1]+"-"+ip[0]+"-"+port
                        clients.append(try_connect_popen(command, filename))
        print("[IPERF] Clients launched from ", os.getenv("POD_NAME"))
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
    log_file = open(filename, "wt")
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
    main()
