#!/bin/bash
for FILE in out*; do 
    unable=`grep unable ${FILE}`
    if [ ! -z "$unable" ]
    then
        echo $FILE
    fi
done

echo Aggregate bw Gbit/s
cat out* | grep receiver| awk '{print $7}'| awk '{s+=$1}END{print s}'

echo Aggregate bw per interface
for i in $(ls out*); do echo $(echo ${i}| cut -d ':' -f 2) "${i##*_}" $(cat ${i} | grep receiver| awk '{print $7}'| awk '{s+=$1}END{print s}');done 

echo Unreachable servers printed above, if any
count=$(cat out* | grep unable | wc -l)
echo Unreachable servers count: $count

echo Cleanup...
rm out*