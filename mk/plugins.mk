# Copyright 2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

TOPDIR ?= $(CURDIR)

PLUGINS_DIR := $(TOPDIR)/plugins
PLUGINS_BIN := $(PLUGINS_DIR)/bin

# Plugin names, discovered from plugins/<kind>/<name>/plugin/main.go
SCHEME_PLUGINS := $(patsubst $(PLUGINS_DIR)/scheme/%/plugin/main.go,%,\
                    $(wildcard $(PLUGINS_DIR)/scheme/*/plugin/main.go))
STORE_PLUGINS  := $(patsubst $(PLUGINS_DIR)/store/%/plugin/main.go,%,\
                    $(wildcard $(PLUGINS_DIR)/store/*/plugin/main.go))

ifeq ($(filter nvat,$(BUILD_TAGS)),)
SCHEME_PLUGINS := $(filter-out nvidia,$(SCHEME_PLUGINS))
endif

SCHEME_PLUGIN_BINS := $(foreach p,$(SCHEME_PLUGINS),$(PLUGINS_BIN)/scheme/scheme-$(p).plugin)
STORE_PLUGIN_BINS  := $(foreach p,$(STORE_PLUGINS),$(PLUGINS_BIN)/store/store-$(p).plugin)

SCHEME_PLUGIN_NAMES := $(addprefix scheme-,$(SCHEME_PLUGINS))
STORE_PLUGIN_NAMES  := $(addprefix store-,$(STORE_PLUGINS))

.PHONY: $(SCHEME_PLUGIN_NAMES) $(STORE_PLUGIN_NAMES)
$(SCHEME_PLUGIN_NAMES): scheme-%: $(PLUGINS_BIN)/scheme/scheme-%.plugin
$(STORE_PLUGIN_NAMES):  store-%:  $(PLUGINS_BIN)/store/store-%.plugin

.PHONY: plugins scheme-plugins store-plugins clean-plugins
plugins: scheme-plugins store-plugins
scheme-plugins: $(SCHEME_PLUGIN_BINS)
store-plugins:  $(STORE_PLUGIN_BINS)

$(PLUGINS_BIN)/scheme/scheme-%.plugin: ALWAYS
	@mkdir -p $(@D)
	go build $(GO_TAGS_FLAG) -o $@ $(PLUGINS_DIR)/scheme/$*/plugin

$(PLUGINS_BIN)/store/store-%.plugin: ALWAYS
	@mkdir -p $(@D)
	go build $(GO_TAGS_FLAG) -o $@ $(PLUGINS_DIR)/store/$*/plugin

.PHONY: clean-plugins
clean-plugins:
	rm -rf $(PLUGINS_BIN)

.PHONY: ALWAYS
ALWAYS:
