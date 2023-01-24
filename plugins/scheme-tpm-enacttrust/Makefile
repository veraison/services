# Copyright 2021 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

PLUGIN := ../bin/scheme-tpm-enacttrust.plugin
GOPKG := github.com/veraison/services/plugins/scheme-tpm-enacttrust
SRCS := main.go

test-hook-pre:
	$(MAKE) -C test/

SUBDIR := test

include ../../mk/common.mk
include ../../mk/plugin.mk
include ../../mk/lint.mk
include ../../mk/test.mk
include ../../mk/subdir.mk
