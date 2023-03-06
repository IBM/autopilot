FROM python:slim

RUN apt -y update && apt -y upgrade && \
        apt install -y --no-install-recommends build-essential vim iperf3 && \
        apt -y clean && apt -y autoremove

RUN pip install kubernetes netifaces

COPY entrypoint.py /entrypoint.py

CMD python entrypoint.py