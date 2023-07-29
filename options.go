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

import "github.com/gsalomao/akira/packet"

// Options contains the Server options to be used.
type Options struct {
	// Config contains the server configuration.
	Config *Config

	// Listeners is the list of Listener to be added into the server.
	Listeners []Listener

	// Hooks is the list of Hook to be added into the server.
	Hooks []Hook

	// SessionStore is the store responsible for persist the Session in a non-volatile memory.
	SessionStore SessionStore
}

// OptionsFunc is the function called by NewServer factory method to set the Options.
type OptionsFunc func(opts *Options)

// WithConfig sets the config into the Options.
func WithConfig(c *Config) OptionsFunc {
	return func(opts *Options) {
		opts.Config = c
	}
}

// WithListeners sets the listeners into the Options.
func WithListeners(l []Listener) OptionsFunc {
	return func(opts *Options) {
		opts.Listeners = l
	}
}

// WithHooks sets the hooks into the Options.
func WithHooks(h []Hook) OptionsFunc {
	return func(opts *Options) {
		opts.Hooks = h
	}
}

// WithSessionStore sets the session store into the Options.
func WithSessionStore(s SessionStore) OptionsFunc {
	return func(opts *Options) {
		opts.SessionStore = s
	}
}

// NewDefaultOptions creates a default Options.
func NewDefaultOptions() *Options {
	return &Options{Config: NewDefaultConfig()}
}

// Config contains the Server configuration.
type Config struct {
	// MaxClientIDSize is the maximum length for client identifier, in bytes, allowed by the server.
	MaxClientIDSize int `json:"max_client_id_size"`

	// ReadBufferSize is the number of bytes for the read buffer.
	ReadBufferSize int `json:"read_buffer_size"`

	// MaxPacketSize indicates the maximum packet size, in bytes, allowed by the server.
	MaxPacketSize uint32 `json:"max_packet_size"`

	// MaxSessionExpiryIntervalSec indicates the maximum session expire interval, in seconds, allowed by the server.
	MaxSessionExpiryIntervalSec uint32 `json:"max_session_expiry_interval_sec"`

	// ConnectTimeoutMs is the maximum number of milliseconds to wait for the CONNECT Packet after the Client
	// has established the network connection.
	ConnectTimeoutMs uint32 `json:"connect_timeout_ms"`

	// MaxKeepAliveSec is the maximum Keep Alive value, in seconds, allowed by the server.
	MaxKeepAliveSec uint16 `json:"max_keep_alive_sec"`

	// MaxInflightMessages indicates maximum number of QoS 1 or 2 messages that can be processed simultaneously per
	// client.
	MaxInflightMessages uint16 `json:"max_inflight_messages"`

	// TopicAliasMax indicates the highest value that the sever accepts as a topic alias sent by the client. The server
	// uses this value to limit the number of topic aliases that it's willing to hold for each client.
	TopicAliasMax uint16 `json:"topic_alias_max"`

	// MaxQoS indicates the maximum QoS for PUBLISH Packets accepted by the server.
	MaxQoS byte `json:"max_qos"`

	// RetainAvailable indicates whether the server allows retained messages or not.
	RetainAvailable bool `json:"retain_available"`

	// WildcardSubscriptionAvailable indicates whether the server allows wildcard subscription or not.
	WildcardSubscriptionAvailable bool `json:"wildcard_subscription_available"`

	// SubscriptionIDAvailable indicates whether the server allows subscription identifier or not.
	SubscriptionIDAvailable bool `json:"subscription_id_available"`

	// SharedSubscriptionAvailable indicates whether the server allows shared subscription or not.
	SharedSubscriptionAvailable bool `json:"shared_subscription_available"`
}

// NewDefaultConfig creates a default Config.
func NewDefaultConfig() *Config {
	c := Config{
		MaxClientIDSize:               23,
		ReadBufferSize:                1024,
		ConnectTimeoutMs:              10000,
		MaxQoS:                        byte(packet.QoS2),
		RetainAvailable:               true,
		WildcardSubscriptionAvailable: true,
		SubscriptionIDAvailable:       true,
		SharedSubscriptionAvailable:   true,
	}
	return &c
}
