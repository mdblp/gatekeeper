// Package v1 authz api
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
	"time"

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

is_our_request {
	input.request.path == "/data/v1/ranges"
	common.is_get(input.request)
	common.have_session_token(input.request)
}

# If we can see our data
allow {
	is_our_request
	common.claims.svr != "yes"
	common.claims.roles[_] == "patient"
	not common.have_user_ids
}
# Else if we ask for others users
set_our_groups := { g | g := data.users[common.claims.usr].groupsList[_] }

result_by_users[user_id] = allowed{
	is_our_request
	common.have_user_ids
	user_id := input.userIds[_]
	set_patient_groups := {g | g = data.users[user_id].groupsList[_]}
	have_common_group := count(set_patient_groups & set_our_groups) > 0
	is_patient := data.users[user_id].roles[_] == "patient"
	allowed := have_common_group == is_patient
}

allow {
	# We have at least one userId which can be see
	result_by_users[_] == true
}

results["allow"] = allow
results["by_users"] = result_by_users
results["claims"] = common.claims
# results["users"] = data.users
`

var staticRules = map[string]string{
	"get.data.v1.ranges": staticRuleGetDataV1Ranges,
}

func (a *API) fetchUserGroupsUpdate() {
	a.b.Logger.Print("Updating users & groups from portal-api: started")
	const nRetry = 4
	for i := 0; i < nRetry; i++ {
		if i > 0 {
			time.Sleep(time.Duration(3*i) * time.Second)
		}
		opaGroupsAsJSON, err := a.b.PortalClient.OpaGroups()
		if err != nil {
			a.b.Logger.Printf("Updating users & groups from portal-api: Failed to fetch data: %s", err)
			continue
		}

		var opaGroups map[string]interface{}
		err = json.Unmarshal(opaGroupsAsJSON, &opaGroups)
		if err != nil {
			a.b.Logger.Printf("Updating users & groups from portal-api: Failed to parse json: %v", err)
			continue
		}

		a.usersAndGroups = inmem.NewFromObject(opaGroups)
		a.b.Logger.Print("Updating users & groups from portal-api: done")
		return
	}

	a.b.Logger.Printf("Updating users & groups from portal-api: failed %d times, giving up", nRetry)
}

func (a *API) initRego() bool {
	var err error
	var b bytes.Buffer
	var module *ast.Module
	var modules = make(map[string]*ast.Module)

	// Default empty users & groups
	opaGroups := make(map[string]interface{}, 2)
	opaGroups["users"] = make(map[string]interface{}, 0)
	opaGroups["groups"] = make(map[string]interface{}, 0)
	a.usersAndGroups = inmem.NewFromObject(opaGroups)

	go a.fetchUserGroupsUpdate()

	a.ctx = context.Background()

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

	a.compiler = ast.NewCompiler()
	a.compiler.Compile(modules)

	if a.compiler.Failed() {
		a.b.Logger.Printf("Failed to compile modules: %v", a.compiler.Errors)
		return false
	}

	return true
}
