#!/bin/bash
# Copyright 2022-2024 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

set -eu
set -o pipefail

THIS_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
GEN_CORIM="$THIS_DIR/../../../common/scripts/gen-corim"

SUBATTESTERS=(
	cca_platform
	cca_realm
)

CORIM_REALM_TEMPLATES=(
	corimCcaRealm
)

COMID_REALM_TEMPLATES=(
	comidCcaRealm
	comidCcaRealmNoClass
	comidCcaRealmNoInstance
	comidCcaRealmInvalidInstance
	comidCcaRealmInvalidClass
)

CORIM_PLATFORM_TEMPLATES=(
	corimCca
	corimCcaNoProfile
)

COMID_PLATFORM_TEMPLATES=(
	comidCcaRefValOne
	comidCcaRefValFour
)

# function to generate test vectors for the supplied CCA Platform or Realm
# $1 passed argument whose templates needs to be constructed
generate_templates() {
	local sub_at=$1

	echo "generating templates for subattester $sub_at"

	if [ "$sub_at" == "cca_platform" ]; then
		COMID_TEMPLATES=("${COMID_PLATFORM_TEMPLATES[@]}")
		CORIM_TEMPLATES=("${CORIM_PLATFORM_TEMPLATES[@]}")
	else
		COMID_TEMPLATES=("${COMID_REALM_TEMPLATES[@]}")
		CORIM_TEMPLATES=("${CORIM_REALM_TEMPLATES[@]}")
	fi

	for corim in "${CORIM_TEMPLATES[@]}"
	do
		for comid in "${COMID_TEMPLATES[@]}"
		do
			"$GEN_CORIM" "$THIS_DIR" "$comid" "$corim" "unsigned"
		done
	done

}

for at in "${SUBATTESTERS[@]}"
do
	generate_templates "$at"
done

echo "done"
