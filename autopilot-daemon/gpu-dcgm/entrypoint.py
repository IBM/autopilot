import json
import subprocess
import os
import argparse

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
    result = subprocess.run(command, check=True, text=True, capture_output=True)
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
                    print(test['name'], ":", test['results'][0]['status'])
                    success = False
                    output+=(test['name']+" ")
        if success:
            print("[[ DCGM ]] SUCCESS")
        else:
            print("Host ", os.getenv("NODE_NAME"))
            print("[[ DCGM ]] FAIL")
            print(output.strip())

if __name__ == '__main__':
    main()