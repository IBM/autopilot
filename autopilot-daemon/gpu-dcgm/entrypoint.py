import json
import subprocess
import os
import argparse
import datetime
from kubernetes import client, config
from kubernetes.client.rest import ApiException

config.load_incluster_config()
v1 = client.CoreV1Api()
nodename = os.getenv("NODE_NAME")

parser = argparse.ArgumentParser()
parser.add_argument('-r', '--run', type=str, default='1')
parser.add_argument('-l', '--label_node', action='store_true')
args = parser.parse_args()
def main():
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
    result = subprocess.run(command, text=True, capture_output=True)
    return_code = result.returncode  # 0 for success
    if return_code != 0:
        print("[[ DCGM ]] DCGM process terminated with errors. Other processes might be running on GPUs. ABORT")
        command = ['nvidia-smi', '--query-gpu=utilization.gpu', '--format=csv']
        try:
            proc = subprocess.run(command, check=True, text=True, capture_output=True)
        except subprocess.CalledProcessError:
            print("[[ DCGM ]] nvidia-smi check terminated with errors. ABORT")
            exit()
        if proc.stdout:
            print("[[ DCGM ]] GPUs currently utilized:\n", proc.stdout)
    
    if result.stderr:
       print(result.stderr)
       print("[[ DCGM ]] exited with error: " + result.stderr + " ERR")
    else:
        dcgm_dict = json.loads(result.stdout)
        tests_dict = dcgm_dict['DCGM GPU Diagnostic']['test_categories']
        success = True
        output = ""
        for category in tests_dict:
            for test in category['tests']:
                if test['results'][0]['status'] == 'Fail':
                    success = False
                    print(test['name'], ":", test['results'][0]['status'])
                    if test['name'] == "GPU Memory":
                        output+=(test['name'].replace(" ","")+"_")
                        for entry in test['results']:
                            output+=("."+entry['gpu_id'])
        if success:
            print("[[ DCGM ]] SUCCESS")
        else:
            print("Host", nodename)
            print("[[ DCGM ]] FAIL")
        if args.label_node:
            patch_node(success, output)
    

def patch_node(success, output):
    now = datetime.datetime.now(datetime.timezone.utc)
    #  ADD UTC
    timestamp = now.strftime("%Y-%m-%d_%H.%M.%SUTC")
    result = ""
    if success:
        result = "PASS_"+timestamp
    else:
        result = "ERR_"+timestamp+"_"+output

    label = {
        "metadata": {
            "labels": {
                "autopilot/dcgm.level.3": result}
        }
    }
    print("label: ", result)
    try:
        api_response = v1.patch_node(nodename, label)
    except ApiException as e:
        print("Exception when calling corev1api->patch_node: %s\n" % e)
        exit()

if __name__ == '__main__':
    main()
