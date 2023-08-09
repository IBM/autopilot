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
    args = vars(parser.parse_args())

    nodelist = args['nodes'].replace(' ', '').split(',') # list of nodes
    job = args['job']
    
    
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
    
    ifaces = []
    
    try:
        autopilot_pods = v1.list_namespaced_pod(namespace=namespace_self)
    except ApiException as e:
        print("Exception when calling CoreV1Api->list_namespaced_pod: %s\n" % e)

    for pod in autopilot_pods.items:
        if pod.spec.node_name != node_name_self and (allnodes or (pod.spec.node_name in nodemap.keys())):
            entry = json.loads(pod.metadata.annotations['k8s.v1.cni.cncf.io/network-status'])
            if len(entry) > 1:
                for i in entry[1]['ips']:
                    ifaces.append(i)
            else:
                print("[IPERF] Secondary nic not found. FAIL")
                return

    print(ifaces)
    timeout_s = 20
    
    command = ['iperf3', '-c', '', '-p', '5101', '-t', '5']
    for iface in ifaces:
        command[2] = iface
        try:
            result = subprocess.run(command, text=True, capture_output=True, timeout=timeout_s)
        except subprocess.TimeoutExpired:
            print("[IPERF] server unreachable - network reachability test cannot run. FAIL")
            return

        if result.stderr:
            print(result.stderr)
            print("[IPERF] server unreachable - network reachability test cannot run. FAIL")
            return
        else:
            output = result.stdout
            print(output)

if __name__ == '__main__':
    main()
