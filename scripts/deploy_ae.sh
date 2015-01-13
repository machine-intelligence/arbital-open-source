#!/bin/bash

# Copy over the config into the AE app's directory, since it otherwise
# won't be copied over in the deployment.
cp -v config.yaml src/site/

# Deploy the app.
goapp deploy \
	src/site/app.yaml
