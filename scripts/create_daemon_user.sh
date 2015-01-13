#!/bin/bash
#
# Sets up 'xelaiedaemon' user, which runs the monitoring
# dashboards. The password needs to be entered interactively (TODO:
# fix this, read from config.yaml), and should be the value in
# config.vm.daemon.password.
#
# This script assumes that the environment on this machine is set up
# correctly for your user, i.e. that you can run stage.sh /
# run_monitoring.sh and everything works.
#
# Note that the xelaiedaemon will be able to authenticate as you to
# gcloud since your credentials are copied over. TODO: use service
# account here if it can be made to work.

source init.sh || exit
DAEMON=xelaiedaemon
sudo adduser ${DAEMON}

for t in ".bashrc" "go_appengine" "go" "google-cloud-sdk" "src"; do
  sudo cp -vr ~/${t} /home/${DAEMON}/
  sudo chown -vR ${DAEMON}:${DAEMON} /home/${DAEMON}/${t}
done

