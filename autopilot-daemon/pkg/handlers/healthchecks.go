package handlers

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
	checks := utils.GetPeriodicChecks()
	runAllTestsLocal("all", checks, "1", "None", "None", nil)
}

func InvasiveCheck() {
	klog.Info("Trying to run an invasive check")
	utils.HealthcheckLock.Lock()
	defer utils.HealthcheckLock.Unlock()
	if utils.GPUsAvailability() {
		klog.Info("Begining invasive health checks, updating node label =TESTING for node ", os.Getenv("NODE_NAME"))
		label := `
		{
			"metadata": {
				"labels": {
					"autopilot.ibm.com/gpuhealth": "TESTING"
					}
			}
		}
		`
		utils.PatchNode(label, os.Getenv("NODE_NAME"))
		err := utils.CreateJob("dcgm")
		if err != nil {
			klog.Info("Invasive health checks failed, updating node label for node ", os.Getenv("NODE_NAME"))
			label := `
		{
			"metadata": {
				"labels": {
					"autopilot.ibm.com/gpuhealth": ""
					}
			}
		}
		`
			utils.PatchNode(label, os.Getenv("NODE_NAME"))
		}
	}
}

func runAllTestsLocal(nodes string, checks string, dcgmR string, jobName string, nodelabel string, r *http.Request) (*[]byte, error) {
	out := []byte("")
	var tmp *[]byte
	var err error
	gpufailures := false
	start := time.Now()
	if strings.Contains(checks, "all") {
		checks = utils.GetPeriodicChecks()
	}
	klog.Info("Health checks ", checks)
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
			if strings.Contains(string(*tmp), "FAIL") {
				klog.Info("GPU failure reported, updating node label for node ", os.Getenv("NODE_NAME"))
				gpufailures = true
			}
			out = append(out, *tmp...)

		case "pciebw":
			klog.Info("Running health check: ", check)
			tmp, err = runPCIeBw()
			if err != nil {
				klog.Error(err.Error())
				return tmp, err
			}
			if strings.Contains(string(*tmp), "FAIL") {
				klog.Info("GPU failure reported, updating node label for node ", os.Getenv("NODE_NAME"))
				gpufailures = true
			}
			out = append(out, *tmp...)

		case "remapped":
			klog.Info("Running health check: ", check)
			tmp, err = runRemappedRows()
			if err != nil {
				klog.Error(err.Error())
				return tmp, err
			}
			if strings.Contains(string(*tmp), "FAIL") {
				klog.Info("GPU failure reported, updating node label for node ", os.Getenv("NODE_NAME"))
				gpufailures = true
			}
			out = append(out, *tmp...)

		case "gpupower":
			klog.Info("Running health check: ", check)
			tmp, err = runGPUPower()
			if err != nil {
				klog.Error(err.Error())
				return tmp, err
			}
			if strings.Contains(string(*tmp), "FAIL") {
				klog.Info("GPU failure reported, updating node label for node ", os.Getenv("NODE_NAME"))
				gpufailures = true
			}
			out = append(out, *tmp...)

		case "gpumem":
			klog.Info("Running health check: ", check)
			tmp, err = runGPUMem()
			if err != nil {
				klog.Error(err.Error())
				return tmp, err
			}
			out = append(out, *tmp...)

		case "pvc":
			klog.Info("Running health check: ", check)
			tmp, err = runCreateDeletePVC()
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
	label := ""
	if gpufailures {
		label = `
		{
			"metadata": {
				"labels": {
					"autopilot.ibm.com/gpuhealth": "ERR"
					}
			}
		}
		`
	} else {
		// In case a previous run failed but the current one succeeds, we can reset the label
		label = `
		{
			"metadata": {
				"labels": {
					"autopilot.ibm.com/gpuhealth": "PASS"
					}
			}
		}
		`
		utils.PatchNode(label, os.Getenv("NODE_NAME"))
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
				utils.HchecksGauge.WithLabelValues("remapped", os.Getenv("NODE_NAME"), utils.CPUModel, utils.GPUModel, strconv.Itoa(gpuid)).Set(rm)
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
			klog.Info("Observation: ", os.Getenv("NODE_NAME"), " 1")
			utils.HchecksGauge.WithLabelValues("gpumem", os.Getenv("NODE_NAME"), utils.CPUModel, utils.GPUModel, "0").Set(1)
		}

		if strings.Contains(string(out[:]), "ABORT") {
			klog.Info("GPU Memory check cannot be run. ", string(out[:]))
			return &out, nil
		}

		klog.Info("Observation: ", os.Getenv("NODE_NAME"), " 0")
		utils.HchecksGauge.WithLabelValues("gpumem", os.Getenv("NODE_NAME"), utils.CPUModel, utils.GPUModel, "0").Set(0)
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
				utils.HchecksGauge.WithLabelValues("pciebw", os.Getenv("NODE_NAME"), utils.CPUModel, utils.GPUModel, strconv.Itoa(gpuid)).Set(bw)
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
		klog.Info("Ping test completed:")

		if strings.Contains(string(out[:]), "FAIL") {
			klog.Info("Ping test failed.", string(out[:]))
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
						utils.HchecksGauge.WithLabelValues("ping", os.Getenv("NODE_NAME"), utils.CPUModel, utils.GPUModel, entry[1]).Set(float64(1))
						klog.Info("Observation: ", entry[1], " ", entry[2], " ", entry[3], " Unreachable")
						unreach_nodes[entry[1]] = append(unreach_nodes[entry[1]], entry[2])
					} else {
						utils.HchecksGauge.WithLabelValues("ping", os.Getenv("NODE_NAME"), utils.CPUModel, utils.GPUModel, entry[1]).Set(float64(0))
					}
				}
			}
		}
		klog.Info("Unreachable nodes count: ", len(unreach_nodes))
	}
	return &out, nil
}

