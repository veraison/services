#!/bin/bash
# Copyright 2022-2024 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

set -eu
set -o pipefail

# function generate_go_test_vector constructs CBOR test vector using
# supplied comid and corim json template and saves them in a file
# $1 file name for comid json template, example one of COMID_TEMPLATES
# $2 file name for corim json template, example CORIM_TEMPLATE
# $3 a qualifier for each cbor test vector name
# $4 name of the file where the generated CBOR test vectors are aggregated
generate_go_test_vector () {
	echo "generating test vector using $1 $2"
	cocli comid create -t $1.json
	cocli corim create -m $1.cbor -t $2.json -o corim$1.cbor
	echo "// automatically generated from:" >> $4
	echo "// $1.json and $2.json" >> $4
	echo "var $3$2$1 = "'`' >> $4
	cat corim$1.cbor | xxd -p >> $4
	echo '`' >> $4
}

CORIM_REALM_TEMPLATE="corimCcaRealm"

COMID_REALM_TEMPLATES=
COMID_REALM_TEMPLATES="${COMID_REALM_TEMPLATES} comidCcaRealm"
COMID_REALM_TEMPLATES="${COMID_REALM_TEMPLATES} comidCcaRealmNoClass"
COMID_REALM_TEMPLATES="${COMID_REALM_TEMPLATES} comidCcaRealmNoInstance"
COMID_REALM_TEMPLATES="${COMID_REALM_TEMPLATES} comidCcaRealmInvalidInstance"
COMID_REALM_TEMPLATES="${COMID_REALM_TEMPLATES} comidCcaRealmInvalidClass"

# CORIM CCA PLATFORM TEMPLATES
CORIM_PLATFORM_TEMPLATE="corimCca"
CORIM_PLATFORM_TEMPLATE="${CORIM_PLATFORM_TEMPLATE} corimCcaNoProfile"

# COMID CCA PLATFORM TEMPLATES
COMID_PLATFORM_TEMPLATES=
COMID_PLATFORM_TEMPLATES="${COMID_PLATFORM_TEMPLATES} comidCcaRefValOne"
COMID_PLATFORM_TEMPLATES="${COMID_PLATFORM_TEMPLATES} comidCcaRefValFour"

TV_DOT_GO=${TV_DOT_GO?must be set in the environment.}

printf "package cca\n\n" > ${TV_DOT_GO}

# function to generate test vectors for the supplied CCA Platform or Realm 
# $1 passed argument whose templates needs to be constructed
generate_templates() {

	echo "generating templates for subattester $1"
	printf "" >> ${TV_DOT_GO}

	if [ "$1" == "cca_platform" ]; then
		COMID_TEMPLATES=$COMID_PLATFORM_TEMPLATES
		CORIM_TEMPLATE=$CORIM_PLATFORM_TEMPLATE
	else
		COMID_TEMPLATES=$COMID_REALM_TEMPLATES
		CORIM_TEMPLATE=$CORIM_REALM_TEMPLATE
	fi
	
	for r in ${CORIM_TEMPLATE}
	do
		for t in ${COMID_TEMPLATES}
		do
			generate_go_test_vector $t $r "unsigned" $TV_DOT_GO
		done
	done

}

SUBATTESTER=
SUBATTESTER="${SUBATTESTER} cca_platform"
SUBATTESTER="${SUBATTESTER} cca_realm" 

for at in ${SUBATTESTER}
do
	generate_templates $at
done

gofmt -w $TV_DOT_GO