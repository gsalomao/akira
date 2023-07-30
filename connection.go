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
	"fmt"
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
	netConn       net.Conn
	sendTimeoutMs int

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

func (c *Connection) readFixedHeader(r *bufio.Reader, h *packet.FixedHeader) (n int, err error) {
	if c == nil {
		return 0, io.EOF
	}

	r.Reset(c.netConn)
	err = c.setReadDeadline()
	if err != nil {
		return 0, fmt.Errorf("failed to set read deadline: %w", err)
	}

	n, err = h.Read(r)
	if err != nil {
		return n, fmt.Errorf("failed to read fixed header: %w", err)
	}

	return n, nil
}

func (c *Connection) receivePacket(r *bufio.Reader, h packet.FixedHeader) (p Packet, n int, err error) {
	if c == nil {
		return nil, 0, io.EOF
	}

	var pd PacketDecodable

	switch h.PacketType {
	case packet.TypeConnect:
		pd = &packet.Connect{}
	default:
		return nil, n, fmt.Errorf("%w: %v: unsupported packet", packet.ErrProtocolError, h.PacketType)
	}

	buf := make([]byte, h.RemainingLength)
	if _, err = io.ReadFull(r, buf); err != nil {
		return nil, n, fmt.Errorf("failed to read remaining bytes: %w", err)
	}

	p = pd
	n += h.RemainingLength

	var dSize int

	dSize, err = pd.Decode(buf, h)
	if err != nil {
		return nil, n, fmt.Errorf("decode error: %s: %w", p.Type(), err)
	}
	if dSize != h.RemainingLength {
		return nil, n, fmt.Errorf("%w: %s: packet size mismatch", packet.ErrMalformedPacket, p.Type())
	}

	return p, n, nil
}

func (c *Connection) sendPacket(p PacketEncodable) (n int, err error) {
	if c == nil {
		return 0, io.EOF
	}

	buf := make([]byte, p.Size())

	_, err = p.Encode(buf)
	if err != nil {
		return n, fmt.Errorf("failed to encode packet: %w", err)
	}

	err = c.setWriteDeadline()
	if err != nil {
		return 0, fmt.Errorf("failed to set write deadline: %w", err)
	}

	return c.netConn.Write(buf)
}

func (c *Connection) setReadDeadline() error {
	var deadline time.Time
	if c.KeepAliveMs > 0 {
		timeout := math.Ceil(float64(c.KeepAliveMs) * 1.5)
		deadline = time.Now().Add(time.Duration(timeout) * time.Millisecond)
	}
	return c.netConn.SetReadDeadline(deadline)
}

func (c *Connection) setWriteDeadline() error {
	var deadline time.Time
	if c.sendTimeoutMs > 0 {
		deadline = time.Now().Add(time.Duration(c.sendTimeoutMs) * time.Millisecond)
	}
	return c.netConn.SetWriteDeadline(deadline)
}
