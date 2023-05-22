from kubernetes import client, config
from kubernetes.client.rest import ApiException
from datetime import datetime
import os


def main():
    config.load_incluster_config()

    api = client.CustomObjectsApi()

    output = os.popen('bash ./gpubw/briefings.sh')
    result = output.read()
    print(result)

    if "ABORT" not in result:
        print("Briefings completed. Continue with pci-e bw evaluation.")
        bw_threshold = "4"
        output = os.popen('./gpubw/gpuLocalBandwidthTest.sh -t ' + bw_threshold)
        result = output.read()

        if "FAIL" not in result:
            print("Health Check successful. No report will be issued")
            print(result)
            fhand = open('./gpubw/gpuBandwidthTest.log')
            bws = ""
            for line in fhand:
                if "Bandwidth =" in line:
                    x = line.split("= ", 2)
                    y = x[1].split(" GB/s")
                    bws += y[0] + " "
            print(bws.strip())
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
            'name': "pciebw-"+nodename+"-"+dt.strftime("%d-%m-%Y-%H.%M.%S.%f")
        },
        'spec': {
            'node': nodename,
            'report': result,
            'issuer': "gpu-pciebw"
        }
    }
    group = "my.domain"
    v = "v1alpha1"
    plural = "healthcheckreports"
    try:
        api.create_namespaced_custom_object(group, v, namespace, plural, hcr_manifest)
    except ApiException as e:
        print("Exception when calling create health check report:\n", e)
   
    # raise TypeError("Failing init container.")
    print("Health Check unsuccessful. Here is the result:")
    print(result)
    return 0 
    # all_reports = api.list_namespaced_custom_object(group, v, namespace, plural)

if __name__ == '__main__':
    main()