func runIperf(workload string, pclients string, startport string, cleanup string) (*[]byte, error) {

	if workload == "" || pclients == "" || startport == "" {
		klog.Error("Must provide arguments \"workload\", \"pclients\" and \"startport\".")
		return nil, nil
	}

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

func startIperfServers(numservers string, startport string) (*[]byte, error) {
	if numservers == "" || startport == "" {
		klog.Error("Must provide arguments \"numservers\" and \"startport\".")
		return nil, nil
	}
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

func stopAllIperfServers() (*[]byte, error) {
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

func startIperfClients(dstip string, dstport string, numclients string) (*[]byte, error) {
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
		var res float64
		res = 0
		if strings.Contains(split[len(split)-1], "SUCCESS") {
			klog.Info("Observation: ", os.Getenv("NODE_NAME"), " Pass ", res)
		} else {
			res = 1
			klog.Info("Observation: ", os.Getenv("NODE_NAME"), " Fail ", res)
		}
		utils.HchecksGauge.WithLabelValues("dcgm", os.Getenv("NODE_NAME"), utils.CPUModel, utils.GPUModel, "").Set(res)
	}
	return &out, nil
}

func runGPUPower() (*[]byte, error) {
	out, err := exec.Command("bash", "./gpu-power/power-throttle.sh").Output()
	if err != nil {
		klog.Error(err.Error())
		return nil, err
	}
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
		}
		klog.Info("Observation: ", os.Getenv("NODE_NAME"), " ", strconv.Itoa(gpuid), " ", pw)
		utils.HchecksGauge.WithLabelValues("power-slowdown", os.Getenv("NODE_NAME"), utils.CPUModel, utils.GPUModel, strconv.Itoa(gpuid)).Set(pw)

	}
	return &out, nil
}

func runCreateDeletePVC() (*[]byte, error) {
	_, exists := os.LookupEnv("PVC_TEST_STORAGE_CLASS")
	if !exists {
		b := []byte("Storage class not set. Cannot run. ABORT")
		return &b, errors.New("storage class not set")
	}
	err := utils.CreatePVC()
	if err != nil {
		klog.Error(err.Error())
		b := []byte("Create PVC Failed. ABORT")
		return &b, err
	}
	// Wait a few seconds before start checking
	waitonpvc := time.NewTicker(30 * time.Second)
	defer waitonpvc.Stop()
	<-waitonpvc.C
	out, err := utils.ListPVC()
	if err != nil {
		klog.Error(err.Error())

	}
	b := []byte(out)
	return &b, nil
}
