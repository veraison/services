#!/bin/bash
# Copyright 2024 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
# shellcheck disable=SC2086
set -e

TEMP_DIR=/tmp/veraison-end-to-end
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# Check if we have cocli installed.
# Otherwise try to use docker deployment aliases
if ! type cocli >/dev/null 2>&1; then
  shopt -s expand_aliases
  source  ${SCRIPT_DIR}/../../../deployments/docker/env.bash
fi

mkdir -p $TEMP_DIR

for scheme in psa cca cca-realm; do
    cocli comid create --template ${SCRIPT_DIR}/comid-${scheme}-ta.json \
                       --template ${SCRIPT_DIR}/comid-${scheme}-refval.json \
                       --output-dir $TEMP_DIR
    cocli corim create --template ${SCRIPT_DIR}/corim-${scheme}.json \
                       --comid ${TEMP_DIR}/comid-${scheme}-refval.cbor \
                       --comid ${TEMP_DIR}/comid-${scheme}-ta.cbor \
                       --output ${SCRIPT_DIR}/../${scheme}-endorsements.cbor
done

evcli psa create --claims ${SCRIPT_DIR}/psa-evidence.json \
	    --key ${SCRIPT_DIR}/../ec256.json \
	    --token ${SCRIPT_DIR}/../psa-evidence.cbor

evcli cca create --claims ${SCRIPT_DIR}/cca-evidence.json \
	    --iak ${SCRIPT_DIR}/../cca-platform-key.json --rak ${SCRIPT_DIR}/../cca-realm-key.json \
	    --token ${SCRIPT_DIR}/../cca-evidence.cbor

echo "done."
