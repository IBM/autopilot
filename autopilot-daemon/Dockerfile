FROM pytorch/pytorch:1.12.1-cuda11.3-cudnn8-devel as cudabuild

RUN apt -y update && apt -y upgrade && apt -y clean && apt -y autoremove \
        && apt install -y --no-install-recommends build-essential git wget openssh-server && \
        apt -y clean && apt -y autoremove

RUN git clone https://github.com/NVIDIA/cuda-samples.git
WORKDIR cuda-samples/Samples/1_Utilities/bandwidthTest

RUN make SMS="60 70 80 86" 

FROM golang:1.19 AS gobuild

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

WORKDIR /workspace
COPY . /workspace/

RUN go build -o bin/autopilot pkg/cmd/main.go

####################### Final Image

# FROM python:3.9.15-slim
FROM nvidia/cuda:12.1.1-runtime-ubuntu20.04
RUN  apt -y update && apt -y upgrade &&  DEBIAN_FRONTEND="noninteractive" TZ="America/New_York" apt install -y --no-install-recommends \ 
        build-essential iperf3 iputils-ping \
        python3 \
        pip \
        pciutils \
        wget \
        net-tools \
        software-properties-common \
        && apt -y clean && apt -y autoremove

# Add capabilities for ping
RUN setcap cap_net_raw,cap_net_admin+p /bin/ping
RUN add-apt-repository "deb https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2004/x86_64/ /" && apt -y update && apt install -y datacenter-gpu-manager
# add ca-certificates (Alpine commands, previous base image)
# RUN apk update && apk --no-cache  add ca-certificates
# RUN adduser -s /bin/bash -D -h /home/autopilot autopilot -G root 

# FROM pytorch/pytorch:1.12.1-cuda11.3-cudnn8-runtime
RUN useradd -ms /bin/bash autopilot && usermod -g root autopilot 

# set working directory
WORKDIR /home/autopilot

COPY --from=gobuild /workspace/bin/autopilot /usr/local/bin/autopilot

# PCIe tests files
COPY --from=cudabuild /workspace/cuda-samples/Samples/1_Utilities/bandwidthTest/bandwidthTest /home/autopilot/gpu-bw/bandwidthTest
COPY gpu-bw/gpuLocalBandwidthTest.sh /home/autopilot/gpu-bw/gpuLocalBandwidthTest.sh
COPY gpu-bw/entrypoint.py /home/autopilot/gpu-bw/entrypoint.py

# Network tests files
# COPY network/metrics-entrypoint.py /home/autopilot/network/metrics-entrypoint.py
COPY network/iperf3-entrypoint.py /home/autopilot/network/iperf3-entrypoint.py
COPY network/ping-entrypoint.py /home/autopilot/network/ping-entrypoint.py
COPY network/start-iperf-servers.py /home/autopilot/network/start-iperf-servers.py
COPY network/iperf3-debrief.sh /home/autopilot/network/iperf3-debrief.sh

# Copy Python script to read Multi-NIC Health Checker status
RUN cd network && wget https://raw.githubusercontent.com/foundation-model-stack/multi-nic-cni/main/health-check/example-client/read_status.py


# Remapped Rows test files
COPY gpu-remapped/entrypoint.py /home/autopilot/gpu-remapped/entrypoint.py
COPY gpu-remapped/remapped-rows.sh /home/autopilot/gpu-remapped/remapped-rows.sh

COPY utils /home/autopilot/utils

# DGCM test files and dependencies 
COPY gpu-dcgm/entrypoint.py /home/autopilot/gpu-dcgm/entrypoint.py
# RUN wget https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2004/x86_64/datacenter-gpu-manager_3.1.8_amd64.deb && dpkg --install datacenter-gpu-manager_3.1.8_amd64.deb

# GPU Power cap
COPY gpu-power/power-throttle.sh /home/autopilot/gpu-power/power-throttle.sh

# Last touches
RUN pip install --upgrade pip && pip install kubernetes netifaces aiohttp[speedups]
RUN apt -y update && apt install -y vim && apt -y clean && apt -y autoremove
RUN chmod 755 /usr/local/bin/autopilot && chown -hR autopilot /home/autopilot && chmod -R g=u /home/autopilot
CMD ["/usr/local/bin/autopilot"]