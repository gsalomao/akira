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
	"sync/atomic"
)

var errListenerFailed = errors.New("listener failed")

type mockListener struct {
	listenCB     func(f Handler) error
	closeCB      func() error
	_listenCalls atomic.Int32
	_closeCalls  atomic.Int32
	handler      Handler
}

func (m *mockListener) Listen(h Handler) error {
	m._listenCalls.Add(1)
	m.handler = h
	if m.listenCB == nil {
		return nil
	}
	return m.listenCB(h)
}

func (m *mockListener) Close() error {
	m._closeCalls.Add(1)
	if m.closeCB == nil {
		return nil
	}
	return m.closeCB()
}

func (m *mockListener) listenCalls() int {
	return int(m._listenCalls.Load())
}

func (m *mockListener) closeCalls() int {
	return int(m._closeCalls.Load())
}
