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
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	v0 "github.com/mdblp/gatekeeper/server/v0"

	mux "github.com/gorilla/mux"
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
	PortalURL          string
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
	// handler := &apiHandler{
	// 	logger: srv.logger,
	// } // srv.httpServer.Handler
	// srv.mux = http.NewServeMux()
	// srv.apiHandleV0(handler)
	srv.logger.Printf("Starting the server on %s", srv.httpServer.Addr)
	srv.setRouter()

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

func (srv *Server) setRouter() {
	mux := mux.NewRouter()
	srv.httpServer.Handler = mux
	v0 := v0.New(srv.logger, srv.config.PortalURL)
	v0.Init(mux)
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
