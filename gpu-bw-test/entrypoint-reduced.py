
import os

def main():
    output = os.popen('bash ./briefings.sh')
    result = output.read()
    print(result)

    if "ABORT" not in result:
        print("Briefings completed. Continue with pci-e bw evaluation.")
        bw_threshold = os.getenv("BW")
        if not bw_threshold:
            bw_threshold = "4"
        output = os.popen('./gpuLocalBandwidthTest.sh -t ' + bw_threshold)
        result = output.read()
        print(result)



if __name__ == '__main__':
    main()