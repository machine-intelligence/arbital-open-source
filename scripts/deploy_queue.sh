#!/bin/bash

# Copy over the config into the AE app's directory, since it otherwise
# won't be copied over in the deployment.
cp -v config.yaml src/go/queue_daemon/

# Deploy the app.
goapp deploy src/go/queue_daemon/module.yaml
