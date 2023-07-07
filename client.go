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
	"container/list"
	"math"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// ClientPending indicates that the network connection is open but the MQTT connection flow is pending.
	ClientPending ClientState = iota

	// ClientClosed indicates that the client is closed and not able to receive or send any packet.
	ClientClosed
)

// ClientState represents the client state.
type ClientState byte

// Client represents a MQTT client.
type Client struct {
	connection connection    // 48 bytes
	server     *Server       // 8 bytes
	done       chan struct{} // 8 bytes
	closeOnce  sync.Once     // 12 bytes
	state      atomic.Uint32 // 4 bytes
}

type connection struct {
	netConn        net.Conn    // 16 bytes
	listener       Listener    // 16 bytes
	outboundStream chan []byte // 8 bytes
	keepAlive      uint16      // 2 bytes
}

func newClient(nc net.Conn, s *Server, l Listener) *Client {
	c := Client{
		connection: connection{
			netConn:        nc,
			outboundStream: make(chan []byte, s.config.OutboundStreamSize),
			listener:       l,
			keepAlive:      s.config.ConnectTimeout,
		},
		server: s,
		done:   make(chan struct{}),
	}
	return &c
}

// Close closes the Client.
func (c *Client) Close(err error) {
	c.closeOnce.Do(func() {
		c.server.hooks.onConnectionClose(c.server, c.connection.listener, err)

		c.state.Store(uint32(ClientClosed))
		_ = c.connection.netConn.Close()
		close(c.done)

		c.server.hooks.onConnectionClosed(c.server, c.connection.listener, err)
	})
}

// State returns the current client state.
func (c *Client) State() ClientState {
	return ClientState(c.state.Load())
}

// Done returns a channel which is closed when the Client is closed.
func (c *Client) Done() <-chan struct{} {
	return c.done
}

func (c *Client) packetToSend() <-chan []byte {
	return c.connection.outboundStream
}

func (c *Client) refreshDeadline() {
	var deadline time.Time

	if c.connection.keepAlive > 0 {
		timeout := math.Ceil(float64(c.connection.keepAlive) * 1.5)
		deadline = time.Now().Add(time.Duration(timeout) * time.Second)
	}

	_ = c.connection.netConn.SetDeadline(deadline)
}

type clients struct {
	mu      sync.RWMutex
	pending list.List
}

func newClients() *clients {
	c := clients{}
	return &c
}

func (c *clients) add(client *Client) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.pending.PushBack(client)
}

func (c *clients) closeAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem := c.pending.Front()

	for {
		if elem == nil {
			break
		}

		elem.Value.(*Client).Close(nil)
		elem = elem.Next()
	}
}
