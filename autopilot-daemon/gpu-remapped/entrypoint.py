import os


def main():
    output = os.popen('bash ./utils/briefings.sh')
    result = output.read()
    print(result)

    if "ABORT" not in result:
        print("[[ REMAPPED ROWS ]] Briefings completed. Continue with remapped rows evaluation.")
        output = os.popen('./gpu-remapped/remapped-rows.sh')
        result = output.read()
        if "FAIL" not in result:
            print("[[ REMAPPED ROWS ]] SUCCESS")
        else:
            print("[[ REMAPPED ROWS ]] FAIL")
            print("Host ", os.getenv("NODE_NAME"))
            print(result.strip())
            return 0
        print("Host ", os.getenv("NODE_NAME"))
        print(result.strip())
    else:
        print("[[ REMAPPED ROWS ]] ABORT")
        print(result.strip())

if __name__ == '__main__':
    main()