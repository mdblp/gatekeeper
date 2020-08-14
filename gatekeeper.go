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

package main

import (
	"log"
	"os"

	"github.com/mdblp/gatekeeper/server"
)

func main() {
	logger := log.New(os.Stdout, "gatekeeper:", log.LstdFlags|log.LUTC|log.Lshortfile)
	logger.Print("Starting service")

	serverConfig := server.NewConfig()
	httpServer, err := server.NewServer(serverConfig, logger)
	if err != nil {
		logger.Fatalf("%v", err)
	}

	done := make(chan bool)
	go httpServer.WaitOSSignals(done)

	httpServer.Start()
	// if err != nil {
	// 	logger.Fatalf("Failed to start the server: %v", err)
	// }
	<-done

	logger.Print("Service stopped")
}
