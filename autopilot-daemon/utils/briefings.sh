#!/bin/bash
exists=`which nvidia-smi`
if [[ -z $exists ]]
then
	echo !! nvidia-smi not present. ABORT.
	killall5 
fi

CMD="$(nvidia-smi)"
errors="$(echo ${CMD} | grep -i err)"
if [[ -n $errors ]]
then
	echo !! nvidia-smi failed to start. ABORT.
	killall5
fi

CMD="$(nvidia-smi --query-gpu=mig.mode.current --format=csv)"
mig="$(echo ${CMD} | grep Enabled)"
if [[ -n $mig ]]
then
	echo !! MIG enabled. ABORT.
	exit 
fi

CMD="$(dcgmi --version)"
errors="$(echo ${CMD} | grep -i 'fail|error')"
if [[ -n $errors ]]
then
	echo !! dcgmi failed to start. ABORT.
	exit 
fi