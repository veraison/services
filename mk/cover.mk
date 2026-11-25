# Copyright 2023-2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

GO_TEST_COVERAGE_VERSION = v2.23.0
GO_TEST_COVERAGE = $(TOPDIR)/tools-bin/coverage/$(GO_TEST_COVERAGE_VERSION)/go-test-coverage

$(GO_TEST_COVERAGE):
	GOBIN=$(@D) go install github.com/vladopajic/go-test-coverage/v2@$(GO_TEST_COVERAGE_VERSION)

.PHONY: coverage cover test-cover
coverage cover test-cover $(TOPDIR)/coverage.out: $(GO_TEST_COVERAGE) $(MOCKGEN)
	@go generate $(TOPDIR)/pkg/... $(TOPDIR)/plugins/...
	@go test $(TOPDIR)/pkg/... $(TOPDIR)/plugins/... \
		-coverprofile=$(TOPDIR)/coverage.out -covermode=atomic -coverpkg=./... > /dev/null
	@$(GO_TEST_COVERAGE) --config=$(TOPDIR)/.testcoverage.yml --format=packages-above,packages-below

.PHONY: cover-report coverage-report report
cover-report coverage-report report: $(TOPDIR)/coverage.out
	go tool cover -html=coverage.out
