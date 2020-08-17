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
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	"github.com/mdblp/gatekeeper/portal"
)

type (
	apiStatus struct {
		Status  string `json:"status"`
		Version string `json:"version"`
	}

	clinicWhomHaveAccessTo struct {
		Team    portal.Team     `json:"team"`
		Members []portal.Member `json:"members"`
	}

	permission       map[string]interface{}
	permissions      map[string]permission
	usersPermissions map[string]permissions
)

// APIv0 data
type APIv0 struct {
	logger    *log.Logger
	portalURL *url.URL
}

// XTidepoolSessionToken in the HTTP header
const XTidepoolSessionToken = "x-tidepool-session-token"

// XTidepoolTraceSession in the HTTP header
const XTidepoolTraceSession = "x-tidepool-trace-session"

// New Create a new APIv0
func New(logger *log.Logger, portalAPIHost string) *APIv0 {
	portalURL, err := url.Parse(portalAPIHost)
	if err != nil {
		logger.Fatalf("Invalid portal-api host")
	}
	return &APIv0{
		logger:    logger,
		portalURL: portalURL,
	}
}

// Init the API v0 HTTP handlers
func (api *APIv0) Init(mux *mux.Router) {
	mux.HandleFunc("/access/status", api.status).Methods("GET")
	// List of users sharing data with one subject
	mux.HandleFunc("/access/groups/{userID}", api.clinicToWhomIHaveAccessTo).Methods("GET")
	// List of users one subject is sharing data with
	// "/access/{userid}" "GET"
	// Check whether one subject is sharing data with one other user
	// "/access/{userid}/{granteeid}" GET
	// Assign permission to one user to view subject's data
	// "/access/{userid}/{granteeid}" POST
}

// @Summary Get the api status
// @Description Get the api status
// @ID gatekeeper-api-v0-getstatus
// @Produce json
// @Success 200 {object} apiStatus
// @Failure 500 {object} apiStatus
// @Router /status [get]
func (api *APIv0) status(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")

	api.logger.Printf("GET %s\n", r.URL.String())

	jStatus := apiStatus{
		Status:  "OK",
		Version: "0.0.0",
	}
	res, err := json.Marshal(jStatus)
	if err != nil {
		api.logger.Printf("Failed to create JSON for %s: %v", r.URL.String(), err)
		w.WriteHeader(500)
		w.Write([]byte("{\"status\": \"KO\", \"version\": \"0.0.0\"}"))
		return
	}

	w.WriteHeader(200)
	w.Write(res)
}

// @Summary List of users sharing data with one subject
// @ID gatekeeper-api-v0-clinic-access-to
// @Security TidepoolAuth
// @Produce json
// @Success 200 {object} usersPermissions
// @Failure 400 {object} portal.APIFailure
// @Failure 403 {object} portal.APIFailure
// @Failure 500 {string} Internal Server Error
// @Failure 503 {string} Service Unavailable
// @Router /access/groups/{userID} [get]
func (api *APIv0) clinicToWhomIHaveAccessTo(w http.ResponseWriter, r *http.Request) {
	start := time.Now().UTC()

	vars := mux.Vars(r) // Decode route parameter
	userID := vars["userID"]
	token := r.Header.Get(XTidepoolSessionToken)
	trace := r.Header.Get(XTidepoolTraceSession)

	if token == "" {
		apiFailure := portal.APIFailure{
			Message: "Missing token",
		}
		res, err := json.Marshal(apiFailure)
		if err != nil {
			w.Header().Add("Content-Type", "application/json; charset=utf-8")
			w.Write(res)
		}
		w.WriteHeader(403)
		return
	}

	portalURL := api.portalURL.String() + "/teams/v1/members/clinic-my-teams"
	request, err := http.NewRequest("GET", portalURL, nil)
	if err != nil {
		api.logger.Printf("Failed to create a new GET HTTP request: %v", err)
		w.WriteHeader(500)
		w.Write([]byte("Internal Server Error"))
		return
	}

	request.Header.Add(XTidepoolSessionToken, token)
	if trace != "" {
		// Forward the trace session id
		request.Header.Add(XTidepoolTraceSession, trace)
	}

	c := http.Client{}
	response, err := c.Do(request)
	if err != nil {
		api.logger.Printf("Failed to send the HTTP request: %v", err)
		w.WriteHeader(503)
		w.Write([]byte("Service Unavailable"))
		return
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		w.WriteHeader(response.StatusCode)
		body, err := ioutil.ReadAll(response.Body)
		if err == nil {
			w.Header().Add("Content-Type", "application/json; charset=utf-8")
			w.Write(body)
		}
		return
	}

	var results []clinicWhomHaveAccessTo
	if err = json.NewDecoder(response.Body).Decode(&results); err != nil {
		api.logger.Printf("Failed to parse portal-api response to JSON: %v", err)
		w.WriteHeader(500)
		w.Write([]byte("Internal Server Error"))
		return
	}

	perms := make(usersPermissions)
	perms[userID] = permissions{
		"root": permission{},
	}
	for _, result := range results {
		// teamID := result.Team.ID
		for _, member := range result.Members {
			if _, exists := perms[member.UserID]; exists == false {
				perms[member.UserID] = permissions{
					"node": permission{},
					"vew":  permission{},
				}
			}
		}
	}

	jsonResponse, err := json.Marshal(perms)
	if err != nil {
		api.logger.Printf("Failed to encode response to JSON: %v", err)
		w.WriteHeader(500)
		w.Write([]byte("Internal Server Error"))
		return
	}

	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.Write(jsonResponse)

	end := time.Now().UTC()
	dur := end.Sub(start).Milliseconds()
	api.logger.Printf("%s - %s %s HTTP/%d.%d 200 - %d ms", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.ProtoMajor, r.ProtoMinor, dur)
}
