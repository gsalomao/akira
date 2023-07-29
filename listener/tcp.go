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
	"context"
	"crypto/tls"
	"net"
	"sync"

	"github.com/gsalomao/akira"
)

// TCP is a Listener responsible for listen and accept TCP connections.
type TCP struct {
	address   string
	listener  net.Listener
	tlsConfig *tls.Config
	wg        sync.WaitGroup
}

// NewTCP creates a new instance of the TCP.
func NewTCP(address string, tlsConfig *tls.Config) *TCP {
	return &TCP{address: address, tlsConfig: tlsConfig}
}

// Listen starts the listener. When the listener starts listening, it starts to accept any incoming TCP connection,
// and calls the Handler with the new connection. If the listener fails to start listening, it returns the error.
// This function does not block the caller and returns immediately after the listener is ready to accept incoming
// connections.
func (t *TCP) Listen(h akira.Handler) error {
	var err error

	if t.tlsConfig == nil {
		t.listener, err = net.Listen("tcp", t.address)
	} else {
		t.listener, err = tls.Listen("tcp", t.address, t.tlsConfig)
	}
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.wg.Add(1)

	go func() {
		defer t.wg.Done()
		cancel()

		for {
			var nc net.Conn

			nc, err = t.listener.Accept()
			if err != nil {
				break
			}

			_ = h(akira.NewConnection(t, nc))
		}
	}()

	<-ctx.Done()
	return nil
}

// Close closes the listener. Once the listener is closed, it does not accept any other incoming TCP connection. This
// function blocks and returns only after the listener has closed.
func (t *TCP) Close() error {
	if t.listener == nil {
		return nil
	}

	err := t.listener.Close()
	if err != nil {
		return err
	}

	t.wg.Wait()
	return nil
}
