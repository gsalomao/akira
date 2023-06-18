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

package melitte

import (
	"net"
	"sync"
	"sync/atomic"
)

// Client represents a MQTT client.
type Client struct {
	connection connection    // 40 bytes
	server     *Server       // 8 bytes
	done       chan struct{} // 8 bytes
	closeOnce  sync.Once     // 12 bytes
	closed     atomic.Bool   // 4 bytes
}

type connection struct {
	netConn        net.Conn    // 16 bytes
	listener       Listener    // 16 bytes
	outboundStream chan []byte // 8 bytes
}

func newClient(nc net.Conn, s *Server, l Listener) *Client {
	c := Client{
		connection: connection{
			netConn:        nc,
			outboundStream: make(chan []byte, s.Options.Config.OutboundStreamSize),
			listener:       l,
		},
		server: s,
		done:   make(chan struct{}),
	}
	return &c
}

// Close closes the Client.
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		c.closed.Store(true)
		c.server.hooks.onConnectionClose(c.server, c.connection.listener)

		_ = c.connection.netConn.Close()
		c.connection.netConn = nil

		close(c.done)
		c.server.hooks.onConnectionClosed(c.server, c.connection.listener)
	})
}

// Closed returns whether the Client is closed or not.
func (c *Client) Closed() bool {
	return c.closed.Load()
}

// Done returns a channel which is closed when the Client is closed.
func (c *Client) Done() <-chan struct{} {
	return c.done
}

func (c *Client) packetToSend() <-chan []byte {
	return c.connection.outboundStream
}
