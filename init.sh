#!/bin/bash

echo "" > log.txt
for i in {0..9}
do
    go run src/main.go server 4040,4041,4042,4043,4044 $i &>> "log.txt" &
done
go run src/main.go client 4040,4041,4042,4043,4044 localhost
./kill.sh
