# Copyright 2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

SUBDIR := plugins
SUBDIR += trustedservices
SUBDIR += pluginmanager
SUBDIR += policymanager
SUBDIR += cmd/vts-service

# Create directories for packaging (TODO: May be a better way to do this)
install:
	mkdir -p $(VTS_DEPLOY_PREFIX)$(BIN_DIR)
	mkdir -p $(VTS_DEPLOY_PREFIX)$(LOG_DIR)
	mkdir -p $(VTS_DEPLOY_PREFIX)$(PLUGIN_DIR)
	mkdir -p $(VTS_DEPLOY_PREFIX)/usr/share/veraison/stores
	install $(TOPDIR)/vts/cmd/vts-service/vts-service $(VTS_DEPLOY_PREFIX)$(BIN_DIR)/vts-service
	install -D $(TOPDIR)/vts/plugins/bin/* $(VTS_DEPLOY_PREFIX)$(PLUGIN_DIR)
	install $(TOPDIR)/vts/cmd/vts-service/config.yaml $(VTS_DEPLOY_PREFIX)/config.yaml
	install $(TOPDIR)/vts/test-harness/init-kvstores.sh $(VTS_DEPLOY_PREFIX)

include ../mk/subdir.mk