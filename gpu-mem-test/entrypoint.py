# bash precheck.sh && ./gpucheck

from kubernetes import client, config
from kubernetes.client.rest import ApiException
from pprint import pprint
from datetime import datetime
import os


def main():
    config.load_incluster_config()

    api = client.CustomObjectsApi()
    output = os.popen('bash ./briefings.sh')
    result = output.read()
    print(result)

    if "ABORT" not in result:
        print("Briefings completed. Continue with memory evaluation.")
        output = os.popen('./gpucheck')
        result = output.read()
        if "NONE" in result:
            print(result)
            return 0 

# apiVersion: my.domain/v1alpha1
# kind: HealthCheckReport
# metadata:
#   labels:
#     name: healthcheckreport
#   name: healthcheckreport-sample
# spec:
#   node: "worker-0"
#   report: <the output>


    nodename = os.getenv("NODE_NAME")
    podname = os.getenv("POD_NAME")
    namespace = os.getenv("NAMESPACE")
    # api_instance = client.CoreV1Api()

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

    dt = datetime.now()
    
    hcr_manifest = {
        'apiVersion': 'my.domain/v1alpha1',
        'kind': 'HealthCheckReport',
        'metadata': {
            'name': "memcheck-"+nodename+"-"+dt.strftime("%d-%m-%Y-%H.%M.%S.%f")
        },
        'spec': {
            'node': nodename,
            'report': result,
            'issuer': "gpu-mem"
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