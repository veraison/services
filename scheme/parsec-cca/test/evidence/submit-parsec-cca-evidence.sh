#!/usr/bin/env bash
# Copyright 2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail

EVIDENCE_FILE=evidence.cbor
EVIDENCE_CONTENT_TYPE=application/vnd.parallaxsecond.key-attestation.cca

THIS_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

session_path=$(curl -k -X POST -D -  https://localhost:8443/challenge-response/v1/newSession?nonce=byTWuWNaLIu_WOkIuU4Ewb-zroDN6-gyQkV4SZ_jF2Hn9eHYvOASGET1Sr36UobaiPU6ZXsVM1yTlrQyklS8XA== 2>/dev/null | grep "Location:" | cut -f2 -d" " | tr -d '\r')
echo "session: $session_path"

set +e
echo "----> post"
curl -k -X POST -D - -H "content-type: $EVIDENCE_CONTENT_TYPE" --data-binary @"$THIS_DIR/$EVIDENCE_FILE" https://localhost:8443/challenge-response/v1/"$session_path"
echo ""
echo ""
set -e

echo "----> delete $session_path"
curl -k -X DELETE -D - https://localhost:8443/challenge-response/v1/"$session_path"
echo "done."
