from kubernetes import client, config, watch
from kubernetes.client.rest import ApiException
from pprint import pprint
import os
from subprocess import Popen
import json
from datetime import datetime
import time

def main():
   
    config.load_incluster_config()

    v1 = client.CoreV1Api()
    count = int(os.getenv("COUNT"))
    w = watch.Watch()
    namespace = os.getenv("NAMESPACE")
    selector = os.getenv("SELECTOR")
    ifaces = []
    for event in w.stream(v1.list_namespaced_pod, namespace=namespace,label_selector=selector, timeout_seconds=20):
        entry = json.loads(event['object'].metadata.annotations['k8s.v1.cni.cncf.io/network-status'])
        podName = event['object'].metadata.name
        ips = entry[1]['ips']
        print("\nPod " + podName + " has these IPs:")
        print(ips)
        for i in ips:
            ifaces.append(i)
        count -= 1
        if not count:
            w.stop()
    print("\nFinished with Pod list stream.")
    print(ifaces)

    print("\nReaching out all hosts..")
    unreachableHosts = []
    for host in ifaces:
        maxRetries = 3
        numTry = 0
        unreachable = True
        while (numTry < maxRetries):
            proc = Popen(['ping', host,'-c','1',"-W","2"])
            proc.wait()
            # If response is not 0, ping was unsuccessful 
            if proc.poll():
                time.sleep(3)
            else:
                unreachable = False
                break
            numTry+=1
        if unreachable: 
            print(str(host) + " is unreachable")
            unreachableHosts.append(host)
        
    print("\nTest completed")

    if len(unreachableHosts) != 0:
        print("The following hosts were unreachable ", unreachableHosts)
        api = client.CustomObjectsApi()
 

# # apiVersion: my.domain/v1alpha1
# # kind: HealthCheckReport
# # metadata:
# #   labels:
# #     name: healthcheckreport
# #   name: healthcheckreport-sample
# # spec:
# #   node: "worker-0"
# #   report: <the output>


        nodename = os.getenv("NODE_NAME")
        namespace = os.getenv("NAMESPACE")
        # api_instance = client.CoreV1Api()

        # We probably don't need to deschedule the pod at all costs.. Also, a less aggressive option should be considered instead of cordining the node, in this case, as it should be an issue with the secondary nic operator.

        # body = {
        #     "metadata": {
        #         "labels": {
        #             "deschedule": ""}
        #     }
        # }

        # try:
        #     api_instance.patch_namespaced_pod(namespace=namespace, name=podname, body=body)
        # except ApiException as e:
        #     print("Exception when patching pod:\n", e)

        result = "Cannot reach the following addresses: " 
        for h in unreachableHosts:
            result = result + str(h) + "\n"

        dt = datetime.now()
        hcr_manifest = {
            'apiVersion': 'my.domain/v1alpha1',
            'kind': 'HealthCheckReport',
            'metadata': {
                'name': "netcheck-"+nodename+"-"+dt.strftime("%d-%m-%Y-%H.%M.%S.%f")
            },
            'spec': {
                'node': nodename,
                'report': result,
                'issuer': "net-reach"
            }
        }
        group = "my.domain"
        v = "v1alpha1"
        plural = "healthcheckreports"
        try:
            api.create_namespaced_custom_object(group, v, namespace, plural, hcr_manifest)
        except ApiException as e:
            print("Exception when calling create health check report:\n", e)

        raise TypeError("Failing init container.")
        # all_reports = api.list_namespaced_custom_object(group, v, namespace, plural)

if __name__ == '__main__':
    main()