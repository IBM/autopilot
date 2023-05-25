from kubernetes import client
from kubernetes.client.rest import ApiException
import os
from datetime import datetime
import subprocess
from subprocess import Popen
import sys



def main():
   
    nodename = os.getenv("NODE_NAME")
    namespace = os.getenv("NAMESPACE")
    command = ['python', 'read_status.py', nodename]
    result = subprocess.run(command, capture_output=True, text=True)

    if result.stderr:
        raise SystemExit("Multi-NIC CNI health checker is not reachable - network reachability test cannot run")
    else:
        output = result.stdout
        print(output)

    if "OK" in output:
        print("Node " + nodename + " is reachable")
    else:
        alert = "Node " + nodename + " is not reachable"
        print(alert)
    
    connectable = output.split("Connectable network devices: ")[1]
    devices = int(connectable.split("/")[0])
    if devices == 2:
        lastline = nodename + " 1 1"
    elif devices == 1:
        lastline = nodename + " 1 0"
    elif devices == 0:
        lastline = nodename + " 0 0"
    else:
        lastline = "Cannot determine connectable devices"
    
    print("\n" + lastline)

        # api = client.CustomObjectsApi()

        # dt = datetime.now()
        # hcr_manifest = {
        #     'apiVersion': 'my.domain/v1alpha1',
        #     'kind': 'HealthCheckReport',
        #     'metadata': {
        #         'name': "netcheck-"+nodename+"-"+dt.strftime("%d-%m-%Y-%H.%M.%S.%f")
        #     },
        #     'spec': {
        #         'node': nodename,
        #         'report': alert,
        #         'issuer': "net-reach"
        #     }
        # }
        # group = "my.domain"
        # v = "v1alpha1"
        # plural = "healthcheckreports"
        # try:
        #     api.create_namespaced_custom_object(group, v, namespace, plural, hcr_manifest)
        # except ApiException as e:
        #     print("Exception when calling create health check report:\n", e)  

if __name__ == '__main__':
    main()