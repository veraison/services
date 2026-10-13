# Copyright 2021-2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
#
# variables:
# * GOLINT_ARGS - command line arguments for $(LINT) when run on the lint target
# targets:
# * lint  - run source code linter

GOLINT_ARGS ?= run

# Current version run in CI
GOLINT_VERSION = v2.14.0
GOLINT = $(TOPDIR)/tools-bin/golangci-lint
GOLINT_STAMP = $(TOPDIR)/tools-bin/golangci-lint-$(GOLINT_VERSION).stamp

$(GOLINT): $(GOLINT_STAMP)

ifdef CI_PIPELINE
# Use golangci-lint from CI instead of installing it
$(GOLINT_STAMP):
	mkdir -p $(dir $(GOLINT))
	touch $(GOLINT_STAMP)
	ln -sf "$$(command -v golangci-lint)" $(GOLINT)
else
$(GOLINT_STAMP):
	mkdir -p $(dir $(GOLINT))
	GOBIN=$(abspath $(dir $(GOLINT))) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLINT_VERSION)
	touch $(GOLINT_STAMP)
endif

.PHONY: lint
lint: $(GOLINT) lint-hook-pre reallint

.PHONY: lint-hook-pre
lint-hook-pre:

.PHONY: reallint
reallint: ; $(GOLINT) $(GOLINT_ARGS)

