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
	server, err := akira.NewServer(akira.NewDefaultOptions())
	if err != nil {
		log.Println("Failed to create server")
		os.Exit(1)
	}
	defer server.Close()

	tcp := listener.NewTCPListener("tcp", ":1883", nil)
	defer func() { _ = tcp.Close() }()

	err = server.AddListener(tcp)
	if err != nil {
		log.Println("Failed to add TCP listener into the server")
		os.Exit(1)
	}

	err = server.AddHook(&loggingHook{})
	if err != nil {
		log.Println("Failed to add logging hook into the server")
		os.Exit(1)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		<-signals
	}()

	err = server.Start()
	if err != nil {
		log.Println("Failed to start server")
		os.Exit(1)
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
	log.Println("Connection opening")
	return nil
}

func (h *loggingHook) OnConnectionOpened(_ *akira.Server, _ akira.Listener) {
	log.Println("Connection opened")
}

func (h *loggingHook) OnConnectionClose(_ *akira.Server, _ akira.Listener, _ error) {
	log.Println("Connection closing")
}

func (h *loggingHook) OnConnectionClosed(_ *akira.Server, _ akira.Listener, _ error) {
	log.Println("Connection closed")
}

func (h *loggingHook) OnPacketReceived(c *akira.Client, p akira.Packet) error {
	if c.Connected() {
		log.Printf("Received packet '%s' from client '%s'", p.Type(), c.ID)
	} else {
		log.Printf("Received packet '%s'", p.Type())
	}
	return nil
}

func (h *loggingHook) OnConnect(_ *akira.Client, p *packet.Connect) error {
	log.Printf("Client '%s' connecting\n", p.ClientID)
	return nil
}

func (h *loggingHook) OnConnectError(_ *akira.Client, p *packet.Connect, err error) {
	log.Printf("Failed to connect client '%s': %s\n", p.ClientID, err)
}

func (h *loggingHook) OnConnected(c *akira.Client) {
	log.Printf("Client '%s' connected with success\n", c.ID)
}
