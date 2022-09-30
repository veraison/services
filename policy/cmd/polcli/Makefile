# Copyright 2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL := all

GOPKG := github.com/veraison/services/policy/cmd/polcli
CMD := polcli
SRCS := main.go $(wildcard commands/*.go)

CMD_DEPS += $(wildcard ../../*.go)

cmd-hook-pre test-hook-pre lint-hook-pre:
	$(MAKE) -C ../../../proto protogen

include ../../../mk/common.mk
include ../../../mk/cmd.mk
include ../../../mk/test.mk
include ../../../mk/lint.mk
include ../../../mk/pkg.mk
