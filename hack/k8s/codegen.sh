#!/bin/bash

# Copyright 2021 BlackRock, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

SCRIPT_ROOT=$(dirname $0)
source "${SCRIPT_ROOT}/library.sh"

# generate code for tinyapp
bash  "${CODEGEN_PKG}/generate-groups.sh" \
      "deepcopy,informer,client,lister" \
      "${PACKAGE_NAME}/pkg/k8s/client/tinyapp" \
      "${PACKAGE_NAME}/pkg/k8s/api" \
      "tinyapp:v1alpha1" \
      --go-header-file "${SCRIPT_ROOT}/boilerplate.go.txt" \
      --output-base ../../.. #GOTPATH/src/go.mod.stuff Good because if publish go mod onto brk artifactory, go mod needs path
