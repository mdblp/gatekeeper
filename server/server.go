/*
 * Gatekeeper for Yourloops - Authorizations management
 * HTTP Server
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

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	mux "github.com/gorilla/mux"
	"github.com/mdblp/gatekeeper/server/common"
	v0 "github.com/mdblp/gatekeeper/server/v0"
	v1 "github.com/mdblp/gatekeeper/server/v1"
)

// Config HTTP server configuration
type Config struct {
	Host *string
	Port int
	// TLS cert file
	CertFile *string
	// TLS key file
	KeyFile *string
	// See http.Transport
	DisableCompression bool
	MaxIdleConns       int
	PortalURL          *url.URL
	ShorelineSecret    string
}

// Server needed infos
type Server struct {
	httpServer *http.Server
	tls        bool
	config     *Config
	logger     *log.Logger
}

// BaseAPI infos for all API versions
// type BaseAPI struct {
// 	logger *log.Logger
// }

// NewConfig init a server config
func NewConfig() *Config {
	return &Config{
		Port:               9123,
		DisableCompression: false,
		MaxIdleConns:       10,
	}
}

// NewServer Init a new server
func NewServer(config *Config, logger *log.Logger) (*Server, error) {
	tls := false
	if config.CertFile != nil || config.KeyFile != nil {
		if config.CertFile == nil || config.KeyFile == nil {
			return nil, fmt.Errorf("Certfile and Keyfile must be both set to use HTTPS")
		}
		tls = true
	}
	var addr string
	if config.Host != nil {
		addr = *config.Host
	}
	addr = addr + ":" + strconv.Itoa(config.Port)
	httpServer := &http.Server{
		Addr: addr,
	}

	return &Server{
		httpServer: httpServer,
		tls:        tls,
		config:     config,
		logger:     logger,
	}, nil
}

// Start the http(s) server
func (srv *Server) Start() error {
	srv.logger.Printf("Starting the server on %s", srv.httpServer.Addr)
	if !srv.setRouter() {
		return fmt.Errorf("Failed to init routes")
	}

	if srv.tls {
		return srv.httpServer.ListenAndServeTLS(*srv.config.CertFile, *srv.config.KeyFile)
	}
	return srv.httpServer.ListenAndServe()
}

// Stop (gracefully) the http server
func (srv *Server) Stop() {
	srv.logger.Printf("Stopping the server")
	err := srv.httpServer.Shutdown(context.Background())
	if err != nil {
		srv.logger.Fatalf("Failed to gracefully stop the server: %v", err)
	}
}

func (srv *Server) setRouter() bool {
	router := mux.NewRouter()
	srv.httpServer.Handler = router
	base := &common.Base{
		Logger:          srv.logger,
		PortalURL:       srv.config.PortalURL,
		ShorelineSecret: srv.config.ShorelineSecret,
	}

	apiStatus := base.RequestLogger(srv.status)
	router.HandleFunc("/status", apiStatus)

	apiV0 := v0.New(base)
	apiV1 := v1.New(base)

	if !apiV1.Init(router) {
		return false
	}
	apiV0.Init(router, apiStatus)

	mux.NewRouter().MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		return true
	}).HandlerFunc(srv.notFound)

	return true
}

// WaitOSSignals to stop the server
func (srv *Server) WaitOSSignals(done chan bool) {
	srv.logger.Printf("Waiting for OS signal to stop\n")

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	for {
		<-sigc
		srv.Stop()
		done <- true
	}
}

// @Summary Get the api status
// @Description Get the api status
// @ID gatekeeper-get-status
// @Produce json
// @Success 200 {object} common.APIStatus
// @Failure 500 {object} common.APIStatus
// @Router /status [get]
func (srv *Server) status(w http.ResponseWriter, r *http.Request) int {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")

	srv.logger.Printf("GET %s\n", r.URL.String())

	jStatus := common.APIStatus{
		Status:  "OK",
		Version: "0.0.0",
	}
	res, err := json.Marshal(jStatus)
	if err != nil {
		srv.logger.Printf("Failed to create JSON for %s: %v", r.URL.String(), err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"status\": \"KO\", \"version\": \"0.0.0\"}"))
		return http.StatusInternalServerError
	}

	w.WriteHeader(200)
	w.Write(res)
	return http.StatusOK
}

func (srv *Server) notFound(w http.ResponseWriter, r *http.Request) {
	srv.logger.Printf("Invalid route %s %s", r.Method, r.URL.RequestURI())
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not Found"))
}
