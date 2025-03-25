# Copyright 2023 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL := all

GOPKG := github.com/veraison/services/coserv/cmd/coserv-service
CMD := coserv-service
SRCS := main.go

CMD_DEPS += $(wildcard ../../api/*.go)

include ../../../mk/common.mk
include ../../../mk/cmd.mk
include ../../../mk/test.mk
include ../../../mk/lint.mk
include ../../../mk/pkg.mk
