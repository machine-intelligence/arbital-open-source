rewards
=============

...

Components
=============

The current system consists of two separate parts, along with a shared
config and some helper scripts. It's recommended to set up the
following environment variable in `.bashrc` or similar to make life
easier while working on the codebase:
`export PATH=[this repo's path]/scripts:${PATH}`

AppEngine website
-------------
Source lives in directory `src/site`. Run `stage.sh` to bring up a
development AppEngine server on your local machine. You'll need the Go
AppEngine SDK, as well as some other dependencies. See
`scripts/gcloud_bootstrap.sh` for a script that sets up the
environment on a GCE VM, which may be useful as a reference.

Configuration in config.yaml
-------------
...
