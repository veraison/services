# Copyright 2021-2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
#
# variables:
# * GOLINT_ARGS - command line arguments for $(LINT) when run on the lint target
# targets:
# * lint  - run source code linter

GOLINT_ARGS ?= run --timeout=3m -E dupl -E gocritic -E gosimple -E prealloc

GOLINT_VERSION = v1.64.8
GOLINT = $(TOPDIR)/tools-bin/golangci-lint
GOLINT_STAMP = $(TOPDIR)/tools-bin/golangci-lint-$(GOLINT_VERSION).stamp

$(GOLINT): $(GOLINT_STAMP)

$(GOLINT_STAMP):
	mkdir -p $(dir $(GOLINT))
	touch $(GOLINT_STAMP)
	GOBIN=$(dir $(GOLINT)) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLINT_VERSION)

.PHONY: lint
lint: $(GOLINT) lint-hook-pre reallint

.PHONY: lint-hook-pre
lint-hook-pre:

.PHONY: reallint
reallint: ; $(GOLINT) $(GOLINT_ARGS)

