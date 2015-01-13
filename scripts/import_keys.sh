#!/bin/bash

# Imports GPG keys from ../keys into local keyring.

for k in keys/*.asc; do
		gpg --import $k
done
echo "All done."
