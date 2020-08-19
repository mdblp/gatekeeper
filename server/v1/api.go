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
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/mdblp/gatekeeper/server/common"
	"github.com/open-policy-agent/opa/rego"
)

type (
	// API data
	API struct {
		b         *common.Base
		ctx       context.Context
		portalAPI rego.PreparedEvalQuery
	}
)

const xRequestToken = "x-diabeloop-authorization"
const tokenSignMethod = "HS256"
const moduleCreateTeam = `
package teams.team.create

default allow = false
allow {
	input.userId != null
	trim_space(input.userId) != ""
	input.teamType == "personal"
	input.userRoles[_] == "patient"
}
allow {
	input.userId != null
	trim_space(input.userId) != ""
	input.teamType == "clinic"
	input.userRoles[_] == "clinic"
}
allow {
	input.userId != null
	trim_space(input.userId) != ""
	input.teamType == "trials"
	input.userRoles[_] == "admin"
}
`
const moduleUpdateTeam = `
package teams/team/update

default allow = false
allow {
	input.userId != null
	trim_space(input.userId) != ""
	input.teamType == "personal"
	input.userRoles[_] == "patient"
}
allow {
	input.userId != null
	trim_space(input.userId) != ""
	input.teamType == "clinic"
	input.userRoles[_] == "clinic"
}
allow {
	input.userId != null
	trim_space(input.userId) != ""
	input.teamType == "trials"
	input.userRoles[_] == "admin"
}
`

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
	v1Router.HandleFunc("/create-team", a.b.RequestLogger(a.canCreateTeam)).Methods(http.MethodGet)
	v1Router.HandleFunc("/update-team/{teamID}", a.b.RequestLogger(a.canUpdateTeam)).Methods(http.MethodGet)

	return true
}

func (a *API) initRego() bool {
	var err error
	a.ctx = context.Background()

	a.portalAPI, err = rego.New(
		rego.Query("allow = data.teams.team.create.allow"),
		rego.Module("teams.team.create", moduleCreateTeam),
	).PrepareForEval(a.ctx)

	if err != nil {
		a.b.Logger.Printf("Failed to create rego rules: %s", err)
		return false
	}
	return true
}

func (a *API) canCreateTeam(w http.ResponseWriter, r *http.Request) int {
	type payloadType struct {
		UserID    string   `json:"userId,omitempty"`
		UserRoles []string `json:"userRoles,omitempty"`
		TeamType  string   `json:"teamType,omitempty"`
		jwt.StandardClaims
	}

	request := r.Header.Get(xRequestToken)
	if request == "" {
		w.WriteHeader(http.StatusForbidden)
		return http.StatusForbidden
	}

	keyFunc := func(t *jwt.Token) (interface{}, error) {
		return []byte(a.b.ShorelineSecret), nil
	}

	jwtToken, err := jwt.ParseWithClaims(request, &payloadType{}, keyFunc)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return http.StatusBadRequest
	}
	if !jwtToken.Valid {
		w.WriteHeader(http.StatusBadRequest)
		return http.StatusBadRequest
	}
	if jwtToken.Method.Alg() != tokenSignMethod {
		w.WriteHeader(http.StatusForbidden)
		return http.StatusForbidden
	}

	payload := jwtToken.Claims.(*payloadType)
	if payload.Issuer != "portal-api" {
		w.WriteHeader(http.StatusBadRequest)
		return http.StatusBadRequest
	}
	a.b.Logger.Printf("Payload: %v", payload)

	results, err := a.portalAPI.Eval(a.ctx, rego.EvalInput(payload))
	a.b.Logger.Printf("canCreateTeam: %v", results)
	if err != nil {
		a.b.Logger.Printf("Failed eval query: %s", err)
		w.WriteHeader(http.StatusForbidden)
		return http.StatusForbidden
	}
	if len(results) == 0 {
		a.b.Logger.Printf("Result is empty")
		w.WriteHeader(http.StatusForbidden)
		return http.StatusForbidden
	}
	allow, ok := results[0].Bindings["allow"].(bool)
	if !ok {
		a.b.Logger.Printf("Result allow not found")
		w.WriteHeader(http.StatusForbidden)
		return http.StatusForbidden
	}
	if !allow {
		a.b.Logger.Printf("Not authorized")
		w.WriteHeader(http.StatusUnauthorized)
		return http.StatusUnauthorized
	}

	return http.StatusOK
}

func (a *API) canUpdateTeam(w http.ResponseWriter, r *http.Request) int {
	return http.StatusOK
}
