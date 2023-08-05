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
	"context"
	"errors"
	"sync/atomic"
)

var errSessionStoreFailed = errors.New("session store failed")

type mockSessionStore struct {
	_getCalls    atomic.Int32
	_saveCalls   atomic.Int32
	_deleteCalls atomic.Int32
	getCB        func([]byte, *Session) error
	saveCB       func([]byte, *Session) error
	deleteCB     func([]byte) error
}

func (m *mockSessionStore) GetSession(_ context.Context, clientID []byte, s *Session) error {
	m._getCalls.Add(1)
	if m.getCB != nil {
		return m.getCB(clientID, s)
	}
	return ErrSessionNotFound
}

func (m *mockSessionStore) SaveSession(_ context.Context, clientID []byte, s *Session) error {
	m._saveCalls.Add(1)
	if m.saveCB != nil {
		return m.saveCB(clientID, s)
	}
	return nil
}

func (m *mockSessionStore) DeleteSession(_ context.Context, clientID []byte) error {
	m._deleteCalls.Add(1)
	if m.deleteCB != nil {
		return m.deleteCB(clientID)
	}
	return nil
}

func (m *mockSessionStore) getCalls() int {
	return int(m._getCalls.Load())
}

func (m *mockSessionStore) saveCalls() int {
	return int(m._saveCalls.Load())
}

func (m *mockSessionStore) deleteCalls() int {
	return int(m._deleteCalls.Load())
}
