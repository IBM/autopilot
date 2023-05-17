.PHONY: gpu-bw-image
gpu-bw-image:
	docker build -t pcie-test:dev -f gpu-bw-test/Dockerfile gpu-bw-test/

.PHONY: gpu-bw-minimal
gpu-bw-minimal:
	docker build -t pcie-mini:dev -f gpu-bw-test/Dockerfile.reduced gpu-bw-test/

.PHONY: gpu-mem-image
gpu-mem-image:
	docker build -t gpu-memcheck:dev -f gpu-mem-test/Dockerfile gpu-mem-test/

.PHONY: net-reach-image
net-reach-image:
	docker build -t network-test:dev -f network-reach-test/Dockerfile network-reach-test/

.PHONY: install
install:
	helm install autopilot autopilot-daemon/helm-charts/autopilot

.PHONY: uninstall
uninstall:
	helm uninstall autopilot

.PHONY: submodule
submodule-init:
	git submodule update --init --recursive
