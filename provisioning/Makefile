# Copyright 2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

SUBDIR := decoder
SUBDIR += api
SUBDIR += cmd/provisioning-service

# Create directories for packaging (TODO: May be a better way to do this)
install:
	mkdir -p $(PROVISIONING_DEPLOY_PREFIX)$(BIN_DIR)
	mkdir -p $(PROVISIONING_DEPLOY_PREFIX)$(LOG_DIR)
	mkdir -p $(PROVISIONING_DEPLOY_PREFIX)$(PLUGIN_DIR)
	mkdir -p $(PROVISIONING_DEPLOY_PREFIX)$(INPUT_FILE_DIR)
	install $(TOPDIR)/provisioning/cmd/provisioning-service/provisioning-service $(PROVISIONING_DEPLOY_PREFIX)$(BIN_DIR)/provisioning-service
	if [[ "x$(COMBINED_PLUGINS)" == "x" ]]; then \
		install -D $(TOPDIR)/scheme/bin/*-endorsement-decoder.plugin $(PROVISIONING_DEPLOY_PREFIX)$(PLUGIN_DIR); \
	else \
		install -D $(TOPDIR)/scheme/bin/*.plugin $(PROVISIONING_DEPLOY_PREFIX)$(PLUGIN_DIR); \
	fi
	install $(TOPDIR)/provisioning/cmd/provisioning-service/config.yaml $(PROVISIONING_DEPLOY_PREFIX)/config.yaml

include ../mk/subdir.mk
