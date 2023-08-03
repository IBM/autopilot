import os
import subprocess
import sys



def main():
    print("[[ MULTINIC-CNI-STATUS ]] Evaluating reachability of Multi-NIC CNI.")
    nodename = os.getenv("NODE_NAME")
    command = ['python3', './network/read_status.py', nodename]
    timeout_s = 30
    try:
        result = subprocess.run(command, text=True, capture_output=True, timeout=timeout_s)
    except subprocess.TimeoutExpired:
        print("Multi-NIC CNI health checker is not reachable - network reachability test cannot run")
        sys.exit(0)

    if result.stderr:
        print(result.stderr)
        print("Multi-NIC CNI health checker is not reachable - network reachability test cannot run")
        sys.exit(0)
    else:
        output = result.stdout
        print(output)
        if "OK" in output:
            print("[[ MULTINIC-CNI-STATUS ]] SUCCESS")
        else:
            print("[[ MULTINIC-CNI-STATUS ]] FAIL")
            print("Host ", os.getenv("NODE_NAME"))
        if "cannot" in output:
            print("Multi-NIC CNI health checker is not reachable - network reachability test cannot run")
            sys.exit(0)
            
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

if __name__ == '__main__':
    main()