# Copyright 2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
THISDIR := $(realpath $(dir $(lastword $(MAKEFILE_LIST))))
CHECK := $(shell $(THISDIR)/check-nvat.sh; echo $$?)
ifeq ($(CHECK),0)

.DEFAULT_GOAL := test

GOPKG := github.com/veraison/services/scheme/nvidia
SRCS := $(wildcard *.go)

SUBDIR += plugin

include ../../mk/common.mk
include ../../mk/lint.mk
include ../../mk/pkg.mk
include ../../mk/subdir.mk
include ../../mk/test.mk

else

all:
test:
lint:

$(info skipping nvidia scheme; could not find nvat.h in the system)

endif
