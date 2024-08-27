from iperf3_utils import *


def kill_all_iperf_servers():
    try:
        result = subprocess.run(
            ["ps", "aux"], text=True, capture_output=True, check=True
        )
    except subprocess.CalledProcessError as e:
        print(f"Error occurred while listing processes: {e}")
        sys.exit(1)

    processes = result.stdout.splitlines()

    for process in processes:
        try:
            # Don't combine the strings...this won't work if "-s" is placed in a different position...
            if "iperf3" in process and "-s" in process:
                parts = process.split()
                if len(parts) > 1:
                    pid = int(parts[1])
                    if pid > 1:
                        try:
                            os.kill(pid, signal.SIGTERM)
                        except PermissionError:
                            log.error(
                                f"Permission denied: Could not kill process with PID {pid} in {CURR_POD_NAME} on {CURR_WORKER_NODE_NAME}."
                            )
                            sys.exit(1)
                        except ProcessLookupError:
                            log.error(
                                f"Process with PID {pid} does not exist in {CURR_POD_NAME} on {CURR_WORKER_NODE_NAME}."
                            )
                            sys.exit(1)
                        except Exception as e:
                            log.error(
                                f"Failed to kill process with PID {pid}: {e} in {CURR_POD_NAME} on {CURR_WORKER_NODE_NAME}"
                            )
                            sys.exit(1)
                else:
                    log.error(
                        f"Unexpected format in process line: {process} in {CURR_POD_NAME} on {CURR_WORKER_NODE_NAME}"
                    )
                    sys.exit(1)
        except ValueError:
            log.error(
                f"Could not convert PID to an integer: {process} in {CURR_POD_NAME} on {CURR_WORKER_NODE_NAME}"
            )
            sys.exit(1)
        except Exception as e:
            log.error(
                f"An unexpected error occurred: {e} in {CURR_POD_NAME} on {CURR_WORKER_NODE_NAME}"
            )
            sys.exit(1)
    log.info(f"All iperf servers have been removed (not deleting default iperf server)")


if __name__ == "__main__":
    kill_all_iperf_servers()

