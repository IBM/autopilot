#!/bin/bash
#
# MH:
# This file is supposed to be used for GPU instance in LLM cluster with PXB topology. Test the localhost only
#
# Requirement: pre-compiled bandwidthTest from cuda_samples on instances.
#
# Usage:
# A. Update PROG and FN in the script. Threshold T is set to 7 by default
# B. run gpuLocalBandwidthTest.sh
# C. Check the last line of output: SUCCESS or FAIL
#
# Note: some parameters are hard coded. You may want to change them for different environment.
#
# Find me at minghungchen@ibm.com if any questions
#
# Ver. 1.2

PROG="/home/autopilot/bandwidthTest"
FN="gpuBandwidthTest.log"
T="7"

numre='^[0-9]+$'
D=0
if [[ -e "/dev/nvidia0" ]]; then
    for d in $(ls -d /dev/nvidia*); do 
    d=${d#*"nvidia"*}
    if [[ "$d" =~ $numre ]]; then
        D=$((D+1))
    fi
    done
fi
if [[ "$D" == "0" ]]; then
  echo "No NVIDIA GPU detected. Skipping the bandwidth test."
  echo "SKIP"
  exit 0
fi
if [[ "$D" != "8" ]]; then
  echo "Not all 8 x NVIDIA GPU were detected. Skipping the bandwidth test to avoid inaccurate evaluation."
  echo "ABORT"
  exit 0
fi

#echo "bandwidthTest Program: $PROG Devices: $D"
D=$((D-1))
for i in $(seq 0 1 $D) ; do
  EXEC=$($PROG --htod --memory=pinned --device=$i --csv | tee -a $FN |grep bandwidthTest-H2D-Pinned | awk -v T=$T -v C=$i '{if ($4 < T) print " GPU " C " has low memory bandwidth: " $4 " GB/s"; }')
  if [[ "$EXEC" != "" ]]; then
    F=1
    echo $EXEC
  fi
done
if [[ "$F" -eq "1" ]]; then
  echo "Please attach this bandwidthTest report to Cloud Support:"
  cat $FN
  echo "FAIL"
  exit 1
else
  rm -f $FN
  echo "SUCCESS"
fi
