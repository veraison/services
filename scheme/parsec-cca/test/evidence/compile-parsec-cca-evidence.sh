#!/bin/bash
# Copyright 2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail

THIS_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
SRC_DIR="$THIS_DIR/src"
BUILD_DIR="$THIS_DIR/__build"

rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

diag2cbor.rb < "$SRC_DIR/kat.diag" > "$BUILD_DIR/kat.cbor"

KAT_HASH=$(sha512sum "$BUILD_DIR/kat.cbor" | cut -f1 -d" " | xxd -p -r | base64 -w0)
export KAT_HASH

envsubst < "$SRC_DIR/cca-claims.json.template" > "$BUILD_DIR/cca-claims.json"

evcli cca create --iak="$SRC_DIR/ec.p256.jwk" --rak="$SRC_DIR/ec.p384.jwk" \
	--claims="$BUILD_DIR/cca-claims.json" --token="$BUILD_DIR/cca-token.cbor"

#     a2         -- map(2)
#       63       -- tstr(3)                      63       -- tstr(3)
#         6b6174 -- "kat"                          706174 -- "pat"
echo "a2636b6174$(xxd -p < "$BUILD_DIR/kat.cbor")63706174$(xxd -p < "$BUILD_DIR/cca-token.cbor")" | \
    xxd -p -r > evidence.cbor
