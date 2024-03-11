from kubernetes import client, config
from kubernetes.client.rest import ApiException
import os
import json
import argparse
import asyncio
import subprocess
import time

parser = argparse.ArgumentParser()
parser.add_argument('--job', type=str, default='None', help='Workload node discovery w/ given namespace and label. Ex: \"--job=namespace:label-key=label-value\". Default is set to None.')
parser.add_argument('--nodelabel', type=str, default='None', help='Node label to select nodes. Ex: \"label-key=label-value\". Default is set to None.')
parser.add_argument('--nodes', type=str, default='all', help='Node(s) running autopilot that will be reached out by ping. Can be a comma separated list. Default is \"all\". Servers are reached out sequentially')
args = vars(parser.parse_args())

job = args['job']
nodemap = {}
namespace_self = os.getenv("NAMESPACE")
nodename_self  = os.getenv("NODE_NAME")
config.load_incluster_config()
kubeapi = client.CoreV1Api()

async def main():
    nodelist = args['nodes'].replace(' ', '').split(',') # list of nodes
    job = args['job']
    nodelabel = args['nodelabel']
    nodemap = {}
    allnodes = False
    if 'all' in nodelist and job == 'None' and nodelabel == 'None':
        allnodes = True
    else:
        nodemap = get_job_nodes(nodelist)

    nodes={}
    ifaces=set()
    print("[PING] Pod running ping: ", os.getenv("POD_NAME"))
    print("[PING] Starting: collecting node list")
    try:
        retries = 0
        daemonset_size = expectedPods()
        autopilot_pods = kubeapi.list_namespaced_pod(namespace=namespace_self, label_selector="app=autopilot")
        while len(autopilot_pods.items) < daemonset_size or retries > 100:
            print("[PING] Waiting for all Autopilot pods to run")
            time.sleep(5)
            autopilot_pods = kubeapi.list_namespaced_pod(namespace=namespace_self, label_selector="app=autopilot")
            retries +=1
        if retries > 100 and len(autopilot_pods.items) < daemonset_size:
            print("[PING] Reached max retries of 100. ABORT")
            exit()

    except ApiException as e:
        print("Exception when calling CoreV1Api->list_namespaced_pod: %s\n" % e)
        exit()

    for pod in autopilot_pods.items:
        if not 'k8s.v1.cni.cncf.io/network-status' in pod.metadata.annotations:
            print("[PING] Pod", pod.metadata.name, "misses network annotation. Skip node", pod.spec.node_name)

    # run through all pods and create a map of all interfaces
    print("Creating a list of interfaces and IPs")
    for pod in autopilot_pods.items:
        if pod.spec.node_name != nodename_self and (allnodes or (pod.spec.node_name in nodemap.keys())):
            try:
                entrylist = json.loads(pod.metadata.annotations['k8s.v1.cni.cncf.io/network-status'])
            except KeyError:
                print("Key k8s.v1.cni.cncf.io/network-status not found on pod", pod.metadata.name, "- Skipping node", pod.spec.node_name)
                continue
            else:
                node={}
                nodes[pod.spec.node_name] = node
                for entry in entrylist:
                    try:
                        iface=entry['interface']
                    except KeyError:
                        print("Interface key name not found, assigning 'k8s-pod-network'.")
                        iface = "k8s-pod-network"
                    ifaces = ifaces | {iface}
                    node[iface] = {
                        'ips': entry['ips'],
                        'pod': pod.metadata.name
                    }

    if len(nodes.keys()) == 0:
        print("[PING] No nodes found. ABORT")
        exit(0)
    # run ping tests to each pod on each interface
    print("[PING] Running ping tests for every interface")
    conn_dict = dict()
    clients = []
    for nodename in nodes.keys():
        conn_dict[nodename] = {}
        for iface in ifaces:
            try:
                ips = nodes[nodename][iface]['ips']
            except KeyError:
                print("Interface", iface, "not found, skipping.")
                continue
            for index, ip in enumerate(ips):
                command = ['ping',ip,'-t','45','-c','10']
                clients.append((subprocess.Popen(command, start_new_session=True, text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE), nodename, ip, "net-"+str(index)))
    for c in clients:
        try:
            c[0].wait(50)
        except:
            print("Timeout while waiting for", c[2], "on node", c[1])
            continue
    fail = False
    for c in clients:
        stdout, stderr = c[0].communicate()
        if stderr:
            print("[PING] output parse exited with error: " + stderr)
            print("FAIL")
        else:
            if "Unreachable" in stdout or "100% packet loss" in stdout:
                print("Node", c[1], c[2], c[3], "1")
                fail = True
            else:
                print("Node", c[1], c[2], c[3], "0")
    if fail:
        print("[PING] At least one node unreachable. FAIL")
    else:
        print("[PING] all nodes reachable. success")
            
def get_job_nodes(nodelist):
    v1 = client.CoreV1Api()
    # get nodes from job is specified
    nodemap = {}
    node_name_self = os.getenv("NODE_NAME")
    job = args['job']
    if job != 'None':
        job = args['job'].split(':') 
        job_ns = job[0] # ex: "default"
        job_label = job[1] # ex: "job-name=my-job" or "app=my-app"]
        try:
            job_pods = v1.list_namespaced_pod(namespace=job_ns, label_selector=job_label)
        except ApiException as e:
            print("[PING] Exception when calling CoreV1Api->list_namespaced_pod: %s\n" % e)

        print('[PING] Workload:', ': '.join(job))
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


def expectedPods():
    v1 = client.AppsV1Api()
    try:
        autopilot = v1.list_namespaced_daemon_set(namespace=namespace_self, label_selector="app=autopilot")
    except ApiException as e:
        print("[PING] Exception when calling fetching Autopilot by corev1api->list_namespaced_daemon_set", e)
        return 0
    return autopilot.items[0].status.desired_number_scheduled

if __name__ == '__main__':
    asyncio.run(main())