# Copyright 2022 The KubeVela Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Set this to 1 to enable debugging output.
DBG_MAKEFILE ?=
ifeq ($(DBG_MAKEFILE),1)
    $(warning ***** starting Makefile for goal(s) "$(MAKECMDGOALS)")
    $(warning ***** $(shell date))
else
    # If we're not debugging the Makefile, don't echo recipes.
    MAKEFLAGS += -s
endif

# No, we don't want builtin rules.
MAKEFLAGS += --no-builtin-rules
MAKEFLAGS += --warn-undefined-variables
# Get rid of .PHONY everywhere.
MAKEFLAGS += --always-make

# Use bash explicitly
SHELL := /usr/bin/env bash -o errexit -o pipefail -o nounset

# If user has not defined target, set some default value, same as host machine.
OS          := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
ARCH        := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))
# Use git tags to set the version string
VERSION     ?= $(shell git describe --tags --always --dirty)
IMG_VERSION ?= $(shell bash -c " \
if [[ ! $(VERSION) =~ ^v[0-9]{1,2}\.[0-9]{1,2}\.[0-9]{1,2}(-(alpha|beta)\.[0-9]{1,2})?$$ ]]; then \
  echo latest;     \
else               \
  echo $(VERSION); \
fi")

BIN_EXTENSION :=
ifeq ($(OS), windows)
    BIN_EXTENSION := .exe
endif

DBG_BUILD   ?=
FULL_NAME   ?=
GOFLAGS     ?=
GOPROXY     ?=

# Registry to push to
REGISTRY := rinp # ghcr.io/bupt-narc
