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

// SessionStore is responsible for Session persistence.
type SessionStore interface {
	// GetSession gets the session from the SessionStore. The session is identified by the given clientID.
	// If the SessionStore fails to get the session, it returns the error. If it fails due to the session does not
	// exist, it returns ErrSessionNotFound.
	GetSession(clientID []byte, s *Session) error

	// SaveSession saves the given session. If there's any existing session associated with the clientID, the existing
	// session is overridden with tne new session. If the SessionStore fails to save the session, it returns the error.
	SaveSession(clientID []byte, s *Session) error

	// DeleteSession deletes the session associated wht the given client identifier. If the SessionStore fails to
	// delete the session, it returns the error.
	DeleteSession(clientID []byte) error
}

type store struct {
	sessionStore SessionStore
}

func (st store) getSession(clientID []byte, s *Session) error {
	if st.sessionStore == nil {
		return ErrSessionNotFound
	}
	return st.sessionStore.GetSession(clientID, s)
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
