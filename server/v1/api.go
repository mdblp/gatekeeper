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
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/mdblp/gatekeeper/server/common"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage"
)

type (
	teamGroup struct {
		TeamType string   `json:"type"`
		UserID   string   `json:"userId"`
		ACLs     []string `json:"acls,omitempty"`
	}
	teamMember struct {
		Role   string `json:"role"`
		UserID string `json:"userId"`
	}

	userGroups map[string][]string

	// API data
	API struct {
		b   *common.Base
		ctx context.Context

		compiler *ast.Compiler
		// term           *ast.Term
		// usersAndGroups *portal.OPAUsersAndGroups
		usersAndGroups storage.Store
	}
)

// // ** Custom headers **

// // The route to test for the authorizations
// const xDiabeloopAuthzRoute = "x-diabeloop-authz-route"

// // The method (GET/POST/PUT) to test for the authorizations
// const xDiabeloopAuthzMethod = "x-diabeloop-authz-method"

// // The token to test for the authorizations
// const xDiabeloopAuthzToken = "x-diabeloop-authz-token"

// New Create a new API
func New(base *common.Base) *API {
	return &API{
		b: base,
	}
}

// Init the API v1 HTTP handlers
func (a *API) Init(baseRouter *mux.Router) bool {
	if !a.initRego() {
		return false
	}

	v1Router := baseRouter.PathPrefix("/authz/v1").Subrouter()
	// v1Router.HandleFunc("/portalapi")
	v1Router.HandleFunc("/groups-update", a.b.RequestLogger(a.groupsUpdate)).Methods(http.MethodPost)
	// v1Router.HandleFunc("/is-allowed", a.b.RequestLogger(a.isAllowedHead)).Methods(http.MethodHead)
	v1Router.HandleFunc("/is-allowed", a.b.RequestLogger(a.isAllowedPut)).Methods(http.MethodPut)
	// v1Router.HandleFunc("/auth-token/{type}", a.b.RequestLogger(a.authTokenVerification)).Methods(http.MethodGet)
	// v1Router.HandleFunc("/create-team", a.b.RequestLogger(a.canCreateTeam)).Methods(http.MethodGet)
	// v1Router.HandleFunc("/update-team/{teamID}", a.b.RequestLogger(a.canUpdateTeam)).Methods(http.MethodGet)

	return true
}

func (a *API) groupsUpdate(w http.ResponseWriter, r *http.Request) {
	// 	data := `{
	//   "d61b6ac828": ["5f3c373ba12c04079d754811", "5f3a94a7f147b0dc6e22d721", "5f39939921db5314d62362a5"]
	// }`
}

// func (a *API) isAllowedHead(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Add("Access-Control-Allow-Methods", http.MethodGet)
// 	w.Header().Add("Access-Control-Request-Method", http.MethodGet)
// 	headers := []string{
// 		shoreline.XTidepoolSessionToken,
// 		shoreline.XTidepoolTraceSession,
// 		xDiabeloopAuthzMethod,
// 		xDiabeloopAuthzRoute,
// 	}
// 	w.Header().Add("Access-Control-Allow-Headers", strings.Join(headers, ", "))
// 	w.WriteHeader(200)
// }

func (a *API) isAllowedPut(w http.ResponseWriter, r *http.Request) {
	var err error
	// Disabled for POC, simpler
	// callerPackedToken := r.Header.Get(shoreline.XTidepoolSessionToken)
	// if callerToken, err := shoreline.UnpackAndVerifyToken(callerPackedToken, a.b.ShorelineSecret); err != nil || callerToken.IsServer != "yes" {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }

	serviceQuery := &inputQuery{}
	jsonDecoder := json.NewDecoder(r.Body)
	jsonDecoder.DisallowUnknownFields()
	err = jsonDecoder.Decode(&serviceQuery)
	if err != nil {
		a.b.Logger.Printf("Invalid serviceQuery received")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	a.b.Logger.Printf("isAllowedPut serviceQuery: %v", serviceQuery)

	query := "result = data."
	query = query + strings.ToLower(serviceQuery.Request.Method)
	query = query + strings.ReplaceAll(serviceQuery.Request.Path, "/", ".")

	regoQuery, err := rego.New(
		rego.Query(query),
		rego.Compiler(a.compiler),
		rego.Store(a.usersAndGroups),
	).PrepareForEval(a.ctx)
	if err != nil {
		a.b.Logger.Printf("Failed to prepare query for evaluation: %s", err)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	regoEval := rego.EvalInput(serviceQuery)

	regoResult, err := regoQuery.Eval(a.ctx, regoEval)
	if err != nil {
		a.b.Logger.Printf("Failed eval query: %s", err)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	ret, err := json.Marshal(regoResult[0].Bindings["result"])
	if err != nil {
		a.b.Logger.Printf("Failed to marshall JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json; charset=utf-8")

	var results map[string]interface{}
	results = regoResult[0].Bindings["result"].(map[string]interface{})
	if !results["allow"].(bool) {
		w.WriteHeader(http.StatusUnauthorized)
	}

	w.Write(ret)
}

// func (a *API) authTokenVerification(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r) // Decode route parameter
// 	queryType := vars["type"]
// 	query, ok := a.regoAuth[queryType]
// 	if !ok {
// 		w.WriteHeader(http.StatusNotFound)
// 		return
// 	}
// }

// func (a *API) canCreateTeam(w http.ResponseWriter, r *http.Request) {
// 	type payloadType struct {
// 		UserID    string   `json:"userId,omitempty"`
// 		UserRoles []string `json:"userRoles,omitempty"`
// 		TeamType  string   `json:"teamType,omitempty"`
// 		jwt.StandardClaims
// 	}

// 	request := r.Header.Get(xRequestToken)
// 	if request == "" {
// 		w.WriteHeader(http.StatusForbidden)
// 		return
// 	}

// 	keyFunc := func(t *jwt.Token) (interface{}, error) {
// 		return []byte(a.b.ShorelineSecret), nil
// 	}

// 	jwtToken, err := jwt.ParseWithClaims(request, &payloadType{}, keyFunc)
// 	if err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}
// 	if !jwtToken.Valid {
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}
// 	if jwtToken.Method.Alg() != tokenSignMethod {
// 		w.WriteHeader(http.StatusForbidden)
// 		return
// 	}

// 	payload := jwtToken.Claims.(*payloadType)
// 	if payload.Issuer != "portal-api" {
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}
// 	a.b.Logger.Printf("Payload: %v", payload)

// 	results, err := a.portalAPI.Eval(a.ctx, rego.EvalInput(payload))
// 	a.b.Logger.Printf("canCreateTeam: %v", results)
// 	if err != nil {
// 		a.b.Logger.Printf("Failed eval query: %s", err)
// 		w.WriteHeader(http.StatusForbidden)
// 		return
// 	}
// 	if len(results) == 0 {
// 		a.b.Logger.Printf("Result is empty")
// 		w.WriteHeader(http.StatusForbidden)
// 		return
// 	}
// 	allow, ok := results[0].Bindings["allow"].(bool)
// 	if !ok {
// 		a.b.Logger.Printf("Result allow not found")
// 		w.WriteHeader(http.StatusForbidden)
// 		return
// 	}
// 	if !allow {
// 		a.b.Logger.Printf("Not authorized")
// 		w.WriteHeader(http.StatusUnauthorized)
// 		return
// 	}
// }

// func (a *API) canUpdateTeam(w http.ResponseWriter, r *http.Request) {
// 	w.WriteHeader(http.StatusNotImplemented)
// }
