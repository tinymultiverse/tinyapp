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
VENDOR_ROOT="${SCRIPT_ROOT}/../../vendor"

# Install controller-gen to generate yaml definnitions for k8s resources
go install -mod=vendor ${VENDOR_ROOT}/sigs.k8s.io/controller-tools/cmd/controller-gen
go install -mod=vendor ${VENDOR_ROOT}/github.com/itchyny/gojq/cmd/gojq

PACKAGE_NAME=$(go mod edit -json | gojq -r '.Module.Path')
CODEGEN_PKG="${VENDOR_ROOT}/k8s.io/code-generator"