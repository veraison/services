#!/usr/bin/env bash
# Copyright 2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

cocli corim submit --corim-file corim-enacttrust-valid.cbor --api-server http://localhost:9443/endorsement-provisioning/v1/submit --media-type='application/rim+cbor; profile="https://enacttrust.com/veraison/1.0.0"' --auth=none

