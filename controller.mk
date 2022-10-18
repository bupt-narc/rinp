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

include makefiles/consts.mk

# CLI entry file
ENTRY        := cmd/controller/main.go

# Binary targets that we support.
# When doing all-build, these targets will be built.
BIN_PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64
IMG_PLATFORMS := linux/amd64 linux/arm64

# Binary basename, without extension
BIN           := controller

# Docker image tag
IMGTAGS  ?= $(addsuffix /$(BIN):$(IMG_VERSION),$(REGISTRY))

include makefiles/common-targets.mk

manifests: # @HELP Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects
manifests: controller-gen
	$(CONTROLLER_GEN) rbac:roleName=rinp-controller-role crd webhook paths="./..." output:crd:artifacts:config=config/crd output:rbac:dir=config/controller

generate: # @HELP Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations
generate: controller-gen
	$(CONTROLLER_GEN) object paths="./..."
	go generate ./...

install: # @HELP Install CRDs into the K8s cluster specified in ~/.kube/config
install: manifests
	kubectl apply -f config/crd

uninstall: # @HELP Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion
uninstall: manifests
	kubectl delete --ignore-not-found -f config/crd

deploy: # @HELP Deploy controller to the K8s cluster specified in ~/.kube/config
deploy: manifests
	kubectl apply -f config/controller

undeploy: # @HELP Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion
undeploy:
	kubectl delete --ignore-not-found -f config/controller

# Location to install dependencies to
bin:
	mkdir -p bin

# Tool Binaries
CONTROLLER_GEN ?= bin/controller-gen

# Tool Versions
CONTROLLER_TOOLS_VERSION ?= v0.9.0

controller-gen: bin
	[ -f $(CONTROLLER_GEN) ] || GOBIN=$(PWD)/bin go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)
