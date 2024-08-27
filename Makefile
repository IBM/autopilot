TAG=tyler-dev
IMAGE=quay.io/autopilot/autopilot

image-build:
	@docker build -t ${IMAGE}:v${TAG} -f autopilot-daemon/Dockerfile autopilot-daemon/

image-push:
	@docker push ${IMAGE}:v${TAG}

all: image-build image-push
