#!/bin/bash
#
# Connects to DB with interactive mysql shell.
#
# CAUTION: The session has the same permissions as the app user. Use
# with caution, especially on live.

source init.sh || exit

if [ "$#" -ne 1 ]; then
  echo "Usage: $0 [localhost|live]"
  exit
fi

if [ $1 == "localhost" ]; then
  HOST=localhost
else
  HOST=$(cfg "mysql.${1}.address")
fi

mysql --host ${HOST} -u $(cfg mysql.user) -p$(cfg mysql.password) $(cfg mysql.database)
