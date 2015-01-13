#!/bin/bash
#
# Runs the "gomon" Docker image.
#
source init.sh || exit

sudo docker run -t -p 8083:8083 -p 8086:8086 hkjn/gomon:v1 /etc/init.d/influxdb start