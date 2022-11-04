#!/bin/bash
# Copyright 2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

set -eu
set -o pipefail

CORIM_CCA_TEMPLATE=corimMini.json

COMID_TEMPLATES=
COMID_TEMPLATES="${COMID_TEMPLATES} ComidPsaIakPubOne"
COMID_TEMPLATES="${COMID_TEMPLATES} ComidPsaIakPubTwo"
COMID_TEMPLATES="${COMID_TEMPLATES} ComidPsaRefValOne"
COMID_TEMPLATES="${COMID_TEMPLATES} ComidPsaRefValThree"
COMID_TEMPLATES="${COMID_TEMPLATES} ComidPsaMultIak"
COMID_TEMPLATES="${COMID_TEMPLATES} ComidPsaRefValMultDigest"
COMID_TEMPLATES="${COMID_TEMPLATES} ComidPsaRefValOnlyMandIDAttr"
COMID_TEMPLATES="${COMID_TEMPLATES} ComidPsaRefValNoMkey"
COMID_TEMPLATES="${COMID_TEMPLATES} ComidPsaRefValNoImplID"
COMID_TEMPLATES="${COMID_TEMPLATES} ComidPsaIakPubNoUeID"
COMID_TEMPLATES="${COMID_TEMPLATES} ComidPsaIakPubNoImplID"

TV_DOT_GO=${TV_DOT_GO?must be set in the environment.}

printf "package main\n\n" > ${TV_DOT_GO}

generate_cbor () {
	echo "generating cbor using $1 $2"
	echo $1
	echo $2
	cocli comid create -t ${1}.json
	cocli corim create -m ${1}.cbor -t $2 -o corim${1}.cbor
	echo "// automatically generated from $t.json" >> ${TV_DOT_GO}
	echo "var ${3}${1} = "'`' >> ${TV_DOT_GO}
	cat corim${1}.cbor | xxd -p >> ${TV_DOT_GO}
	echo '`' >> ${TV_DOT_GO}
	gofmt -w ${TV_DOT_GO}
}

for t in ${COMID_TEMPLATES}
do
generate_cbor $t $CORIM_CCA_TEMPLATE "unsignedCorim"
done


CORIM_CCA_TEMPLATE=corimCca.json
CORIM_CCA_TEMPLATE1=corimCcaNoProfile.json

COMID_CCA_TEMPLATES=
COMID_CCA_TEMPLATES="${COMID_CCA_TEMPLATES} ComidCcaRefValOne"
COMID_CCA_TEMPLATES="${COMID_CCA_TEMPLATES} ComidCcaRefValFour"

for t in ${COMID_CCA_TEMPLATES}
do
generate_cbor $t $CORIM_CCA_TEMPLATE "unsignedCorim"
done

for t in ${COMID_CCA_TEMPLATES}
do
generate_cbor $t $CORIM_CCA_TEMPLATE1 "unsignedCorimnoprofile"
done