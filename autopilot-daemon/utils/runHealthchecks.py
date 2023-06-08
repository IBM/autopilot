########################################################################
# Python program that uses the Python Client Library for Kubernetes to
# run the autopilot health checks on all nodes or a specific node. 
########################################################################
import argparse
from kubernetes import client, config
import requests
import time
# from multiprocessing import Process


# load in cluster kubernetes config for access to cluster
config.load_incluster_config()
v1 = client.CoreV1Api()

# get arguments for service, namespace, node(s), and check (test type)
parser = argparse.ArgumentParser()
parser.add_argument('--service', type=str, help='Autopilot healthchecks service name')
parser.add_argument('--namespace', type=str, help='Autopilot healthchecks namespace')
parser.add_argument('--node', type=str, help='Node that will run a healthcheck')
parser.add_argument('--check', type=str, help='The specific test that will run: \"all\", \"pciebw\", \"nic\", or \"remapped\".')
args = vars(parser.parse_args())
service = args['service']
namespace = args['namespace']
node = args['node']
check = args['check']
if check == None:
    check = 'all'

node_status = {} # updates after each node is tested


# get addresses in desired endpointslice (autopilot-healthchecks) based on which node(s) the user chooses
def get_addresses():
    endpoints = v1.list_namespaced_endpoints(namespace=namespace).items
    for endpointslice in endpoints:
        if endpointslice.metadata.name == service:
            print("EndpointSlice: " + str(endpointslice.metadata.name)) 
            addresses = endpointslice.subsets[0].addresses
            if node == 'all':
                return addresses
            else:
                address_list = []
                for address in addresses:
                    if address.node_name == node:
                        address_list.append(address)
                        return address_list
                raise Exception('Error: Issue with --node parameter. Choices include: \"all\" OR a specific node name.')
    raise Exception('Error: Issue with --service or --namespace parameter. Check that they are correct.')


# runs healthchecks at each endpoint (there is one endpoint in each node)
def run_tests(addresses):
    for address in addresses:
        print("\nEndpoint: " + str(address.ip)) # debug
        daemon_node = str(address.node_name)
        print("\nNode: ", daemon_node) # debug
        # create url for test. ex: http://10.128.11.100:3333/status?host=dev-ppv5g-worker-3-with-secondary-thlkf&check=nic
        url = create_url(address, daemon_node)
        print('\nurl: ', url) #debug
        # run test
        response = get_requests(url)
        print('\nResponse: \n', response)
        get_node_status(response, daemon_node)
        print('\nNode Status: ', ', '.join(node_status[daemon_node]))
        print("\n-------------------------------------\n") # separator
    # print list of nodes that were tested and their status
    print('Node Summary: ')
    for node in node_status: 
        print('\n', node, ': ', ', '.join(node_status[node]))


# create url for test
def create_url(address, daemon_node):
    if check == 'all':
        return 'http://' + str(address.ip) + ':3333/status?host=' + daemon_node
    elif check == 'nic':
        return 'http://' + str(address.ip) + ':3333/status?host=' + daemon_node + '&check=nic'
    elif check == 'remapped':
        return 'http://' + str(address.ip) + ':3333/status?host=' + daemon_node + '&check=remapped'
    elif check == 'pciebw':
        return 'http://' + str(address.ip) + ':3333/status?host=' + daemon_node + '&check=pciebw'
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
    run_tests(get_addresses())
