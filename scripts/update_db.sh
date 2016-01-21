#!/bin/bash
#
# Updates the tables with the schema changes.

source init.sh || exit

HOST=localhost

DB_NAME=$(cfg mysql.database)
DB_USER=$(cfg mysql.user)
ROOT_PW=$(cfg mysql.root.password)
USER_PW=$(cfg mysql.password)

cat schemas/helpers/changes.sql | mysql -f --host ${HOST} -u ${DB_USER} -p${USER_PW} ${DB_NAME}

echo "All done."
