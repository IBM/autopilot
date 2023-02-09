FROM python:slim

RUN apt -y update && apt -y upgrade && \
        apt install -y --no-install-recommends build-essential curl vim && \
        apt -y clean && apt -y autoremove

RUN pip install kubernetes requests

COPY entrypoint.py /entrypoint.py

CMD python entrypoint.py