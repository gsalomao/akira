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

// Options contains the Server options to be used.
type Options struct {
	// Config contains the server configuration.
	Config *Config

	// Listeners is the list of Listener to be added into the server.
	Listeners []Listener

	// Hooks is the list of Hook to be added into the server.
	Hooks []Hook
}

// NewDefaultOptions creates a default Options.
func NewDefaultOptions() *Options {
	return &Options{
		Config: NewDefaultConfig(),
	}
}

// Config contains the Server configuration.
type Config struct {
	// OutboundStreamSize is the number of bytes of each Client's outbound stream.
	OutboundStreamSize uint32 `json:"outbound_stream_size"`

	// ReadBufferSize is the number of bytes for the read buffer.
	ReadBufferSize uint32 `json:"read_buffer_size"`

	// ConnectTimeout is the number of seconds to wait for the CONNECT Packet after the Client has established the
	// network connection.
	ConnectTimeout uint16 `json:"connect_timeout"`
}

// NewDefaultConfig creates a default Config.
func NewDefaultConfig() *Config {
	c := Config{
		OutboundStreamSize: 8 * 1024,
		ReadBufferSize:     1024,
		ConnectTimeout:     10,
	}
	return &c
}
