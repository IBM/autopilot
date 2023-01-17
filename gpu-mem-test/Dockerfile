FROM nvidia/cuda:11.8.0-cudnn8-devel-ubuntu20.04 as build

RUN apt -y update && apt -y upgrade && apt -y clean && apt -y autoremove \
        && apt install -y --no-install-recommends build-essential vim && \
        apt -y clean && apt -y autoremove

WORKDIR /workspace

COPY gpucheck.cu .

RUN nvcc -gencode arch=compute_60,code=sm_60 -gencode arch=compute_70,code=sm_70 -gencode arch=compute_80,code=sm_80 -gencode arch=compute_86,code=sm_86 gpucheck.cu -o gpucheck -lcublas --linker-options -lnvidia-ml -O3


FROM nvidia/cuda:11.8.0-cudnn8-runtime-ubuntu20.04


RUN apt -y update && apt -y upgrade \
        && apt install -y python3 pip  \
        && apt -y clean && apt -y autoremove 

RUN useradd -ms /bin/bash autopilot && usermod -g root autopilot 

WORKDIR /home/autopilot

COPY --from=build /workspace/gpucheck /home/autopilot/gpucheck
COPY precheck.sh /home/autopilot/precheck.sh
COPY entrypoint.py /home/autopilot/entrypoint.py
RUN chown -hR autopilot /home/autopilot && chmod -R g=u /home/autopilot && chmod -R g=u /home/autopilot
RUN python3 -m pip install --upgrade pip && pip install kubernetes

CMD python3 entrypoint.py