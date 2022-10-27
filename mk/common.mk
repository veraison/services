# Copyright 2021 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

SHELL := /bin/bash

# Pass this to sub-make
export GO111MODULE := on

# Used to set the ServerVersion reported by  services
VERSION_FROM_GIT := $(shell git describe --tags --exact-match 2>/dev/null || echo -n "commit-$(shell git rev-parse --revs-only --short HEAD)")

