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
	"math"
	"net"
	"sync/atomic"
	"time"

	"github.com/gsalomao/akira/packet"
)

// Client represents a MQTT client.
type Client struct {
	// ID represents the client identifier.
	ID []byte `json:"id"`

	// Connection represents the client's connection.
	Connection *Connection `json:"connection"`

	// Session represents the client's session.
	Session *Session `json:"session"`

	connected atomic.Bool
}

// Connected returns whether the client is connected or not.
func (c *Client) Connected() bool {
	return c.connected.Load()
}

func (c *Client) refreshDeadline(keepAlive uint16) {
	var deadline time.Time

	if keepAlive > 0 {
		timeout := math.Ceil(float64(keepAlive) * 1.5)
		deadline = time.Now().Add(time.Duration(timeout) * time.Second)
	}

	_ = c.Connection.netConn.SetDeadline(deadline)
}

// Connection represents the network connection with the client.
type Connection struct {
	netConn  net.Conn
	listener Listener

	// KeepAlive is a time interval, measured in seconds, that is permitted to elapse between the point at which
	// the client finishes transmitting one control packet and the point it starts sending the next.
	KeepAlive uint16

	// Version represents the MQTT version.
	Version packet.Version `json:"version"`
}

// Session represents the MQTT session.
type Session struct {
	// Properties represents the properties of the session (MQTT V5.0).
	Properties *SessionProperties `json:"properties,omitempty"`

	// LastWill represents the MQTT Will Message to be published by the server.
	LastWill *LastWill `json:"last_will,omitempty"`

	// ConnectedAt indicates the timestamp, in milliseconds, of the last time the client has connected.
	ConnectedAt int64 `json:"connected_at"`

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
