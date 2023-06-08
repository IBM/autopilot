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
  echo -n "Detected NVIDIA GPU: "
  for d in $RES; do 
    d=${d#*"nvidia"*}
    if [[ "$d" =~ $numre ]]; then
      echo -n "$d "
      D=$((D+1))
    fi
  done
  echo "Total: $D"
else
  echo "No NVIDIA GPU detected. Skipping the Remapped Rows check."
  echo "SKIP"
  exit 0
fi
RESULT=""
FAIL=0
for i in $(seq 0 1 $((D-1))) ; do
  OUT=$(nvidia-smi -q -i $i| grep -A 10 "Remapped Rows")
  REMAPPED=$(echo $OUT | egrep "Pending\s*:\s+Yes")
  [[ -z "$REMAPPED" ]] && RESULT+="0 " || RESULT+="1 "; FAIL=1
done
if [[ $FAIL -eq 0 ]]; then
  echo FAIL
fi
echo $RESULT