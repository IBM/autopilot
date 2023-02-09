from kubernetes import client, config, watch
from kubernetes.client.rest import ApiException
from pprint import pprint
import os
import requests
import json

def main():
    config.load_incluster_config()

    v1 = client.CoreV1Api()
    count = int(os.getenv("COUNT"))
    w = watch.Watch()
    namespace = os.getenv("NAMESPACE")
    selector = os.getenv("SELECTOR")
    ifaces = []
    for event in w.stream(v1.list_namespaced_pod, namespace=namespace,label_selector=selector, timeout_seconds=120):
        entry = json.loads(event['object'].metadata.annotations['k8s.v1.cni.cncf.io/network-status'])
        for i in entry[1]['ips']:
            ifaces.append(i)
        count -= 1
        if not count:
            w.stop()
    print("Finished services stream.")
    print(ifaces)

    
#     api = client.CustomObjectsApi()
 

# # apiVersion: my.domain/v1alpha1
# # kind: HealthCheckReport
# # metadata:
# #   labels:
# #     name: healthcheckreport
# #   name: healthcheckreport-sample
# # spec:
# #   node: "worker-0"
# #   report: <the output>


#     nodename = os.getenv("NODE_NAME")
#     podname = os.getenv("POD_NAME")
#     namespace = os.getenv("NAMESPACE")
#     api_instance = client.CoreV1Api()

#     body = {
#         "metadata": {
#             "labels": {
#                 "deschedule": ""}
#         }
#     }

#     try:
#         api_instance.patch_namespaced_pod(namespace=namespace, name=podname, body=body)
#     except ApiException as e:
#         print("Exception when patching pod:\n", e)

#     hrr_manifest = {
#         'apiVersion': 'my.domain/v1alpha1',
#         'kind': 'HealthCheckReport',
#         'metadata': {
#             'name': "hrr-netcheck-"+nodename
#         },
#         'spec': {
#             'node': nodename,
#             'report': result
#         }
#     }
#     group = "my.domain"
#     v = "v1alpha1"
#     plural = "healthcheckreports"
#     namespace = "default"
#     try:
#         api.create_namespaced_custom_object(group, v, namespace, plural, hrr_manifest)
#     except ApiException as e:
#         print("Exception when calling create health check report:\n", e)

#     # all_reports = api.list_namespaced_custom_object(group, v, namespace, plural)

if __name__ == '__main__':
    main()