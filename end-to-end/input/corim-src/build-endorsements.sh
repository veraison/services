#!/bin/bash
# Copyright 2024 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
set -e

TEMP_DIR=/tmp/veraison-end-to-end
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

mkdir -p $TEMP_DIR

for scheme in psa cca; do
    cocli comid create --template ${SCRIPT_DIR}/comid-${scheme}-ta.json \
                       --template ${SCRIPT_DIR}/comid-${scheme}-refval.json \
                       --output-dir $TEMP_DIR
    cocli corim create --template ${SCRIPT_DIR}/corim-${scheme}.json \
                       --comid ${TEMP_DIR}/comid-${scheme}-refval.cbor \
                       --comid ${TEMP_DIR}/comid-${scheme}-ta.cbor \
                       --output ${SCRIPT_DIR}/../${scheme}-endorsements.cbor
done

