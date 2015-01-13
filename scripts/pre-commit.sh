#!/bin/bash
#
# Precommit scripts for xelaie.

if [ -z "${TRUST_ME_I_KNOW_WHAT_I_AM_DOING}" ]; then
	 scripts/needs_gofmt.sh
else # Allow for a failsafe.
	echo "Okay, if you say so. Skipping pre-commit checks. Have fun." >&2
fi
