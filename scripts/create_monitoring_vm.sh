#!/bin/bash
#
# Creates the monitoring GCE VM from container manifest.
#
source init.sh || exit

PROJECT_NAME="exemplary-cycle-688"
INSTANCE="monitoring-5"
ZONE="europe-west1-b"
CONTAINER_MANIFEST="containers/monitoring.yaml"
if [ ! -e ${CONTAINER_MANIFEST} ]; then
		echo "Missing manifest file ${CONTAINER_MANIFEST}." >&2
		exit 1
fi

# A list of container VM images can be gotten with gcloud compute
# images list --project google-containers, via
# https://cloud.google.com/compute/docs/containers/container_vms.
gcloud compute --project ${PROJECT_NAME} instances create ${INSTANCE} \
    --image "container-vm-v20140929" \
    --image-project "google-containers" \
    --metadata-from-file "google-container-manifest=${CONTAINER_MANIFEST}" \
    --tags "http-server" \
    --zone ${ZONE} \
    --network "staging" \
    --machine-type "f1-micro" \
    --scopes "https://www.googleapis.com/auth/compute" "https://www.googleapis.com/auth/devstorage.read_only"

echo "Sleeping 60s to allow VM to boot.." >&2
sleep 60

gcloud_bootstrap_init.sh ${INSTANCE} ${ZONE}
echo "All done." >&2

