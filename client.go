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
	"sync/atomic"
	"time"

	"github.com/gsalomao/akira/packet"
)

// Client represents a MQTT client.
type Client struct {
	connected atomic.Bool

	// ID represents the client identifier.
	ID []byte `json:"id"`

	// Session represents the client's session.
	Session Session `json:"session"`

	// Connection represents the client's connection.
	Connection *Connection `json:"connection"`

	// SessionPresent indicates whether the session was present before the current connection or not.
	SessionPresent bool `json:"session_present"`

	// PersistentSession indicates whether the session is persisted in the session store or not.
	PersistentSession bool `json:"persistent_session"`
}

// Connected returns whether the client is connected or not.
func (c *Client) Connected() bool {
	return c.connected.Load()
}

// Session represents the MQTT session.
type Session struct {
	// Properties represents the properties of the session (MQTT V5.0).
	Properties *SessionProperties `json:"properties,omitempty"`

	// LastWill represents the MQTT Will Message to be published by the server.
	LastWill *LastWill `json:"last_will,omitempty"`

	// ConnectedAt indicates the Unix timestamp, in milliseconds, of the last time the client connected.
	ConnectedAt int64 `json:"connected_at"`

	// DisconnectedAt indicates the Unix timestamp, in milliseconds, of the last time the client disconnected.
	DisconnectedAt int64 `json:"disconnected_at"`
}

// Expired returns whether the session is expired or not. A session is considered expired only if it is not connected
// anymore and the timestamp when the session disconnected plus the session expiry interval is older than the current
// timestamp.
func (s *Session) Expired() bool {
	if s == nil || s.Properties == nil || !s.Properties.Has(packet.PropertySessionExpiryInterval) {
		return false
	}
	expiryAt := s.ConnectedAt + int64(s.Properties.SessionExpiryInterval)*1000
	return time.Now().After(time.UnixMilli(expiryAt))
}

// SessionProperties represents the properties of the session (MQTT V5.0 only).
type SessionProperties struct {
	// Flags indicates which properties are present.
	Flags packet.PropertyFlags `json:"flags"`

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
