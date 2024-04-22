FROM nvidia/cuda:12.1.0-devel-ubuntu22.04 as base

ENV HOME=/homedir \
    PYTHON_VERSION=3.11 \
    PATH=/opt/conda/envs/ai/bin:/opt/conda/bin:${PATH} \
    MOFED_VER=23.04-1.1.3.0 \
    OS_VER=ubuntu22.04 \
    PLATFORM=x86_64

WORKDIR /app

RUN apt-get -y update && \
    apt-get install -y make git git-lfs curl wget unzip libaio-dev libnvidia-compute-525 && \
    apt-get -y clean

RUN wget --quiet http://content.mellanox.com/ofed/MLNX_OFED-${MOFED_VER}/MLNX_OFED_LINUX-${MOFED_VER}-${OS_VER}-${PLATFORM}.tgz && \
    tar -xvzf MLNX_OFED_LINUX-23.04-1.1.3.0-ubuntu22.04-x86_64.tgz && \
    ./MLNX_OFED_LINUX-23.04-1.1.3.0-ubuntu22.04-x86_64/mlnxofedinstall --user-space-only --without-fw-update --all --force && \
    rm -rf MLNX_OFED_LINUX-23.04-1.1.3.0-ubuntu22.04-x86_64 MLNX_OFED_LINUX-23.04-1.1.3.0-ubuntu22.04-x86_64.tgz

# taken form pytorch's dockerfile
RUN curl -L -o ./miniconda.sh -O https://repo.anaconda.com/miniconda/Miniconda3-latest-Linux-x86_64.sh && \
    bash miniconda.sh -b -p /opt/conda && \
    rm ./miniconda.sh

# create conda env
RUN conda create -n ai python=${PYTHON_VERSION} pip -y

FROM base as conda

# update conda
RUN conda update -n base -c defaults conda -y
# cmake
RUN conda install -c anaconda cmake -y

# necessary packages


RUN pip install --pre torch==2.3.0.dev20240311+cu121 torchvision torchaudio --index-url https://download.pytorch.org/whl/nightly/cu121 --no-cache-dir
RUN pip install transformers \
    numpy \
    fire \
    sentencepiece \
    --no-cache-dir

# clean conda env
RUN conda clean -ya

RUN mkdir -p ~/.cache ~/.local ~/.triton && \
    chmod -R g+w /app ~/.cache ~/.local ~/.triton /opt/conda

RUN mkdir /apps
WORKDIR /apps
COPY nccl_envs.sh allreduce-loop.py /apps/

CMD ["torchrun", "/apps/allreduce-loop.py"]
