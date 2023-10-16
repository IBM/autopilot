import os


def main():
    output = os.popen('bash ./gpu-remapped/briefings.sh')
    result = output.read()
    print(result)

    if "ABORT" not in result:
        print("[[ REMAPPED ROWS ]] Briefings completed. Continue with remapped rows evaluation.")
        output = os.popen('./gpu-remapped/remapped-rows.sh')
        result = output.read()
        if "FAIL" not in result:
            print("[[ REMAPPED ROWS ]] SUCCESS")
            # print(result)
        else:
            print("[[ REMAPPED ROWS ]] FAIL")
            print("Host ", os.getenv("NODE_NAME"))
            return 0
        print("Host ", os.getenv("NODE_NAME"))
        print(result.strip())
    else:
        print("[[ REMAPPED ROWS ]] ABORT")
        print(result)

if __name__ == '__main__':
    main()