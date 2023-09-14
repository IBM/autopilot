#!/bin/bash
for FILE in out-*; do 
    unable=`grep unable ${FILE}`
    if [ ! -z "$unable" ]
    then
        echo $FILE
    fi
done

echo Total bw Gbit/s
cat out-* | grep receiver| awk '{print $7}'| awk '{s+=$1}END{print s}'

echo Unreachable servers printed above
count=$(cat out-* | grep unable | wc -l)
echo Count: $count

echo Cleanup...
rm out-*