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
	"bytes"
	"io"
	"testing"

	"github.com/gsalomao/akira/packet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadPacketSuccess(t *testing.T) {
	testCases := []struct {
		name   string
		data   []byte
		packet Packet
	}{
		{
			name: "CONNECT - V3",
			data: []byte{
				byte(packet.TypeConnect) << 4, 14, // Fixed header.
				0, 4, 'M', 'Q', 'T', 'T', // Protocol name.
				4,      // Protocol version.
				2,      // Packet flags (Clean session).
				0, 255, // Keep alive.
				0, 2, 'a', 'b', // Client ID.
			},
			packet: &packet.Connect{
				Version:   packet.MQTT311,
				KeepAlive: 255,
				Flags:     packet.ConnectFlags(0x02), // Clean session flag.
				ClientID:  []byte("ab"),
			},
		},
		{
			name: "CONNECT - V5",
			data: []byte{
				byte(packet.TypeConnect) << 4, 20, // Fixed header.
				0, 4, 'M', 'Q', 'T', 'T', // Protocol name.
				5,      // Protocol version.
				2,      // Packet flags (Clean session).
				0, 255, // Keep alive.
				5,               // Property length.
				17, 0, 0, 0, 30, // Session Expiry Interval.
				0, 2, 'a', 'b', // Client ID.
			},
			packet: &packet.Connect{
				Version:   packet.MQTT50,
				KeepAlive: 255,
				Flags:     packet.ConnectFlags(0x02), // Clean session flag.
				ClientID:  []byte("ab"),
				Properties: &packet.ConnectProperties{
					Flags:                 packet.PropertyFlags(0).Set(packet.PropertySessionExpiryInterval),
					SessionExpiryInterval: 30,
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			reader := bufio.NewReader(bytes.NewReader(test.data))

			p, n, err := readPacket(reader)
			require.NoError(t, err)
			assert.Equal(t, len(test.data), n)
			assert.Equal(t, test.packet, p)
		})
	}
}

func TestReadPacketError(t *testing.T) {
	testCases := []struct {
		name string
		data []byte
		err  error
	}{
		{name: "Invalid packet type", data: []byte{0, 0}, err: packet.ErrProtocolError},
		{name: "Invalid packet", data: []byte{byte(packet.TypeConnect) << 4, 0}, err: packet.ErrMalformedPacket},
		{name: "Missing remaining length", data: []byte{byte(packet.TypeConnect) << 4, 10}, err: io.EOF},
		{
			name: "Unexpected packet length",
			data: []byte{0x10, 15, 0, 4, 'M', 'Q', 'T', 'T', 4, 2, 0, 255, 0, 2, 'a', 'b', 0},
			err:  packet.ErrMalformedPacket,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			reader := bufio.NewReader(bytes.NewReader(test.data))

			_, _, err := readPacket(reader)
			require.ErrorIs(t, err, test.err)
		})
	}
}

func BenchmarkReadPacket(b *testing.B) {
	testCases := []struct {
		name string
		data []byte
	}{
		{
			name: "CONNECT-V3",
			data: []byte{
				byte(packet.TypeConnect) << 4, 14, // Fixed header.
				0, 4, 'M', 'Q', 'T', 'T', // Protocol name.
				4,      // Protocol version.
				2,      // Packet flags (Clean Session).
				0, 255, // Keep alive.
				0, 2, 'a', 'b', // Client ID.
			},
		},
		{
			name: "CONNECT-V5",
			data: []byte{
				byte(packet.TypeConnect) << 4, 15, // Fixed header.
				0, 4, 'M', 'Q', 'T', 'T', // Protocol name.
				5,      // Protocol version.
				2,      // Packet flags (Clean Session).
				0, 255, // Keep alive.
				0,              // Property length.
				0, 2, 'a', 'b', // Client ID.
			},
		},
	}

	for _, test := range testCases {
		b.Run(test.name, func(b *testing.B) {
			data := bytes.NewReader(test.data)
			reader := bufio.NewReaderSize(data, 1024)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, _, err := readPacket(reader)
				if err != nil {
					b.Fatal(err)
				}

				data.Reset(test.data)
				reader.Reset(data)
			}
		})
	}
}
