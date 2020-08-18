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
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mdblp/gatekeeper/portal"
	"github.com/mdblp/gatekeeper/server/common"
)

type (
	permission       map[string]interface{}
	permissions      map[string]permission
	usersPermissions map[string]permissions

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

// Init the API v0 HTTP handlers
func (api *API) Init(mux *mux.Router, apiStatus func(http.ResponseWriter, *http.Request)) {
	mux.HandleFunc("/access/status", api.status(apiStatus)).Methods(http.MethodGet)
	// List of users sharing data with one subject
	mux.HandleFunc("/access/groups/{userID}", api.b.RequestLogger(api.clinicToWhomIHaveAccessTo)).Methods(http.MethodGet)
	mux.HandleFunc("/access/{userID}", api.b.RequestLogger(api.patientShares)).Methods(http.MethodGet)
	mux.HandleFunc("/access/{groupID}/{userID}", api.b.RequestLogger(api.userInGroupOf)).Methods(http.MethodGet)
	mux.HandleFunc("/access/{groupID}/{userID}", api.b.RequestLogger(api.invalidRoute)).Methods(http.MethodPost)
	// /access/:userid/:granteeid

	// List of users one subject is sharing data with
	// "/access/{userid}" "GET"
	// Check whether one subject is sharing data with one other user
	// "/access/{userid}/{granteeid}" GET
	// Assign permission to one user to view subject's data
	// "/access/{userid}/{granteeid}" POST
}

// @Summary Get the api status
// @Description Get the api status
// @ID gatekeeper-get-access-status
// @Produce json
// @Success 200 {object} common.APIStatus
// @Failure 500 {object} common.APIStatus
// @Router /access/status [get]
func (api *API) status(apiStatus func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return apiStatus
}

// FIXME how to match all other routes?
func (api *API) invalidRoute(w http.ResponseWriter, r *http.Request) int {
	api.b.Logger.Printf("Invalid route %s %s", r.Method, r.URL.RequestURI())
	w.WriteHeader(http.StatusNotImplemented)
	return http.StatusNotImplemented
}

// @Summary List of users sharing data with one subject
// @ID gatekeeper-get-access-group-userid
// @Security TidepoolAuth
// @Produce json
// @Success 200 {object} usersPermissions
// @Failure 400 {object} portal.APIFailure
// @Failure 403 {object} portal.APIFailure
// @Failure 500 {string} Internal Server Error
// @Failure 503 {string} Service Unavailable
// @Router /access/groups/{userID} [get]
func (api *API) clinicToWhomIHaveAccessTo(w http.ResponseWriter, r *http.Request) int {
	vars := mux.Vars(r) // Decode route parameter
	userID := vars["userID"]

	portalClient := portal.New(api.b.Logger, api.b.PortalURL, api.b.ShorelineSecret)
	results, status, err := portalClient.ClinicalShares(w, r, userID)
	if err != nil {
		return status
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
		api.b.Logger.Printf("Failed to encode response to JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return http.StatusInternalServerError
	}

	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.Write(jsonResponse)
	return http.StatusOK
}

// @Summary Check whether one subject is sharing data with one other user
// @ID gatekeeper-api-v0-clinic-access-to-with-ids
// @Security TidepoolAuth
// @Produce json
// @Success 200 {object} usersPermissions
// @Failure 400 {object} portal.APIFailure
// @Failure 403 {object} portal.APIFailure
// @Failure 500 {string} Internal Server Error
// @Failure 503 {string} Service Unavailable
// @Router /access/{groupID}/{userID} [get]
func (api *API) userInGroupOf(w http.ResponseWriter, r *http.Request) int {
	var status int
	vars := mux.Vars(r) // Decode route parameter
	groupID := vars["groupID"]
	userID := vars["userID"]
	// api.b.Logger.Printf("TOTO userInGroupOf: groupID{%s} userID{%s}", vars["groupID"], vars["userID"])

	portalClient := portal.New(api.b.Logger, api.b.PortalURL, api.b.ShorelineSecret)
	results, status, err := portalClient.ClinicalShares(w, r, userID)
	if err != nil {
		return status
	}

	perm := permissions{}
	found := false
	for _, result := range results {
		for _, member := range result.Members {
			if member.UserID == groupID {
				perm = permissions{
					"root": permission{},
					"vew":  permission{},
				}
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	jsonResponse, err := json.Marshal(perm)
	if err != nil {
		api.b.Logger.Printf("Failed to encode response to JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return http.StatusInternalServerError
	}

	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.Write(jsonResponse)
	return http.StatusOK
}

func (api *API) patientShares(w http.ResponseWriter, r *http.Request) int {
	var status int
	vars := mux.Vars(r) // Decode route parameter
	userID := vars["userID"]

	portalClient := portal.New(api.b.Logger, api.b.PortalURL, api.b.ShorelineSecret)
	results, status, err := portalClient.ClinicalShares(w, r, userID)
	if err != nil {
		return status
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
		api.b.Logger.Printf("Failed to encode response to JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return http.StatusInternalServerError
	}

	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.Write(jsonResponse)
	return http.StatusOK
}
