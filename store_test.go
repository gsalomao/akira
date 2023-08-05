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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/gsalomao/akira/packet"
)

func TestInMemorySessionStoreGetSession(t *testing.T) {
	store := newInMemorySessionStore()
	session, encoded := newSession(t)
	store.sessions["abc"] = encoded

	var s Session
	err := store.GetSession(context.Background(), []byte("abc"), &s)
	if err != nil {
		t.Fatalf("Unexpected error\n%s", err)
	}
	if !reflect.DeepEqual(session, s) {
		t.Errorf("Unexpected session\nwant: %+v\ngot:  %+v", session, s)
	}
	if !reflect.DeepEqual(session.Properties, s.Properties) {
		t.Errorf("Unexpected session properties\nwant: %+v\ngot:  %+v", session.Properties, s.Properties)
	}
	if !reflect.DeepEqual(session.LastWill, s.LastWill) {
		t.Errorf("Unexpected session Last Will\nwant: %+v\ngot:  %+v", session.LastWill, s.LastWill)
	}
}

func TestInMemorySessionStoreGetSessionReturnsErrorWhenNotFound(t *testing.T) {
	var s Session
	store := newInMemorySessionStore()

	err := store.GetSession(context.Background(), []byte("abc"), &s)
	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("Unexpected error\nwant: %v\ngot:  %v", ErrSessionNotFound, err)
	}
}

func TestInMemorySessionStoreSaveSession(t *testing.T) {
	store := newInMemorySessionStore()
	session, encoded := newSession(t)

	err := store.SaveSession(context.Background(), []byte("abc"), &session)
	if err != nil {
		t.Fatalf("Unexpected error\n%s", err)
	}
	if !bytes.Equal(encoded, store.sessions["abc"]) {
		t.Errorf("Unexpected encoded session\nwant: %+v\ngot:  %+v", encoded, store.sessions["abc"])
	}
}

func TestInMemorySessionStoreDeleteSession(t *testing.T) {
	store := newInMemorySessionStore()
	_, encoded := newSession(t)
	store.sessions["abc"] = encoded

	err := store.DeleteSession(context.Background(), []byte("abc"))
	if err != nil {
		t.Fatalf("Unexpected error\n%s", err)
	}
	if !bytes.Equal(nil, store.sessions["abc"]) {
		t.Errorf("Unexpected encoded session\nwant: %+v\ngot:  %+v", nil, store.sessions["abc"])
	}
}

func TestInMemorySessionStoreDeleteSessionReturnsErrorWhenNotFound(t *testing.T) {
	store := newInMemorySessionStore()

	err := store.DeleteSession(context.Background(), []byte("abc"))
	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("Unexpected error\nwant: %v\ngot:  %v", ErrSessionNotFound, err)
	}
}

func BenchmarkInMemorySessionStoreGetSession(b *testing.B) {
	store := newInMemorySessionStore()
	numOfSessions := 10000

	for i := 0; i < numOfSessions; i++ {
		id := fmt.Sprintf("client-%v", i)
		_, encoded := newSession(b)
		store.sessions[id] = encoded
	}

	var s Session
	ctx := context.Background()
	existingID := []byte(fmt.Sprintf("client-%v", rand.Intn(numOfSessions)))
	nonExistingID := []byte(fmt.Sprintf("client-%v", numOfSessions))
	b.ResetTimer()

	b.Run("Found", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := store.GetSession(ctx, existingID, &s)
			if err != nil {
				b.Fatalf("Unexpected error\n%s", err)
			}
		}
	})

	b.Run("Not Found", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := store.GetSession(ctx, nonExistingID, &s)
			if !errors.Is(err, ErrSessionNotFound) {
				b.Fatalf("Unexpected error\nwant: %v\ngot:  %v", ErrSessionNotFound, err)
			}
		}
	})
}

func BenchmarkInMemorySessionStoreSaveSession(b *testing.B) {
	store := newInMemorySessionStore()
	session, _ := newSession(b)
	ctx := context.Background()
	id := []byte("abc")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := store.SaveSession(ctx, id, &session)
		if err != nil {
			b.Fatalf("Unexpected error\n%s", err)
		}
	}
}

func BenchmarkInMemorySessionStoreDeleteSession(b *testing.B) {
	store := newInMemorySessionStore()
	numOfSessions := 10000

	for i := 0; i < numOfSessions; i++ {
		id := fmt.Sprintf("client-%v", i)
		_, encoded := newSession(b)
		store.sessions[id] = encoded
	}

	_, encoded := newSession(b)
	ctx := context.Background()
	idStr := fmt.Sprintf("client-%v", numOfSessions)
	id := []byte(idStr)
	nonExistingID := []byte(fmt.Sprintf("client-%v", numOfSessions+1))
	b.ResetTimer()

	b.Run("Found", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			store.sessions[idStr] = encoded

			err := store.DeleteSession(ctx, id)
			if err != nil {
				b.Fatalf("Unexpected error\n%s", err)
			}
		}
	})

	b.Run("Not Found", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := store.DeleteSession(ctx, nonExistingID)
			if !errors.Is(err, ErrSessionNotFound) {
				b.Fatalf("Unexpected error\nwant: %v\ngot:  %v", ErrSessionNotFound, err)
			}
		}
	})
}

func newSession(tb testing.TB) (s Session, encoded []byte) {
	tb.Helper()

	s = Session{Connected: true, ConnectedAt: time.Now().UnixMilli()}
	s.Properties = &SessionProperties{
		Flags:                 packet.PropertyFlags(0).Set(packet.PropertySessionExpiryInterval),
		SessionExpiryInterval: 100,
	}

	var err error
	encoded, err = json.Marshal(s)
	if err != nil {
		tb.Fatalf("Unexpected error\n%s", err)
	}

	return s, encoded
}
