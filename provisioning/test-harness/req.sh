#!/bin/bash

set -o pipefail
set -eux

T=${T?must be set in the environment to one of psa or tpm-enacttrust.}
B=${B?must be set in the environment to one of trustanchor or refvalue.}
P=${P?must be set in the environment for psa be to one of p1 or p2.}

if [ "${T}" == "psa" ]
then
	if [ "${P}" == "p1" ]; 
	then
		CORIM_FILE=corim-${T}-${P}-${B}.cbor
		CONTENT_TYPE="Content-Type: application/corim-unsigned+cbor; profile=PSA_IOT_PROFILE_1"
	elif [ "${P}" == "p2" ];
	then
		CORIM_FILE=corim-${T}-${P}-${B}.cbor
		CONTENT_TYPE="Content-Type: application/corim-unsigned+cbor; profile=http://arm.com/psa/2.0.0"
	else
		echo "unknown profile set for ${T}"
		exit
	fi
elif [ "${T}" == "tpm-enacttrust" ];
then
	CONTENT_TYPE="Content-Type: application/corim-unsigned+cbor; profile=http://enacttrust.com/veraison/1.0.0"
	CORIM_FILE=corim-${T}-${B}.cbor
else
	echo "unknown Type: Please set T=psa or T=tpm-enacttrust"
	exit
fi

echo "Value of CORIM FILE: = \n, $CORIM_FILE"
echo "Value of CONTENT TYPE: = \n, $CONTENT_TYPE"

curl --include \
	--data-binary "@${CORIM_FILE}" \
	--header "${CONTENT_TYPE}" \
	--header "Accept: application/vnd.veraison.provisioning-session+json" \
	--request POST \
	http://localhost:8888/endorsement-provisioning/v1/submit 
