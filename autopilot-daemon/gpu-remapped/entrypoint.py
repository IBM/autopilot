import os


def main():
    output = os.popen('bash ./gpu-power/briefings.sh')
    result = output.read()
    print(result)

    if "ABORT" not in result:
        print("[[ POWER THROTTLE ]] Briefings completed. Continue with power throttle evaluation.")
        output = os.popen('./gpu-power/power-throttle.sh')
        result = output.read()
        if "FAIL" not in result:
            print("[[ POWER THROTTLE ]] SUCCESS")
            # print(result)
        else:
            print("[[ POWER THROTTLE ]] FAIL")
            print("Host ", os.getenv("NODE_NAME"))
            return 0
        print("Host ", os.getenv("NODE_NAME"))
        print(result.strip())
    else:
        print("[[ POWER THROTTLE ]] ABORT")
        print(result)

if __name__ == '__main__':
    main()