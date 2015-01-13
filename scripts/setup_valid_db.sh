#!/bin/bash
#
# Inserts a static tweet with some associated rewards into the DB at localhost.

source init.sh || exit

HOST=localhost

HOST_PARAM=$1
DB_NAME=$(cfg mysql.database)
DB_USER=$(cfg mysql.user)
USER_PW=$(cfg mysql.password)

echo "==== Inserting custom rewards"

mysql --host ${HOST} -u ${DB_USER} -p${USER_PW} ${DB_NAME} -e "INSERT IGNORE INTO rewards(description, type, company) VALUES ('\$100 Amazon Gift Card', 'giftCard', 'Amazon');"
mysql --host ${HOST} -u ${DB_USER} -p${USER_PW} ${DB_NAME} -e "INSERT IGNORE INTO rewards(description, type, company) VALUES ('\$25 Amazon Gift Card', 'giftCard', 'Amazon');"
mysql --host ${HOST} -u ${DB_USER} -p${USER_PW} ${DB_NAME} -e "INSERT IGNORE INTO rewards(description, type, company) VALUES ('\$5 Amazon Gift Card', 'giftCard', 'Amazon');"

import-rewards() {
		echo "==== Importing rewards"
		cd src/py
		./import_rewards.py "../../data/discounts.csv" $HOST_PARAM
}
import-rewards

echo "==== Inserting contests"

mysql --host ${HOST} -u ${DB_USER} -p${USER_PW} ${DB_NAME} -e "INSERT INTO contests (ownerId, ownerName, ownerScreenName, hashtag, text) VALUES (2804496104, 'Xelaie Sweepstakes', 'Xelaie', '#rt2win', 'It\'s EASY to win with @Xelaie! Just RT this message and INSTANTLY win prizes including a \$100 Amazon gift card! #rt2win');"
mysql --host ${HOST} -u ${DB_USER} -p${USER_PW} ${DB_NAME} -e "INSERT INTO contests (ownerId, ownerName, ownerScreenName) VALUES (2804496104, 'Xelaie', 'Xelaie Sweepstakes');"
