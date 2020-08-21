/*
 * Gatekeeper for Yourloops - Authorizations management
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

// @title Gatekeeper API
// @version 1.0.0
// @description The purpose of this API is to provide authorizations for end users and other tidepool Services
// @license.name BSD 2-Clause "Simplified" License
// @host localhost
// @BasePath /
// @accept json
// @produce json
// @schemes https
// @securityDefinitions.apikey TidepoolAuth
// @in header
// @name x-tidepool-session-token

package main

import (
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/mdblp/gatekeeper/server"
)

const (
	defaultPortalURL       = "http://localhost:9507"
	defaultShorelineSecret = "This is a local API secret for everyone. BsscSHqSHiwrBMJsEGqbvXiuIUPAjQXU"
	defaultServerSecret    = "This needs to be the same secret everywhere. YaHut75NsK1f9UKUXuWqxNN0RUwHFBCy"
)

func main() {
	var err error
	logger := log.New(os.Stdout, "gatekeeper:", log.LstdFlags|log.LUTC|log.Lshortfile)
	logger.Printf("Starting service: %v\n", os.Environ())

	serverConfig := server.NewConfig()
	serverConfig.ShorelineSecret = os.Getenv("SHORELINE_SECRET")
	if serverConfig.ShorelineSecret == "" {
		serverConfig.ShorelineSecret = defaultShorelineSecret
	}
	portalURL := os.Getenv("PORTAL_API_HOST")
	if !strings.HasPrefix(portalURL, "http") {
		// Default value
		portalURL = defaultPortalURL
	}
	serverConfig.PortalURL, err = url.Parse(portalURL)
	if err != nil {
		logger.Fatalf("Invalid portal-api host")
	}
	logger.Printf("Using portal-api url: %s\n", serverConfig.PortalURL.String())

	serverPort := os.Getenv("PORT")
	if serverPort != "" {
		if port, err := strconv.Atoi(serverPort); err == nil && port >= 80 && port < 65536 {
			serverConfig.Port = port
		} else {
			logger.Fatalf("Invalid PORT value: %s", serverPort)
		}
	}

	httpServer, err := server.NewServer(serverConfig, logger)
	if err != nil {
		logger.Fatalf("%v", err)
	}

	done := make(chan bool)
	go httpServer.WaitOSSignals(done)

	err = httpServer.Start()
	if err != nil {
		logger.Fatal(err)
	}
	<-done

	logger.Print("Service stopped")
}
