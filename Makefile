# Copyright 2021 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

export TOPDIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))

SUBDIR += builtin
SUBDIR += config
SUBDIR += handler
SUBDIR += kvstore
SUBDIR += log
SUBDIR += plugin
SUBDIR += policy
SUBDIR += proto
SUBDIR += provisioning
SUBDIR += scheme
SUBDIR += trustedservices
SUBDIR += verification

COVERAGE_THRESHOLD := 60.0
# plugin coverage is low because it is mostly tested via plugin/test, a
# separate package (this is necessary due to to the nature of the code being
# tested. plugin/test coverage is low because it's purely test code).
IGNORE_COVERAGE += github.com/veraison/services/plugin
IGNORE_COVERAGE += github.com/veraison/services/plugin/test

include mk/cover.mk
include mk/subdir.mk
