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
	"sync/atomic"

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
	listening atomic.Bool
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

// Listen starts the listener. The listener calls the OnConnectionFunc for any received incoming TCP connection.
// This function does not block the caller and returns a channel, which an event is sent when the listener is
// ready for accept incoming connection, and closed when the listener has stopped.
// It returns an error if the Listener fails to start.
func (l *TCPListener) Listen(f akira.OnConnectionFunc) (<-chan bool, error) {
	var err error

	if l.tlsConfig == nil {
		l.listener, err = net.Listen("tcp", l.address)
	} else {
		l.listener, err = tls.Listen("tcp", l.address, l.tlsConfig)
	}
	if err != nil {
		return nil, err
	}

	l.handle = f
	l.wg.Add(1)
	listening := make(chan bool, 1)

	go func() {
		defer l.wg.Done()
		defer close(listening)

		l.listening.Store(true)
		listening <- true

		for {
			c, err := l.listener.Accept()
			if err != nil {
				break
			}

			if f != nil {
				f(l, c)
			}
		}

		l.listening.Store(false)
	}()

	return listening, nil
}

// Stop stops the listener. When the listener is stopped, the channel returned by the Listen method is closed,
// and the listener does not accept any other incoming connection.
func (l *TCPListener) Stop() {
	_ = l.Close()
	l.wg.Wait()
}

// Listening returns whether the listener is listening for incoming connection or not.
func (l *TCPListener) Listening() bool {
	return l.listening.Load()
}

// Close closes the listener.
func (l *TCPListener) Close() error {
	if l.listener == nil {
		return nil
	}
	return l.listener.Close()
}
