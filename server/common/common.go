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

	"github.com/mdblp/gatekeeper/portal"
)

type (
	// Base API structure
	Base struct {
		Logger          *log.Logger
		PortalURL       *url.URL
		ShorelineSecret string
		PortalClient    *portal.Client
	}

	// APIStatus for /status route
	APIStatus struct {
		Status  string `json:"status"`
		Version string `json:"version"`
	}

	httpResponseWriter struct {
		w          http.ResponseWriter
		statusCode *int
	}
)

// RequestLoggerFunc type to simplify func signatures
type RequestLoggerFunc func(http.HandlerFunc) http.HandlerFunc

func (hrw httpResponseWriter) Header() http.Header {
	return hrw.w.Header()
}

func (hrw httpResponseWriter) Write(v []byte) (int, error) {
	return hrw.w.Write(v)
}

func (hrw httpResponseWriter) WriteHeader(statusCode int) {
	*hrw.statusCode = statusCode
	hrw.w.WriteHeader(statusCode)
}

// RequestLogger middleware (finalware?) to log received requests
func (a *Base) RequestLogger(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		statusCode := http.StatusOK
		hrw := httpResponseWriter{
			w:          w,
			statusCode: &statusCode,
		}

		start := time.Now().UTC()
		fn(hrw, r)
		end := time.Now().UTC()
		dur := end.Sub(start).Milliseconds()
		a.Logger.Printf("%s - %s %s HTTP/%d.%d %d - %d ms", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.ProtoMajor, r.ProtoMinor, statusCode, dur)
	}
}
