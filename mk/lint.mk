# Copyright 2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
# Do NOT EDIT: synced from veraison/ci configs/mk/lint.mk

TOPDIR ?= $(CURDIR)

GOLINT_VERSION = v2.14.0
GOLINT = $(TOPDIR)/tools-bin/$(GOLINT_VERSION)/golangci-lint
GOLINT_ARGS ?= run
GOMODS ?= .

.PHONY: lint lint-hook-pre
lint: $(GOLINT) lint-hook-pre
	set -e; for mod in $(GOMODS); do (cd $$mod && $(GOLINT) $(GOLINT_ARGS)); done

# projects add prerequisites
lint-hook-pre:

$(GOLINT):
	GOBIN=$(@D) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLINT_VERSION)
