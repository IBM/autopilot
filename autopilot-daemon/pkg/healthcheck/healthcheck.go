package healthcheck

import (
	"errors"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/IBM/autopilot/pkg/utils"
	"k8s.io/klog/v2"
)

func PeriodicCheck() {
	klog.Info("Running a periodic check")
	utils.HealthcheckLock.Lock()
	defer utils.HealthcheckLock.Unlock()
	checks := GetPeriodicChecks()
	RunHealthLocalNode(checks, "1", "None", "None", nil)
	hasFailures := GetNodeStatus()
	klog.Info("Errors after running periodic health checks: ", hasFailures)
	if hasFailures {
		utils.PatchNode(utils.GPUHealthWarnLabel, utils.NodeName, true)
	} else {
		utils.PatchNode(utils.GPUHealthPassLabel, utils.NodeName, true)
	}
}

func InvasiveCheck() {
	klog.Info("Trying to run an invasive check")
	utils.HealthcheckLock.Lock()
	defer utils.HealthcheckLock.Unlock()
	if utils.GPUsAvailability() {
		klog.Info("Starting invasive health checks, updating node label =TESTING for node ", utils.NodeName)
		utils.PatchNode(utils.GPUHealthTestingLabel, utils.NodeName, true)
		err := utils.CreateJob("dcgm")
		if err != nil {
			klog.Info("Invasive health checks Job creation failed, reset node label for node ", utils.NodeName)
			utils.PatchNode(utils.GPUHealthEmptyLabel, utils.NodeName, true)
		}
	}
}

