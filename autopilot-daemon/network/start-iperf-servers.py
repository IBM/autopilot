import subprocess
import argparse
import netifaces
import os

parser = argparse.ArgumentParser()
parser.add_argument('--replicas', type=int, default=1, help='Number of iperf3 servers per node. If #replicas is less than the number of secondary nics, it will create #replicas server per nic. Otherwise, it will spread #replicas servers as evenly as possible on all interfaces')
args = vars(parser.parse_args())

server_replicas = args['replicas']

def main():
    # interfaces = netifaces.interfaces()
    interfaces = [iface for iface in netifaces.interfaces() if "net" in iface] ## VERY TEMPORARY
    if len(interfaces)==0:
        print("[IPERF] Cannot launch servers -- secondary nics not found ", os.getenv("POD_NAME"), ". ABORT")
        return

    secondary_nics_count = len(interfaces) # quite a lame bet.. excluding eth0 and lo assuming all the other ones are what we want.
    if server_replicas > secondary_nics_count:
        subset = int(server_replicas/secondary_nics_count) 
    else:
        subset = server_replicas
    print("[IPERF] Number of servers per interface: " + str(subset) + " -- " + str(server_replicas) + " / " + str(secondary_nics_count))
    for iface in interfaces:
        if iface != "lo" and iface != "eth0":
            address = netifaces.ifaddresses(iface)
            ip = address[netifaces.AF_INET] 
            for r in range(subset):
                if r <= 9:
                    port = '510'+str(r)
                else:
                    port = '51'+str(r)
                command = ['iperf3', '-s', '-B', str(ip[0]['addr']), '-p', port, '-D', '-1']
                # print("Start server on " + str(ip[0]['addr']) + " - iface " + str(iface))
                result = subprocess.run(command, text=True, capture_output=True)
                if result.stderr:
                    print(result.stderr)
                    print("Server fail to start on " + str(ip[0]['addr']) + " - iface " + str(iface) + ". Exited with error: " + result.stderr)
                # else:
                #     output = result.stdout
                #     print(output)

if __name__ == '__main__':
    main()