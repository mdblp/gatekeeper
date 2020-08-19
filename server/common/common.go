// Package common : for Gatekeeper APIs
/*
 * Gatekeeper for Yourloops - Authorizations management
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
package common

import (
	"log"
	"net/http"
	"net/url"
	"time"
)

type (
	// Base API structure
	Base struct {
		Logger          *log.Logger
		PortalURL       *url.URL
		ShorelineSecret string
	}

	// APIStatus for /status route
	APIStatus struct {
		Status  string `json:"status"`
		Version string `json:"version"`
	}

	// // API common interface
	// API interface {
	// 	Init(b *Base) bool
	// }
)

// RequestLogger middleware (finalware?) to log received requests
func (a *Base) RequestLogger(fn func(w http.ResponseWriter, r *http.Request) int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now().UTC()
		status := fn(w, r)
		end := time.Now().UTC()
		dur := end.Sub(start).Milliseconds()
		a.Logger.Printf("%s - %s %s HTTP/%d.%d %d - %d ms", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.ProtoMajor, r.ProtoMinor, status, dur)
	}
}
