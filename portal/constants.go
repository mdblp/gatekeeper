/*
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

package portal

const (
	// DefaultPortalURL used to connect to portal
	DefaultPortalURL = "http://localhost:9507"

	// TeamTypeClinic a team type
	TeamTypeClinic = "clinic"
	// TeamTypeTrials a team type
	TeamTypeTrials = "trials"
	// TeamTypePersonal a team type
	TeamTypePersonal = "personal"
	// MemberRolePatient in a team
	MemberRolePatient = "patient"
	// MemberRoleMember in a team
	MemberRoleMember = "member"
	// MemberRoleAdmin in a team
	MemberRoleAdmin = "admin"
)

var (
	// TeamTypes as an array
	TeamTypes = [...]string{TeamTypeClinic, TeamTypeTrials, TeamTypePersonal}
	// MemberRoles as an array
	MemberRoles = [...]string{MemberRolePatient, MemberRoleMember, MemberRoleAdmin}
)
