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

	"github.com/gsalomao/akira/packet"
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
	// Connection represents the client's connection.
	Connection Connection `json:"connection"`

	// Server represents the current server.
	Server *Server `json:"-"`

	// Session represents the client's session.
	Session *Session `json:"session"`

	done      chan struct{}
	closeOnce sync.Once
	state     atomic.Uint32
}

// Connection represents the network connection with the client.
type Connection struct {
	netConn        net.Conn
	listener       Listener
	outboundStream chan Packet

	// KeepAlive is a time interval, measured in seconds, that is permitted to elapse between the point at which
	// the client finishes transmitting one control packet and the point it starts sending the next.
	KeepAlive uint16
}

// Session represents the MQTT session.
type Session struct {
	// ClientID represents the client identifier.
	ClientID []byte `json:"client_id"`

	// Properties represents the properties of the session (MQTT V5.0).
	Properties *SessionProperties `json:"properties,omitempty"`

	// LastWill represents the MQTT Will Message to be published by the server.
	LastWill *LastWill `json:"last_will,omitempty"`

	// Version represents the MQTT version.
	Version packet.Version `json:"version"`

	// Connected indicates that the client associated with the session is connected.
	Connected bool `json:"connected"`
}

// SessionProperties represents the properties of the session (MQTT V5.0 only).
type SessionProperties struct {
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

// LastWill represents the MQTT Will Message.
type LastWill struct {
	// Topic represents the topic which the Will Message will be published.
	Topic []byte `json:"topic"`

	// Payload represents the message payload of the Will Message.
	Payload []byte `json:"payload"`

	// Properties represents the properties of the Will Message.
	Properties *packet.PropertiesWill `json:"properties,omitempty"`

	// QoS represents the QoS level to be used when publishing the Will Message.
	QoS packet.QoS `json:"qos"`

	// Retain indicates whether the Will Message will be published as a retained message or not.
	Retain bool `json:"retain"`
}

func newClient(nc net.Conn, s *Server, l Listener) *Client {
	c := Client{
		Connection: Connection{
			netConn:        nc,
			outboundStream: make(chan Packet, s.config.OutboundStreamSize),
			listener:       l,
			KeepAlive:      s.config.ConnectTimeout,
		},
		Server: s,
		done:   make(chan struct{}),
	}
	return &c
}

// Close closes the Client.
func (c *Client) Close(err error) {
	c.closeOnce.Do(func() {
		c.Server.hooks.onConnectionClose(c.Server, c.Connection.listener, err)

		c.state.Store(uint32(ClientClosed))
		_ = c.Connection.netConn.Close()
		close(c.done)

		c.Server.hooks.onConnectionClosed(c.Server, c.Connection.listener, err)
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

func (c *Client) packetToSend() <-chan Packet {
	return c.Connection.outboundStream
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
