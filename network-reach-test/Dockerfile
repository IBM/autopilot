FROM python:latest

# Install iperf and ping
RUN apt -y update && apt -y upgrade && \
        apt install -y --no-install-recommends build-essential vim iperf3 iputils-ping && \
        apt -y clean && apt -y autoremove

# Add capabilities for ping
RUN setcap cap_net_raw,cap_net_admin+p /bin/ping

# Install required Python libraries
RUN pip install kubernetes

# Copy main Python script
COPY metrics-entrypoint.py /metrics-entrypoint.py

# Copy Python script to read Multi-NIC Health Checker status
RUN mkdir network
RUN cd network && wget https://raw.githubusercontent.com/foundation-model-stack/multi-nic-cni/main/health-check/example-client/read_status.py

# Run main Python script
WORKDIR /
CMD python metrics-entrypoint.py