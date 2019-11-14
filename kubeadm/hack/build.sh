#!/usr/bin/env bash

# Copyright 2019 The Kubernetes Authors.
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

KUBE_ROOT="$(dirname "$0")"/..
source "${KUBE_ROOT}"/hack/lib/version.sh

pushd "${KUBE_ROOT}" > /dev/null
LDFLAGS=$(kube::version::ldflags)

COMMAND="go build -ldflags \"${LDFLAGS}\""
echo "${COMMAND}"
eval "${COMMAND}"
