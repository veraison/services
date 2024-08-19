#!/bin/bash
# Copyright 2024 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail

ADMIN_USER=root
VERAISON_USER=veraison
VERAISON_DB=veraison

MAX_COL_WIDTH=21,844

[[ "$(type -p mariadb)" == "" ]] && CLIENT=mysql || CLIENT=mariadb

NOPASS=${NOPASS:-}
if [[ "$NOPASS" == "" ]]; then
	ADMIN_OPTS="-u $ADMIN_USER -p"
else
	ADMIN_OPTS="-u $ADMIN_USER"
fi

echo "Creating user $VERAISON_USER..."
while : ; do
	read -r -s -p "Enter password for new user $VERAISON_USER:" VERAISON_PASSWD
	echo ""
	read -r -s -p "And again:" VERAISON_PASSWD_AGAIN
	echo ""
	[[ "$VERAISON_PASSWD" != "$VERAISON_PASSWD_AGAIN" ]] || break
	echo "passwords did not match"
done

# shellcheck disable=SC2086
$CLIENT $ADMIN_OPTS <<EOF
CREATE USER IF NOT EXISTS '$VERAISON_USER'@'%' IDENTIFIED BY '$VERAISON_PASSWD';
EOF

echo "Creating database $VERAISON_DB..."
# shellcheck disable=SC2086
$CLIENT $ADMIN_OPTS <<EOF
CREATE DATABASE IF NOT EXISTS $VERAISON_DB;
GRANT ALL PRIVILEGES ON $VERAISON_DB.* TO '$VERAISON_USER'@'%';
EOF

echo "Initialing store tables..."
$CLIENT --user=$VERAISON_USER --password="$VERAISON_PASSWD" $VERAISON_DB <<EOF
CREATE TABLE IF NOT EXISTS endorsements (
    kv_key TEXT NOT NULL,
    kv_val TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS trust_anchors (
    kv_key TEXT NOT NULL,
    kv_val TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS policies (
    kv_key TEXT NOT NULL,
    kv_val TEXT NOT NULL
);
EOF

echo "done"
