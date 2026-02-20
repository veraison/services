#!/usr/bin/env bash
# Copyright 2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
evcli psa create --allow-invalid --claims=psa.good.json --key=ec.p256.jwk --token=psa.good.cose
