import argparse
import os


def main():
    
    parser = argparse.ArgumentParser()
    parser.add_argument('-t', '--threshold', type=str, default='4')
    args = parser.parse_args()
    output = os.popen('bash ./utils/briefings.sh')
    result = output.read()
    # print(result)

    if "ABORT" not in result:
        print("[[ PCIEBW ]] Briefings completed. Continue with PCIe Bandwidth evaluation.")
        output = os.popen('./gpu-bw/gpuLocalBandwidthTest.sh -t ' + args.threshold)
        result = output.read()

        if "ABORT" in result or "SKIP" in result:
            print("[[ PCIEBW ]] ABORT")
            print(result)
            exit()

        print("SUCCESS")
        print("Host ", os.getenv("NODE_NAME"))
        splitres = result.split("\n")
        bws = ""
        for line in splitres:
            if "Bandwidth =" in line:
                x = line.split("= ", 2)
                y = x[1].split(" GB/s")
                bws += y[0] + " "
        print(bws.strip())
    else:
        print("[[ PCIEBW ]] ABORT")
        print(result)


if __name__ == '__main__':
    main()