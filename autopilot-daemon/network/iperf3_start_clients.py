import argparse
import asyncio
import json
from decimal import Decimal
from iperf3_utils import *

parser = argparse.ArgumentParser()
parser.add_argument("--dstip", type=str, default="", help="IP for iperf3 server")
parser.add_argument("--dstport", type=int, default=5200, help="Port for iperf3 server")
parser.add_argument("--numclients", type=int, default=1, help="Number of clients")
args = parser.parse_args()


async def run_iperf_client(dstip, dstport, iteration, duration_seconds):
    dstport += iteration
    command = [
        "iperf3",
        "-c",
        dstip,
        "-p",
        str(dstport),
        "-t",
        duration_seconds,
        "-i",
        "1.0",
        "-Z",
    ]

    default_res = {
        "interface": {"ip": dstip, "port": dstport},
        "results": {
            "sender": {
                "transfer": {"rate": 0.0, "units": "n/a"},
                "bitrate": {"rate": 0.0, "units": "n/a"},
            },
            "receiver": {
                "transfer": {"rate": 0.0, "units": "n/a"},
                "bitrate": {"rate": 0.0, "units": "n/a"},
            },
        },
    }

    try:
        process = await asyncio.wait_for(
            asyncio.create_subprocess_exec(
                *command, stdout=asyncio.subprocess.PIPE, stderr=asyncio.subprocess.PIPE
            ),
            timeout=60,
        )
        stdout, stderr = await process.communicate()
        output_filename = f"{dstip}_{dstport}_client.log"
        with open(output_filename, "w") as f:
            f.write(stdout.decode())
    except asyncio.TimeoutError:
        log.error(
            f"iperf3 client {iteration} on {dstip}:{dstport} failed, time-out exceeded: {stderr.decode()}"
        )
        return {"interface": {"ip": dstip, "port": dstport}, "results": default_res}
    except Exception as e:
        log.error(
            f"iperf3 client {iteration} on {dstip}:{dstport} failed with {e}: {stderr.decode()}"
        )
        return {"interface": {"ip": dstip, "port": dstport}, "results": default_res}

    # In theory this should not occur since we catch this above...but just to be safe let's ensure
    # the return code is zero...
    if process.returncode != 0:
        log.error(
            f"iperf3 client {iteration} on {dstip}:{dstport} failed: {stderr.decode()}"
        )
        return {"interface": {"ip": dstip, "port": dstport}, "results": default_res}

    iperf3_stdout = stdout.decode().strip().splitlines()
    for line in iperf3_stdout:
        result = {
            "sender": {"transfer": {}, "bitrate": {}},
            "receiver": {"transfer": {}, "bitrate": {}},
        }
        line = line.lower()
        if "sender" in line:
            parts = line.split()
            result["sender"]["transfer"] = {"rate": parts[4], "units": parts[5]}
            result["sender"]["bitrate"] = {"rate": parts[6], "units": parts[7]}
        elif "receiver" in line:
            parts = line.split()
            result["receiver"]["transfer"] = {"rate": parts[4], "units": parts[5]}
            result["receiver"]["bitrate"] = {"rate": parts[6], "units": parts[7]}

    return {"interface": {"ip": dstip, "port": dstport}, "results": result}


def calculate_stats(values, num_clients):
    return {
        "aggregate": {
            "transfer": str(round(Decimal(sum(values["transfer"])), 2)),
            "bitrate": str(round(Decimal(sum(values["bitrate"])), 2)),
        },
        "mean": {
            "transfer": str(round(Decimal(sum(values["transfer"]) / num_clients), 2)),
            "bitrate": str(round(Decimal(sum(values["bitrate"]) / num_clients), 2)),
        },
        "min": {
            "transfer": str(round(Decimal(min(values["transfer"])), 2)),
            "bitrate": str(round(Decimal(min(values["bitrate"]))), 2),
        },
        "max": {
            "transfer": str(round(Decimal(max(values["transfer"])), 2)),
            "bitrate": str(round(Decimal(max(values["bitrate"]))), 2),
        },
    }


async def main():
    dstip, dstport, numclients = args.dstip, args.dstport, args.numclients
    duration_seconds = "5"

    tasks = [
        asyncio.create_task(run_iperf_client(dstip, dstport, i, duration_seconds))
        for i in range(numclients)
    ]
    results = await asyncio.gather(*tasks)

    sender_values = {"transfer": [], "bitrate": []}
    receiver_values = {"transfer": [], "bitrate": []}

    total_results = {}
    for idx, result in enumerate(results):
        total_results[str(idx)] = result
        sender_values["transfer"].append(
            float(result["results"]["sender"]["transfer"]["rate"])
        )
        sender_values["bitrate"].append(
            float(result["results"]["sender"]["bitrate"]["rate"])
        )
        receiver_values["transfer"].append(
            float(result["results"]["receiver"]["transfer"]["rate"])
        )
        receiver_values["bitrate"].append(
            float(result["results"]["receiver"]["bitrate"]["rate"])
        )

    stats = {
        "sender": calculate_stats(sender_values, numclients),
        "receiver": calculate_stats(receiver_values, numclients),
    }

    total_results["stats"] = stats
    summary_file = f"{dstip}_summary.json"
    with open(summary_file, "w") as f:
        json.dump(total_results, f, indent=4)

    print(json.dumps(stats, indent=4))


if __name__ == "__main__":
    asyncio.run(main())
