TAG=1.4.0
IMAGE=hypervisorepo:5000/autopilot

image:
	@docker build -t ${IMAGE}:v${TAG} -f autopilot-daemon/Dockerfile autopilot-daemon/
	@docker push ${IMAGE}:v${TAG}

publish:
	@git checkout gh-pages
	@git merge main
	@helm repo index --url https://raw.github.ibm.com/hybrid-cloud-infrastructure-research/autopilot/gh-pages .
	@helm package autopilot-daemon/helm-charts/*
	@echo -e "User-Agent: *\nDisallow: /" > robots.txt
	@git add robots.txt index.yaml autopilot-daemon-${TAG}.tgz
	@git commit -m "update helm package"
	@git push
	@git checkout main
