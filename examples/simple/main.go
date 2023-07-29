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
)

func main() {
	server, err := akira.NewServer()
	if err != nil {
		log.Println("Failed to create server")
		os.Exit(1)
	}
	defer server.Close()

	tcp := listener.NewTCP(":1883", nil)
	defer func() { _ = tcp.Close() }()

	err = server.AddListener(tcp)
	if err != nil {
		log.Println("Failed to add TCP listener into the server")
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
