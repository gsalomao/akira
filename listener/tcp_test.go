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

package listener

import (
	"crypto/tls"
	"net"
	"os"
	"testing"

	"github.com/gsalomao/akira"
)

func TestTCPListen(t *testing.T) {
	tcp := newTCP(t)

	err := tcp.Listen(func(_ *akira.Connection) error { return nil })
	if err != nil {
		t.Fatalf("Unexpected error\n%s", err)
	}
	closeTCP(t, tcp)
}

func TestTCPListenWithTLSConfig(t *testing.T) {
	crt, err := os.ReadFile("testdata/test.crt")
	if err != nil {
		t.Fatalf("Unexpected error\n%s", err)
	}
	if len(crt) == 0 {
		t.Fatal("A certificate was expected")
	}

	key, err := os.ReadFile("testdata/test.key")
	if err != nil {
		t.Fatalf("Unexpected error\n%s", err)
	}
	if len(key) == 0 {
		t.Fatal("A key was expected")
	}

	cert, err := tls.X509KeyPair(crt, key)
	if err != nil {
		t.Fatalf("Unexpected error\n%s", err)
	}

	tlsConfig := tls.Config{MinVersion: tls.VersionTLS12, Certificates: []tls.Certificate{cert}}

	tcp := NewTCP(":1883", &tlsConfig)
	if tcp == nil {
		t.Fatal("A TCP listener was expected")
	}

	err = tcp.Listen(func(_ *akira.Connection) error { return nil })
	if err != nil {
		t.Fatalf("Unexpected error\n%s", err)
	}
	closeTCP(t, tcp)
}

func TestTCPListenErrorInvalidAddress(t *testing.T) {
	tcp := NewTCP(":abc", nil)
	if tcp == nil {
		t.Fatal("A TCP listener was expected")
	}

	err := tcp.Listen(func(_ *akira.Connection) error { return nil })
	if err == nil {
		t.Fatal("An error was expected")
	}
}

func TestTCPListenCallsHandlerWhenAcceptsConnection(t *testing.T) {
	tcp := newTCP(t)

	var conn *akira.Connection
	done := make(chan struct{})

	err := tcp.Listen(func(c *akira.Connection) error {
		conn = c
		close(done)
		return nil
	})
	if err != nil {
		t.Fatalf("Unexpected error\n%s", err)
	}
	defer closeTCP(t, tcp)

	client, err := net.Dial("tcp", ":1883")
	if err != nil {
		t.Fatalf("Unexpected error\n%s", err)
	}
	if client == nil {
		t.Fatal("A client was expected")
	}
	defer func() { _ = client.Close() }()

	<-done
	if conn == nil {
		t.Fatal("A Connection was expected")
	}
}

func TestTCPClose(t *testing.T) {
	tcp := newTCP(t)
	listenTCP(t, tcp)

	err := tcp.Close()
	if err != nil {
		t.Fatalf("Unexpected error\n%s", err)
	}
}

func TestTCPCloseWhenNotListening(t *testing.T) {
	tcp := newTCP(t)

	err := tcp.Close()
	if err != nil {
		t.Fatalf("Unexpected error\n%s", err)
	}
}

func TestTCPCloseTwiceReturnsError(t *testing.T) {
	tcp := newTCP(t)
	listenTCP(t, tcp)
	closeTCP(t, tcp)

	err := tcp.Close()
	if err == nil {
		t.Fatal("An error was expected")
	}
}

func TestTCPListenAfterClosed(t *testing.T) {
	tcp := newTCP(t)
	listenTCP(t, tcp)
	closeTCP(t, tcp)

	err := tcp.Listen(func(_ *akira.Connection) error { return nil })
	if err != nil {
		t.Fatalf("Unexpected error\n%s", err)
	}
}

func newTCP(tb testing.TB) *TCP {
	tb.Helper()
	tcp := NewTCP(":1883", nil)
	if tcp == nil {
		tb.Fatal("A TCP listener was expected")
	}
	return tcp
}

func listenTCP(tb testing.TB, tcp *TCP) {
	tb.Helper()
	err := tcp.Listen(func(_ *akira.Connection) error { return nil })
	if err != nil {
		tb.Fatalf("Unexpected error\n%s", err)
	}
}

func closeTCP(tb testing.TB, tcp *TCP) {
	tb.Helper()
	err := tcp.Close()
	if err != nil {
		tb.Fatalf("Unexpected error\n%s", err)
	}
}
