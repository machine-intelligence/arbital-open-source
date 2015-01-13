#!/bin/bash
#
# Symlinks 

cd .git/hooks/
ln -s ../../scripts/pre-commit.sh pre-commit
ln -s ../../scripts/pre-push.sh pre-push
