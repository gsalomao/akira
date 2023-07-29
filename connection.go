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
	"bufio"
	"errors"
	"io"
	"math"
	"net"
	"time"

	"github.com/gsalomao/akira/packet"
)

// ErrInvalidConnection indicates that the Connection is invalid.
var ErrInvalidConnection = errors.New("invalid connection")

// Connection represents the network connection with the client.
type Connection struct {
	netConn net.Conn

	// Listener is the Listener which accepted the connection.
	Listener Listener `json:"-"`

	// KeepAliveMs is a time interval, measured in milliseconds, that is permitted to elapse between the point
	// at which the client finishes transmitting one control packet and the point it starts sending the next.
	KeepAliveMs uint32 `json:"keep_alive_ms"`

	// Version represents the MQTT version.
	Version packet.Version `json:"version"`
}

// NewConnection creates a new Connection.
func NewConnection(l Listener, nc net.Conn) *Connection {
	return &Connection{Listener: l, netConn: nc}
}

func (c *Connection) receivePacket(r *bufio.Reader) (p Packet, n int, err error) {
	if c == nil {
		return nil, 0, io.EOF
	}

	r.Reset(c.netConn)
	c.setReadDeadline()
	return readPacket(r)
}

func (c *Connection) sendPacket(p PacketEncodable) (n int, err error) {
	if c == nil {
		return 0, io.EOF
	}
	return writePacket(c.netConn, p)
}

func (c *Connection) setReadDeadline() {
	var deadline time.Time
	if c.KeepAliveMs > 0 {
		timeout := math.Ceil(float64(c.KeepAliveMs) * 1.5)
		deadline = time.Now().Add(time.Duration(timeout) * time.Millisecond)
	}
	_ = c.netConn.SetReadDeadline(deadline)
}
