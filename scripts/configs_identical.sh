#!/bin/bash
#
# Utility for checking whether the config.yaml is up to date, i.e. if
# the decrypted config.yaml.pgp equals it in contents.
#
# If it does not, this either means that:
# 1. more changes arrived in config.yaml.pgp since your config.yaml was decrypted
# 2. you've changed config.yaml locally and haven't re-encrypted and
#   pushed the changes to config.yaml.pgp
source init.sh || exit

cp -iv config.yaml config.yaml.bak
echo "Decrypting config.yaml.gpg -> config.yaml.." >&2
decrypt_config.sh
if diff config.yaml.bak config.yaml ; then
  echo "Identical." >&2
  rm config.yaml.bak
  exit 0
else
  echo "The config.yaml file differs from your local copy, moved your changes to config.yaml.bak." >&2
	echo "Please merge and commit your changes, or remove if they're not needed." >&2
  exit 1
fi
