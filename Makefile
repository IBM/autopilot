.PHONY: gpu-bw-image
gpu-bw-image:
	docker build -t healthcheck:dev -f gpu-bw-test/Dockerfile scripts/

.PHONY: submodule-init
submodule-init:
	git submodule update --init --recursive
