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
    nodemap = {}
    allnodes = False
    if 'all' in nodelist and job == 'None':
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

        # print("Expecting ", daemonset_size, " pods, got ", len(autopilot_pods.items))
    except ApiException as e:
        print("Exception when calling CoreV1Api->list_namespaced_pod: %s\n" % e)
        exit()

    for pod in autopilot_pods.items:
        if not 'k8s.v1.cni.cncf.io/network-status' in pod.metadata.annotations:
            print("[PING] Pod", pod.metadata.name, "misses network annotation. ABORT.")

    # run through all pods and create a map of all interfaces
    print("Creating a list of interfaces and IPs")
    for pod in autopilot_pods.items:
        if pod.spec.node_name != nodename_self and (allnodes or (pod.spec.node_name in nodemap.keys())):
            node={}
            nodes[pod.spec.node_name] = node
            entrylist = json.loads(pod.metadata.annotations['k8s.v1.cni.cncf.io/network-status'])
            for entry in entrylist:
                iface=entry['interface']
                ifaces = ifaces | {iface}
                node[iface] = {
                    'ips': entry['ips'], 
                    'pod': pod.metadata.name
                }

    # run ping tests to each pod on each interface
    print("[PING] Running ping tests for every interface")
    conn_dict = dict()
    clients = []
    for nodename in nodes.keys():
        conn_dict[nodename] = {}
        for iface in ifaces:
            for ip in nodes[nodename][iface]['ips']:
            # ip=nodes[nodename][iface]['ips']
            # r = ping(ip, timeout=1, count=1, verbose=False)
            # conn_dict[nodename][iface] = r.success()
                command = ['ping',ip,'-t','1','-c','1']
                clients.append((subprocess.Popen(command, start_new_session=True, text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE), nodename, ip))
    # [c[0].wait() for c in clients]
    for c in clients:
        try:
            c[0].wait(30)
        except:
            print("Timeout while waiting for", c[2], "on node", c[1])
            continue
    for c in clients:
        stdout, stderr = c[0].communicate()
        if stderr:
            print("[PING] output parse exited with error: " + stderr + " FAIL")
        else:
            print("Node", c[1], c[2], "1") if "Unreachable" in stdout or "0 received" in stdout else print("Node", c[1], c[2], "0")
            
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