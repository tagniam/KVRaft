#!/bin/bash

echo "" > log.txt
for i in {0..4}
do
    go run src/main.go server 4040,4041,4042,4043,4044 $i &>> "log.txt" &
    echo "launched server $i with pid $!"
done
go run src/main.go client 4040,4041,4042,4043,4044 
./kill.sh
