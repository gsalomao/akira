// Copyright 2023 Gustavo Salomao
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/gsalomao/akira"
	"github.com/gsalomao/akira/listener"
)

func main() {
	server, err := akira.NewServer(akira.NewDefaultOptions())
	if err != nil {
		log.Fatal("Failed to create server")
	}
	defer server.Close()

	tcp := listener.NewTCPListener("tcp", ":1883", nil)
	defer func() { _ = tcp.Close() }()

	err = server.AddListener(tcp)
	if err != nil {
		log.Fatal("Failed to add TCP listener into the server")
	}

	err = server.AddHook(&loggingHook{})
	if err != nil {
		log.Fatal("Failed to add logging hook into the server")
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		<-signals
	}()

	err = server.Start(context.Background())
	if err != nil {
		log.Fatal("Failed to start server")
	}

	wg.Wait()
	_ = server.Stop(context.Background())
}

type loggingHook struct {
}

func (h *loggingHook) Name() string {
	return "logging"
}

func (h *loggingHook) OnServerStart(_ *akira.Server) error {
	fmt.Println("Server starting")
	return nil
}

func (h *loggingHook) OnServerStarted(_ *akira.Server) {
	fmt.Println("Server started")
}

func (h *loggingHook) OnServerStartFailed(_ *akira.Server, err error) {
	fmt.Println("Failed to start server: " + err.Error())
}

func (h *loggingHook) OnServerStop(_ *akira.Server) {
	fmt.Println("Server stopping")
}

func (h *loggingHook) OnServerStopped(_ *akira.Server) {
	fmt.Println("Server stopped")
}

func (h *loggingHook) OnConnectionOpen(_ *akira.Server, _ akira.Listener) error {
	fmt.Println("Client open")
	return nil
}

func (h *loggingHook) OnConnectionOpened(_ *akira.Server, _ akira.Listener) {
	fmt.Println("Client opened")
}

func (h *loggingHook) OnConnectionClose(_ *akira.Server, _ akira.Listener) {
	fmt.Println("Client close")
}

func (h *loggingHook) OnConnectionClosed(_ *akira.Server, _ akira.Listener) {
	fmt.Println("Client closed")
}