func RunHealthLocalNode(checks string, dcgmR string, jobName string, nodelabel string, r *http.Request) (*[]byte, error) {
	out := []byte("")
	var tmp *[]byte
	var err error
	start := time.Now()
	if strings.Contains(checks, "all") {
		checks = GetPeriodicChecks()
	}
	klog.Info("Health checks ", checks)
	for _, check := range strings.Split(checks, ",") {
		switch check {
		case string(Ping):
			klog.Info("Running health check: ", check)
			pingnodes := "all"
			if r != nil {
				pingnodes = r.URL.Query().Get("pingnodes")
				if pingnodes == "" {
					pingnodes = "all"
				}
			}
			tmp, err = RunPing(pingnodes, jobName, nodelabel)
			if err != nil {
				klog.Error(err.Error())
				return tmp, err
			}
			out = append(out, *tmp...)

		case string(DCGM):
			klog.Info("Running health check: ", check, " -r ", dcgmR)
			tmp, err = RunDCGM(dcgmR)
			if err != nil {
				klog.Error(err.Error())
				return tmp, err
			}
			out = append(out, *tmp...)

		case string(PCIeBW):
			klog.Info("Running health check: ", check)
			tmp, err = RunPCIeBW()
			if err != nil {
				klog.Error(err.Error())
				return tmp, err
			}
			out = append(out, *tmp...)

		case string(RowRemap):
			klog.Info("Running health check: ", check)
			tmp, err = RunRemappedRows()
			if err != nil {
				klog.Error(err.Error())
				return tmp, err
			}
			out = append(out, *tmp...)

		case string(GPUPower):
			klog.Info("Running health check: ", check)
			tmp, err = RunGPUPower()
			if err != nil {
				klog.Error(err.Error())
				return tmp, err
			}
			out = append(out, *tmp...)

		case string(GPUMem):
			klog.Info("Running health check: ", check)
			tmp, err = RunGPUMem()
			if err != nil {
				klog.Error(err.Error())
				return tmp, err
			}
			out = append(out, *tmp...)

		case string(PVC):
			klog.Info("Running health check: ", check)
			tmp, err = RunCreateDeletePVC()
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

func RunHealthRemoteNodes(host string, check string, batch string, jobName string, dcgmR string, nodelabel string) (*[]byte, error) {
	klog.Info("About to run command:\n", "./utils/runHealthchecks.py", " --nodes="+host, " --check="+check, " --batchSize="+batch, " --wkload="+jobName, " --dcgmR="+dcgmR, " --nodelabel="+nodelabel)

	out, err := exec.Command("python3", "./utils/runHealthchecks.py", "--service=autopilot-healthchecks", "--namespace="+utils.Namespace, "--nodes="+host, "--check="+check, "--batchSize="+batch, "--wkload="+jobName, "--dcgmR="+dcgmR, "--nodelabel="+nodelabel).Output()
	if err != nil {
		klog.Info(string(out))
		klog.Error(err.Error())
		return nil, err
	}

	return &out, nil
}

func RunRemappedRows() (*[]byte, error) {
	HealthCheckStatus[RowRemap] = false
	out, err := exec.Command("python3", "./gpu-remapped/entrypoint.py").CombinedOutput()
	if err != nil {
		klog.Info("Out:", string(out))
		klog.Error(err.Error())
		return nil, err
	} else {
		klog.Info("Remapped Rows check test completed:")

		if strings.Contains(string(out[:]), "FAIL") {
			klog.Info("Remapped Rows test failed.", string(out[:]))
			HealthCheckStatus[RowRemap] = true
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
				klog.Info("Observation: ", utils.NodeName, " ", strconv.Itoa(gpuid), " ", rm)
				utils.HchecksGauge.WithLabelValues(string(RowRemap), utils.NodeName, utils.CPUModel, utils.GPUModel, strconv.Itoa(gpuid)).Set(rm)
			}
		}
	}
	return &out, nil
}

func RunGPUMem() (*[]byte, error) {
	HealthCheckStatus[GPUMem] = false
	out, err := exec.Command("python3", "./gpu-mem/entrypoint.py").CombinedOutput()
	if err != nil {
		klog.Info("Out:", string(out))
		klog.Error(err.Error())
		return nil, err
	} else {
		klog.Info("GPU Memory check completed:")

		if strings.Contains(string(out[:]), "FAIL") {
			klog.Info("GPU Memory check failed.", string(out[:]))
			klog.Info("Observation: ", utils.NodeName, " 1")
			utils.HchecksGauge.WithLabelValues(string(GPUMem), utils.NodeName, utils.CPUModel, utils.GPUModel, "0").Set(1)
			HealthCheckStatus[GPUMem] = true
		}

		if strings.Contains(string(out[:]), "ABORT") {
			klog.Info("GPU Memory check cannot be run. ", string(out[:]))
			return &out, nil
		}

		klog.Info("Observation: ", utils.NodeName, " 0")
		utils.HchecksGauge.WithLabelValues(string(GPUMem), utils.NodeName, utils.CPUModel, utils.GPUModel, "0").Set(0)
	}
	return &out, nil
}

func RunPCIeBW() (*[]byte, error) {
	HealthCheckStatus[PCIeBW] = false
	out, err := exec.Command("python3", "./gpu-bw/entrypoint.py", "-t", strconv.Itoa(utils.UserConfig.BWThreshold)).CombinedOutput()
	if err != nil {
		klog.Info("Out:", string(out))
		klog.Error(err.Error())
		return nil, err
	} else {
		klog.Info("GPU PCIe BW test completed:")

		if strings.Contains(string(out[:]), "FAIL") {
			klog.Info("PCIe BW test failed.", string(out[:]))
			HealthCheckStatus[PCIeBW] = true
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
				logline := "Observation: " + utils.NodeName + " " + strconv.Itoa(gpuid) + " " + v
				if bw < float64(utils.UserConfig.BWThreshold) {
					logline += "  [[ LOW PCIE -- Below expected threshold of " + strconv.Itoa(utils.UserConfig.BWThreshold) + " Gb/s ]]"
					HealthCheckStatus[PCIeBW] = true
				}
				klog.Info(logline)
				utils.HchecksGauge.WithLabelValues(string(PCIeBW), utils.NodeName, utils.CPUModel, utils.GPUModel, strconv.Itoa(gpuid)).Set(bw)
			}
		}
	}
	return &out, nil
}

func RunPing(nodelist string, jobName string, nodelabel string) (*[]byte, error) {
	HealthCheckStatus[Ping] = false
	out, err := exec.Command("python3", "./network/ping-entrypoint.py", "--nodes", nodelist, "--job", jobName, "--nodelabel", nodelabel).CombinedOutput()
	if err != nil {
		klog.Info(string(out))
		klog.Error(err.Error())
		return nil, err
	} else {
		klog.Info("Ping test completed:")

		if strings.Contains(string(out[:]), "FAIL") {
			klog.Info("Ping test failed.", string(out[:]))
			HealthCheckStatus[PCIeBW] = true
		}

		if strings.Contains(string(out[:]), "ABORT") {
			klog.Info("Ping cannot be run. ", string(out[:]))
			return &out, nil
		}

		output := strings.TrimSuffix(string(out[:]), "\n")
		lines := strings.Split(output, "\n")
		unreach_nodes := make(map[string][]string)
		for _, line := range lines {
			if strings.HasPrefix(line, "Node") {
				entry := strings.Split(line, " ")
				if _, exists := unreach_nodes[entry[1]]; !exists {
					if entry[len(entry)-1] == "1" {
						utils.HchecksGauge.WithLabelValues(string(Ping), utils.NodeName, utils.CPUModel, utils.GPUModel, entry[1]).Set(float64(1))
						klog.Info("Observation: ", entry[1], " ", entry[2], " ", entry[3], " Unreachable")
						unreach_nodes[entry[1]] = append(unreach_nodes[entry[1]], entry[2])
					} else {
						utils.HchecksGauge.WithLabelValues(string(Ping), utils.NodeName, utils.CPUModel, utils.GPUModel, entry[1]).Set(float64(0))
					}
				}
			}
		}
		klog.Info("Unreachable nodes count: ", len(unreach_nodes))
	}
	return &out, nil
}

func RunIperf(workload string, pclients string, startport string, cleanup string) (*[]byte, error) {

	args := []string{"./network/iperf3_entrypoint.py", "--workload", workload, "--pclients", pclients, "--startport", startport}

	if cleanup != "" {
		args = append(args, cleanup)
	}
	out, err := exec.Command("python3", args...).CombinedOutput()
	if err != nil {
		return nil, err
	}
	klog.Info("iperf3 test completed:\n", string(out))
	return &out, nil
}

func StartIperfServers(numservers string, startport string) (*[]byte, error) {
	out, err := exec.Command("python3", "./network/iperf3_start_servers.py", "--numservers", numservers, "--startport", startport).CombinedOutput()
	if err != nil {
		klog.Info(string(out))
		klog.Error(err.Error())
		return nil, err
	} else {
		klog.Info("iperf3 servers started.")
	}
	return &out, nil
}

func StopAllIperfServers() (*[]byte, error) {
	out, err := exec.Command("python3", "./network/iperf3_stop_servers.py").CombinedOutput()
	if err != nil {
		klog.Info(string(out))
		klog.Error(err.Error())
		return nil, err
	} else {
		klog.Info("iperf3 servers stopped.")
	}
	return &out, nil
}

func StartIperfClients(dstip string, dstport string, numclients string) (*[]byte, error) {
	if dstip == "" || dstport == "" || numclients == "" {
		klog.Error("Must provide arguments \"dstip\", \"dstport\", and \"startport\".")
		return nil, nil
	}

	out, err := exec.Command("python3", "./network/iperf3_start_clients.py", "--dstip", dstip, "--dstport", dstport, "--numclients", numclients).CombinedOutput()
	if err != nil {
		klog.Info(string(out))
		klog.Error(err.Error())
		return nil, err
	} else {
		klog.Info("iperf3 clients started.")
	}
	return &out, nil
}

func RunDCGM(dcgmR string) (*[]byte, error) {
	HealthCheckStatus[DCGM] = false
	out, err := exec.Command("python3", "./gpu-dcgm/entrypoint.py", "-r", dcgmR, "-l").Output()
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
		var res float64
		res = 0
		if strings.Contains(split[len(split)-1], "SUCCESS") {
			klog.Info("Observation: ", utils.NodeName, " Pass ", res)
		} else {
			res = 1
			klog.Info("Observation: ", utils.NodeName, " Fail ", res)
			HealthCheckStatus[DCGM] = true
		}
		utils.HchecksGauge.WithLabelValues(string(DCGM), utils.NodeName, utils.CPUModel, utils.GPUModel, "").Set(res)
	}
	return &out, nil
}

