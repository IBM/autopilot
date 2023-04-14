FROM python:latest

RUN apt -y update && apt -y upgrade && \
        apt install -y --no-install-recommends build-essential vim iperf3 iputils-ping && \
        apt -y clean && apt -y autoremove

RUN pip install kubernetes netifaces

COPY ping-entrypoint.py /ping-entrypoint.py

RUN setcap cap_net_raw,cap_net_admin+p /bin/ping

CMD python ping-entrypoint.py