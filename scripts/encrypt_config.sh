#!/bin/bash
#
# Encrypts config.yaml as config.yaml.pgp, decryptable by trusted GPG keys.
#
# Note that your local pgp keyring needs to know about everyone on the
# --recipient line below before you'll be able to use their public
# keys for encryption. Run `gpg --import keys/foo.asc` for the
# relevant key foo.asc that you don't have to import.

source init.sh || exit

gpg --output config.yaml.pgp --encrypt --armor --recipient me@hkjn.me --recipient alexei@xelaie.com config.yaml
