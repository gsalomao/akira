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
	"sync"
	"time"

	"github.com/gsalomao/akira/packet"
)

const (
	// ClientPending indicates that the network connection is open but the MQTT connection flow is pending.
	ClientPending ClientState = iota

	// ClientConnected indicates that the client is currently connected.
	ClientConnected

	// ClientClosed indicates that the client is closed and not able to receive or send any packet.
	ClientClosed
)

// ClientState represents the client state.
type ClientState byte

// Client represents a MQTT client.
type Client struct {
	sync.RWMutex

	// Connection represents the client's connection.
	Connection *Connection `json:"connection"`

	// Server represents the current server.
	Server *Server `json:"-"`

	// Session represents the client's session.
	Session *Session `json:"session"`

	closeOnce sync.Once
	state     ClientState
}

// Connection represents the network connection with the client.
type Connection struct {
	netConn  net.Conn
	listener Listener

	// Timestamp which the client connected.
	ConnectedAt time.Time

	// KeepAlive is a time interval, measured in seconds, that is permitted to elapse between the point at which
	// the client finishes transmitting one control packet and the point it starts sending the next.
	KeepAlive uint16

	// Version represents the MQTT version.
	Version packet.Version `json:"version"`
}

// Session represents the MQTT session.
type Session struct {
	// ClientID represents the client identifier.
	ClientID []byte `json:"client_id"`

	// Properties represents the properties of the session (MQTT V5.0).
	Properties *SessionProperties `json:"properties,omitempty"`

	// LastWill represents the MQTT Will Message to be published by the server.
	LastWill *LastWill `json:"last_will,omitempty"`

	// Connected indicates that the client associated with the session is connected.
	Connected bool `json:"connected"`
}

// SessionProperties represents the properties of the session (MQTT V5.0 only).
type SessionProperties struct {
	// Flags indicates which properties are present.
	Flags packet.PropertyFlags

	// UserProperties is a list of user properties.
	UserProperties []packet.UserProperty `json:"user_properties"`

	// SessionExpiryInterval represents the time, in seconds, which the server must store the Session State after
	// the network connection is closed.
	SessionExpiryInterval uint32 `json:"session_expiry_interval"`

	// MaximumPacketSize represents the maximum packet size, in bytes, the client is willing to accept.
	MaximumPacketSize uint32 `json:"maximum_packet_size"`

	// ReceiveMaximum represents the maximum number of inflight messages with QoS > 0.
	ReceiveMaximum uint16 `json:"receive_maximum"`

	// TopicAliasMaximum represents the highest number of Topic Alias that the client accepts.
	TopicAliasMaximum uint16 `json:"topic_alias_maximum"`

	// RequestResponseInfo indicates if the server can send Response Information with the CONNACK Packet.
	RequestResponseInfo bool `json:"request_response_info"`

	// RequestProblemInfo indicates whether the Reason String or User Properties can be sent to the client in case
	// of failures on any packet.
	RequestProblemInfo bool `json:"request_problem_info"`
}

// Has returns whether the property is present or not.
func (p *SessionProperties) Has(id packet.PropertyID) bool {
	if p != nil {
		return p.Flags.Has(id)
	}
	return false
}

// Set sets the property indicating that it's present.
func (p *SessionProperties) Set(id packet.PropertyID) {
	if p != nil {
		p.Flags = p.Flags.Set(id)
	}
}

// LastWill represents the MQTT Will Message.
type LastWill struct {
	// Topic represents the topic which the Will Message will be published.
	Topic []byte `json:"topic"`

	// Payload represents the message payload of the Will Message.
	Payload []byte `json:"payload"`

	// Properties represents the properties of the Will Message.
	Properties *packet.WillProperties `json:"properties,omitempty"`

	// QoS represents the QoS level to be used when publishing the Will Message.
	QoS packet.QoS `json:"qos"`

	// Retain indicates whether the Will Message will be published as a retained message or not.
	Retain bool `json:"retain"`
}

func newClient(nc net.Conn, s *Server, l Listener) *Client {
	c := Client{
		Connection: &Connection{
			netConn:     nc,
			listener:    l,
			ConnectedAt: time.Now(),
			KeepAlive:   s.config.ConnectTimeout,
		},
		Server: s,
	}
	return &c
}

// Close closes the client. When the client is closed, it's not able to receive or send any other packet.
func (c *Client) Close(err error) {
	c.closeOnce.Do(func() {
		c.Server.hooks.onConnectionClose(c.Server, c.Connection.listener, err)

		c.Lock()
		defer c.Unlock()

		c.state = ClientClosed
		_ = c.Connection.netConn.Close()

		c.Server.hooks.onConnectionClosed(c.Server, c.Connection.listener, err)
	})
}

// State returns the current client state.
func (c *Client) State() ClientState {
	c.RLock()
	defer c.RUnlock()

	return c.state
}

func (c *Client) setState(s ClientState) {
	c.Lock()
	defer c.Unlock()

	c.state = s
}

func (c *Client) receivePacket(r *bufio.Reader) (Packet, error) {
	if err := c.Server.hooks.onPacketReceive(c); err != nil {
		return nil, err
	}

	p, _, err := readPacket(r)
	if err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) || c.State() == ClientClosed {
			return nil, err
		}

		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return nil, err
		}

		err = c.Server.hooks.onPacketReceiveError(c, err)
		return nil, err
	}

	if err = c.Server.hooks.onPacketReceived(c, p); err != nil {
		return nil, err
	}

	return p, err
}

func (c *Client) sendPacket(p PacketEncodable) error {
	err := c.Server.hooks.onPacketSend(c, p)
	if err != nil {
		return err
	}

	buf := make([]byte, p.Size())

	_, err = p.Encode(buf)
	if err != nil {
		c.Server.hooks.onPacketSendError(c, p, err)
		return err
	}

	_, err = c.Connection.netConn.Write(buf)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			c.Server.hooks.onPacketSendError(c, p, err)
		}
		return err
	}

	return c.Server.hooks.onPacketSent(c, p)
}

func (c *Client) refreshDeadline() {
	var deadline time.Time

	if c.Connection.KeepAlive > 0 {
		timeout := math.Ceil(float64(c.Connection.KeepAlive) * 1.5)
		deadline = time.Now().Add(time.Duration(timeout) * time.Second)
	}

	_ = c.Connection.netConn.SetDeadline(deadline)
}

type clients struct {
	mutex     sync.RWMutex
	pending   []*Client
	connected map[string]*Client
}

func newClients() *clients {
	c := clients{
		pending:   make([]*Client, 0),
		connected: make(map[string]*Client),
	}
	return &c
}

func (c *clients) add(client *Client) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.addLocked(client)
}

func (c *clients) remove(client *Client) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.removePendingLocked(client)
}

func (c *clients) update(client *Client) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.removePendingLocked(client)
	c.addLocked(client)
}

func (c *clients) closeAll() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, client := range c.pending {
		client.Close(nil)
		c.removePendingLocked(client)
	}
	for id, client := range c.connected {
		client.Close(nil)
		delete(c.connected, id)
	}
}

func (c *clients) addLocked(client *Client) {
	switch client.State() {
	case ClientPending:
		c.pending = append(c.pending, client)
	case ClientConnected:
		c.connected[string(client.Session.ClientID)] = client
	default:
	}
}

func (c *clients) removePendingLocked(client *Client) {
	var i int
	size := len(c.pending)

	for ; i < size; i++ {
		if c.pending[i] == client {
			break
		}
	}

	if i < size {
		c.pending[i] = c.pending[size-1]
		c.pending = c.pending[:size-1]
	}
}
