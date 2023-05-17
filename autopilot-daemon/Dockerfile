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

# Final Image
# FROM python:3.9.15-slim
FROM nvidia/cuda:12.1.1-runtime-ubuntu20.04
RUN  apt -y update && apt -y upgrade && apt install -y --no-install-recommends \ 
        python3 \
        pip \
        && apt -y clean && apt -y autoremove
# add ca-certificates (Alpine commands, previous base image)
# RUN apk update && apk --no-cache  add ca-certificates
# RUN adduser -s /bin/bash -D -h /home/autopilot autopilot -G root 

# FROM pytorch/pytorch:1.12.1-cuda11.3-cudnn8-runtime
RUN useradd -ms /bin/bash autopilot && usermod -g root autopilot 

# set working directory
WORKDIR /home/autopilot

COPY --from=gobuild /workspace/bin/autopilot /usr/local/bin/autopilot
COPY --from=cudabuild /workspace/cuda-samples/Samples/1_Utilities/bandwidthTest/bandwidthTest /home/autopilot/gpubw/bandwidthTest
COPY gpubw/gpuLocalBandwidthTest.sh /home/autopilot/gpubw/gpuLocalBandwidthTest.sh
COPY gpubw/entrypoint.py /home/autopilot/gpubw/entrypoint.py
COPY gpubw/briefings.sh /home/autopilot/gpubw/briefings.sh
RUN pip install --upgrade pip && pip install kubernetes
RUN apt install -y vim && apt -y clean && apt -y autoremove
RUN chmod 755 /usr/local/bin/autopilot && chown -hR autopilot /home/autopilot && chmod -R g=u /home/autopilot

CMD ["/usr/local/bin/autopilot"]
