package handlers

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.ibm.com/hybrid-cloud-infrastructure-research/autopilot-daemon/pkg/utils"
	"k8s.io/klog/v2"
)

func TimerRun() {
	klog.Info("Running a periodic check")
	runAllTestsLocal("pciebw,remapped,dcgm", "1")
}

func runAllTestsLocal(checks string, dcgmR string) (error, *[]byte) {
	out := []byte("")
	var tmp *[]byte
	var err error
	start := time.Now()
	for _, check := range strings.Split(checks, ",") {
		switch check {
		case "dcgm":
			klog.Info("Running health check: ", check)
			err, tmp = runDCGM(dcgmR)
			if err != nil {
				klog.Error(err.Error())
				return err, tmp
			}
			out = append(out, *tmp...)

		case "pciebw":
			klog.Info("Running health check: ", check)
			err, tmp = runPCIeBw()
			if err != nil {
				klog.Error(err.Error())
				return err, tmp
			}
			out = append(out, *tmp...)

		case "remapped":
			klog.Info("Running health check: ", check)
			err, tmp = runRemappedRows()
			if err != nil {
				klog.Error(err.Error())
				return err, nil
			}
			out = append(out, *tmp...)

		case "nic":
			klog.Info("Running health check: ", check, " -- DISABLED")

		case "all":
			klog.Info("Run all health checks\n")
			err, tmp := runPCIeBw()
			if err != nil {
				klog.Error(err.Error())
				return err, tmp
			}
			out = append(out, *tmp...)
			err, tmp = runRemappedRows()
			if err != nil {
				klog.Error(err.Error())
				return err, nil
			}
			out = append(out, *tmp...)
			err, tmp = runDCGM(dcgmR)
			if err != nil {
				klog.Error(err.Error())
				return err, tmp
			}
			out = append(out, *tmp...)

			// err, tmp = netReachability()
			// if err != nil {
			// 	klog.Error(err.Error())
			// 	return err, nil
			// }
			// out = append(out, *tmp...)
		default:
			notsupported := "check not supported: " + check
			out = append(out, []byte(notsupported)...)
		}
	}
	end := time.Now()
	diff := end.Sub(start)
	klog.Info("Total time (s) for all checks: ", diff.Seconds())
	return nil, &out
}

func runAllTestsRemote(host string, check string, batch string, jobName string, dcgmR string) (error, *[]byte) {
	klog.Info("About to run command:\n", "./utils/runHealthchecks.py", " --service=autopilot-healthchecks", " --namespace="+os.Getenv("NAMESPACE"), " --nodes="+host, " --check="+check, " --batchSize="+batch, " --wkload="+jobName, " --dcgmR="+dcgmR)

	out, err := exec.Command("python3", "./utils/runHealthchecks.py", "--service=autopilot-healthchecks", "--namespace="+os.Getenv("NAMESPACE"), "--nodes="+host, "--check="+check, "--batchSize="+batch, "--wkload="+jobName, "--dcgmR="+dcgmR).Output()
	if err != nil {
		klog.Info("Out:", string(out))
		klog.Error(err.Error())
		return err, nil
	}

	return nil, &out
}

func netReachability() (error, *[]byte) {
	out, err := exec.Command("python3", "./network/metrics-entrypoint.py").CombinedOutput()
	if err != nil {
		klog.Info("Out:", string(out))
		klog.Error(err.Error())
		return err, nil
	} else {
		klog.Info("Secondary NIC health check test completed:")

		if strings.Contains(string(out[:]), "FAIL") {
			klog.Info("Multi-NIC CNI reachability test failed.", string(out[:]))
		} else if strings.Contains(string(out[:]), "cannot") {
			klog.Info("Unable to determine the reachability of the node.", string(out[:]))
		} else {
			output := strings.TrimSuffix(string(out[:]), "\n")
			split := strings.Split(output, "\n")
			lastline := split[len(split)-1]
			final := strings.Split(lastline, " ")
			var nicid1, nicid2 int = 1, 2
			if reachable1, err := strconv.ParseFloat(final[1], 32); err == nil {
				klog.Info("Observation: ", os.Getenv("NODE_NAME"), " ", strconv.Itoa(nicid1), " ", reachable1)
				utils.HchecksGauge.WithLabelValues("net-reach", os.Getenv("NODE_NAME"), strconv.Itoa(nicid1)).Set(reachable1)
			}
			if reachable2, err := strconv.ParseFloat(final[2], 32); err == nil {
				klog.Info("Observation: ", os.Getenv("NODE_NAME"), " ", strconv.Itoa(nicid2), " ", reachable2)
				utils.HchecksGauge.WithLabelValues("net-reach", os.Getenv("NODE_NAME"), strconv.Itoa(nicid2)).Set(reachable2)
			}
		}

	}
	return nil, &out
}

