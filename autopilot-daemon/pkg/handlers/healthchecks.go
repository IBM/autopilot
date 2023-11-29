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
	runAllTestsLocal("pciebw,remapped,dcgm,ping,gpupower", "1")
}

func runAllTestsLocal(checks string, dcgmR string) (error, *[]byte) {
	out := []byte("")
	var tmp *[]byte
	var err error
	start := time.Now()
	for _, check := range strings.Split(checks, ",") {
		switch check {
		case "ping":
			klog.Info("Running health check: ", check)
			err, tmp = runPing("all", "None")
			if err != nil {
				klog.Error(err.Error())
				return err, tmp
			}
			out = append(out, *tmp...)

		case "dcgm":
			klog.Info("Running health check: ", check, " -r ", dcgmR)
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

		case "gpupower":
			klog.Info("Running health check: ", check)
			err, tmp = runGPUPower()
			if err != nil {
				klog.Error(err.Error())
				return err, nil
			}
			out = append(out, *tmp...)

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
			err, tmp = runPing("all", "None")
			if err != nil {
				klog.Error(err.Error())
				return err, tmp
			}
			out = append(out, *tmp...)
			err, tmp = runGPUPower()
			if err != nil {
				klog.Error(err.Error())
				return err, tmp
			}
			out = append(out, *tmp...)
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

func runAllTestsRemote(host string, check string, batch string, jobName string, dcgmR string, nodelabel string) (error, *[]byte) {
	klog.Info("About to run command:\n", "./utils/runHealthchecks.py", " --nodes="+host, " --check="+check, " --batchSize="+batch, " --wkload="+jobName, " --dcgmR="+dcgmR, " --nodelabel="+nodelabel)

	out, err := exec.Command("python3", "./utils/runHealthchecks.py", "--service=autopilot-healthchecks", "--namespace="+os.Getenv("NAMESPACE"), "--nodes="+host, "--check="+check, "--batchSize="+batch, "--wkload="+jobName, "--dcgmR="+dcgmR, "--nodelabel="+nodelabel).Output()
	if err != nil {
		klog.Info(string(out))
		klog.Error(err.Error())
		return err, nil
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
	out, err := exec.Command("python3", "./gpu-bw/entrypoint.py", "-t", utils.UserConfig.BWThreshold).CombinedOutput()
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

func runPing(nodelist string, jobName string) (error, *[]byte) {
	out, err := exec.Command("python3", "./network/ping-entrypoint.py", "--nodes", nodelist, "--job", jobName).CombinedOutput()
	// klog.Info("Running command: ./network/ping-entrypoint.py ", " --nodes ", nodelist, " --job ", jobName)
	if err != nil {
		klog.Info(string(out))
		klog.Error(err.Error())
		return err, nil
	} else {
		output := strings.TrimSuffix(string(out[:]), "\n")
		lines := strings.Split(output, "\n")
		unreach_nodes := make(map[string][]string)
		reach_nodes := make(map[string][]string)
		for _, line := range lines {
			if strings.HasPrefix(line, "Node") {
				entry := strings.Split(line, " ")
				if entry[len(entry)-1] == "1" {
					utils.HchecksGauge.WithLabelValues("ping", entry[1], entry[2]).Set(1)
					klog.Info("Observation: ", entry[1], " ", entry[2], " ", entry[3], " Unreachable")
					unreach_nodes[entry[1]] = append(unreach_nodes[entry[1]], entry[2])
				} else {
					utils.HchecksGauge.WithLabelValues("ping", entry[1], entry[2]).Set(0)
					reach_nodes[entry[1]] = append(reach_nodes[entry[1]], entry[2])
				}
			}
		}
		klog.Info("Observation: ", len(reach_nodes)-len(unreach_nodes), "/", len(reach_nodes)+len(unreach_nodes), " remote nodes are reachable")
	}
	return nil, &out
}

func runIperf(nodelist string, jobName string, plane string, clients string, servers string, sourceNode string, nodelabel string) (error, *[]byte) {
	out, err := exec.Command("python3", "./network/iperf3-entrypoint.py", "--nodes", nodelist, "--job", jobName, "--plane", plane, "--clients", clients, "--servers", servers, "--source", sourceNode, "--nodelabel", nodelabel).CombinedOutput()
	klog.Info("Running command: ./network/iperf3-entrypoint.py ", " --nodes ", nodelist, " --job ", jobName, " --plane ", plane, " --clients ", clients, " --servers ", servers, " --source ", sourceNode, " --nodelabel", nodelabel)
	if err != nil {
		klog.Info(string(out))
		klog.Error(err.Error())
		return err, nil
	} else {
		klog.Info("iperf3 test completed:")
		klog.Info("iperf3 result:\n", string(out))
		if clients == "1" && servers == "1" {
			output := strings.TrimSuffix(string(out[:]), "\n")
			line := strings.Split(output, "\n")
			var bw float64
			for i := len(line) - 3; i > 0; i-- {
				if strings.Contains(line[i], "Aggregate") {
					break
				}
				entries := strings.Split(line[i], " ")
				if len(entries) == 2 {
					bw = 0
				} else {
					bw, err = strconv.ParseFloat(entries[2], 64)
				}
				if err != nil {
					klog.Error(err.Error())
					return err, nil
				}
				klog.Info("Observation: ", entries[0], " ", entries[1], " ", bw)
				// utils.HchecksGauge.WithLabelValues("iperf", entries[0], entries[1]).Set(bw)
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

func runGPUPower() (error, *[]byte) {
	out, err := exec.Command("bash", "./gpu-power/power-throttle.sh").Output()
	if err != nil {
		klog.Error(err.Error())
		return err, nil
	} else {
		klog.Info("Power Throttle check test completed:")

		if strings.Contains(string(out[:]), "FAIL") {
			klog.Info("Power Throttle test failed.", string(out[:]))
		}

		if strings.Contains(string(out[:]), "ABORT") {
			klog.Info("Power Throttle cannot be run. ", string(out[:]))
			return nil, &out
		}

		output := strings.TrimSuffix(string(out[:]), "\n")
		split := strings.Split(output, "\n")
		pwrs := split[len(split)-1]
		final := strings.Split(pwrs, " ")

		for gpuid, v := range final {
			pw, err := strconv.ParseFloat(v, 64)
			if err != nil {
				klog.Error(err.Error())
				return err, nil
			} else {
				klog.Info("Observation: ", os.Getenv("NODE_NAME"), " ", strconv.Itoa(gpuid), " ", pw)
				utils.HchecksGauge.WithLabelValues("power-slowdown", os.Getenv("NODE_NAME"), strconv.Itoa(gpuid)).Set(pw)
			}
		}
	}
	return nil, &out
}
