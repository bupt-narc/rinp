MAKEFLAGS += --no-print-directory
MAKEFLAGS += --always-make

# Use bash explicitly
SHELL := /usr/bin/env bash -o errexit -o pipefail -o nounset

BIN = $(wildcard *.mk)

all: $(addprefix mk-all_,$(BIN))

build: $(addprefix mk-build_,$(BIN))

all-build: $(addprefix mk-all-build_,$(BIN))

all-package: $(addprefix mk-all-package_,$(BIN))

all-docker-build-push: $(addprefix mk-all-docker-build-push_,$(BIN))

docker-build: $(addprefix mk-docker-build_,$(BIN))

docker-push: $(addprefix mk-docker-push_,$(BIN))

version: $(addprefix mk-version_,$(BIN))

imageversion: $(addprefix mk-imageversion_,$(BIN))

binary-name: $(addprefix mk-binary-name_,$(BIN))

variables: $(addprefix mk-variables_,$(BIN))

help: $(addprefix mk-help_,$(BIN))

mk-%:
	$(MAKE) -f $(lastword $(subst _, ,$*)) $(firstword $(subst _, ,$*))
