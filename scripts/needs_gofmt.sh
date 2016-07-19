#!/bin/sh
#
# Checks if any files about to be committed need gofmt'ing.

echo "Checking if any files need gofmt.." >&2
IFS="
"
if git rev-parse HEAD >/dev/null 2>&1; then
		FILES=$(git diff --cached --name-only | grep -e '\.go$');
else
		FILES=$(git ls-files -c | grep -e '\.go$');
fi
for file in $FILES; do
		badfile="$(git --no-pager show :"$file" | gofmt -s -l)"
		if test -n "$badfile" ; then
				echo "git pre-commit check failed: file needs 'gofmt -s': $file" >&2
				exit 1
		fi
done
