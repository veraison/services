# Copyright 2021-2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

TOPDIR ?= $(CURDIR)

GO_TAGS_FLAG = $(if $(BUILD_TAGS_CSV),-tags=$(BUILD_TAGS_CSV))

MOCKGEN_VERSION = v1.7.0-rc.1
MOCKGEN = $(TOPDIR)/tools-bin/mockgen/$(MOCKGEN_VERSION)/mockgen

$(MOCKGEN):
	GOBIN=$(@D) go install github.com/golang/mock/mockgen@$(MOCKGEN_VERSION)

export PATH := $(dir $(MOCKGEN)):$(PATH)

ifdef CI_PIPELINE
COPYRIGHT_FLAGS += --no-year-check
endif

test: $(MOCKGEN) check-copyrights
	go generate $(TOPDIR)/pkg/... $(TOPDIR)/plugins/...
	go test $(GO_TAGS_FLAG) $(TOPDIR)/pkg/... $(TOPDIR)/plugins/...

.PHONY: check-copyrights
check-copyrights:
	python3 $(TOPDIR)/scripts/check-copyright $(COPYRIGHT_FLAGS) .

recursive_wildcard = $(foreach d,$(wildcard $(addsuffix /*,$(1))),\
              $(call recursive_wildcard,$(d),$(2)) \
              $(if $(filter $(subst *,%,$(2)),$(notdir $(d))),$(d)))

MOCK_FILES = $(strip $(call recursive_wildcard,pkg,mock_*.go))

.PHONY: clean-mocks
clean-mocks:
	rm -f $(MOCK_FILES)
