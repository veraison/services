# Copyright 2021 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
#
# variables:
# * GOLINT - name of the linter package
# * GOLINT_ARGS - command line arguments for $(LINT) when run on the lint target
# targets:
# * lint  - run source code linter

GOLINT ?= golangci-lint

GOLINT_ARGS ?= run --timeout=3m -E dupl -E gocritic -E gosimple -E prealloc

.PHONY: lint
lint: lint-hook-pre reallint

.PHONY: lint-hook-pre
lint-hook-pre:

.PHONY: reallint
reallint: ; $(GOLINT) $(GOLINT_ARGS)

