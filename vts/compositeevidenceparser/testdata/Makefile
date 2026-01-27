# Copyright 2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

DIAG_TESTDATA := $(wildcard *.diag)
CBOR_TESTDATA := $(DIAG_TESTDATA:.diag=.cbor)

%.cbor: %.diag ; diag2cbor.rb $< > $@

CLEANFILES += $(CBOR_TESTDATA)

all: $(CBOR_TESTDATA)

clean: ; $(RM) -f $(CLEANFILES)
