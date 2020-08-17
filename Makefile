
# Copyright 2020 Diabeloop
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

GOPATH ?= ~/go
GO111MODULE = on
GOCC = go

DEPLOY_DOC = docs/soup

all: dist doc soup

dist: build
	mkdir dist
	mv gatekeeper dist/
	cp -a start.sh dist/

build: clean
	GOPATH=$(GOPATH) $(GOCC) mod tidy
	GOPATH=$(GOPATH) $(GOCC) build

doc: $(GOPATH)/bin/swag
	mkdir -p doc/openapi
	$(GOPATH)/bin/swag --version
	$(GOPATH)/bin/swag init --generalInfo gatekeeper.go --output doc/openapi

soup:
	mkdir -p doc/soup
	go list -f '## {{printf "%s \n\t* description: \n\t* version: %s\n\t* webSite: https://%s\n\t* sources:" .Path .Version .Path}}' -m all >> doc/soup/soup.md

clean:
	rm -f gatekeeper
	rm -rf dist
	rm -rf doc
	rm -rf soup/gatekeeper

test:
	GOPATH=$(GOPATH) $(GOCC) test
