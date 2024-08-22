#!/bin/bash
# Copyright 2022-2023 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

set -o pipefail
set -eux

T=${T?must be set in the environment to one of psa, tpm-enacttrust, parsec-tpm.}
B=${B?must be set in the environment to one of trustanchor or refvalue.}

CORIM_FILE=corim-${T}-${B}.cbor
CORIM_FILE=corim-${T}-${B}.cbor

if [ "${T}" == "psa" ]
then
	CONTENT_TYPE='Content-Type: application/corim-unsigned+cbor; profile="http://arm.com/psa/iot/1"'
elif [ "${T}" == "tpm-enacttrust" ];
then
	CONTENT_TYPE='Content-Type: application/corim-unsigned+cbor; profile="http://enacttrust.com/veraison/1.0.0"'
elif [ "${T}" == "parsec-tpm" ]
then
	CONTENT_TYPE='Content-Type: application/corim-unsigned+cbor; profile="tag:github.com/parallaxsecond,2023-03-03:tpm"'
else
	echo "unknown type ${T}"
	exit 1
fi

curl --include \
	--insecure \
	--data-binary "@${CORIM_FILE}" \
	--header "${CONTENT_TYPE}" \
	--header "Accept: application/vnd.veraison.provisioning-session+json" \
	--request POST \
	https://localhost:8888/endorsement-provisioning/v1/submit
