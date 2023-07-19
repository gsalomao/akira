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
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gsalomao/akira"
	"github.com/gsalomao/akira/listener"
	"github.com/gsalomao/akira/packet"
)

func main() {
	var exitCode int
	defer func() { os.Exit(exitCode) }()

	server, err := akira.NewServer(akira.NewDefaultOptions())
	if err != nil {
		log.Println("Failed to create server")
		exitCode = 1
		return
	}
	defer server.Close()

	tcp := listener.NewTCPListener("tcp", ":1883", nil)
	defer func() { _ = tcp.Close() }()

	err = server.AddListener(tcp)
	if err != nil {
		log.Println("Failed to add TCP listener into the server")
		exitCode = 1
		return
	}

	err = server.AddHook(&loggingHook{})
	if err != nil {
		log.Println("Failed to add logging hook into the server")
		exitCode = 1
		return
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		<-signals
	}()

	err = server.Start(context.Background())
	if err != nil {
		log.Println("Failed to start server")
		exitCode = 1
		return
	}

	wg.Wait()
	_ = server.Stop(context.Background())
}

type loggingHook struct{}

func (h *loggingHook) Name() string {
	return "logging"
}

func (h *loggingHook) OnServerStart(_ *akira.Server) error {
	log.Println("Server starting")
	return nil
}

func (h *loggingHook) OnServerStarted(_ *akira.Server) {
	log.Println("Server started")
}

func (h *loggingHook) OnServerStartFailed(_ *akira.Server, err error) {
	log.Println("Failed to start server: " + err.Error())
}

func (h *loggingHook) OnServerStop(_ *akira.Server) {
	log.Println("Server stopping")
}

func (h *loggingHook) OnServerStopped(_ *akira.Server) {
	log.Println("Server stopped")
}

func (h *loggingHook) OnConnectionOpen(_ *akira.Server, _ akira.Listener) error {
	log.Println("Client open")
	return nil
}

func (h *loggingHook) OnConnectionOpened(_ *akira.Server, _ akira.Listener) {
	log.Println("Client opened")
}

func (h *loggingHook) OnConnectionClose(_ *akira.Server, _ akira.Listener, _ error) {
	log.Println("Client close")
}

func (h *loggingHook) OnConnectionClosed(_ *akira.Server, _ akira.Listener, _ error) {
	log.Println("Client closed")
}

func (h *loggingHook) OnConnect(_ *akira.Client, _ *packet.Connect) error {
	log.Println("Client connecting")
	return nil
}

func (h *loggingHook) OnConnectError(_ *akira.Client, _ *packet.Connect, _ error) {
	log.Println("Connection failed")
}

func (h *loggingHook) OnConnected(_ *akira.Client) {
	log.Println("Client connected")
}
