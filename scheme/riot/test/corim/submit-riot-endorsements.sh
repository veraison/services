#!/usr/bin/env bash
# Copyright 2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

cocli corim submit --corim-file corim-riot-valid.cbor --api-server http://localhost:9443/endorsement-provisioning/v1/submit --media-type='application/rim+cbor; profile="tag:veraison-project.com,2026:riot"' --auth=none

