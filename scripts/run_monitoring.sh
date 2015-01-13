#!/bin/bash
#
# Runs monitoring jobs on localhost.
source init.sh || exit

go build xelaie/src/go/monitoring/dash/xelaiemon
./xelaiemon -alsologtostderr