func RunGPUPower() (*[]byte, error) {
	HealthCheckStatus[GPUPower] = false
	out, err := exec.Command("bash", "./gpu-power/power-throttle.sh").Output()
	if err != nil {
		klog.Error(err.Error())
		return nil, err
	}
	klog.Info("Power Throttle check test completed:")

	if strings.Contains(string(out[:]), "FAIL") {
		klog.Info("Power Throttle test failed.", string(out[:]))
		HealthCheckStatus[GPUPower] = true
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
		}
		klog.Info("Observation: ", utils.NodeName, " ", strconv.Itoa(gpuid), " ", pw)
		utils.HchecksGauge.WithLabelValues("power-slowdown", utils.NodeName, utils.CPUModel, utils.GPUModel, strconv.Itoa(gpuid)).Set(pw)

	}
	return &out, nil
}

func RunCreateDeletePVC() (*[]byte, error) {
	_, exists := os.LookupEnv("PVC_TEST_STORAGE_CLASS")
	if !exists {
		b := []byte("Storage class not set. Cannot run. ABORT")
		return &b, errors.New("storage class not set")
	}
	HealthCheckStatus[PVC] = false
	err := createPVC()
	if err != nil {
		klog.Error(err.Error())
		b := []byte("Create PVC Failed. ABORT")
		return &b, err
	}
	// Wait a few seconds before start checking
	waitonpvc := time.NewTicker(30 * time.Second)
	defer waitonpvc.Stop()
	<-waitonpvc.C
	out, err := ListPVC()
	if err != nil {
		klog.Error(err.Error())
	}
	b := []byte(out)
	return &b, nil
}
