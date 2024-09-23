from iperf3_utils import *

parser = argparse.ArgumentParser()
parser.add_argument(
    "--dstip",
    type=str,
    default="",
    help=("IP for iperf3 server"),
)

parser.add_argument(
    "--dstport",
    type=int,
    default=5200,
    help=("Port corresponding to the dstip"),
)

parser.add_argument(
    "--numclients",
    type=int,
    default=1,
    help=("Number of clients to run simultaneously against dstip:dstport"),
)
args = vars(parser.parse_args())


async def run_iperf_client(dstip, dstport, iteration, duration_seconds):
    dstport = dstport + iteration
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

    process = await asyncio.create_subprocess_exec(
        *command, stdout=asyncio.subprocess.PIPE, stderr=asyncio.subprocess.PIPE
    )
    stdout, stderr = await process.communicate()

    output_filename = f"{dstip}_{dstport}_client.log"

    if process.returncode == 0:
        iperf3_stdout = stdout.decode()
        with open(output_filename, "w") as f:
            f.write(iperf3_stdout)
        iperf3_stdout = iperf3_stdout.strip().splitlines()
        sender_transfer = ""
        sender_transfer_units = ""
        sender_bitrate = ""
        sender_bitrate_units = ""
        receiver_transfer = ""
        receiver_transfer_units = ""
        receiver_bitrate = ""
        receiver_bitrate_units = ""
        for line in iperf3_stdout:
            line = line.lower()
            if "sender" in line:
                line = line.split()
                sender_transfer = line[4]
                sender_transfer_units = line[5]
                sender_bitrate = line[6]
                sender_bitrate_units = line[7]
            elif "receiver" in line:
                line = line.split()
                receiver_transfer = line[4]
                receiver_transfer_units = line[5]
                receiver_bitrate = line[6]
                receiver_bitrate_units = line[7]
        iperf_result = {
            "sender": {
                "transfer": {"rate": sender_transfer, "units": sender_transfer_units},
                "bitrate": {
                    "rate": sender_bitrate,
                    "units": sender_bitrate_units,
                },
            },
            "receiver": {
                "transfer": {
                    "rate": receiver_transfer,
                    "units": receiver_transfer_units,
                },
                "bitrate": {
                    "rate": receiver_bitrate,
                    "units": receiver_bitrate_units,
                },
            },
        }
        iperf_result = {
            "interface": {"ip": dstip, "port": dstport},
            "results": iperf_result,
        }
        return iperf_result
    else:
        log.error(
            f"iperf3 client number {iteration} on {dstip}:{dstport} failed with error: {stderr.decode()}"
        )
        with open(output_filename, "w") as f:
            f.write(stderr.decode())


async def main():
    dstip = args["dstip"]
    dstport = args["dstport"]
    numclients = args["numclients"]
    duration_seconds = "5"

    tasks = [
        asyncio.create_task(run_iperf_client(dstip, dstport, i, duration_seconds))
        for i in range(numclients)
    ]

    results = await asyncio.gather(*tasks)

    sender_transfer_sum = 0.0
    sender_min_transfer = 1000.0
    sender_max_transfer = 0.0

    sender_bitrate_sum = 0.0
    sender_min_bitrate = 1000.0
    sender_max_bitrate = 0.0

    receiver_transfer_sum = 0.0
    receiver_min_transfer = 1000.0
    receiver_max_transfer = 0.0

    receiver_bitrate_sum = 0.0
    receiver_min_bitrate = 1000.0
    receiver_max_bitrate = 0.0
    total_results = {}
    for index, el in enumerate(results):
        total_results[str(index)] = el
        sender_transfer = float(el["results"]["sender"]["transfer"]["rate"])
        sender_bitrate = float(el["results"]["sender"]["bitrate"]["rate"])

        sender_transfer_sum = sender_transfer_sum + sender_transfer
        sender_bitrate_sum = sender_bitrate_sum + sender_bitrate

        if sender_transfer > sender_max_transfer:
            sender_max_transfer = sender_transfer
        elif sender_transfer < sender_min_transfer:
            sender_min_transfer = sender_transfer

        if sender_bitrate > sender_max_bitrate:
            sender_max_bitrate = sender_bitrate
        elif sender_bitrate < sender_min_bitrate:
            sender_min_bitrate = sender_bitrate

        receiver_transfer = float(el["results"]["receiver"]["transfer"]["rate"])
        receiver_bitrate = float(el["results"]["receiver"]["bitrate"]["rate"])

        receiver_transfer_sum = receiver_transfer_sum + receiver_transfer
        receiver_bitrate_sum = receiver_bitrate_sum + receiver_bitrate

        if receiver_transfer > receiver_max_transfer:
            receiver_max_transfer = receiver_transfer
        elif receiver_transfer < receiver_min_transfer:
            receiver_min_transfer = receiver_transfer

        if receiver_bitrate > receiver_max_bitrate:
            receiver_max_bitrate = receiver_bitrate
        elif receiver_bitrate < receiver_min_bitrate:
            receiver_min_bitrate = receiver_bitrate

    stats = {
        "sender": {
            "aggregate": {
                "transfer": str(round(Decimal(sender_transfer_sum), 2)),
                "bitrate": str(round(Decimal(sender_bitrate_sum), 2)),
            },
            "mean": {
                "transfer": str(round(Decimal(sender_transfer_sum / numclients), 2)),
                "bitrate": str(round(Decimal(sender_bitrate_sum / numclients), 2)),
            },
            "min": {
                "transfer": str(round(Decimal(sender_min_transfer), 2)),
                "bitrate": str(round(Decimal(sender_min_bitrate), 2)),
            },
            "max": {
                "transfer": str(round(Decimal(sender_max_transfer), 2)),
                "bitrate": str(round(Decimal(sender_max_bitrate), 2)),
            },
        },
        "receiver": {
            "aggregate": {
                "transfer": str(round(Decimal(receiver_transfer_sum), 2)),
                "bitrate": str(round(Decimal(receiver_bitrate_sum), 2)),
            },
            "mean": {
                "transfer": str(round(Decimal(receiver_transfer_sum / numclients), 2)),
                "bitrate": str(round(Decimal(receiver_bitrate_sum / numclients), 2)),
            },
            "min": {
                "transfer": str(round(Decimal(receiver_min_transfer), 2)),
                "bitrate": str(round(Decimal(receiver_min_bitrate), 2)),
            },
            "max": {
                "transfer": str(round(Decimal(receiver_max_transfer), 2)),
                "bitrate": str(round(Decimal(receiver_max_bitrate), 2)),
            },
        },
    }

    total_results["stats"] = stats
    summary_file = f"{dstip}_summary.json"
    with open(summary_file, "w") as f:
        f.write(json.dumps(total_results, indent=4))
        f.write("\n")

    print(json.dumps(total_results["stats"], indent=4))


if __name__ == "__main__":
    asyncio.run(main())
