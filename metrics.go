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
	"encoding/json"
	"sync/atomic"
)

// Metrics contains the server metrics since the server has started.
type Metrics struct {
	// PacketReceived contains the number of packets received.
	PacketReceived MetricCounter `json:"packet_received"`

	// PacketSent contains the number of packets sent.
	PacketSent MetricCounter `json:"packet_sent"`

	// BytesReceived contains the number of bytes received.
	BytesReceived MetricCounter `json:"bytes_received"`

	// BytesSent contains the number of bytes sent.
	BytesSent MetricCounter `json:"bytes_sent"`

	// ClientsConnected contains the number of clients currently connected.
	ClientsConnected MetricGauge `json:"clients_connected"`

	// ClientMaximum contains the maximum number of clients connected at the same time.
	ClientsMaximum MetricCounter `json:"clients_maximum"`

	// PersistentSessions contains the number of persistent sessions (sessions stored on the session store).
	PersistentSessions MetricGauge `json:"persistent_sessions"`
}

// MetricCounter is a metric that only goes up.
type MetricCounter struct {
	// Avoid false sharing by adding 64-byte pads before and after the value.
	_pad0 [8]uint64 //nolint:unused
	value atomic.Uint64
	_pad1 [8]uint64 //nolint:unused
}

// Value returns the value of the metric.
func (m *MetricCounter) Value() uint64 {
	return m.value.Load()
}

// MarshalJSON returns a JSON encoding of the metric.
func (m *MetricCounter) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Value())
}

// MetricGauge is a metric that can go up and down.
type MetricGauge struct {
	// Avoid false sharing by adding 64-byte pads before and after the value.
	_pad0 [8]uint64 //nolint:unused
	value atomic.Int64
	_pad1 [8]uint64 //nolint:unused
}

// Value returns the value of the metric.
func (m *MetricGauge) Value() int64 {
	return m.value.Load()
}

// MarshalJSON returns a JSON encoding of the metric.
func (m *MetricGauge) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Value())
}
