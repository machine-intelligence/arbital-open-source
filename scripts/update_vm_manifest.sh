#!/bin/bash
#
# Updates the monitoring VM from container manifest.
#
CONTAINER_MANIFEST="containers/monitoring.yaml"
ZONE="europe-west1-a"
if [ ! -e ${CONTAINER_MANIFEST} ]; then
		echo "Missing manifest file ${CONTAINER_MANIFEST}." >&2
		exit 1
fi

gcloud compute instances add-metadata monitoring \
    --metadata-from-file "google-container-manifest=${CONTAINER_MANIFEST}" \
    --zone ${ZONE}
echo "All done." >&2
