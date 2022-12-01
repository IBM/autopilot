from kubernetes import client, config
from kubernetes.client.rest import ApiException
from pprint import pprint
import os


def main():
    config.load_incluster_config()

    api = client.CustomObjectsApi()
    output = os.popen('./gpuLocalBandwidthTest.sh')
    result = output.read()
    print(result)

    if "FAIL" not in result:
        print("No report will be issued")
        return 0 

#
# apiVersion: my.domain/v1alpha1
# kind: HealthCheckReport
# metadata:
#   labels:
#     name: healthcheckreport
#   name: healthcheckreport-sample
# spec:
#   node: "worker-0"
#   bandwidth: "6GB/s"


    nodename = os.getenv("NODE_NAME")

    hrr_manifest = {
        'apiVersion': 'my.domain/v1alpha1',
        'kind': 'HealthCheckReport',
        'metadata': {
            'name': "hrr-"+nodename
        },
        'spec': {
            'node': nodename,
            'bandwidth': result
        }
    }
    group = "my.domain"
    v = "v1alpha1"
    plural = "healthcheckreports"
    namespace = "default"
    try:
        api.create_namespaced_custom_object(group, v, namespace, plural, hrr_manifest)
    except ApiException as e:
        print("Exception when calling create health check report:\n", e)

    # all_reports = api.list_namespaced_custom_object(group, v, namespace, plural)

if __name__ == '__main__':
    main()