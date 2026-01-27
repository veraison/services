#!/usr/bin/env bash
# Copyright 2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

cocli corim submit --corim-file corim-cca-platform-valid.cbor --api-server http://localhost:9443/endorsement-provisioning/v1/submit --media-type='application/rim+cbor; profile="http://arm.com/cca/ssd/1"' --auth=none
cocli corim submit --corim-file corim-cca-realm-valid.cbor --api-server http://localhost:9443/endorsement-provisioning/v1/submit --media-type='application/rim+cbor; profile="http://arm.com/cca/realm/1"' --auth=none

