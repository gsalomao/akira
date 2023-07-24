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
	tcp := NewTCP(":1883", nil)
	if tcp == nil {
		t.Fatal("A TCP listener was expected")
	}

	err := tcp.Listen(func(_ akira.Listener, _ net.Conn) {})
	if err != nil {
		t.Fatalf("Unexpected error\n%s", err)
	}
	_ = tcp.Close()
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

	err = tcp.Listen(func(_ akira.Listener, _ net.Conn) {})
	if err != nil {
		t.Fatalf("Unexpected error\n%s", err)
	}
	_ = tcp.Close()
}

func TestTCPListenError(t *testing.T) {
	tcp := NewTCP(":abc", nil)
	if tcp == nil {
		t.Fatal("A TCP listener was expected")
	}

	err := tcp.Listen(func(akira.Listener, net.Conn) {})
	if err == nil {
		t.Fatal("An error was expected")
	}
}

func TestTCPListenOnConnection(t *testing.T) {
	tcp := NewTCP(":1883", nil)
	if tcp == nil {
		t.Fatal("A TCP listener was expected")
	}

	var lsn akira.Listener
	var nc net.Conn
	done := make(chan struct{})

	err := tcp.Listen(func(l akira.Listener, c net.Conn) {
		lsn = l
		nc = c
		close(done)
	})
	if err != nil {
		t.Fatalf("Unexpected error\n%s", err)
	}
	defer func() { _ = tcp.Close() }()

	client, err := net.Dial("tcp", ":1883")
	if err != nil {
		t.Fatalf("Unexpected error\n%s", err)
	}
	if client == nil {
		t.Fatal("A client was expected")
	}
	defer func() { _ = client.Close() }()

	<-done
	if lsn == nil {
		t.Fatal("A listener was expected")
	}
	if nc == nil {
		t.Fatal("A network connection was expected")
	}
}

func TestTCPClose(t *testing.T) {
	tcp := NewTCP(":1883", nil)
	if tcp == nil {
		t.Fatal("A TCP listener was expected")
	}

	err := tcp.Listen(func(_ akira.Listener, _ net.Conn) {})
	if err != nil {
		t.Fatalf("Unexpected error\n%s", err)
	}

	err = tcp.Close()
	if err != nil {
		t.Fatalf("Unexpected error\n%s", err)
	}
}

func TestTCPCloseWhenNotListening(t *testing.T) {
	tcp := NewTCP(":1883", nil)
	if tcp == nil {
		t.Fatal("A TCP listener was expected")
	}

	err := tcp.Close()
	if err != nil {
		t.Fatalf("Unexpected error\n%s", err)
	}
}
