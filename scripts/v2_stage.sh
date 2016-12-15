#!/bin/bash
#
# Stages the App Engine website locally.

source init.sh || exit

check_deps.go || exit

# Copy over the config into the AE app's directory, since we need it
# there for deployment.
cp -v config.yaml src/v2/

# Start dev server, allowing access from *any* IP address. (Don't run
# this if such access is undesirable, e.g. if your machine's IP is
# publicly accessible and pages that you're working on that are super
# secret aren't properly guarded by other auth mechanisms).
dev_appserver.py \
	src/v2/app.yaml \
	--admin_port 8011 --port 8012 --host 0.0.0.0 --enable_sendmail=yes &
appserver_PID=$!

# Start webpack-dev-server to serve webpack bundles. The dev server
# will watch for updates to files that the bundles depends on and
# hot-reload them in the browser.
#
# Keep the port in sync with pageHandler.go.
npm run webpack-dev-server -- \
    --inline \
    --progress \
    --color \
    --port 8014 \
    --output-public-path "http://localhost:8014/static/js/" \
    --hot \
    --config dev.config.js \
    &
webpack_server_PID=$!

# Kill both dev servers on ctrl-c.
trap ctrl_c INT
function ctrl_c() {
    kill $appserver_PID
    kill $webpack_server_PID
}

# https://stackoverflow.com/questions/2935183/bash-infinite-sleep-infinite-blocking
cat
