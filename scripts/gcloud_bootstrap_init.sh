#!/bin/bash
#
# Installs dependencies on a GCE VM to act as a monitoring host for
# xelaie.
#
# This script copies over and executes gcloud_bootstrap.sh on the
# remote VM.
#
# Prerequisites:
# 0. gcloud SDK is installed locally.
# 1. SSH agent has unlocked keys to the VM, so passwordless SSH is possible
#   (`ssh-add ~/.ssh/google_compute_engine or similar).

source init.sh || exit

if [ "$#" -ne 2 ]; then
		echo "Usage: $0 [instance name] [instance zone]" >&2
		exit 1
fi

INSTANCE=${1}
ZONE=${2}

echo "Bootstrapping ${INSTANCE} VM in zone ${ZONE}.." >&2

gcloud compute copy-files \
    scripts/gcloud_bashrc.sh \
    ${INSTANCE}:.bashrc \
    --zone ${ZONE}
gcloud compute copy-files \
    scripts/gcloud_bootstrap.sh \
    ${INSTANCE}: \
    --zone ${ZONE}

# Include a small .emacs to display tabs sanely.
gcloud compute ssh ${INSTANCE} \
    --zone ${ZONE} \
    --command "echo '(setq tab-width 2)' > .emacs"
gcloud compute ssh ${INSTANCE} \
    --zone ${ZONE} \
    --command "~/gcloud_bootstrap.sh"
