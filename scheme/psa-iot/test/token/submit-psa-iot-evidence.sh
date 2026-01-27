#!/usr/bin/env bash
# Copyright 2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail

session_path=$(curl -X POST -D -  http://localhost:8443/challenge-response/v1/newSession?nonce=QUp8F0FBs9DpodKK8xUg8NQimf6sQAfe2J1ormzZLxk= 2>/dev/null | grep "Location:" | cut -f2 -d" " | tr -d '\r')
echo "session: $session_path"

set +e
echo "----> post"
curl -X POST -D - -H "Content-Type: application/psa-attestation-token" --data-binary @psa.good.cose http://localhost:8443/challenge-response/v1/"$session_path"
echo ""
echo ""
set -e

echo "----> delete $session_path"
curl -X DELETE -D - http://localhost:8443/challenge-response/v1/"$session_path"
echo "done."