func runRemappedRows() (error, *[]byte) {
	out, err := exec.Command("python3", "./gpu-remapped/entrypoint.py").CombinedOutput()
	if err != nil {
		klog.Info("Out:", string(out))
		klog.Error(err.Error())
		return err, nil
	} else {
		klog.Info("Remapped Rows check test completed:")

		if strings.Contains(string(out[:]), "FAIL") {
			klog.Info("Remapped Rows test failed.", string(out[:]))
		}

		if strings.Contains(string(out[:]), "ABORT") {
			klog.Info("Remapped Rows cannot be run. ", string(out[:]))
			return nil, &out
		}

		output := strings.TrimSuffix(string(out[:]), "\n")
		split := strings.Split(output, "\n")
		rmr := split[len(split)-1]
		final := strings.Split(rmr, " ")

		for gpuid, v := range final {
			rm, err := strconv.ParseFloat(v, 64)
			if err != nil {
				klog.Error(err.Error())
				return err, nil
			} else {
				klog.Info("Observation: ", os.Getenv("NODE_NAME"), " ", strconv.Itoa(gpuid), " ", rm)
				// utils.Hchecks.WithLabelValues("remapped", os.Getenv("NODE_NAME"), strconv.Itoa(gpuid)).Observe(rm)
				utils.HchecksGauge.WithLabelValues("remapped", os.Getenv("NODE_NAME"), strconv.Itoa(gpuid)).Set(rm)
			}
		}
	}
	return nil, &out
}

func runPCIeBw() (error, *[]byte) {
	out, err := exec.Command("python3", "./gpubw/entrypoint.py", "-t", utils.UserConfig.BWThreshold).CombinedOutput()
	if err != nil {
		klog.Info("Out:", string(out))
		klog.Error(err.Error())
		return err, nil
	} else {
		klog.Info("GPU PCIe BW test completed:")

		if strings.Contains(string(out[:]), "FAIL") {
			klog.Info("PCIe BW test failed.", string(out[:]))
		}

		if strings.Contains(string(out[:]), "ABORT") {
			klog.Info("PCIe BW cannot be run. ", string(out[:]))
			return nil, &out
		}

		output := strings.TrimSuffix(string(out[:]), "\n")
		split := strings.Split(output, "\n")

		bws := split[len(split)-1]
		final := strings.Split(bws, " ")

		for gpuid, v := range final {
			bw, err := strconv.ParseFloat(v, 64)
			if err != nil {
				klog.Error(err.Error())
				return err, nil
			} else {
				klog.Info("Observation: ", os.Getenv("NODE_NAME"), " ", strconv.Itoa(gpuid), " ", bw)
				utils.HchecksGauge.WithLabelValues("pciebw", os.Getenv("NODE_NAME"), strconv.Itoa(gpuid)).Set(bw)
			}
		}
	}
	return nil, &out
}

func runIperf(nodelist string, jobName string, plane string, clients string, servers string, sourceNode string) (error, *[]byte) {
	out, err := exec.Command("python3", "./network/iperf3-entrypoint.py", "--nodes", nodelist, "--job", jobName, "--plane", plane, "--clients", clients, "--servers", servers, "--source", sourceNode).CombinedOutput()
	klog.Info("Running command: ./network/iperf3-entrypoint.py ", " --nodes ", nodelist, " --job ", jobName, " --plane ", plane, " --clients ", clients, " --servers ", servers, " --source ", sourceNode)
	if err != nil {
		klog.Info(string(out))
		klog.Error(err.Error())
		return err, nil
	} else {
		klog.Info("iperf3 test completed:")
		klog.Info("iperf3 result:\n", string(out))
		if (clients == "1" && servers == "1") {
			output := strings.TrimSuffix(string(out[:]), "\n")
			line := strings.Split(output, "\n")
			for i := len(line) - 3; i > 0; i-- {
				if strings.Contains(line[i], "Aggregate") {
					break
				}
				entries := strings.Split(line[i], " ")
				bw, err := strconv.ParseFloat(entries[2], 64)
				if err != nil {
					klog.Error(err.Error())
					return err, nil
				} else {
					klog.Info("Observation: ", entries[0], " ", entries[1], " ", bw)
					utils.HchecksGauge.WithLabelValues("iperf", entries[0], entries[1]).Set(bw)
				}
			}
		}
		
	}
	return nil, &out
}

func startIperfServers(replicas string) (error, *[]byte) {
	out, err := exec.Command("python3", "./network/start-iperf-servers.py", "--replicas", replicas).CombinedOutput()
	klog.Info("Running command: ./network/start-iperf-servers.py --replicas ", replicas)
	if err != nil {
		klog.Info(string(out))
		klog.Error(err.Error())
		return err, nil
	} else {
		klog.Info("iperf3 servers started.")
	}
	return nil, &out
}

func runDCGM(dcgmR string) (error, *[]byte) {
	out, err := exec.Command("python3", "./gpu-dcgm/entrypoint.py", "-r", dcgmR).Output()
	if err != nil {
		klog.Error(err.Error())
		return err, nil
	} else {
		klog.Info("DCGM test completed:")

		if strings.Contains(string(out[:]), "ERR") {
			klog.Info("DCGM test exited with errors.", string(out[:]))
		}

		if strings.Contains(string(out[:]), "ABORT") {
			klog.Info("DCGM cannot be run. ", string(out[:]))
			return nil, &out
		}
		output := strings.TrimSuffix(string(out[:]), "\n")
		split := strings.Split(output, "\n")
		dcgmtests := split[len(split)-1]
		var res float64
		res = 0
		if strings.Contains(split[len(split)-1], "SUCCESS") {
			klog.Info("Observation: ", os.Getenv("NODE_NAME"), " ", "result", " ", res)
		} else if strings.Contains(split[len(split)-2], "FAIL") {
			res = 1
			for _, v := range strings.Split(dcgmtests, " ") {
				klog.Info("Observation: ", os.Getenv("NODE_NAME"), " ", "Fail ", v, " ", res)
			}
		}
		utils.HchecksGauge.WithLabelValues("dcgm", os.Getenv("NODE_NAME"), "").Set(res)
	}
	return nil, &out
}
