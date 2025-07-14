#!/bin/bash

set -o pipefail
set -eu

A=${A?must be set in the environment to one of rv or ta}

base64url_encode() {
  if [ "$(uname)" == "Darwin" ]
  then
    _base64="base64"
  else
    _base64="base64 -w0"
  fi

  ${_base64} | tr '+/' '-_' | tr -d '=';
}

# ref-value query
# impl-id: 7f454c4602010100000000000000000003003e00010000005058000000000000 
function rv_query() {
cat << EOF | diag2cbor.rb | base64url_encode
{
  / profile / 0: "tag:arm.com,2023:cca_platform#1.0.0",
  / query /   1: {
    / artifact-type /         0: 2, / reference-values /
    / environment-selector /  1: {
      / class / 0: [ [
        {
          / class-id /  0: 600(h'7f454c4602010100000000000000000003003e00010000005058000000000000')  / tagged-impl-id-type /
        }
      ] ]
    },
    / timestamp /   2: 0("2030-12-01T18:30:01Z"),
    / result-type / 3: 0 / collected material /
  }
}
EOF
}

# ta query
# inst-id: 0107060504030201000f0e0d0c0b0a090817161514131211101f1e1d1c1b1a1918
function ta_query() {
cat << EOF | diag2cbor.rb | base64url_encode
{
  / profile / 0: "tag:arm.com,2023:cca_platform#1.0.0",
  / query /   1: {
    / artifact-type /         0: 1, / trust-anchors /
    / environment-selector /  1: {
      / instance / 1: [
        [ 550(h'0107060504030201000f0e0d0c0b0a090817161514131211101f1e1d1c1b1a1918') ] / UEID /
      ]
    },
    / timestamp /   2: 0("2030-12-01T18:30:01Z"),
    / result-type / 3: 0 / collected material /
  }
}
EOF
}

if [ "${A}" == "rv" ]; then
  q=$(rv_query)
elif [ "${A}" == "ta" ]; then
  q=$(ta_query)
fi

curl https://localhost:11443/endorsement-distribution/v1/coserv/$q -s \
  --insecure \
  --header 'Accept: application/coserv+cbor; profile="tag:arm.com,2023:cca_platform#1.0.0"' \
  | cbor-edn cbor2diag
