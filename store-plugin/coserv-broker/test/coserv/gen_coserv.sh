#!/bin/bash
# Copyright 2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

set -o pipefail
set -eu

# ref-value query
# impl-id: 7f454c4602010100000000000000000003003e00010000005058000000000000 
function rv_result() {
cat << EOF | diag2cbor.rb
{
  / profile / 0: "tag:arm.com,2025:cca_platform#1.0.0",
  / query / 1: {
    / artifact-type / 0: 2 / reference-values / ,
    / environment-selector / 1: {
      / class / 0: [
        [
          {
           / class-id / 0: 560(h'7F454C4602010100000000000000000003003E00010000005058000000000000')
          }
        ]
      ]
    },
    / result-type / 2: 0 / collected values /
  },
  / results / 2: {
    / rvq / 0: [
      {
        / authorities / 1: [
          558({
            1: 2,
            3: -7,
            -1: 1,
            -2: h'CF4B4DB205072B54D90B7DBD8EAAF7DECFC6B9A7EAD640FE4A3015C258C107C6',
            -3: h'5C8AD59BE377FF25AC642B70C1A3D541AEABF12911741B724A8E97857674F852'
          })
        ],
        / rv-triple / 2: [
          {
            0: {
              0: 560(h'7F454C4602010100000000000000000003003E00010000005058000000000000')
            }
          },
          [
            {
              0: "cca.software-component",
              1: {
                2: [
                  [1, h'9A271F2A916B0B6EE6CECB2426F0B3206EF074578BE55D9BC94F6F3FE3AB86AA']
                ],
                11: "RSE_BL1_2"
              }
            },
            {
              0: "cca.software-component",
              1: {
                2: [
                  [1, h'53C234E5E8472B6AC51C1AE1CAB3FE06FAD053BEB8EBFD8977B010655BFDD3C3']
                ],
                11: "RSE_BL2"
              }
            },
            {
              0: "cca.software-component",
              1: {
                2: [
                  [1, h'1121CFCCD5913F0A63FEC40A6FFD44EA64F9DC135C66634BA001D10BCF4302A2']
                ],
                11: "RSE_S"
              }
            },
            {
              0: "cca.software-component",
              1: {
                2: [
                  [1, h'1571B5EC78BD68512BF7830BB6A2A44B2047C7DF57BCE79EB8A1C0E5BEA0A501']
                ],
                11: "AP_BL1"
              }
            },
            {
              0: "cca.software-component",
              1: {
                2: [
                  [1, h'10159BAF262B43A92D95DB59DAE1F72C645127301661E0A3CE4E38B295A97C58']
                ],
                11: "AP_BL2"
              }
            },
            {
              0: "cca.software-component",
              1: {
                2: [
                  [1, h'10122E856B3FCD49F063636317476149CB730A1AA1CFAAD818552B72F56D6F68']
                ],
                11: "SCP_BL1"
              }
            },
            {
              0: "cca.software-component",
              1: {
                2: [
                  [1, h'AA67A169B0BBA217AA0AA88A65346920C84C42447C36BA5F7EA65F422C1FE5D8']
                ],
                11: "SCP_BL2"
              }
            },
            {
              0: "cca.software-component",
              1: {
                2: [
                  [1, h'2E6D31A5983A91251BFAE5AEFA1C0A19D8BA3CF601D0E8A706B4CFA9661A6B8A']
                ],
                11: "AP_BL31"
              }
            },
            {
              0: "cca.software-component",
              1: {
                2: [
                  [1, h'A1FB50E6C86FAE1679EF3351296FD6713411A08CF8DD1790A4FD05FAE8688164']
                ],
                11: "RMM"
              }
            },
            {
              0: "cca.software-component",
              1: {
                2: [
                  [1, h'1A252402972F6057FA53CC172B52B9FFCA698E18311FACD0F3B06ECAAEF79E17']
                ],
                11: "HW_CONFIG"
              }
            },
            {
              0: "cca.software-component",
              1: {
                2: [
                  [1, h'9A92ADBC0CEE38EF658C71CE1B1BF8C65668F166BFB213644C895CCB1AD07A25']
                ],
                11: "FW_CONFIG"
              }
            },
            {
              0: "cca.software-component",
              1: {
                2: [
                  [1, h'238903180CC104EC2C5D8B3F20C5BC61B389EC0A967DF8CC208CDC7CD454174F']
                ],
                11: "TB_FW_CONFIG"
              }
            },
            {
              0: "cca.software-component",
              1: {
                2: [
                  [1, h'E6C21E8D260FE71882DEBDB339D2402A2CA7648529BC2303F48649BCE0380017']
                ],
                11: "SOC_FW_CONFIG"
              }
            },
            {
              0: "cca.platform-config",
              1: {
                4: 560(h'CFCFCFCF')
              }
            }
          ]
        ]
      }
    ],
    / expiry / 10: 0("2099-06-16T13:40:20+01:00")
  }
}
EOF
}

# ta query
# inst-id: 0107060504030201000f0e0d0c0b0a090817161514131211101f1e1d1c1b1a1918
function ta_result() {
cat << EOF | diag2cbor.rb
{
  / profile / 0: "tag:arm.com,2025:cca_platform#1.0.0",
  / query / 1: {
    / artifact-type / 0: 1 / trust-anchors /,
    / environment-selector / 1: {
      / instance / 1: [
        [
          550(h'0107060504030201000F0E0D0C0B0A090817161514131211101F1E1D1C1B1A1918')
        ]
      ]
    },
    / result-type / 2: 0 / collected material /
  },
  / results / 2: {
    / akq / 3: [
      {
        / authorities / 1: [
          558({
            1: 2,
            3: -7,
            -1: 1,
            -2: h'CF4B4DB205072B54D90B7DBD8EAAF7DECFC6B9A7EAD640FE4A3015C258C107C6',
            -3: h'5C8AD59BE377FF25AC642B70C1A3D541AEABF12911741B724A8E97857674F852'
          })
        ],
        / ak-triple / 2: [
          {
            0: {
              0: 560(h'7F454C4602010100000000000000000003003E00010000005058000000000000')
            },
            1: 550(h'0107060504030201000F0E0D0C0B0A090817161514131211101F1E1D1C1B1A1918')
          },
          [
            554("-----BEGIN PUBLIC KEY-----\nMHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEIShnxS4rlQiwpCCpBWDzlNLfqiG911FP\n8akBr+fh94uxHU5m+Kijivp2r2oxxN6MhM4tr8mWQli1P61xh3T0ViDREbF26DGO\nEYfbAjWjGNN7pZf+6A4OTHYqEryz6m7U\n-----END PUBLIC KEY-----\n")
          ]
        ]
      }
    ],
    / expiry / 10: 0("2099-06-16T13:39:18+01:00")
  }
}
EOF
}

ta_result > coserv_ta_result.cbor
rv_result > coserv_rv_result.cbor