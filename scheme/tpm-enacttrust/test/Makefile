# Copyright 2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
#
OUTPUT := ../test_vectors.go

SUBDIR := cmd/gen-token/
SUBDIR += tokens/


DEPS := $(wildcard Comid*.json)

all: $(OUTPUT)

$(OUTPUT): $(DEPS)
	env TV_DOT_GO=$(OUTPUT) ./build-test-vectors.sh

clean: ; $(RM) -f *.cbor

include ../../../mk/subdir.mk
