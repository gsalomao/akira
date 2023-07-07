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

package akira

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
)

type mockListener struct {
	name      string
	address   string
	done      chan struct{}
	wg        sync.WaitGroup
	handle    OnConnectionFunc
	listening atomic.Bool
}

func newMockListener(name, address string) *mockListener {
	m := mockListener{
		name:    name,
		address: address,
		done:    make(chan struct{}),
	}
	return &m
}

func (l *mockListener) Name() string {
	return l.name
}

func (l *mockListener) Address() string {
	return l.address
}

func (l *mockListener) Protocol() string {
	return "mock"
}

func (l *mockListener) Listen(f OnConnectionFunc) (<-chan bool, error) {
	if _, _, err := net.SplitHostPort(l.address); err != nil {
		return nil, errors.New("invalid address")
	}

	listening := make(chan bool, 1)

	l.done = make(chan struct{})
	l.handle = f
	l.wg.Add(1)

	go func() {
		defer l.wg.Done()
		defer close(listening)

		l.listening.Store(true)
		listening <- true

		<-l.done
		l.listening.Store(false)
	}()

	return listening, nil
}

func (l *mockListener) Stop() {
	close(l.done)
	l.wg.Wait()
}

func (l *mockListener) Listening() bool {
	return l.listening.Load()
}

func (l *mockListener) onConnection(c net.Conn) {
	l.handle(l, c)
}
