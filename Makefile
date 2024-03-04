TAG=dev
IMAGE=containerregistry:5000/autopilot

image-build:
	@docker build -t ${IMAGE}:v${TAG} -f autopilot-daemon/Dockerfile autopilot-daemon/

image-push:
	@docker push ${IMAGE}:v${TAG}
