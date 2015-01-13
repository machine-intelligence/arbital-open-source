#!/bin/bash
#
# Runs the prober agains rewards.xelaie.com locally.
go build src/go/prober/prober.go
./prober -alsologtostderr
