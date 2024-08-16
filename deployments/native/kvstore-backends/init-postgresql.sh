#!/bin/bash
# Copyright 2024 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail

ADMIN_USER=postgres
VERAISON_USER=veraison
VERAISON_DB=veraison

set +e
read -r -d '' CREATE_TABLES_SQL <<EOF
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

CREATE INDEX ON endorsements(kv_key);
CREATE INDEX ON trust_anchors(kv_key);
CREATE INDEX ON policies(kv_key);
EOF
set -e

echo "Creating user $VERAISON_USER..."
createuser -U $ADMIN_USER -P $VERAISON_USER

echo "Creating database $VERAISON_DB..."
createdb -U $ADMIN_USER -O $VERAISON_USER $VERAISON_DB

echo "Initialing store tables..."
psql -U $VERAISON_USER -d $VERAISON_DB -c "$CREATE_TABLES_SQL"

echo "done"
