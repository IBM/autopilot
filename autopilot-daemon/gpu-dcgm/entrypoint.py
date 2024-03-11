import json
import subprocess
import os
import argparse
import datetime
from kubernetes import client, config
from kubernetes.client.rest import ApiException

# load in cluster kubernetes config for access to cluster
config.load_incluster_config()
v1 = client.CoreV1Api()


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('-r', '--run', type=str, default='1')
    args = parser.parse_args()

    output = os.popen('bash ./utils/briefings.sh')
    result = output.read()
    print(result)

    if "ABORT" not in result:
        print("[[ DCGM ]] Briefings completed. Continue with dcgm evaluation.")
        command = ['dcgmi', 'diag', '-j', '-r', args.run]
        try_dcgm(command)
    else:
        print("[[ DCGM ]] ABORT")
        print(result)

def try_dcgm(command):
    try:
        result = subprocess.run(command, check=True, text=True, capture_output=True)
    except subprocess.CalledProcessError:
        print("[[ DCGM ]] DCGM process terminated with errors. Other processes might be running on GPUs. ABORT")
        command = ['nvidia-smi', '--query-gpu=utilization.gpu', '--format=csv']
        try:
            proc = subprocess.run(command, check=True, text=True, capture_output=True)
        except subprocess.CalledProcessError:
            print("[[ DCGM ]] nvidia-smi check terminated with errors. ABORT")
            exit()
        if proc.stdout:
            print("[[ DCGM ]] GPUs currently utilized:\n", proc.stdout)
        # exit()

    if result.stderr:
       print(result.stderr)
       print("[[ DCGM ]] exited with error: " + result.stderr + " ERR")
    #    exit()
    else:
        dcgm_dict = json.loads(result.stdout)
        tests_dict = dcgm_dict['DCGM GPU Diagnostic']['test_categories']
        success = True
        output = ""
        for category in tests_dict:
            for test in category['tests']:
                if test['results'][0]['status'] == 'Fail':
                    print(test['name'], ":", test['results'][0]['status'])
                    success = False
                    output+=(test['name']+" ")
        if success:
            print("[[ DCGM ]] SUCCESS")
        else:
            print("Host ", os.getenv("NODE_NAME"))
            print("[[ DCGM ]] FAIL")
            print(output.strip())

def label_node(success):
    now = datetime.datetime.now()
    timestamp = now.strftime("%Y/%m/%d-%H:%M:%S")
    try:
        node = v1.list_node(field_selector=nodelabel)
    except ApiException as e:
        print("Exception when calling CoreV1Api->list_node: %s\n" % e)
        exit()
    if success:
        print("labeling node with success")
    else:
        print("labeling node with dcgm issue")


if __name__ == '__main__':
    main()