#!/bin/bash

set -o pipefail
set -eu

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
    0:"tag:github.com/veraison,2023:nvidia_coserv_proxy#1.0.0",
    1:{
        0:2,
        1:{
            0:[
                [
                    {
                        1:"NVIDIA",
                        2:"NV_NIC_FIRMWARE_CX7_28.39.4082-LTS_MCX713104AC-ADA"
                    }
                ]
            ]
        },
        2:0("2025-07-11T17:23:31+01:00"),
        3:2
    }
}
EOF
}

q=$(rv_query)

curl https://localhost:11443/endorsement-distribution/v1/coserv/$q -s \
  --insecure \
  --header 'Accept: application/coserv+cbor; profile="tag:github.com/veraison,2023:nvidia_coserv_proxy#1.0.0"' \
  | cbor-edn cbor2diag
