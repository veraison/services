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
    0:"tag:github.com/veraison,2023:amd_kds_coserv_proxy#1.0.0",
    1:{
        0:1,
        1:{
            1:[
                [
                    560(h'a4751aa669180bf5fe715cbe7afd5dff88c8a474a81694b87cb0360e0f476b4bef2c7caca24ef6adb4e41e4f0de8cb3171c0ac440794e7e429e668dcf7816c1c'),
                    [
                        {0: 647, 1: {1: 552(15354741454542995460)}},
                        {0: 648, 1: {4: 560(h'19'), 5: h'ff'}},
                        {0: 649, 1: {4: 560(h'01'), 5: h'ff'}},
                        {0: 650, 1: {4: 560(h'01'), 5: h'ff'}}
                    ]
                ]
            ]
        },
        2:0("2025-07-13T13:55:06+01:00"),
        3:2
    }
}
EOF
}

q=$(rv_query)

curl https://localhost:11443/endorsement-distribution/v1/coserv/$q -s \
  --insecure \
  --header 'Accept: application/coserv+cbor; profile="tag:github.com/veraison,2023:amd_kds_coserv_proxy#1.0.0"' \
  | cbor-edn cbor2diag
