/*
 * Gatekeeper for Yourloops - Authorizations management
 * Old Gatekeeper API
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

package v0

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// APIv0 data
type APIv0 struct {
	logger *log.Logger
}

// New Create a new APIv0
func New(logger *log.Logger) *APIv0 {
	return &APIv0{
		logger: logger,
	}
}

// Init the API v0 HTTP handlers
func (api *APIv0) Init(mux *mux.Router) {

	mux.HandleFunc("/access/status", api.status).Methods("GET")
	mux.HandleFunc("/access/groups", api.groups).Methods("GET")
}

func (api *APIv0) status(w http.ResponseWriter, r *http.Request) {
	type JSONStatus struct {
		status string
	}

	api.logger.Printf("GET %s\n", r.URL.String())

	jStatus := &JSONStatus{
		status: "OK",
	}
	res, err := json.Marshal(jStatus)
	if err != nil {
		api.logger.Printf("Failed to create /status JSON: %v", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(res)
}

func (api *APIv0) groups(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}
