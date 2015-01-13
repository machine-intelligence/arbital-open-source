#!/bin/bash
#
# Decrypts local config.yaml.gpg and saves it on remote host.
#
# This script allows your private GPG key to stay on your local
# machine, but to be used to decrypt the config remotely.
#
# TODO: gpg-agent (https://www.gnupg.org/documentation/manuals/gnupg/Invoking-GPG_002dAGENT.html)
# or keychain (http://funtoo.org/Package:Keychain) should be able to
# forward GPG keys, if this can be made to work that might be a
# cleaner setup.

source init.sh || exit

if [ "$#" -ne 2 ]; then
		echo "Usage: $0 [vm, e.g. 'monitoring'] [remote user]"
		exit
fi

INSTANCE=$(cfg "vm.${1}.instance")
ZONE=$(cfg "vm.${1}.zone")
REMOTE_USER=${2}
FILE="src/xelaie/config.yaml"
CONFIG=/home/${REMOTE_USER}/${FILE}

echo "Decrypting config.yaml.pgp and copying to ${REMOTE_USER}@${HOST}:${CONFIG}.."
gpg --decrypt config.yaml.pgp | gcloud compute ssh ${INSTANCE} \
    --zone ${ZONE} --command \
    "sudo sh -c 'cat - > ${CONFIG} && sudo chmod 600 ${CONFIG}' && sudo chown ${REMOTE_USER}:${REMOTE_USER} ${CONFIG}"
