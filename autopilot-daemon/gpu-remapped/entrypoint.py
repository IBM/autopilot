import os


def main():
    output = os.popen('bash ./gpu-remapped/briefings.sh')
    result = output.read()
    print(result)

    if "ABORT" not in result:
        print("Briefings completed. Continue with remapped rows evaluation.")
        output = os.popen('./gpu-remapped/remapped-rows.sh')
        result = output.read()

        print(result.strip())

    return 0 

if __name__ == '__main__':
    main()