.PHONY: gpu-bw-image
gpu-bw-image:
	docker build -t pcie-test:dev -f gpu-bw-test/Dockerfile gpu-bw-test/

gpu-mem-image:
	docker build -t gpu-memcheck:dev -f gpu-mem-test/Dockerfile gpu-mem-test/


.PHONY: submodule-init
submodule-init:
	git submodule update --init --recursive
