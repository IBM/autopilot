########################################################################
# Python program that uses the Python Client Library for Kubernetes to
# run autopilot health checks on all nodes or a specific node(s). 
########################################################################
import argparse
from kubernetes import client, config
import requests
import time
from multiprocessing import Pool


# load in cluster kubernetes config for access to cluster
config.load_incluster_config()
v1 = client.CoreV1Api()

# get arguments for service, namespace, node(s), and check (test type)
parser = argparse.ArgumentParser()
parser.add_argument('--service', type=str, default='autopilot-healthchecks', help='Autopilot healthchecks service name. Default is \"autopilot-healthchecks\".')
parser.add_argument('--namespace', type=str, default='autopilot', help='Autopilot healthchecks namespace. Default is \"autopilot\".')
parser.add_argument('--nodes', type=str, default='all', help='Node(s) that will run a healthcheck. Default is \"all\".')
parser.add_argument('--check', type=str, default='all', help='The specific test that will run: \"all\", \"pciebw\", \"nic\", or \"remapped\". Default is \"all\".')
parser.add_argument('--batchSize', type=int, default=1, help='Number of nodes running in parallel at a time. Default is \"1\".')
args = vars(parser.parse_args())
service = args['service']
namespace = args['namespace']
node = args['nodes'].replace(' ', '').split(',') # list of nodes
check = args['check']
batch_size = args['batchSize']

node_status = {} # updates after each node is tested


# get addresses in desired endpointslice (autopilot-healthchecks) based on which node(s) the user chooses
def get_addresses():
    endpoints = v1.list_namespaced_endpoints(namespace=namespace).items
    for endpointslice in endpoints:
        if endpointslice.metadata.name == service:
            print("EndpointSlice: " + str(endpointslice.metadata.name)) 
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
                raise Exception('Error: Issue with --node parameter. Choices include: \"all\", a specific node name, or a comma separated list of node names.')
    raise Exception('Error: Issue with --service or --namespace parameter. Check that they are correct.')


# runs healthchecks at each endpoint (there is one endpoint in each node)
def run_tests(address):
    daemon_node = str(address.node_name)
    url = create_url(address, daemon_node)
    response = get_requests(url)
    get_node_status(response, daemon_node)
    output = '\nEndpoint: {ip}\nNode: {daemon_node}\nurl: {url}\nResponse:\n{response}\nNode Status: {status}'.format(ip=address.ip, daemon_node=daemon_node, url=url, response=response, status=', '.join(node_status[daemon_node]))
    output += "\n-------------------------------------\n" # separator
    return output



# create url for test
def create_url(address, daemon_node):
    if check == 'all':
        return 'http://' + str(address.ip) + ':3333/status?host=' + daemon_node
    elif (check == 'nic' or check == 'remapped' or check == 'pciebw'):
        return 'http://' + str(address.ip) + ':3333/status?host=' + daemon_node + '&check=' + check
    else:
        raise Exception('Error: Issue with --check parameter. Options are \"all\", \"pciebw\", \"nic\", or \"remapped\"')


# rest api calls for healthcheck
def get_requests(url):
    page = ''
    while page == '':
        try:
            page = requests.get(url)
            break
        except:
            print('Connection refused by server..')
            print('sleeping for 5 seconds')
            time.sleep(5)
            continue
    return page.text


# check and print status of each node
def get_node_status(response, daemon_node):
    node_status[daemon_node] = []
    response_list = response.split('\n')
    for line in response_list:
        if ('FAIL' in line):
            if ('PCIE' in line):
                node_status[daemon_node].append('PCIE Failed')
            elif ('NETWORK' in line):
                node_status[daemon_node].append('MULTI-NIC CNI Failed')
            elif('REMAPPED ROWS' in line):
                node_status[daemon_node].append('REMAPPED ROWS Failed')
    if len(node_status[daemon_node]) < 1:
        node_status[daemon_node].append('Ok')


# start program
if __name__ == "__main__":
    # run_tests(get_addresses())
    addresses = get_addresses()
    total_nodes = len(addresses)

    # set max number of processes
    if (total_nodes < 4):
        max_processes = total_nodes
    else:
        max_processes = 4

    # multiprocessing for parallelism in batches
    with Pool(processes=max_processes) as pool:
        for result in pool.map(run_tests, addresses, chunksize=batch_size):
            print(result)

