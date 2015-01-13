#!/bin/bash
#
# Decrypts config.yaml from config.yaml.gpg.

gpg --output config.yaml --decrypt config.yaml.pgp

chmod 600 config.yaml
