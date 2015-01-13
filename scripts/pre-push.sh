#!/bin/bash
#
# Prepush hooks for xelaie.

if [ -z "${TRUST_ME_I_KNOW_WHAT_I_AM_DOING}" ]; then
	 scripts/run_tests.sh
else # Allow for a failsafe.
	echo "Okay, if you say so. Skipping pre-push checks. Have fun." >&2
fi
