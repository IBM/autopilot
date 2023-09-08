import subprocess
import argparse
import netifaces

parser = argparse.ArgumentParser()
parser.add_argument('--replicas', type=int, default=1, help='Number of iperf3 servers per node')
args = vars(parser.parse_args())

server_replicas = args['replicas']

def main():
    interfaces = netifaces.interfaces()
    subset = int(server_replicas/(len(interfaces)-2)) # quite a lame bet.. excluding eth0 and lo assuming all the other ones are what we want.
    print("Number of servers per interface: " + str(subset) + " -- " + str(server_replicas) + " / " + str(len(interfaces)-2))
    for iface in interfaces:
        if iface != "lo" and iface != "eth0":
            address = netifaces.ifaddresses(iface)
            ip = address[netifaces.AF_INET] 
            for r in range(subset):
                if r > 9:
                    port = '51'+str(r)
                else:
                    port = '510'+str(r)
                command = ['iperf3', '-s', '-B', str(ip[0]['addr']), '-p', port, '-D']
                print("Start server on " + str(ip[0]['addr']) + " - iface " + str(iface))
                result = subprocess.run(command, text=True, capture_output=True)
                if result.stderr:
                    print(result.stderr)
                    print("Server fail to start on " + str(ip[0]['addr']) + " - iface " + str(iface) + ". Exited with error: " + result.stderr)
                else:
                    output = result.stdout
                    print(output)

if __name__ == '__main__':
    main()