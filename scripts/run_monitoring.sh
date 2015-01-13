#!/bin/bash
#
# Runs monitoring jobs on localhost.
source init.sh || exit

go build zanaduu3/src/monitoring/dash/xelaiemon
./xelaiemon -alsologtostderr
