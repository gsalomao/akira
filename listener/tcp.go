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
	"sync"

	"github.com/gsalomao/akira"
)

// TCPListener is a Listener responsible for listen and accept TCP connections.
type TCPListener struct {
	name      string
	address   string
	listener  net.Listener
	tlsConfig *tls.Config
	handle    akira.OnConnectionFunc
	wg        sync.WaitGroup
}

// NewTCPListener creates a new instance of the TCPListener.
func NewTCPListener(name, address string, tlsConfig *tls.Config) *TCPListener {
	l := TCPListener{
		name:      name,
		address:   address,
		tlsConfig: tlsConfig,
	}
	return &l
}

// Name returns the name of the listener.
func (l *TCPListener) Name() string {
	return l.name
}

// Listen starts the listener. When the listener starts listening, it starts to accept any incoming TCP connection,
// and calls f with the new TCP connection. If the listener fails to start listening, it returns the error.
// This function does not block the caller and returns immediately after the listener is ready to accept incoming
// connections.
func (l *TCPListener) Listen(f akira.OnConnectionFunc) error {
	var err error

	if l.tlsConfig == nil {
		l.listener, err = net.Listen("tcp", l.address)
	} else {
		l.listener, err = tls.Listen("tcp", l.address, l.tlsConfig)
	}
	if err != nil {
		return err
	}

	l.handle = f
	l.wg.Add(1)
	listening := make(chan struct{})

	go func() {
		defer l.wg.Done()
		close(listening)

		for {
			var c net.Conn

			c, err = l.listener.Accept()
			if err != nil {
				break
			}

			if f != nil {
				f(l, c)
			}
		}
	}()

	return nil
}

// Close closes the listener. Once the listener is closed, it does not accept any other incoming TCP connection. This
// function blocks and returns only after the listener has closed.
func (l *TCPListener) Close() error {
	if l.listener == nil {
		return nil
	}

	err := l.listener.Close()
	if err != nil {
		return err
	}

	l.wg.Wait()
	return nil
}
