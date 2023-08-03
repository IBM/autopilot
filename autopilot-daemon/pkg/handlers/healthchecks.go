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
	runAllTestsLocal("")
}

func runAllTestsLocal(checks string) (error, *[]byte) {
	out := []byte("")
	var tmp *[]byte
	var err error
	start := time.Now()
	for _, check := range strings.Split(checks, ",") {

		switch check {
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
			klog.Info("Running health check: ", check)
			err, tmp = netReachability()
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
			err, tmp = netReachability()
			if err != nil {
				klog.Error(err.Error())
				return err, nil
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

func runAllTestsRemote(host string, check string, batch string, jobName string) (error, *[]byte) {
	klog.Info("About to run command:\n", "./utils/runHealthchecks.py", " --service=autopilot-healthchecks", " --namespace="+os.Getenv("NAMESPACE"), " --nodes="+host, " --check="+check, " --batchSize="+batch, " --wkload="+jobName)

	out, err := exec.Command("python3", "./utils/runHealthchecks.py", "--service=autopilot-healthchecks", "--namespace="+os.Getenv("NAMESPACE"), "--nodes="+host, "--check="+check, "--batchSize="+batch, "--wkload="+jobName).Output()
	if err != nil {
		klog.Error(err.Error())
		return err, nil
	}

	return nil, &out
}

func netReachability() (error, *[]byte) {
	out, err := exec.Command("python3", "./network/metrics-entrypoint.py").Output()
	if err != nil {
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
	out, err := exec.Command("python3", "./gpu-remapped/entrypoint.py").Output()
	if err != nil {
		klog.Error(err.Error())
		return err, nil
	} else {
		klog.Info("Remapped Rows check test completed:")

		if strings.Contains(string(out[:]), "FAIL") {
			klog.Info("Remapped Rows test failed.", string(out[:]))
		}

		if strings.Contains(string(out[:]), "ABORT") {
			klog.Info("PCIe BW cannot be run. ", string(out[:]))
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
	out, err := exec.Command("python3", "./gpubw/entrypoint.py", "-t", utils.UserConfig.BWThreshold).Output()
	if err != nil {
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
				// utils.Hchecks.WithLabelValues("pciebw", os.Getenv("NODE_NAME"), strconv.Itoa(gpuid)).Observe(bw)
				utils.HchecksGauge.WithLabelValues("pciebw", os.Getenv("NODE_NAME"), strconv.Itoa(gpuid)).Set(bw)
			}
		}
	}
	return nil, &out
}
