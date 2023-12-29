import os

def main():
    output = os.popen('bash ./utils/briefings.sh')
    result = output.read()

    if "ABORT" not in result:
        print("[[ GPU-MEM ]] Briefings completed. Continue with memory evaluation.")
        output = os.popen('./gpu-mem/gpucheck')
        result = output.read()
        if "NONE" in result:
            print("[[ GPU-MEM ]] Health Check successful")
            exit()

    print("[[ GPU-MEM ]] Health Check unsuccessful. FAIL.")
    print(result)
    exit()

if __name__ == '__main__':
    main()