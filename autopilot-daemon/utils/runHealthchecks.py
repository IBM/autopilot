##################################################################################
# Python program that uses the Python Client Library for Kubernetes to
# run autopilot health checks on all nodes or a specific node(s).
# Healchecks include PCIEBW, MULTI-NIC CNI Reachability, and GPU REMAPPED ROWS.
# Image: us.icr.io/cil15-shared-registry/gracek/run-healthchecks:3.0.1
##################################################################################
import argparse
import os
import requests
import time
import asyncio
import aiohttp
from itertools import islice
import pprint
from kubernetes import client, config
from kubernetes.client.rest import ApiException
from multiprocessing import Pool

# load in cluster kubernetes config for access to cluster
config.load_incluster_config()
v1 = client.CoreV1Api()

# get arguments for service, namespace, node(s), and check (test type)
parser = argparse.ArgumentParser()
parser.add_argument('--service', type=str, default='autopilot-healthchecks', help='Autopilot healthchecks service name. Default is \"autopilot-healthchecks\".')
parser.add_argument('--namespace', type=str, default='autopilot', help='Namespace where autopilot DaemonSet is deployed. Default is \"autopilot\".')
parser.add_argument('--nodes', type=str, default='all', help='Node(s) that will run a healthcheck. Can be a comma separated list. Default is \"all\" unless --wkload is provided, then set to None. Specific nodes can be provided in addition to --wkload.')
parser.add_argument('--check', type=str, default='all', help='The specific test(s) that will run: \"all\", \"pciebw\", \"nic\", \"remapped\" or \"iperf\". Default is \"all\". Can be a comma separated list.')
parser.add_argument('--batchSize', default='0', type=str, help='Number of nodes to check in parallel. Default is set to the number of the worker nodes.')
parser.add_argument('--wkload', type=str, default='None', help='Workload node discovery w/ given namespace and label. Ex: \"--wkload=namespace:label-key=label-value\". Default is set to None.')
parser.add_argument('--dcgmR', type=str, default='1', help='Run a diagnostic in dcgmi. Run a diagnostic. (Note: higher numbered tests include all beneath.)\n\t1 - Quick (System Validation ~ seconds)\n\t2 - Medium (Extended System Validation ~ 2 minutes)\n\t3 - Long (System HW Diagnostics ~ 15 minutes)\n\t4 - Extended (Longer-running System HW Diagnostics)')

args = vars(parser.parse_args())
service = args['service']
namespace = args['namespace']
node = args['nodes'].replace(' ', '').split(',') # list of nodes
checks = args['check'].replace(' ', '').split(',') # list of checks
batch_size = int(args['batchSize'])
wkload = args['wkload']
if wkload != 'None':
    wkload = args['wkload'].split(':') # ex: --wkload=namespace:label (ex: label='job-name:my-job' or 'app=my-deployment')
    # changing default node value from 'all' to an empty list if there is a workload.
    # this still allows users to include a list of nodes and a workload.
    if (len(wkload) > 1) and (node[0] == 'all'):
        node = []

# debug: runtime
start_time = time.time()


# find workload addresses
def find_wkload():
    node_len = len(node)
    copy = False
    wkload_ns = wkload[0] # ex: "default"
    wkload_label = wkload[1] # ex: "job-name=my-job" or "app=my-app"
    try:
        wkload_pods = v1.list_namespaced_pod(namespace=wkload_ns, label_selector=wkload_label)
    except ApiException as e:
        print("Exception when calling CoreV1Api->list_namespaced_pod: %s\n" % e)
    print('Workload:', ': '.join(wkload))
    for pod in wkload_pods.items:
        pod_name = pod.metadata.name
        node_name = pod.spec.node_name
        if node_name not in node:
            node.append(node_name)
        else:
            copy = True
        # return pod_name
    if (len(node) == node_len) and not copy:
        raise Exception('Error: Issue with --wkload parameter.\nMake sure your workload is spelled correctly and exists in the cluster.')


# get addresses in desired endpointslice (autopilot-healthchecks) based on which node(s) the user chooses
def get_addresses():
    try:
        endpoints = v1.list_namespaced_endpoints(namespace=namespace)
    except ApiException as e:
        print("Exception when calling CoreV1Api->list_namespaced_endpoints: %s\n" % e)
    for endpointslice in endpoints.items:
        if endpointslice.metadata.name == service:
            # print("EndpointSlice: " + str(endpointslice.metadata.name)) 
            addresses = endpointslice.subsets[0].addresses
            if 'all' in node:
                return addresses
            else:
                address_list = []
                for address in addresses:
                    if address.node_name in node:
                        address_list.append(address)
                if len(address_list) > 0:
                    return address_list
                raise Exception('Error: Issue with --node and/or --wkload parameter. Node choices include: \"all\", a specific node name, or a comma separated list of node names. Check that workload is spelled correctly and exists in the cluster.')
    raise Exception('Error: Issue with --service and/or --namespace parameter(s). Check that they are correct.')


