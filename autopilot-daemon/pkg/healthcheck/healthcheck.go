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
	"github.com/IBM/autopilot/pkg/healthcheck/pciebw"
	"github.com/IBM/autopilot/pkg/healthcheck/remapped"
	"github.com/IBM/autopilot/pkg/healthcheck/gpumem"
	"github.com/IBM/autopilot/pkg/healthcheck/dcgm"
	"github.com/IBM/autopilot/pkg/healthcheck/ping"
	"github.com/IBM/autopilot/pkg/healthcheck/gpupower"
	"github.com/IBM/autopilot/pkg/healthcheck/iperf"
	"github.com/IBM/autopilot/pkg/healthcheck/pvc"
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
	
	// Use the new Go implementation
	result, err := remapped.RunRemappedRowsCheck()
	out := []byte(result)
	
	if err != nil {
		klog.Info("Out:", string(out))
		klog.Error(err.Error())
		
		// Check if it's a specific failure we should handle
		if strings.Contains(string(out[:]), "FAIL") {
			klog.Info("Remapped Rows test failed.", string(out[:]))
			HealthCheckStatus[RowRemap] = true
		} else if !strings.Contains(string(out[:]), "ABORT") && !strings.Contains(string(out[:]), "SKIP") {
			// Actual execution failure
			return nil, err
		}
	} else {
		klog.Info("Remapped Rows check test completed:")
	}

	if strings.Contains(string(out[:]), "ABORT") {
		klog.Info("Remapped Rows cannot be run. ", string(out[:]))
		return &out, nil
	}

	if strings.Contains(string(out[:]), "SKIP") {
		klog.Info("Remapped Rows test skipped. ", string(out[:]))
		return &out, nil
	}

	// Parse output for metrics
	output := strings.TrimSuffix(string(out[:]), "\n")
	split := strings.Split(output, "\n")
	
	if len(split) >= 1 {
		// Check if the first line is "FAIL"
		if split[0] == "FAIL" && len(split) >= 2 {
			// Remapped rows line is the second line
			rmr := split[1]
			final := strings.Split(rmr, " ")
			
			for gpuid, v := range final {
				if v != "" {
					rm, err := strconv.ParseFloat(v, 64)
					if err != nil {
						klog.Error(err.Error())
						continue
					} else {
						klog.Info("Observation: ", utils.NodeName, " ", strconv.Itoa(gpuid), " ", rm)
						utils.HchecksGauge.WithLabelValues(string(RowRemap), utils.NodeName, utils.CPUModel, utils.GPUModel, strconv.Itoa(gpuid)).Set(rm)
					}
				}
			}
		} else if split[0] != "FAIL" {
			// Success case - remapped rows line is the first line
			rmr := split[0]
			final := strings.Split(rmr, " ")
			
			for gpuid, v := range final {
				if v != "" {
					rm, err := strconv.ParseFloat(v, 64)
					if err != nil {
						klog.Error(err.Error())
						continue
					} else {
						klog.Info("Observation: ", utils.NodeName, " ", strconv.Itoa(gpuid), " ", rm)
						utils.HchecksGauge.WithLabelValues(string(RowRemap), utils.NodeName, utils.CPUModel, utils.GPUModel, strconv.Itoa(gpuid)).Set(rm)
					}
				}
			}
		}
	}
	
	return &out, nil
}

func RunGPUMem() (*[]byte, error) {
	HealthCheckStatus[GPUMem] = false
	
	// Use the new Go implementation
	result, err := gpumem.RunGPUMemCheck()
	out := []byte(result)
	
	if err != nil {
		klog.Info("Out:", string(out))
		klog.Error(err.Error())
		
		// Check if it's a specific failure we should handle
		if strings.Contains(string(out[:]), "FAIL") {
			klog.Info("GPU Memory check failed.", string(out[:]))
			klog.Info("Observation: ", utils.NodeName, " 1")
			utils.HchecksGauge.WithLabelValues(string(GPUMem), utils.NodeName, utils.CPUModel, utils.GPUModel, "0").Set(1)
			HealthCheckStatus[GPUMem] = true
		} else if !strings.Contains(string(out[:]), "ABORT") && !strings.Contains(string(out[:]), "SKIP") {
			// Actual execution failure
			return nil, err
		}
	} else {
		klog.Info("GPU Memory check completed:")
		
		// Check if the result indicates success
		if strings.Contains(string(out[:]), "Health Check successful") {
			klog.Info("Observation: ", utils.NodeName, " 0")
			utils.HchecksGauge.WithLabelValues(string(GPUMem), utils.NodeName, utils.CPUModel, utils.GPUModel, "0").Set(0)
		} else if strings.Contains(string(out[:]), "Health Check unsuccessful") {
			klog.Info("GPU Memory check failed.", string(out[:]))
			klog.Info("Observation: ", utils.NodeName, " 1")
			utils.HchecksGauge.WithLabelValues(string(GPUMem), utils.NodeName, utils.CPUModel, utils.GPUModel, "0").Set(1)
			HealthCheckStatus[GPUMem] = true
		}
	}
	
	if strings.Contains(string(out[:]), "ABORT") {
		klog.Info("GPU Memory check cannot be run. ", string(out[:]))
		return &out, nil
	}

	if strings.Contains(string(out[:]), "SKIP") {
		klog.Info("GPU Memory check skipped. ", string(out[:]))
		return &out, nil
	}
	
	return &out, nil
}

