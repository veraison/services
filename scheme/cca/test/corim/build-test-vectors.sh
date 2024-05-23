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
	cocli corim create -m $1.cbor -t $2 -o corim$1.cbor
	echo "// automatically generated from:" >> $4
	echo "// $1.json and $2" >> $4
	echo "var $3$1 = "'`' >> $4
	cat corim$1.cbor | xxd -p >> $4
	echo '`' >> $4
}

CORIM_TEMPLATE="corimCcaRealm.json"

COMID_TEMPLATES=
COMID_TEMPLATES="${COMID_TEMPLATES} comidCcaRealm"
COMID_TEMPLATES="${COMID_TEMPLATES} comidCcaRealmNoClass"
COMID_TEMPLATES="${COMID_TEMPLATES} comidCcaRealmNoInstance"
COMID_TEMPLATES="${COMID_TEMPLATES} comidCcaRealmInvalidInstance"
COMID_TEMPLATES="${COMID_TEMPLATES} comidCcaRealmInvalidClass"

TV_DOT_GO=${TV_DOT_GO?must be set in the environment.}

printf "package cca\n\n" > ${TV_DOT_GO}

for t in ${COMID_TEMPLATES}
do
	generate_go_test_vector $t $CORIM_TEMPLATE "unsignedCorim" $TV_DOT_GO
done

gofmt -w $TV_DOT_GO