FROM pytorch/pytorch:1.12.1-cuda11.3-cudnn8-devel as build

RUN apt -y update && apt -y upgrade && apt -y clean && apt -y autoremove \
        && apt install -y --no-install-recommends build-essential git wget openssh-server && \
        apt -y clean && apt -y autoremove

RUN git clone https://github.com/NVIDIA/cuda-samples.git
WORKDIR cuda-samples/Samples/1_Utilities/bandwidthTest

RUN make SMS="60 70 80 86" 


FROM ubuntu:18.04

RUN useradd -ms /bin/bash autopilot && usermod -g root autopilot 

WORKDIR /home/autopilot
COPY gpuLocalBandwidthTest.sh /home/autopilot/gpuLocalBandwidthTest.sh
COPY --from=build /workspace/cuda-samples/Samples/1_Utilities/bandwidthTest/bandwidthTest /home/autopilot/bandwidthTest

RUN chmod 755 /home/autopilot/gpuLocalBandwidthTest.sh && chown -hR autopilot /home/autopilot
