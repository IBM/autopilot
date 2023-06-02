package handlers

import (
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.ibm.com/hybrid-cloud-infrastructure-research/autopilot-daemon/pkg/utils"
	"k8s.io/klog/v2"
)

func runPCIeBw() (error, *[]byte) {
	out, err := exec.Command("python3", "./gpubw/entrypoint.py").Output()
	if err != nil {
		klog.Error(err.Error())
		return err, nil
	} else {
		klog.Info("GPU PCIe BW test completed:")
		// w.Write(out)
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
				utils.Hchecks.WithLabelValues("pciebw", os.Getenv("NODE_NAME"), strconv.Itoa(gpuid)).Observe(bw)
				utils.HchecksGauge.WithLabelValues("pciebw", os.Getenv("NODE_NAME"), strconv.Itoa(gpuid)).Set(bw)
			}
		}
	}
	return nil, &out
}
