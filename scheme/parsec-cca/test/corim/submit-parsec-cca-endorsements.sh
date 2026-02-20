#!/usr/bin/env bash
# Copyright 2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

cocli corim submit -i --corim-file corim-parsec-cca-valid.cbor --api-server https://localhost:9443/endorsement-provisioning/v1/submit --media-type='application/rim+cbor; profile="tag:github.com/parallaxsecond,2023-03-03:cca"' --auth=none

