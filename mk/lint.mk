# Copyright 2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

TOPDIR ?= $(CURDIR)

LINT_TAGS_FLAG = $(if $(BUILD_TAGS_CSV), --build-tags=$(BUILD_TAGS_CSV))

GOLINT_VERSION = v2.14.0
GOLINT = $(TOPDIR)/tools-bin/lint/$(GOLINT_VERSION)/golangci-lint
GOLINT_ARGS ?= run $(LINT_TAGS_FLAG)

.PHONY: lint
lint: $(GOLINT)
	$(GOLINT) $(GOLINT_ARGS) $(TOPDIR)/pkg/... $(TOPDIR)/cmd/... $(TOPDIR)/plugins/...

$(GOLINT): $(MOCKGEN)
	go generate $(TOPDIR)/pkg/... $(TOPDIR)/plugins/...
	GOBIN=$(@D) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLINT_VERSION)
