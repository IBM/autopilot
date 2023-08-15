from kubernetes import client, config
from kubernetes.client.rest import ApiException
import os
import json
import subprocess
import argparse

def main():

    parser = argparse.ArgumentParser()
    parser.add_argument('--nodes', type=str, default='all', help='Node(s) running autopilot that will be reached out by iperf3. Can be a comma separated list. Default is \"all\". Servers are reached out sequentially')
    parser.add_argument('--job', type=str, default='None', help='Workload node discovery w/ given namespace and label. Ex: \"--job=namespace:label-key=label-value\". Default is set to None.')
    parser.add_argument('--iface', type=str, default='eth0', help='Name of the interface to test with iperf3. The client will connect to IP on the selected interface. Can be management plane (eth0) or data plane (e.g., net1). Only one value allowed. Default is set to eth0.')
    args = vars(parser.parse_args())

    nodelist = args['nodes'].replace(' ', '').split(',') # list of nodes
    job = args['job']
    iface = args['iface']
    
    nodemap = {}
    allnodes = False
    namespace_self = os.getenv("NAMESPACE")
    node_name_self = os.getenv("NODE_NAME")

    config.load_incluster_config()
    v1 = client.CoreV1Api()

    if 'all' in nodelist and job == 'None':
        allnodes = True
    else:
        # get nodes from job is specified
        if job != 'None':
            job = args['job'].split(':') 
            job_ns = job[0] # ex: "default"
            job_label = job[1] # ex: "job-name=my-job" or "app=my-app"
            try:
                job_pods = v1.list_namespaced_pod(namespace=job_ns, label_selector=job_label)
            except ApiException as e:
                print("Exception when calling CoreV1Api->list_namespaced_pod: %s\n" % e)

            print('Workload:', ': '.join(job))
            for pod in job_pods.items:
                if pod.spec.node_name != node_name_self:
                    nodemap[pod.spec.node_name] = True
        # get nodes from input list, if any
        if 'all' not in nodelist:
            for i in nodelist:
                if i != node_name_self:
                    nodemap[i] = True

    print("All? ", allnodes)
    print("Nodes: ", nodemap.keys())
    print("iface:", iface)
    
    ifaces = []
    nodenames = []
    
    try:
        autopilot_pods = v1.list_namespaced_pod(namespace=namespace_self, label_selector="app=autopilot")
    except ApiException as e:
        print("Exception when calling CoreV1Api->list_namespaced_pod: %s\n" % e)

    for pod in autopilot_pods.items:
        if pod.spec.node_name != node_name_self and (allnodes or (pod.spec.node_name in nodemap.keys())):
            entrylist = json.loads(pod.metadata.annotations['k8s.v1.cni.cncf.io/network-status'])
            if len(entrylist) > 0:
                for entry in entrylist:
                    if entry['interface'] == iface:
                        ifaces.append(entry['ips'][0])
                        nodenames.append(pod.spec.node_name)
                        print("Adding " + entry['ips'][0] + " from " + pod.spec.node_name + " pod " + pod.metadata.name)
    if len(ifaces) == 0:
        print("No interfaces found. FAIL.")
        return

    print(ifaces)
    timeout_s = 60
    
    command = ['iperf3', '-c', '', '-p', '5101', '-t', '10']
    iter = 0
    for iface in ifaces:
        command[2] = iface
        try:
            print("Connect to iface " + iface + " on " + nodenames[iter])
            result = subprocess.run(command, text=True, capture_output=True, timeout=timeout_s)
        except subprocess.TimeoutExpired:
            print("[IPERF] server " + iface + " on " + nodenames[iter] + " unreachable - timeout error. FAIL")
            continue
        if result.stderr:
            print(result.stderr)
            print("[IPERF] server " + iface + " on " + nodenames[iter] + " exited with error: " + result.stderr + " FAIL")
        else:
            output = result.stdout
            print(output)
        iter = iter + 1

if __name__ == '__main__':
    main()
