#!/usr/bin/env bash
# Copyright 2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail

THIS_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
SRC_DIR="$THIS_DIR/src"

evcli cca create --iak="$SRC_DIR/ec.p256.jwk" --rak="$SRC_DIR/ec.p384.jwk" \
	--claims="$SRC_DIR/cca-good.json" --token="$THIS_DIR/cca-good.cbor"
