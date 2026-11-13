#!/usr/bin/env bash
# Copyright 2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail

THIS_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
SRC_DIR="$THIS_DIR/src"

testcases=(
	cca-good
	cca-bad
)

for case in "${testcases[@]}"; do
	echo "generating ${case}.cbor"
	allow="false"
	if [[ "$case" == *bad* ]]; then
		echo "file is bad"
		allow="true"
	fi

	evcli cca create -I="${allow}" --iak="$SRC_DIR/ec.p256.jwk" --rak="$SRC_DIR/ec.p384.jwk" \
		--claims="$SRC_DIR/${case}.json" --token="$THIS_DIR/${case}.cbor"

done

echo "done"