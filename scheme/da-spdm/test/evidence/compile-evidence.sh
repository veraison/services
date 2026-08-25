#!/usr/bin/env bash
# Copyright 2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail

THIS_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
SCRIPT_DIR=$(realpath "$THIS_DIR/../../../../scripts")
SRC_DIR="$THIS_DIR/src"
CORIM_DIR=$(realpath "$THIS_DIR/../corim")

echo "Generating evidence tokens"
diag2cbor.rb < "$SRC_DIR/good.diag" > "$THIS_DIR/da-spdm-good.cbor"

echo "Generating test_vars.go"
"$SCRIPT_DIR/generate-test-vector-embeds" -o "$(realpath "$THIS_DIR/../../test_vars.go")" \
    -p da_spdm "$THIS_DIR"/da-*.cbor "$CORIM_DIR"/corim-*.cbor
