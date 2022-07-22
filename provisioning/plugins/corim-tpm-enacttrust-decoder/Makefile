# Copyright 2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

PLUGIN := ../bin/veraison-provisining-decoder-corim-tpm-enacttrust
GOPKG := github.com/veraison/services/provisioning/plugins/corim-tpm-enacttrust-decoder
SRCS := $(wildcard *.go)

all-hook-pre all-test-pre all-lint-pre:
	$(MAKE) -C ../../../proto protogen
	$(MAKE) -C ../../decoder protogen

include ../../../mk/common.mk
include ../../../mk/plugin.mk
include ../../../mk/lint.mk
include ../../../mk/test.mk
