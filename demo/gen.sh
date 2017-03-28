#!/bin/bash

i=1; 
for (( ; ; ))
do
  sleep 1
  echo "Number: $((i++))" >> /home/lhj/mnt/localhost/my_topic/bar/writer
done
