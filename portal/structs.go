// Package portal API client fo golang
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

// TeamContactInfo are the contacts for a team (phone, email)
type TeamContactInfo struct {
	Value   string `json:"value" bson:"value"`
	Comment string `json:"comment,omitempty" bson:"comment"`
}

// Team in portal
type Team struct {
	ID           string            `json:"_id,omitempty" bson:"_id"`                           //  MongoDB _id: ObjectID (auto generated)
	UserID       string            `json:"userId" bson:"userId"`                               // Creator userId
	Code         string            `json:"code,omitempty" bson:"code"`                         // Random generated 9 digit code to performs invitations (auto generated)
	Name         string            `json:"name,omitempty" bson:"name,omitempty"`               // Name for clinic and trials teams
	Type         string            `json:"type" bson:"type"`                                   // Type of team: clinic, trials, personal
	Emails       []TeamContactInfo `json:"emails,omitempty" bson:"emails,omitempty"`           // emails to contact the team
	Phones       []TeamContactInfo `json:"phones,omitempty" bson:"phones,omitempty"`           // phone numbers to contact the team
	Description  string            `json:"description,omitempty" bson:"description,omitempty"` // optional description
	CreatedTime  string            `json:"createdTime,omitempty" bson:"createdTime"`
	ModifiedTime string            `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
}

// Member of a tem
type Member struct {
	ID           string `json:"_id,omitempty" bson:"_id"` // MongoDB _id: ObjectID (auto generated)
	UserID       string `json:"userId" bson:"userId"`     // Member user ID
	TeamID       string `json:"teamId" bson:"teamId"`     // Team id (MongoDB _id: ObjectID)
	Role         string `json:"role" bson:"role"`         // "patient" | "member" | "admin"
	CreatedTime  string `json:"createdTime,omitempty" bson:"createdTime"`
	ModifiedTime string `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
}

// Group definition for OPA
type Group struct {
	Type string `json:"group"`
}

// UserGroup a group definition for a user (a team member)
type UserGroup struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

// User definition for OPA
type User struct {
	Roles  []string    `json:"roles"`
	Groups []UserGroup `json:"groups"`
}

// OPAUsersAndGroups user & groups for OPA rules
type OPAUsersAndGroups struct {
	Groups map[string]Group `json:"groups"`
	Users  map[string]User  `json:"users"`
}

// APIFailure returned when an error occur
type APIFailure struct {
	Message string `json:"message"`
}
