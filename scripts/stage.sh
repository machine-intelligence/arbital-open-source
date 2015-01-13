#!/bin/bash
#
# Stages the App Engine website locally.

source init.sh || exit

# Copy over the config into the AE app's directory, since we need it
# there for deployment.
cp -v config.yaml src/site/

# Start dev server, allowing access from *any* IP address. (Don't run
# this if such access is undesirable, e.g. if your machine's IP is
# publicly accessible and pages that you're working on that are super
# secret aren't properly guarded by other auth mechanisms).
dev_appserver.py \
	src/site/app.yaml \
	--admin_port 8011 --port 8012 --host 0.0.0.0
