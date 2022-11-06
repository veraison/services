# Copyright 2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL: all

diag2cbor ?= $(shell command -v diag2cbor.rb)
ifeq ($(strip $(diag2cbor)),)
  $(error diag2cbor.rb not found. To install diag2cbor.rb: 'gem install cbor-diag')
endif

%.cbor: %.diag ; $(diag2cbor) $< > $@

DIAG_FILES = $(wildcard *.diag)
CBOR_FILES = $(DIAG_FILES:.diag=.cbor)

all: $(CBOR_FILES)

clean: ; $(RM) $(CBOR_FILES)
