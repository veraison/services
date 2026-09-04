#!/bin/bash
# Copyright 2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

set -o pipefail
set -eu

A=${A?must be set to one of rv or ta or rim}
O=${O?output file}

# ref-value ELM query
# impl-id: 0000000000000000000000000000000000000000000000000000000000000000 
function rv_query() {
cat << EOF | diag2cbor.rb
{
  / profile /              0: "tag:arm.com,2025:cca_platform#1.0.0",
  / artifact-type /        1: 2, / reference-values /
  / environment-selector / 2: {
    / class / 0: [ [
        {
          / class-id /  0: 560(h'0000000000000000000000000000000000000000000000000000000000000000')  / tagged-bytes /
        }
    ] ]
  }
}
EOF
}

# ta ELM query
# inst-id: 010202020202020202020202020202020202020202020202020202020202020202
function ta_query() {
cat << EOF | diag2cbor.rb
{
  / profile /              0: "tag:arm.com,2025:cca_platform#1.0.0",
  / artifact-type /        1: 1, / trust-anchors /
  / environment-selector / 2: {
    / instance / 1: [ 
      [ 550( h'010202020202020202020202020202020202020202020202020202020202020202' ) ] / UEID /
    ]
  }
}
EOF
}

# rim ELM query
# corim-id: <input>
function rim_query() {
cat << EOF | diag2cbor.rb
{
  / rim-id selector / 3: [
    [ / corim / 2, / CoRIM id / "${C}" ]
  ]
}
EOF
}

if [ "${A}" == "rv" ]; then
  rv_query > ${O}
elif [ "${A}" == "ta" ]; then
  ta_query > ${O}
elif [ "${A}" = "rim" ]; then
  C=${C?C (corim-id) must be set when A=rim}
  rim_query > ${O}
fi
