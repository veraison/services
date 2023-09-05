##############################################################################
# CONFIG
# NOTE: DO NOT modify these here -- edit and source ../deployment.cfg instead.
##############################################################################

# The name of the Veraison docker network
VERAISON_NETWORK ?= veraison-net

# The ports on which services will be listening.
VTS_PORT ?= 50051
PROVISIONING_PORT ?= 8888
VERIFICATION_PORT ?= 8080
MANAGEMENT_PORT ?= 8088
KEYCLOAK_PORT ?= 11111

# Deploy destination is either an absolute path to a directory on the host, or
# the name of a docker volume.
DEPLOY_DEST ?= /tmp/veraison

# Top-level directory for docker image build contexts. This is used for
# "utility" images such as builder and tester. Service images use the
# DEPLOY_DEST as their build contex.
CONTEXT_DIR ?= /tmp/veraison-build-context

# Docker volume that will contain veraison logs.
LOGS_VOLUME ?= veraison-logs

# Docker volume that will contain VTS stores.
STORES_VOLUME ?= veraison-stores

# "docker run" flags used for the debug container
DEBUG_FLAGS ?=
# User for the debug session. Set to "root" to get more control over the container
DEBUG_USER ?= builder

##############################################################################

