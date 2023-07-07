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
	"sync/atomic"
)

type listenerMock struct {
	name         string
	address      string
	listening    chan bool
	onConnection OnConnectionFunc
	running      atomic.Bool
}

func newListenerMock(name, address string) *listenerMock {
	m := listenerMock{
		name:    name,
		address: address,
	}
	return &m
}

func (l *listenerMock) Name() string {
	return l.name
}

func (l *listenerMock) Listen(f OnConnectionFunc) (<-chan bool, error) {
	if _, _, err := net.SplitHostPort(l.address); err != nil {
		return nil, errors.New("invalid address")
	}

	l.onConnection = f
	l.listening = make(chan bool, 1)

	l.running.Store(true)
	l.listening <- true

	return l.listening, nil
}

func (l *listenerMock) Stop() {
	if l.Listening() {
		l.running.Store(false)
		close(l.listening)
	}
}

func (l *listenerMock) Listening() bool {
	return l.running.Load()
}

func (l *listenerMock) handle(c net.Conn) {
	l.onConnection(l, c)
}
