#!/bin/bash
#
# Builds the "gomon" Docker image.
#
source init.sh || exit
TARGET=containers/monitoring/
cp -v scripts/gcloud_bashrc.sh ${TARGET}
cp -v config.yaml ${TARGET}
# cp src/go/monitoring ${TARGET}
sudo docker build -t "hkjn/gomon" ${TARGET}
