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

import "errors"

// ErrSessionNotFound indicates that the session was not found in the SessionStore.
var ErrSessionNotFound = errors.New("session not found")

// SessionStore is responsible for persistence of Session in a non-volatile memory.
type SessionStore interface {
	// GetSession gets the session stored for the given client identifier and the error if it fails to get the session.
	// If the SessionStore does not find the session for the given client identifier, it returns ErrSessionNotFound.
	GetSession(clientID []byte) (s *Session, err error)

	// SaveSession saves the given session. If there's any existing session associated with the client identifier, the
	// existing session is overridden with tne new session. If the SessionStore fails to save the session, the error is
	// returned.
	SaveSession(clientID []byte, s *Session) error

	// DeleteSession deletes the session associated wht the given client identifier. If the SessionStore fails to
	// delete the session, the error is returned.
	DeleteSession(clientID []byte) error
}

type store struct {
	sessionStore SessionStore
}

func (st store) getSession(clientID []byte) (*Session, error) {
	if st.sessionStore == nil {
		return nil, ErrSessionNotFound
	}
	return st.sessionStore.GetSession(clientID)
}

func (st store) saveSession(clientID []byte, s *Session) error {
	if st.sessionStore == nil {
		return nil
	}
	return st.sessionStore.SaveSession(clientID, s)
}

func (st store) deleteSession(clientID []byte) error {
	if st.sessionStore == nil {
		return nil
	}
	return st.sessionStore.DeleteSession(clientID)
}
