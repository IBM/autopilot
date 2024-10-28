# FROM pytorch/pytorch:1.12.1-cuda11.3-cudnn8-devel as cudabuild
FROM pytorch/pytorch:2.1.2-cuda12.1-cudnn8-devel as cudabuild

RUN apt -y update && apt -y upgrade && apt -y clean && apt -y autoremove \
        && apt install -y --no-install-recommends build-essential git wget openssh-server && \
        apt -y clean && apt -y autoremove

RUN git clone --depth 1 --branch v12.4.1 https://github.com/NVIDIA/cuda-samples.git
WORKDIR cuda-samples/Samples/1_Utilities/bandwidthTest

RUN make SMS="80 86 90" 

WORKDIR /workspace

COPY gpu-mem/gpucheck.cu .

RUN nvcc -gencode arch=compute_80,code=sm_80 -gencode arch=compute_86,code=sm_86 -gencode arch=compute_90,code=sm_90 gpucheck.cu -o gpucheck -lcublas --linker-options -lnvidia-ml -O3

FROM golang:1.21 AS gobuild

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
        git \
        && apt -y clean && apt -y autoremove

RUN add-apt-repository "deb https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2004/x86_64/ /" && apt -y update && apt install -y datacenter-gpu-manager
# add ca-certificates (Alpine commands, previous base image)
# RUN apk update && apk --no-cache  add ca-certificates
# RUN adduser -s /bin/bash -D -h /home/autopilot autopilot -G root 

# RDMA ping utils
# RUN apt -y update  && apt -y upgrade && apt -y install build-essential automake autoconf libtool libibverbs-dev librdmacm-dev libibumad-dev pciutils libpci-dev
# RUN git clone https://github.com/linux-rdma/perftest.git && \
#         cd perftest && \
#         ./autogen.sh && ./configure && \
#         make && make install
# Add capabilities for ping
RUN setcap cap_net_raw,cap_net_admin+p /bin/ping

RUN useradd -ms /bin/bash autopilot && usermod -g root autopilot 

# set working directory
WORKDIR /home/autopilot

COPY --from=gobuild /workspace/bin/autopilot /usr/local/bin/autopilot

# PCIe tests files
COPY --from=cudabuild /workspace/cuda-samples/Samples/1_Utilities/bandwidthTest/bandwidthTest /home/autopilot/gpu-bw/bandwidthTest
COPY gpu-bw/gpuLocalBandwidthTest.sh /home/autopilot/gpu-bw/gpuLocalBandwidthTest.sh
COPY gpu-bw/entrypoint.py /home/autopilot/gpu-bw/entrypoint.py

# DGEMM DAXPY test files

COPY --from=cudabuild /workspace/gpucheck /home/autopilot/gpu-mem/gpucheck
COPY gpu-mem/entrypoint.py /home/autopilot/gpu-mem/entrypoint.py


# Network tests files
# COPY network/metrics-entrypoint.py /home/autopilot/network/metrics-entrypoint.py
COPY network/ping-entrypoint.py /home/autopilot/network/ping-entrypoint.py
COPY network/iperf3_entrypoint.py /home/autopilot/network/iperf3_entrypoint.py
COPY network/iperf3_utils.py /home/autopilot/network/iperf3_utils.py
COPY network/network_workload.py /home/autopilot/network/network_workload.py
COPY network/iperf3_start_servers.py /home/autopilot/network/iperf3_start_servers.py
COPY network/iperf3_stop_servers.py /home/autopilot/network/iperf3_stop_servers.py
COPY network/iperf3_start_clients.py /home/autopilot/network/iperf3_start_clients.py

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
RUN apt -y update && apt install -y vim curl && apt -y clean && apt -y autoremove
RUN chmod 755 /usr/local/bin/autopilot && chown -hR autopilot /home/autopilot && chmod -R g=u /home/autopilot
RUN chmod 777 /tmp



CMD ["/usr/local/bin/autopilot"]
