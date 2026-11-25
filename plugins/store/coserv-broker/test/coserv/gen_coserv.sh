#!/bin/bash
# Copyright 2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

set -o pipefail
set -eu

# ref-value query
# impl-id: 7f454c4602010100000000000000000003003e00010000005058000000000000 
diag2cbor.rb coserv_rv_result.diag > coserv_rv_result.cbor

# ta query
# inst-id: 0107060504030201000f0e0d0c0b0a090817161514131211101f1e1d1c1b1a1918
diag2cbor.rb coserv_ta_result.diag > coserv_ta_result.cbor
