from enum import Enum
from decimal import Decimal
import argparse
import asyncio
import logging
import aiohttp
import os
import json
import requests
import netifaces
import subprocess
import sys
import signal

from kubernetes import client, config
from kubernetes.client.rest import ApiException

log = logging.getLogger(__name__)
logging.basicConfig(
    format="[NETWORK] - [IPERF] - [%(levelname)s] : %(message)s",
    level=logging.INFO,
)


#
# TODO: Add this to network_workload.py
#
class SupportedWorkload(Enum):
    RING = "RING"


CURR_POD_NAME = os.getenv("POD_NAME")
CURR_WORKER_NODE_NAME = os.getenv("NODE_NAME")
AUTOPILOT_NAMESPACE = os.getenv("NAMESPACE")
AUTOPILOT_PORT = os.getenv("AUTOPILOT_HEALTHCHECKS_SERVICE_PORT")
