# Copyright 2021 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

export TOPDIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))

SUBDIR += config
SUBDIR += kvstore
SUBDIR += log
SUBDIR += policy
SUBDIR += proto
SUBDIR += provisioning
SUBDIR += scheme
SUBDIR += verification
SUBDIR += vts
SUBDIR += vtsclient

include mk/subdir.mk
