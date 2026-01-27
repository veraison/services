#!/bin/bash
# Copyright 2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail

THIS_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
SRC_DIR="$THIS_DIR/src"

diag2cbor.rb < "$SRC_DIR/evidence.diag" > "$THIS_DIR/evidence.cbor"
