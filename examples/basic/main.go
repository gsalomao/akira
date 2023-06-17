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

	"github.com/gsalomao/melitte"
)

func main() {
	s := melitte.NewServer(melitte.NewDefaultOptions())
	defer s.Close()

	l := loggingHook{}
	err := s.AddHook(&l)
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

	err = s.Start(context.Background())
	if err != nil {
		log.Fatal("Failed to start server")
	}

	wg.Wait()
	_ = s.Stop(context.Background())
}

type loggingHook struct {
}

func (h *loggingHook) Name() string {
	return "logging"
}

func (h *loggingHook) OnServerStart(_ *melitte.Server) error {
	fmt.Println("Server starting")
	return nil
}

func (h *loggingHook) OnServerStarted(_ *melitte.Server) {
	fmt.Println("Server started")
}

func (h *loggingHook) OnServerStartFailed(_ *melitte.Server, err error) {
	fmt.Println("Failed to start server: " + err.Error())
}

func (h *loggingHook) OnServerStop(_ *melitte.Server) {
	fmt.Println("Server stopping")
}

func (h *loggingHook) OnServerStopped(_ *melitte.Server) {
	fmt.Println("Server stopped")
}