func RunPCIeBW() (*[]byte, error) {
	HealthCheckStatus[PCIeBW] = false
	
	// Use the new Go implementation
	result, err := pciebw.RunPCIeBWCheck(utils.UserConfig.BWThreshold)
	out := []byte(result)
	
	if err != nil {
		klog.Info("Out:", string(out))
		klog.Error(err.Error())
		
		// Check if it's a threshold failure
		if strings.Contains(string(out[:]), "SUCCESS") {
			// Even though there's an error, if we have SUCCESS in output,
			// it's a threshold failure, not an execution failure
			klog.Info("PCIe BW test failed due to low bandwidth.", string(out[:]))
			HealthCheckStatus[PCIeBW] = true
		} else {
			// Actual execution failure
			return nil, err
		}
	} else {
		klog.Info("GPU PCIe BW test completed:")
	}

	if strings.Contains(string(out[:]), "ABORT") {
		klog.Info("PCIe BW cannot be run. ", string(out[:]))
		return &out, nil
	}

	if strings.Contains(string(out[:]), "SKIP") {
		klog.Info("PCIe BW test skipped. ", string(out[:]))
		return &out, nil
	}

	// Parse output for metrics
	output := strings.TrimSuffix(string(out[:]), "\n")
	split := strings.Split(output, "\n")

	if len(split) >= 3 {
		// Host line
		hostLine := split[len(split)-2]
		// Bandwidths line
		bws := split[len(split)-1]
		final := strings.Split(bws, " ")

		for gpuid, v := range final {
			if v != "" {
				bw, err := strconv.ParseFloat(v, 64)
				if err != nil {
					klog.Error(err.Error())
					continue
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
	}

	return &out, nil
}

func RunPing(nodelist string, jobName string, nodelabel string) (*[]byte, error) {
	HealthCheckStatus[Ping] = false
	
	// Parse nodelist into a slice
	nodes := strings.Split(strings.ReplaceAll(nodelist, " ", ""), ",")
	
	// Use the new Go implementation
	result, err := ping.RunPingCheck(nodes, jobName, nodelabel)
	out := []byte(result)
	
	if err != nil {
		klog.Info("Out:", string(out))
		klog.Error(err.Error())
		
		// Check if it's a specific failure we should handle
		if !strings.Contains(string(out[:]), "ABORT") {
			// Actual execution failure
			return nil, err
		}
	} else {
		klog.Info("Ping test completed:")
	}
	
	if strings.Contains(string(out[:]), "ABORT") {
		klog.Info("Ping cannot be run. ", string(out[:]))
		return &out, nil
	}

	// Parse output for metrics
	output := strings.TrimSuffix(string(out[:]), "\n")
	lines := strings.Split(output, "\n")
	unreach_nodes := make(map[string][]string)
	for _, line := range lines {
		if strings.HasPrefix(line, "Node") {
			entry := strings.Split(line, " ")
			if len(entry) >= 5 {
				nodeName := entry[1]
				if _, exists := unreach_nodes[nodeName]; !exists {
					if entry[len(entry)-1] == "1" {
						utils.HchecksGauge.WithLabelValues(string(Ping), utils.NodeName, utils.CPUModel, utils.GPUModel, nodeName).Set(float64(1))
						klog.Info("Observation: ", nodeName, " ", entry[2], " ", entry[3], " Unreachable")
						unreach_nodes[nodeName] = append(unreach_nodes[nodeName], entry[2])
					} else {
						utils.HchecksGauge.WithLabelValues(string(Ping), utils.NodeName, utils.CPUModel, utils.GPUModel, nodeName).Set(float64(0))
					}
				}
			}
		}
	}
	klog.Info("Unreachable nodes count: ", len(unreach_nodes))
	
	return &out, nil
}

func RunIperf(workload string, pclients string, startport string, cleanup string) (*[]byte, error) {
	// Use the new Go implementation
	result, err := iperf.RunIperfCheck(workload, pclients, startport)
	if err != nil {
		klog.Error("iperf3 test failed:", err)
		return nil, err
	}
	
	// Handle cleanup if requested
	if cleanup != "" {
		_, err := iperf.StopAllIperfServers()
		if err != nil {
			klog.Error("Failed to stop iperf servers:", err)
			return nil, err
		}
	}
	
	out := []byte(result)
	klog.Info("iperf3 test completed:\n", string(out))
	return &out, nil
}

func StartIperfServers(numservers string, startport string) (*[]byte, error) {
	// Use the new Go implementation
	result, err := iperf.StartIperfServers(numservers, startport)
	if err != nil {
		klog.Error("Failed to start iperf servers:", err)
		return nil, err
	}
	
	out := []byte(result)
	klog.Info("iperf3 servers started.")
	return &out, nil
}

func StopAllIperfServers() (*[]byte, error) {
	// Use the new Go implementation
	result, err := iperf.StopAllIperfServers()
	if err != nil {
		klog.Error("Failed to stop iperf servers:", err)
		return nil, err
	}
	
	out := []byte(result)
	klog.Info("iperf3 servers stopped.")
	return &out, nil
}

func StartIperfClients(dstip string, dstport string, numclients string) (*[]byte, error) {
	if dstip == "" || dstport == "" || numclients == "" {
		klog.Error("Must provide arguments \"dstip\", \"dstport\", and \"startport\".")
		return nil, nil
	}

	// Use the new Go implementation
	result, err := iperf.StartIperfClients(dstip, dstport, numclients)
	if err != nil {
		klog.Error("Failed to start iperf clients:", err)
		return nil, err
	}
	
	out := []byte(result)
	klog.Info("iperf3 clients started.")
	return &out, nil
}

func RunDCGM(dcgmR string) (*[]byte, error) {
	HealthCheckStatus[DCGM] = false
	
	// Use the new Go implementation
	result, err := dcgm.RunDCGMCheck(dcgmR)
	out := []byte(result)
	
	if err != nil {
		klog.Info("Out:", string(out))
		klog.Error(err.Error())
		
		// Check if it's a specific failure we should handle
		if !strings.Contains(string(out[:]), "ABORT") {
			// Actual execution failure
			return nil, err
		}
	} else {
		klog.Info("DCGM test completed:")
	}
	
	if strings.Contains(string(out[:]), "ABORT") {
		klog.Info("DCGM cannot be run. ", string(out[:]))
		return &out, nil
	}

	// Parse output for metrics
	output := strings.TrimSuffix(string(out[:]), "\n")
	split := strings.Split(output, "\n")
	var res float64
	res = 0
	if strings.Contains(output, "SUCCESS") {
		klog.Info("Observation: ", utils.NodeName, " Pass ", res)
	} else {
		res = 1
		klog.Info("Observation: ", utils.NodeName, " Fail ", res)
		HealthCheckStatus[DCGM] = true
	}
	utils.HchecksGauge.WithLabelValues(string(DCGM), utils.NodeName, utils.CPUModel, utils.GPUModel, "").Set(res)
	
	return &out, nil
}

func RunGPUPower() (*[]byte, error) {
	HealthCheckStatus[GPUPower] = false
	
	// Use the new Go implementation
	result, err := gpupower.RunGPUPowerCheck()
	out := []byte(result)
	
	if err != nil {
		klog.Info("Out:", string(out))
		klog.Error(err.Error())
		
		// Check if it's a specific failure we should handle
		if !strings.Contains(string(out[:]), "ABORT") {
			// Actual execution failure
			return nil, err
		}
	} else {
		klog.Info("Power Throttle check test completed:")
	}
	
	if strings.Contains(string(out[:]), "ABORT") {
		klog.Info("Power Throttle cannot be run. ", string(out[:]))
		return &out, nil
	}

	// Parse output for metrics
	output := strings.TrimSuffix(string(out[:]), "\n")
	split := strings.Split(output, "\n")
	
	// Get the last line which contains the power values
	if len(split) > 0 {
		pwrs := split[len(split)-1]
		final := strings.Split(pwrs, " ")
		
		for gpuid, v := range final {
			if v != "" {
				pw, err := strconv.ParseFloat(v, 64)
				if err != nil {
					klog.Error(err.Error())
					continue
				}
				klog.Info("Observation: ", utils.NodeName, " ", strconv.Itoa(gpuid), " ", pw)
				utils.HchecksGauge.WithLabelValues("power-slowdown", utils.NodeName, utils.CPUModel, utils.GPUModel, strconv.Itoa(gpuid)).Set(pw)
			}
		}
	}
	
	// Check if there was a failure
	if strings.Contains(output, "FAIL") {
		klog.Info("Power Throttle test failed.", output)
		HealthCheckStatus[GPUPower] = true
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
	
	// Use the new Go implementation
	result, err := pvc.RunPVCCheck()
	if err != nil {
		klog.Error(err.Error())
		b := []byte(result)
		return &b, err
	}
	b := []byte(result)
	return &b, nil
}
