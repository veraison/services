#!/bin/bash
#
# Copyright 2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

set -x -e

KEY="${1:-key}"

openssl ecparam -genkey -name prime256v1 -noout -out "${KEY}.pem"
openssl ec -in "${KEY}.pem" -pubout -out "${KEY}.pub"
