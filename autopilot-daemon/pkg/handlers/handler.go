package handlers

import (
	"fmt"
	"net/http"
	"os/exec"

	"k8s.io/klog/v2"
)

func PCIeBWHandler(pciebw string) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Requesting pcie test with bw: " + pciebw))

		out, err := exec.Command("python3", "./gpubw/entrypoint.py").Output()
		if err != nil {
			klog.Error(err.Error())
		} else {
			klog.Info("GPU PCIe BW Test completed:")
			output := string(out[:])
			fmt.Println(output)
		}

	}
	return http.HandlerFunc(fn)
}

func GPUMemHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Requesting HBM test "))
	}
	return http.HandlerFunc(fn)
}

func NetReachHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Requesting secondary nics reachability test"))
	}
	return http.HandlerFunc(fn)
}
