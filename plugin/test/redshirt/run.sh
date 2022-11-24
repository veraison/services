#!/bin/bash
# Copyright 2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
export VERAISON_PLUGIN=VERAISON
export PLUGIN_MIN_PORT=10000
export PLUGIN_MAX_PORT=25000
export PLUGIN_PROTOCOL_VERSIONS=1

dlv debug
