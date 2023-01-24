# Copyright 2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

PLUGIN := ../bin/decoder-corim-psa.plugin
GOPKG := github.com/veraison/services/plugins/corim-psa-decoder
SRCS := $(wildcard *.go)

all-hook-pre all-test-pre all-lint-pre:
	$(MAKE) -C ../../proto protogen
	$(MAKE) -C ../../provisioning/decoder protogen

include ../../mk/common.mk
include ../../mk/plugin.mk
include ../../mk/lint.mk
include ../../mk/test.mk
