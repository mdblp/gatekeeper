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
	"bytes"
	"context"
	"encoding/json"
	"text/template"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/storage/inmem"
)

type inputQuery struct {
	// UserIDs is the specific userIDs of the request,
	// omit if the path do not have any, user the one in the token
	UserIDs []string `json:"userIds,omitempty"`
	Request struct {
		Method string `json:"method"`
		Path   string `json:"path"`
		Token  string `json:"token"`
	} `json:"request"`
}

const staticRulesCommon = `
package common
default have_user_ids = false

have_user_ids = is_array(input.userIds)

is_get(r) = {
	r.method == "GET"
}
is_post(r) = true {
	r.method == "POST"
}
is_owner(user_ids, claims) = true {
	count(user_ids) == 0
	claims.svr == "no"
}
have_session_token(r) = true {
	is_string(r.token)
	r.token != ""
}
claims = payload {
	have_session_token(input.request)
	io.jwt.verify_hs256(input.request.token, "{{.ShorelineSecret}}")
	[_, payload, _] := io.jwt.decode(input.request.token)
}
`

const staticRuleGetDataV1Ranges = `
package get.data.v1.ranges
import data.common
default allow = false
claims = common.claims
is_our_request {
	input.request.path == "/data/v1/ranges"
	common.is_get(input.request)
	common.have_session_token(input.request)
}
allow {
	is_our_request
	common.claims["svr"] != "yes"
	common.claims["roles"][_] == "patient"
	not common.have_user_ids
}
results[user_id] = allowed {
	user_id := input.userIds[_]
	# Target user must be a patient
	is_patient := data.users[user_id].roles[_] == "patient"
	# We must have a common group

	allowed = {
		is_patient
	}
}
groups = data.groups
users = data.users
`

// groups := data.groups
// users := data.users

const staticRulesAll = `
package authz.all

import data.common
claims = common.claims
import data.get.data.v1.ranges

default allow = false
allow {
	data.get.data.v1.ranges.allow
}
`

var staticRules = map[string]string{
	"all":                staticRulesAll,
	"get.data.v1.ranges": staticRuleGetDataV1Ranges,
	// "patient":   modelAuthTokenPatient,
	// "clinic":    modelAuthTokenClinic,
}

func (a *API) initRego() bool {
	var err error
	var b bytes.Buffer
	var module *ast.Module
	var modules = make(map[string]*ast.Module)

	a.ctx = context.TODO()

	opaGroupsAsJSON, err := a.b.PortalClient.OpaGroups()
	if err != nil {
		a.b.Logger.Print("Failed to fetch from portal-api")
		a.b.Logger.Print(err)
		return false
	}

	var opaGroups map[string]interface{}
	err = json.Unmarshal(opaGroupsAsJSON, &opaGroups)
	if err != nil {
		a.b.Logger.Printf("Failed to parse json: %v", err)
		return false
	}

	tpl, err := template.New("rego").Parse(staticRulesCommon)
	if err != nil {
		a.b.Logger.Printf("Parse staticRulesCommon template error: %s", err)
		return false
	}
	tpl.Execute(&b, a.b)

	staticRules["common"] = b.String()

	for name, rules := range staticRules {
		module, err = ast.ParseModule(name, rules)
		if err != nil {
			a.b.Logger.Printf("Failed to parse module %s", name)
			a.b.Logger.Print(err)
			return false
		}
		modules[name] = module
	}

	// values, err := ast.InterfaceToValue(opaGroups)
	// if err != nil {
	// 	fmt.Print("Failed to convert portal JSON to OPA value")
	// 	fmt.Print(err)
	// 	return false
	// }
	// a.term = ast.NewTerm(values)
	// fmt.Print(a.term.String())
	// fmt.Print(a.term.Vars().String())

	a.usersAndGroups = inmem.NewFromObject(opaGroups)

	a.compiler = ast.NewCompiler()
	a.compiler.Compile(modules)
	if a.compiler.Failed() {
		a.b.Logger.Printf("Failed to compile modules: %v", a.compiler.Errors)
		return false
	}

	return true
}
