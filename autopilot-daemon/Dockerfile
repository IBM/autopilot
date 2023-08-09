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
RUN  apt -y update && apt -y upgrade && apt install -y --no-install-recommends \ 
        build-essential iperf3 iputils-ping \
        python3 \
        pip \
        pciutils \
        wget \
        net-tools \
        && apt -y clean && apt -y autoremove

# Add capabilities for ping
RUN setcap cap_net_raw,cap_net_admin+p /bin/ping

# add ca-certificates (Alpine commands, previous base image)
# RUN apk update && apk --no-cache  add ca-certificates
# RUN adduser -s /bin/bash -D -h /home/autopilot autopilot -G root 

# FROM pytorch/pytorch:1.12.1-cuda11.3-cudnn8-runtime
RUN useradd -ms /bin/bash autopilot && usermod -g root autopilot 

# set working directory
WORKDIR /home/autopilot

COPY --from=gobuild /workspace/bin/autopilot /usr/local/bin/autopilot

# PCIe tests files
COPY --from=cudabuild /workspace/cuda-samples/Samples/1_Utilities/bandwidthTest/bandwidthTest /home/autopilot/gpubw/bandwidthTest
COPY gpu-bw/gpuLocalBandwidthTest.sh /home/autopilot/gpubw/gpuLocalBandwidthTest.sh
COPY gpu-bw/entrypoint.py /home/autopilot/gpubw/entrypoint.py
COPY gpu-bw/briefings.sh /home/autopilot/gpubw/briefings.sh

# Secondary NIC tests files
COPY network/metrics-entrypoint.py /home/autopilot/network/metrics-entrypoint.py
COPY network/iperf3-entrypoint.py /home/autopilot/network/iperf3-entrypoint.py

# Copy Python script to read Multi-NIC Health Checker status
RUN cd network && wget https://raw.githubusercontent.com/foundation-model-stack/multi-nic-cni/main/health-check/example-client/read_status.py


# Remapped Rows test files
COPY gpu-remapped/entrypoint.py /home/autopilot/gpu-remapped/entrypoint.py
COPY gpu-remapped/briefings.sh /home/autopilot/gpu-remapped/briefings.sh
COPY gpu-remapped/remapped-rows.sh /home/autopilot/gpu-remapped/remapped-rows.sh

COPY utils /home/autopilot/utils

# Last touches
RUN pip install --upgrade pip && pip install kubernetes
RUN apt -y update && apt install -y vim && apt -y clean && apt -y autoremove
RUN chmod 755 /usr/local/bin/autopilot && chown -hR autopilot /home/autopilot && chmod -R g=u /home/autopilot

CMD ["/usr/local/bin/autopilot"]