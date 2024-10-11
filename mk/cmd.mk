# Copyright 2021 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
#
# variables:
# * CMD      - the name of the binary that is built
# * SRCS     - the source files that CMD is build from
# * CMD_DEPS - any extra dependency
#
# targets:
# * all      - build the binary and save it to $(CMD)
# * clean    - remove the generated binary


ifndef CMD
  $(error CMD must be set when including cmd.mk)
endif

SCHEME_LOADER ?= plugins

_MIN_GO_VERSION = 1.22
_GO_VERSION = $(shell go version | sed 's/^[^0-9]*\([0-9]\+\.[0-9]\+\.[0-9]\+\).*/\1/')

.PHONY: _check_version
_check_version:
	@if [[ "$(shell echo -e "$(_GO_VERSION)\n$(_MIN_GO_VERSION)" | sort -V | head -n 1)" != "$(_MIN_GO_VERSION)" ]]; then \
		echo -e "\n\tERROR: Please upgrade Go. Must be at least v$(_MIN_GO_VERSION) (found v$(_GO_VERSION)).\n"; \
		exit 1; \
	fi

.PHONY: _check_scheme_loader
_check_scheme_loader:
	@if [[ "$(SCHEME_LOADER)" != "plugins" && "$(SCHEME_LOADER)" != "builtin" ]]; then \
	    echo 'ERROR: invalid SCHEME_LOADER value: $(SCHEME_LOADER); ' \
	    	 'must be "plugins" or "builtin"'; \
	    exit 1; \
	fi

$(CMD): $(SRCS) $(CMD_DEPS) _check_scheme_loader _check_version
	CGO_ENABLED=1 go build -o $(CMD) -ldflags \
	"-X 'github.com/veraison/services/config.Version=$(VERSION_FROM_GIT)' \
	 -X 'github.com/veraison/services/config.SchemeLoader=$(SCHEME_LOADER)'"

CLEANFILES += $(CMD)

.PHONY: realall
realall: $(CMD)

.PHONY: cmd-hook-pre
cmd-hook-pre:

.PHONY: all
all: cmd-hook-pre realall

.PHONY: debug
debug:
	dlv debug --build-flags "-ldflags '-X github.com/veraison/services/config.SchemeLoader=builtin'"
