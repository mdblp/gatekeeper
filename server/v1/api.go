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
	"github.com/mdblp/gatekeeper/shoreline"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage"
	"github.com/open-policy-agent/opa/topdown"
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

		compiler       *ast.Compiler
		usersAndGroups storage.Store
	}
)

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
	v1Router.HandleFunc("/groups-update", a.b.RequestLogger(a.groupsUpdate)).Methods(http.MethodPost)
	v1Router.HandleFunc("/is-allowed", a.b.RequestLogger(a.isAllowedPut)).Methods(http.MethodPut)

	return true
}

func (a *API) groupsUpdate(w http.ResponseWriter, r *http.Request) {
	callerPackedToken := r.Header.Get(shoreline.XTidepoolSessionToken)
	if callerToken, err := shoreline.UnpackAndVerifyToken(callerPackedToken, a.b.ShorelineSecret); err != nil || callerToken.IsServer != "yes" {
		a.b.Logger.Print(err)
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	go a.fetchUserGroupsUpdate()

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Accepted"))
}

func (a *API) isAllowedPut(w http.ResponseWriter, r *http.Request) {
	var err error
	// Allow only server token to call this route
	callerPackedToken := r.Header.Get(shoreline.XTidepoolSessionToken)
	if callerToken, err := shoreline.UnpackAndVerifyToken(callerPackedToken, a.b.ShorelineSecret); err != nil || callerToken.IsServer != "yes" {
		a.b.Logger.Print(err)
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	serviceQuery := &inputQuery{}
	jsonDecoder := json.NewDecoder(r.Body)
	jsonDecoder.DisallowUnknownFields()
	err = jsonDecoder.Decode(&serviceQuery)
	if err != nil {
		a.b.Logger.Printf("Invalid serviceQuery received")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	buf := topdown.NewBufferTracer()

	query := "results = data."
	query = query + strings.ToLower(serviceQuery.Request.Method)
	query = query + strings.ReplaceAll(serviceQuery.Request.Path, "/", ".")
	query = query + ".results"

	regoQuery, err := rego.New(
		rego.Query(query),
		rego.Compiler(a.compiler),
		rego.Store(a.usersAndGroups),
		rego.QueryTracer(buf),
	).PrepareForEval(a.ctx)
	if err != nil {
		a.b.Logger.Printf("Failed to prepare query for evaluation: %s", err)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	regoEval := rego.EvalInput(serviceQuery)

	regoResult, err := regoQuery.Eval(a.ctx, regoEval)
	topdown.PrettyTraceWithLocation(a.b.Logger.Writer(), *buf)

	if err != nil {
		a.b.Logger.Printf("Failed eval query: %s", err)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	ret, err := json.Marshal(regoResult[0].Bindings["results"])
	if err != nil {
		a.b.Logger.Printf("Failed to marshall JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json; charset=utf-8")

	var results map[string]interface{}
	results = regoResult[0].Bindings["results"].(map[string]interface{})
	if !results["allow"].(bool) {
		w.WriteHeader(http.StatusUnauthorized)
	}

	w.Write(ret)
}
