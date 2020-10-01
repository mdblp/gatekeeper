#!/bin/sh -e

wget -q -O artifact_node.sh 'https://raw.githubusercontent.com/mdblp/tools/feature/review_security_scan/artifact/artifact_node.sh'
chmod +x artifact_node.sh

. ./version.sh
./artifact_node.sh node
