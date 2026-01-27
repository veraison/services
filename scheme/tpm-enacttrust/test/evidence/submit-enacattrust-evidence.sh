#!/usr/bin/env bash
# Copyright 2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail

session_path=$(curl -k -X POST -D -  https://localhost:8443/challenge-response/v1/newSession 2>/dev/null | grep "location:" | cut -f2 -d" " | tr -d '\r')
echo "session: $session_path"

set +e
echo "----> post"
curl -k -X POST -D - -H "content-type: application/vnd.enacttrust.tpm-evidence" --data-binary @basic.token https://localhost:8443/challenge-response/v1/"$session_path"
echo ""
echo ""
set -e

echo "----> delete $session_path"
curl -k -X DELETE -D - https://localhost:8443/challenge-response/v1/"$session_path"
echo "done."
