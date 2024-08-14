#!/bin/bash
# Copyright 2022-2024 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

set -eu
set -o pipefail

THIS_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
GEN_CORIM="$THIS_DIR/../../../common/scripts/gen-corim"

CORIM_TEMPLATE=corimMini

COMID_TEMPLATES=(
	ComidPsaIakPubOne
	ComidPsaIakPubTwo
	ComidPsaRefValOne
	ComidPsaRefValThree
	ComidPsaMultIak
	ComidPsaRefValMultDigest
	ComidPsaRefValOnlyMandIDAttr
	ComidPsaRefValNoMkey
	ComidPsaRefValNoImplID
	ComidPsaIakPubNoUeID
	ComidPsaIakPubNoImplID
)

for comid in "${COMID_TEMPLATES[@]}"
do
	"$GEN_CORIM" "$THIS_DIR" "$comid" "$CORIM_TEMPLATE" "unsigned"
done

echo "done"
