#!/bin/bash -eu

. "${NVM_DIR}/nvm.sh"
. version.sh

service=${1:-'gatekeeper'}

nvm ls "${START_NODE_VERSION}" > /dev/null || { echo "ERROR: Node version ${START_NODE_VERSION} not installed"; exit 1; }
nvm use --delete-prefix "${START_NODE_VERSION}"

. config/env.sh

nohup node index.js > ../$service.log 2> ../$service.error.log <&- &
