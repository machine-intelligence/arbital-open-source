#!/bin/bash

# Copy over the config into the AE app's directory, since it otherwise
# won't be copied over in the deployment.
cp -v config.yaml src/site/
cp -v config.yaml src/queue_daemon/

# Update the queues
#appcfg.py update_queues src/site/

npm run webpack

# Deploy the app.
goapp deploy \
	src/queue_daemon/module.yaml \
	src/site/app.yaml
