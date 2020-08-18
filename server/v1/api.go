// Package v1
/*
 * Gatekeeper for Yourloops - Authorizations management
 * Gatekeeper API v1
 *
 * Copyright 2020 Diabeloop
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1

import (
	"github.com/gorilla/mux"
	"github.com/mdblp/gatekeeper/server/common"
)

type (
	// API data
	API struct {
		b *common.Base
	}
)

// New Create a new API
func New(base *common.Base) *API {
	return &API{
		b: base,
	}
}

// Init the API v1 HTTP handlers
func (a *API) Init(mux *mux.Router) {
	// mux.HandleFunc("/")
}
