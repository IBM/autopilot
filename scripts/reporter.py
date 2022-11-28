from kubernetes import client, config
import os


def main():
    config.load_incluster_config()

    api = client.CustomObjectsApi()
    output = os.popen('./gpuLocalBandwidthTest.sh')
    result = output.read()
    print(result)

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
    api.create_namespaced_custom_object(group, v, namespace, plural, hrr_manifest)

    # all_reports = api.list_namespaced_custom_object(group, v, namespace, plural)

if __name__ == '__main__':
    main()