#!/bin/bash

# Bootstrap script for a GCE VM to act as host for xelaie.
#
# NOTE: Somewhat untested, likely not fully working in non-interactive
# mode. May still be useful as a guide for what is needed on a clean
# host to get dev environment set up.
#
AE_FILE=go_appengine_sdk_linux_amd64-1.9.13.zip
AE_SDK=https://storage.googleapis.com/appengine-sdks/featured/$AE_FILE
GC_PROJECT=examplary-cycle-688

# Fail if any command fails (returns != 0).
set -e
set -o pipefail

# TODO: we should use a service account here, but instructions at
# https://developers.google.com/console/help/new/#serviceaccounts do
# not line up with actual options under
# https://console.developers.google.com/project/apps~exemplary-cycle-688/apiui/credential..
GOOGLE_ACCOUNT=me@hkjn.me

# Install required/useful packages.
sudo apt-get -y update
sudo apt-get -y install tmux unzip git emacs mysql-client mercurial
export GOPATH=~/

# Set upp AppEngine SDK.
wget $AE_SDK
unzip $AE_FILE

# Set up GCloud SDK, initialize repo.
# TODO: This shouldn't be necessary, as gcloud already exists in the
# base GCE image, but default version seemingly has a bug:
# $ gcloud init exemplary-cycle-688
# Initialized gcloud directory in [/home/zero/exemplary-cycle-688/.gcloud].
# [..]
# /usr/local/bin/git-credential-gcloud.sh: 9: cd: can't cd to ../share/google/google-cloud-sdk/bin
# /usr/local/bin/git-credential-gcloud.sh: 11: /usr/local/bin/git-credential-gcloud.sh: /gcloud: not found
# fatal: 'credential-gcloud.sh' appears to be a git command, but we were not
# able to execute it. Maybe git-credential-gcloud.sh is broken?
# Username for 'https://source.developers.google.com': ^C
#
# So, let's move away existing and broken gcloud install and get a fresh one..
sudo mv /usr/local/share/google/google-cloud-sdk/ ~/old-cloud-sdk/
wget https://dl.google.com/dl/cloudsdk/release/google-cloud-sdk.zip
unzip google-cloud-sdk.zip
./google-cloud-sdk/install.sh \
    --usage-reporting true \
    --bash-completion true \
    --rc-path ~/.bashrc \
    --path-update true \
    --disable-installation-options

google-cloud-sdk/bin/gcloud config set account $GOOGLE_ACCOUNT
google-cloud-sdk/bin/gcloud auth login
google-cloud-sdk/bin/gcloud init $GC_PROJECT

# Move repo and GCloud metadata from default directory to ~/src, which
# is required to have Go packages be importable from other code:
# https://golang.org/doc/code.html
mkdir ~/src
mv -v ${GC_PROJECT}/default/ ~/src/xelaie
mv -v ${GC_PROJECT}/.gcloud ~/src/xelaie
rm -rfv ${GC_PROJECT}

# Install Go from source, since Debian is way behind (1.02, current
# release is 1.3.3).
hg clone -u release https://code.google.com/p/go
cd go/src
./make.bash

# Install Go dependencies.
go get github.com/garyburd/go-oauth/oauth github.com/hkjn/pages github.com/hkjn/timeutils github.com/go-sql-driver/mysql gopkg.in/yaml.v2 github.com/golang/glog github.com/gorilla/sessions github.com/gorilla/mux code.google.com/p/goauth2/oauth github.com/imdario/mergo github.com/influxdb/influxdb/client github.com/sendgrid/sendgrid-go

# Enable git highlighting.
git config --global color.ui "auto"

# Stage the AE app.
cd ~/src/xelaie && ./stage.sh

