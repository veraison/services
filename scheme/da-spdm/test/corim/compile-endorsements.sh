#!/bin/bash
# Copyright 2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail

THIS_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
SCRIPT_DIR=$(realpath "$THIS_DIR/../../../../scripts")
SRC_DIR="$THIS_DIR/src"
EVIDENCE_DIR=$(realpath "$THIS_DIR/../evidence")

echo "Generating CoRIMs"
"$SCRIPT_DIR/generate-corims" "$SRC_DIR/corims.yaml"

echo "Generating test_vars.go"
"$SCRIPT_DIR/generate-test-vector-embeds" -o "$(realpath "$THIS_DIR/../../test_vars.go")" \
    -p da_spdm "$EVIDENCE_DIR"/da-*.cbor "$THIS_DIR"/corim-*.cbor
