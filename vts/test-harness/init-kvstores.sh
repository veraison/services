# Copyright 2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.
#
#!/bin/bash

set -eux
set -o pipefail

for t in en ta
do
    echo "CREATE TABLE kvstore ( key text NOT NULL, vals text NOT NULL );" | \
        sqlite3 $t-store.sql
done
