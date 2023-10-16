#!/bin/bash
RES=$(ls -d /dev/nvidia* 2>1)
numre='^[0-9]+$'
D=-1
for d in $RES; do
  d=${d#*"nvidia"*}
  if [[ "$d" =~ $numre ]]; then
    D=0
    break
  fi
done
if [[ $D -eq 0 ]]; then
  echo -n "[GPU POWER] Detected NVIDIA GPU: "
  for d in $RES; do
    d=${d#*"nvidia"*}
    if [[ "$d" =~ $numre ]]; then
      echo -n "$d "
      D=$((D+1))
    fi
  done
  echo "Total: $D"
else
  echo "[GPU POWER] No NVIDIA GPU detected. Skipping the Power Throttle check."
  echo "ABORT"
  exit 0
fi
RESULT=""
FAIL=0
for i in $(seq 0 1 $((D-1))) ; do
  OUT=$(nvidia-smi --format=csv -i $i --query-gpu=clocks_event_reasons.hw_slowdown)
  NOTACTIVE=$(echo $OUT | grep "Not Active")
  if [[ ! -z "$NOTACTIVE" ]]; then
    RESULT+="0 "
  else
    RESULT+="1 "
    FAIL=1
  fi
done
if [[ $FAIL -ne 0 ]]; then
  echo "[GPU POWER] FAIL"
else
  echo "[GPU POWER] SUCCESS"
fi
echo $RESULT