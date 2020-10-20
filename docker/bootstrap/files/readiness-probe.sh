#!/bin/bash
#
# Copyright 2019 Orange
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

# Add Jolokia credentials, if defined in environment.
if [[ -z $JOLOKIA_USER ]] || [[ -z $JOLOKIA_PASSWORD ]]; then
    USER_OPT=""
else
    USER_OPT="--user $JOLOKIA_USER:$JOLOKIA_PASSWORD"
fi

# We check when the node is up and in normal state
CURL="/opt/bin/curl $USER_OPT -s --connect-timeout 0.5"
BASE_CMD="http://$POD_IP:8778/jolokia/read/org.apache.cassandra.db:type=StorageService"

if $CURL ${BASE_CMD}/LiveNodes | grep -q $POD_IP; then
  [[ $DEBUG ]] && echo Up
  exit 0
fi

[[ $DEBUG ]] && echo Not Up
exit 1
