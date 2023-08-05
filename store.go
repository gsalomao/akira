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
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

// ErrSessionNotFound indicates that the session was not found in the SessionStore.
var ErrSessionNotFound = errors.New("session not found")

// SessionStore is responsible for Session persistence.
type SessionStore interface {
	// GetSession gets the session from the SessionStore. The session is identified by the given clientID.
	// If the SessionStore fails to get the session, it returns the error. If it fails due to the session does not
	// exist, it returns ErrSessionNotFound.
	GetSession(ctx context.Context, clientID []byte, s *Session) error

	// SaveSession saves the given session. If there's any existing session associated with the clientID, the existing
	// session is overridden with tne new session. If the SessionStore fails to save the session, it returns the error.
	SaveSession(ctx context.Context, clientID []byte, s *Session) error

	// DeleteSession deletes the session associated wht the given client identifier. If the SessionStore fails to
	// delete the session, it returns the error. If it fails due to the session does not exist, it returns
	// ErrSessionNotFound.
	DeleteSession(ctx context.Context, clientID []byte) error
}

type inMemorySessionStore struct {
	mutex    sync.RWMutex
	sessions map[string][]byte
}

func newInMemorySessionStore() *inMemorySessionStore {
	return &inMemorySessionStore{
		sessions: make(map[string][]byte),
	}
}

func (ss *inMemorySessionStore) GetSession(_ context.Context, clientID []byte, s *Session) error {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()

	encoded, ok := ss.sessions[string(clientID)]
	if !ok {
		return ErrSessionNotFound
	}

	err := json.Unmarshal(encoded, s)
	if err != nil {
		// When it fails to unmarshal the session, it means there's a programmer error in the definition of the session.
		// Due to this, it was decided to panic so that this error can be identified sooner.
		panic(fmt.Sprintf("failed to get session: %s", err.Error()))
	}
	return nil
}

func (ss *inMemorySessionStore) SaveSession(_ context.Context, clientID []byte, s *Session) error {
	encoded, err := json.Marshal(s)
	if err != nil {
		// When it fails to marshal the session, it means there's a programmer error in the definition of the session.
		// Due to this, it was decided to panic so that this error can be identified sooner.
		panic(fmt.Sprintf("failed to save session: %s", err.Error()))
	}

	ss.mutex.Lock()
	ss.sessions[string(clientID)] = encoded
	ss.mutex.Unlock()
	return nil
}

func (ss *inMemorySessionStore) DeleteSession(_ context.Context, clientID []byte) error {
	id := string(clientID)

	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	if _, ok := ss.sessions[id]; !ok {
		return ErrSessionNotFound
	}

	delete(ss.sessions, id)
	return nil
}
