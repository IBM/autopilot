package handlers

import (
	"net/http"
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
	runAllTestsLocal("all", "pciebw,remapped,dcgm,ping,gpupower", "1", "None", "None", nil)
}

func runAllTestsLocal(nodes string, checks string, dcgmR string, jobName string, nodelabel string, r *http.Request) (*[]byte, error) {
	out := []byte("")
	var tmp *[]byte
	var err error
	start := time.Now()
	for _, check := range strings.Split(checks, ",") {
		switch check {
		case "ping":
			klog.Info("Running health check: ", check)
			pingnodes := "all"
			if r != nil {
				pingnodes = r.URL.Query().Get("pingnodes")
				if pingnodes == "" {
					pingnodes = "all"
				}
			}
			tmp, err = runPing(pingnodes, jobName, nodelabel)
			if err != nil {
				klog.Error(err.Error())
				return tmp, err
			}
			out = append(out, *tmp...)

		case "dcgm":
			klog.Info("Running health check: ", check, " -r ", dcgmR)
			tmp, err = runDCGM(dcgmR)
			if err != nil {
				klog.Error(err.Error())
				return tmp, err
			}
			out = append(out, *tmp...)

		case "pciebw":
			klog.Info("Running health check: ", check)
			tmp, err = runPCIeBw()
			if err != nil {
				klog.Error(err.Error())
				return tmp, err
			}
			out = append(out, *tmp...)

		case "remapped":
			klog.Info("Running health check: ", check)
			tmp, err = runRemappedRows()
			if err != nil {
				klog.Error(err.Error())
				return nil, err
			}
			out = append(out, *tmp...)

		case "gpupower":
			klog.Info("Running health check: ", check)
			tmp, err = runGPUPower()
			if err != nil {
				klog.Error(err.Error())
				return nil, err
			}
			out = append(out, *tmp...)

		case "gpumem":
			klog.Info("Running health check: ", check)
			tmp, err = runGPUMem()
			if err != nil {
				klog.Error(err.Error())
				return nil, err
			}
			out = append(out, *tmp...)

		case "all":
			klog.Info("Run all health checks\n")
			tmp, err := runPCIeBw()
			if err != nil {
				klog.Error(err.Error())
				return tmp, err
			}
			out = append(out, *tmp...)
			tmp, err = runRemappedRows()
			if err != nil {
				klog.Error(err.Error())
				return nil, err
			}
			out = append(out, *tmp...)
			tmp, err = runDCGM(dcgmR)
			if err != nil {
				klog.Error(err.Error())
				return tmp, err
			}
			out = append(out, *tmp...)
			pingnodes := "all"
			if r != nil {
				pingnodes = r.URL.Query().Get("pingnodes")
				if pingnodes == "" {
					pingnodes = "all"
				}
			}
			tmp, err = runPing(pingnodes, jobName, nodelabel)
			if err != nil {
				klog.Error(err.Error())
				return tmp, err
			}
			out = append(out, *tmp...)
			tmp, err = runGPUPower()
			if err != nil {
				klog.Error(err.Error())
				return tmp, err
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
	return &out, nil
}

func runAllTestsRemote(host string, check string, batch string, jobName string, dcgmR string, nodelabel string) (*[]byte, error) {
	klog.Info("About to run command:\n", "./utils/runHealthchecks.py", " --nodes="+host, " --check="+check, " --batchSize="+batch, " --wkload="+jobName, " --dcgmR="+dcgmR, " --nodelabel="+nodelabel)

	out, err := exec.Command("python3", "./utils/runHealthchecks.py", "--service=autopilot-healthchecks", "--namespace="+os.Getenv("NAMESPACE"), "--nodes="+host, "--check="+check, "--batchSize="+batch, "--wkload="+jobName, "--dcgmR="+dcgmR, "--nodelabel="+nodelabel).Output()
	if err != nil {
		klog.Info(string(out))
		klog.Error(err.Error())
		return nil, err
	}

	return &out, nil
}

func runRemappedRows() (*[]byte, error) {
	out, err := exec.Command("python3", "./gpu-remapped/entrypoint.py").CombinedOutput()
	if err != nil {
		klog.Info("Out:", string(out))
		klog.Error(err.Error())
		return nil, err
	} else {
		klog.Info("Remapped Rows check test completed:")

		if strings.Contains(string(out[:]), "FAIL") {
			klog.Info("Remapped Rows test failed.", string(out[:]))
		}

		if strings.Contains(string(out[:]), "ABORT") {
			klog.Info("Remapped Rows cannot be run. ", string(out[:]))
			return &out, nil
		}

		output := strings.TrimSuffix(string(out[:]), "\n")
		split := strings.Split(output, "\n")
		rmr := split[len(split)-1]
		final := strings.Split(rmr, " ")

		for gpuid, v := range final {
			rm, err := strconv.ParseFloat(v, 64)
			if err != nil {
				klog.Error(err.Error())
				return nil, err
			} else {
				klog.Info("Observation: ", os.Getenv("NODE_NAME"), " ", strconv.Itoa(gpuid), " ", rm)
				// utils.Hchecks.WithLabelValues("remapped", os.Getenv("NODE_NAME"), strconv.Itoa(gpuid)).Observe(rm)
				utils.HchecksGauge.WithLabelValues("remapped", os.Getenv("NODE_NAME"), strconv.Itoa(gpuid)).Set(rm)
			}
		}
	}
	return &out, nil
}

func runGPUMem() (*[]byte, error) {
	out, err := exec.Command("python3", "./gpu-mem/entrypoint.py").CombinedOutput()
	if err != nil {
		klog.Info("Out:", string(out))
		klog.Error(err.Error())
		return nil, err
	} else {
		klog.Info("GPU Memory check completed:")

		if strings.Contains(string(out[:]), "FAIL") {
			klog.Info("GPU Memory check failed.", string(out[:]))
			klog.Info("Observation: ", os.Getenv("NODE_NAME"), " ", "1")
			utils.HchecksGauge.WithLabelValues("gpumem", os.Getenv("NODE_NAME"), "0").Set(1)
		}

		if strings.Contains(string(out[:]), "ABORT") {
			klog.Info("GPU Memory check cannot be run. ", string(out[:]))
			return &out, nil
		}

		klog.Info("Observation: ", os.Getenv("NODE_NAME"), " ", "0")
		utils.HchecksGauge.WithLabelValues("gpumem", os.Getenv("NODE_NAME"), "0").Set(0)
	}
	return &out, nil
}

func runPCIeBw() (*[]byte, error) {
	out, err := exec.Command("python3", "./gpu-bw/entrypoint.py", "-t", utils.UserConfig.BWThreshold).CombinedOutput()
	if err != nil {
		klog.Info("Out:", string(out))
		klog.Error(err.Error())
		return nil, err
	} else {
		klog.Info("GPU PCIe BW test completed:")

		if strings.Contains(string(out[:]), "FAIL") {
			klog.Info("PCIe BW test failed.", string(out[:]))
		}

		if strings.Contains(string(out[:]), "ABORT") {
			klog.Info("PCIe BW cannot be run. ", string(out[:]))
			return &out, nil
		}

		output := strings.TrimSuffix(string(out[:]), "\n")
		split := strings.Split(output, "\n")

		bws := split[len(split)-1]
		final := strings.Split(bws, " ")

		for gpuid, v := range final {
			bw, err := strconv.ParseFloat(v, 64)
			if err != nil {
				klog.Error(err.Error())
				return nil, err
			} else {
				klog.Info("Observation: ", os.Getenv("NODE_NAME"), " ", strconv.Itoa(gpuid), " ", bw)
				utils.HchecksGauge.WithLabelValues("pciebw", os.Getenv("NODE_NAME"), strconv.Itoa(gpuid)).Set(bw)
			}
		}
	}
	return &out, nil
}

func runPing(nodelist string, jobName string, nodelabel string) (*[]byte, error) {
	out, err := exec.Command("python3", "./network/ping-entrypoint.py", "--nodes", nodelist, "--job", jobName, "--nodelabel", nodelabel).CombinedOutput()
	// klog.Info("Running command: ./network/ping-entrypoint.py ", " --nodes ", nodelist, " --job ", jobName)
	if err != nil {
		klog.Info(string(out))
		klog.Error(err.Error())
		return nil, err
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
	return &out, nil
}

func runIperf(nodelist string, jobName string, plane string, clients string, servers string, sourceNode string, nodelabel string) (*[]byte, error) {
	out, err := exec.Command("python3", "./network/iperf3-entrypoint.py", "--nodes", nodelist, "--job", jobName, "--plane", plane, "--clients", clients, "--servers", servers, "--source", sourceNode, "--nodelabel", nodelabel).CombinedOutput()
	klog.Info("Running command: ./network/iperf3-entrypoint.py ", " --nodes ", nodelist, " --job ", jobName, " --plane ", plane, " --clients ", clients, " --servers ", servers, " --source ", sourceNode, " --nodelabel", nodelabel)
	if err != nil {
		klog.Info(string(out))
		klog.Error(err.Error())
		return nil, err
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
					return nil, err
				}
				klog.Info("Observation: ", entries[0], " ", entries[1], " ", bw)
				// utils.HchecksGauge.WithLabelValues("iperf", entries[0], entries[1]).Set(bw)
			}
		}
	}
	return &out, nil
}

func startIperfServers(replicas string) (*[]byte, error) {
	out, err := exec.Command("python3", "./network/start-iperf-servers.py", "--replicas", replicas).CombinedOutput()
	klog.Info("Running command: ./network/start-iperf-servers.py --replicas ", replicas)
	if err != nil {
		klog.Info(string(out))
		klog.Error(err.Error())
		return nil, err
	} else {
		klog.Info("iperf3 servers started.")
	}
	return &out, nil
}

func runDCGM(dcgmR string) (*[]byte, error) {
	out, err := exec.Command("python3", "./gpu-dcgm/entrypoint.py", "-r", dcgmR).Output()
	if err != nil {
		klog.Error(err.Error())
		return nil, err
	} else {
		klog.Info("DCGM test completed:")

		if strings.Contains(string(out[:]), "ERR") {
			klog.Info("DCGM test exited with errors.", string(out[:]))
		}

		if strings.Contains(string(out[:]), "ABORT") {
			klog.Info("DCGM cannot be run. ", string(out[:]))
			return &out, nil
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
	return &out, nil
}

func runGPUPower() (*[]byte, error) {
	out, err := exec.Command("bash", "./gpu-power/power-throttle.sh").Output()
	if err != nil {
		klog.Error(err.Error())
		return nil, err
	} else {
		klog.Info("Power Throttle check test completed:")

		if strings.Contains(string(out[:]), "FAIL") {
			klog.Info("Power Throttle test failed.", string(out[:]))
		}

		if strings.Contains(string(out[:]), "ABORT") {
			klog.Info("Power Throttle cannot be run. ", string(out[:]))
			return &out, nil
		}

		output := strings.TrimSuffix(string(out[:]), "\n")
		split := strings.Split(output, "\n")
		pwrs := split[len(split)-1]
		final := strings.Split(pwrs, " ")

		for gpuid, v := range final {
			pw, err := strconv.ParseFloat(v, 64)
			if err != nil {
				klog.Error(err.Error())
				return nil, err
			} else {
				klog.Info("Observation: ", os.Getenv("NODE_NAME"), " ", strconv.Itoa(gpuid), " ", pw)
				utils.HchecksGauge.WithLabelValues("power-slowdown", os.Getenv("NODE_NAME"), strconv.Itoa(gpuid)).Set(pw)
			}
		}
	}
	return &out, nil
}
