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
            if "iperf3" in process and "-s" in process:
                parts = process.split()
                if len(parts) > 1:
                    pid = int(parts[1])
                    # Not killing default iperf server spun up on entrypoint...
                    if pid > 1:
                        log.info(
                            f"Killing iperf3 server process (PID: {pid}) in {CURR_POD_NAME} on {CURR_WORKER_NODE_NAME}"
                        )
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
                        log.info(
                            f"Nothing left to kill in {CURR_POD_NAME} on {CURR_WORKER_NODE_NAME} (Not killing default entrypoint iperf3 server)."
                        )
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


if __name__ == "__main__":
    kill_all_iperf_servers()
