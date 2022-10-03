# Copyright 2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

SUBDIR := api
SUBDIR += verifier
SUBDIR += sessionmanager
SUBDIR += cmd/verification-service

# Create directories for packaging (TODO: May be a better way to do this)
install:
	mkdir -p $(VERIFICATION_DEPLOY_PREFIX)$(BIN_DIR)
	mkdir -p $(VERIFICATION_DEPLOY_PREFIX)$(LOG_DIR)
	mkdir -p $(VERIFICATION_DEPLOY_PREFIX)$(INPUT_FILE_DIR)
	install $(TOPDIR)/verification/cmd/verification-service/verification-service $(VERIFICATION_DEPLOY_PREFIX)$(BIN_DIR)/verification-service
	install $(TOPDIR)/verification/cmd/verification-service/config.yaml $(VERIFICATION_DEPLOY_PREFIX)/config.yaml
	install $(TOPDIR)/end-to-end/input/psa-claims-profile-2-without-nonce.json $(VERIFICATION_DEPLOY_PREFIX)$(INPUT_FILE_DIR)
	install $(TOPDIR)/end-to-end/input/psa-evidence.cbor $(VERIFICATION_DEPLOY_PREFIX)$(INPUT_FILE_DIR)
	install $(TOPDIR)/end-to-end/input/ec-p256.jwk $(VERIFICATION_DEPLOY_PREFIX)$(INPUT_FILE_DIR)

include ../mk/subdir.mk
