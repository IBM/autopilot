import json
import subprocess
import os
import argparse
import re
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


# translate key-strings into lowercase and strip spaces
def unify_string_format(key: str) -> str:
    to_lower = key.strip().lower()
    res, _ = re.subn('[\/|\s]', '_', to_lower)
    return res

def parse_all_results(result: str):
    dcgm_dict = json.loads(result)
    tests_dict = dcgm_dict['DCGM GPU Diagnostic']['test_categories']
    success = True
    output = ""
    for category in tests_dict:
        for test in category['tests']:
            if test['results'][0]['status'] == 'Fail':
                success = False
                print(test['name'], ":", test['results'][0]['status'])
                output+=f'{unify_string_format(test["name"])}'
                for entry in test['results']:
                    output+=f'{"."+str(entry["gpu_id"]) if "gpu_id" in entry else "NoGPUid"}'
    return success, output


# parsing the json result string based on a comma-separated list of paths (levels separated by '.')
def parse_selected_results(result: str, testpaths: str):
    '''
    follow the list of selected paths down the dcgm json tree

    the specification of the paths: <top_level>.<category>.<name>

    to walk down this example json snippet below your path should be:

       'DCGM GPU Diagnostic.Hardware.GPU Memory'

    for the search, all strings are turned to lowercase and spaces are replaced with '_'
    therefore the following path would achieve the same:

        'dcgm_gpu_diagnostic.HARDWare.gpu Memory'

    "DCGM GPU Diagnostic" : {
        "test_categories" : [ {
            ...
            "category" : "Hardware",
            "tests" : [ {
                "name" : "GPU Memory",
                "results" : [ {
                    "gpu_id" : "0",
                    "status" : "Fail",
         ...


    The paths need to be specified in env variable AUTOPILOT_DCGM_RESULT_PATHS as a comma-separated list
    If the variable is not set, then the regular scan is performed
    '''
    _dcgm_json_levels = [
        ("top_level","dcgm_gpu_diagnostic"),
        ("category","tests"),
        ("name","results")
    ]


    # scan the dictionary and recursively transform all keys using key_update
    def normalize_json_keys(data) -> dict:
        ndata = {}
        if not isinstance(data, dict) and not isinstance(data, list):
            return data
        for key,val in data.items():
            key_n = unify_string_format(key)

            if isinstance(val, dict):
                val_n = normalize_json_keys(data[key])
            elif isinstance(val, list):
                val_n = [ normalize_json_keys(v) for v in val ]
            else:
                val_n = data[key]

            ndata[ key_n ] = val_n

        # unfortunately, the top level of dcgm dict is structured differently from the rest,
        # adjusting by inserting/moving it's sub-dict into top-level and rename
        if _dcgm_json_levels[0][1] in ndata:
            ndata[_dcgm_json_levels[0][0]] = _dcgm_json_levels[0][1] # replace old dcgm_gpu_diagnostics with  'top_level' as a name
            ndata[_dcgm_json_levels[0][1]] = ndata[_dcgm_json_levels[0][1]].pop("test_categories") # move test_categories entry to new 'top_level'
        return ndata


    # recursively dive into the json tree by following a given path
    def dive_to_test(data, jpath: list[str], depth: int):
        assert( 3-len(jpath) == depth )
        assert( depth < 3 )

        jlevel_spec = _dcgm_json_levels[depth]

        if not isinstance(data, list):
            data = [data]
        for entry in data:
            if jlevel_spec[0] in entry and jpath[0] == unify_string_format( entry[jlevel_spec[0]] ):
                if depth == 2:
                    return entry[ jlevel_spec[1] ]
                else:
                    return dive_to_test( entry[ jlevel_spec[1] ], jpath[1:], depth+1 )
        return

    # browses the result section of a single test and extracts info
    def parse_single_test_result(data) -> tuple[bool, str]:
        if not data:
            return False, "No Data"
        if not isinstance(data, list):
            data = [data]

        success = True
        output = []
        for entry in data:
            if "status" in entry:
                good = (unify_string_format(entry['status']) == 'pass')
                success &= good
                if not good:
                    output.append( (
                        entry["gpu_id"] if "gpu_id" in entry else "NoGPU_ID",
                        entry["info"] if "info" in entry else "NoInfo"
                    ))
            else:
                success &= False
                output.append( ("No Status") )
        return success,output

    # create output from the parsed results (can be adjusted to whatever)
    def build_output(output_list: tuple[str, str]) -> str:
        print(output_list)
        output = ""
        for test,result in output_list:
            if len(output):
                output += ";"
            output += f'{unify_string_format(test)}:'
            for result_data in result:
                for r in result_data:
                    output += f'{unify_string_format(r)},'
        return output

    jdata = json.load(result)
    norm_d = normalize_json_keys(jdata)

    result_list = []
    overall_success = True
    for path in testpaths.split(','):
        single_test_result = dive_to_test( norm_d, [ unify_string_format(p) for p in path.split('.') ], 0 )
        test_success,output = parse_single_test_result(single_test_result)
        overall_success &= test_success
        if not test_success:
            result_list.append( (path, output) )
    return overall_success, build_output(result_list)



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
        testpaths = os.getenv("AUTOPILOT_DCGM_RESULT_PATHS")
        if testpaths == None:
            success, output = parse_all_results(result.stdout)
        if success:
            print("[[ DCGM ]] SUCCESS")
        else:
            print("Host", nodename)
            print("[[ DCGM ]] FAIL")
        if args.label_node:
            patch_node(success, output)


def patch_node(success, output):
    now = datetime.datetime.now(datetime.timezone.utc)
    timestamp = now.strftime("%Y-%m-%d_%H.%M.%SUTC")
    result = ""
    general_health = "PASS"
    if success:
        result = "PASS_"+timestamp
    else:
        result = "ERR_"+timestamp+"_"+output
        general_health = "EVICT"

    label = {
        "metadata": {
            "labels": {
                "autopilot.ibm.com/dcgm.level.3": result,
                "autopilot.ibm.com/gpuhealth": general_health}
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