# runs healthchecks at each endpoint (there is one endpoint in each node)
def run_tests(address):
    daemon_node = str(address.node_name)
    pid = os.getpid()
    urls = create_url(address, daemon_node)
    output = '\nAutopilot Endpoint: {ip}\nNode: {daemon_node}\nurl(s): {urls}'.format(ip=address.ip, daemon_node=daemon_node, urls='\n        '.join(urls))
    response = []
    for url in urls:
        response.append(get_requests(url))
    node_status_list = get_node_status(response)
    output += '\nResponse:\n{response}\nNode Status: {status}\n-------------------------------------\n'.format(response='~~\n'.join(response), status=', '.join(node_status_list))
    # output += "\n-------------------------------------\n" # separator
    return output, pid, daemon_node, node_status_list



# create url for test
def create_url(address, daemon_node):
    urls = []
    for check in checks:
        if check == 'all':
            urls.append('http://' + str(address.ip) + ':3333/status?host=' + daemon_node)
            return urls
    urls.append('http://' + str(address.ip) + ':3333/status?host=' + daemon_node + '&check=' + args['check'] + '&r='+args['dcgmR'])
    return urls


# rest api calls for healthcheck
def get_requests(url):
    page = ''
    while page == '':
        try:
            page = requests.get(url)
            break
        except:
            print('Connection refused by server..')
            print('Sleeping for 5 seconds')
            time.sleep(5)
            continue
    return page.text


# check and print status of each node
def get_node_status(responses):
    node_status_list = []
    for response in responses:
        response_list = response.split('\n')
        for line in response_list:
            if (('FAIL' in line) or ('ABORT' in line)):
                if ('PCIE' in line):
                    node_status_list.append('PCIE Failed')
                elif ('MULTINIC-CNI-STATUS' in line):
                    node_status_list.append('MULTI-NIC CNI Failed')
                elif('REMAPPED ROWS' in line):
                    node_status_list.append('REMAPPED ROWS Failed')
                elif('IPERF' in line):
                    node_status_list.append('IPERF Failed')
                elif('DCGM' in line):
                    node_status_list.append('DCGM Failed')
    if len(node_status_list) < 1:
        node_status_list.append('Ok')
    return node_status_list

async def makeconnection(address):
    daemon_node = str(address.node_name)
    pid = os.getpid()
    url = create_url(address, daemon_node)
    output = '\nAutopilot Endpoint: {ip}\nNode: {daemon_node}\nurl(s): {url}'.format(ip=address.ip, daemon_node=daemon_node, url='\n        '.join(url))
    print(f"Initiated connection to {url}.")
    total_timeout=aiohttp.ClientTimeout(total=60*60*24)
    async with aiohttp.ClientSession(timeout=total_timeout) as session:
        async with session.get(url[0]) as resp:
            reply = await resp.text()
    response=[reply]
    node_status_list = get_node_status(response)
    output += '\nResponse:\n{response}\nNode Status: {status}\n-------------------------------------\n'.format(response='~~\n'.join(response), status=', '.join(node_status_list))
    # output += "\n-------------------------------------\n" # separator
    return output, pid, daemon_node, node_status_list


async def main(addresses):
    res = await asyncio.gather(*(makeconnection(addr) for addr in addresses))
    return res

def batch_of_nodes(nodelist, batch_size):
    it = iter(nodelist)
    while True:
        batch = list(islice(it, batch_size))
        if not batch:
            break
        yield batch

# start program
if __name__ == "__main__":
    # initializing some variables
    if wkload != 'None':
        find_wkload()
    addresses = get_addresses()
    total_nodes = len(addresses)
    node_status = {} # updates after each node is tested
    pids_tups = [] # debug: process list
    pids_dict = {} # debug: process list

    if batch_size == 0 or batch_size > total_nodes:
        batch_size = total_nodes
    asyncres = []

    for b in batch_of_nodes(addresses, batch_size):
        asyncres.extend(asyncio.run(main(b)))

    for result, pid, daemon_node, node_status_list in asyncres:
        pids_tups.append((pid, daemon_node))
        node_status[daemon_node] = node_status_list
        print(result)
    
    print("Node Summary:\n")
    pprint.pprint(node_status)
    
    # debug: print each process with the nodes they ran
    for p, n in pids_tups:
        pids_dict.setdefault(p, []).append(n)
    print("\n~~~DEBUGGING BELOW~~~\nProcesses (randomly ordered) and the nodes they ran (process:[nodes]):")
    pprint.pprint(pids_dict, width=1)

    # print runtime
    print('\nruntime:', str(time.time() - start_time), 'sec')
