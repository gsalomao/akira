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
	"testing"

	"github.com/stretchr/testify/suite"
)

type PacketConnectTestSuite struct {
	suite.Suite
}

func (s *PacketConnectTestSuite) TestType() {
	var p PacketConnect
	s.Require().Equal(PacketTypeConnect, p.Type())
}

func (s *PacketConnectTestSuite) TestSize() {
	testCases := []struct {
		name   string
		packet PacketConnect
		size   int
	}{
		{
			name: "V3.1, no will, no username, no password",
			packet: PacketConnect{
				Version:   MQTT31,
				KeepAlive: 0,
				ClientID:  []byte("a"),
			},
			size: 17,
		},
		{
			name: "V3.1.1, no will, no username, no password",
			packet: PacketConnect{
				Version:   MQTT311,
				KeepAlive: 30,
				ClientID:  []byte("ab"),
			},
			size: 16,
		},
		{
			name: "V5.0, no will, no username, no password, no properties",
			packet: PacketConnect{
				Version:   MQTT50,
				KeepAlive: 500,
				ClientID:  []byte("abc"),
			},
			size: 18,
		},
		{
			name: "V5.0, no will, no username, no password, empty properties",
			packet: PacketConnect{
				Version:        MQTT50,
				KeepAlive:      500,
				ClientID:       []byte("abc"),
				Properties:     &PropertiesConnect{},
				WillProperties: &PropertiesWill{},
			},
			size: 18,
		},
		{
			name: "V5.0, with will, no username, no password, and empty properties",
			packet: PacketConnect{
				Version:        MQTT50,
				ClientID:       []byte("abc"),
				Flags:          connectFlagWillFlag,
				Properties:     &PropertiesConnect{},
				WillTopic:      []byte("a"),
				WillPayload:    []byte("b"),
				WillProperties: &PropertiesWill{},
			},
			size: 25,
		},
		{
			name: "V5.0, with will, username, password, and properties",
			packet: PacketConnect{
				Version:  MQTT50,
				ClientID: []byte("abc"),
				Flags:    ConnectFlags(connectFlagWillFlag | connectFlagUsernameFlag | connectFlagPasswordFlag),
				Properties: &PropertiesConnect{
					Flags:                 propertyFlags(0).set(PropertySessionExpiryInterval),
					SessionExpiryInterval: 30,
				},
				WillTopic:   []byte("a"),
				WillPayload: []byte("b"),
				WillProperties: &PropertiesWill{
					Flags:             propertyFlags(0).set(PropertyWillDelayInterval),
					WillDelayInterval: 60,
				},
				Username: []byte("user"),
				Password: []byte("pass"),
			},
			size: 47,
		},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			size := test.packet.Size()
			s.Require().Equal(test.size, size)
		})
	}
}

func (s *PacketConnectTestSuite) TestDecodeSuccess() {
	testCases := []struct {
		name   string
		data   []byte
		packet PacketConnect
	}{
		{
			name: "V3.1",
			data: []byte{
				0, 6, 'M', 'Q', 'I', 's', 'd', 'p', // Protocol name
				3,     // Protocol version
				0,     // Packet flags
				0, 30, // Keep alive
				0, 1, 'a', // Client ID
			},
			packet: PacketConnect{
				Version:   MQTT31,
				KeepAlive: 30,
				ClientID:  []byte("a"),
			},
		},
		{
			name: "V3.1.1",
			data: []byte{
				0, 4, 'M', 'Q', 'T', 'T', // Protocol name
				4,      // Protocol version
				0,      // Packet flags
				0, 255, // Keep alive
				0, 2, 'a', 'b', // Client ID
			},
			packet: PacketConnect{
				Version:   MQTT311,
				KeepAlive: 255,
				ClientID:  []byte("ab"),
			},
		},
		{
			name: "V3.1.1, clean session",
			data: []byte{
				0, 4, 'M', 'Q', 'T', 'T', // Protocol name
				4,      // Protocol version
				2,      // Packet flags (Clean Session)
				0, 255, // Keep alive
				0, 2, 'a', 'b', // Client ID
			},
			packet: PacketConnect{
				Version:   MQTT311,
				KeepAlive: 255,
				Flags:     connectFlagCleanSession,
				ClientID:  []byte("ab"),
			},
		},
		{
			name: "V3.1.1, clean session + no client ID",
			data: []byte{
				0, 4, 'M', 'Q', 'T', 'T', // Protocol name
				4,      // Protocol version
				2,      // Packet flags (Clean Session)
				0, 255, // Keep alive
				0, 0, // Client ID
			},
			packet: PacketConnect{
				Version:   MQTT311,
				KeepAlive: 255,
				Flags:     connectFlagCleanSession,
				ClientID:  []byte{},
			},
		},
		{
			name: "V3.1.1, will flags",
			data: []byte{
				0, 4, 'M', 'Q', 'T', 'T', // Protocol name
				4,    // Protocol version
				0x2c, // Packet flags (Will Retain + Will QoS + Will Flag)
				1, 0, // Keep alive
				0, 2, 'a', 'b', // Client ID
				0, 1, 'a', // Will Topic
				0, 1, 'b', // Will Payload
			},
			packet: PacketConnect{
				Version:   MQTT311,
				KeepAlive: 256,
				Flags: ConnectFlags(
					connectFlagWillRetain | (QoS1 << connectFlagShiftWillQoS) | connectFlagWillFlag,
				),
				ClientID:    []byte("ab"),
				WillTopic:   []byte("a"),
				WillPayload: []byte("b"),
			},
		},
		{
			name: "V3.1.1, will flags + no will payload",
			data: []byte{
				0, 4, 'M', 'Q', 'T', 'T', // Protocol name
				4,    // Protocol version
				0x2c, // Packet flags (Will Retain + Will QoS + Will Flag)
				1, 0, // Keep alive
				0, 2, 'a', 'b', // Client ID
				0, 1, 'a', // Will Topic
				0, 0, // Will Payload
			},
			packet: PacketConnect{
				Version:   MQTT311,
				KeepAlive: 256,
				Flags: ConnectFlags(
					connectFlagWillRetain | (QoS1 << connectFlagShiftWillQoS) | connectFlagWillFlag,
				),
				ClientID:    []byte("ab"),
				WillTopic:   []byte("a"),
				WillPayload: []byte{},
			},
		},
		{
			name: "V3.1.1, username/password",
			data: []byte{
				0, 4, 'M', 'Q', 'T', 'T', // Protocol name
				4,    // Protocol version
				0xc0, // Packet flags (Username + Password)
				0, 1, // Keep alive
				0, 2, 'a', 'b', // Client ID
				0, 1, 'a', // Username
				0, 1, 'b', // Password
			},
			packet: PacketConnect{
				Version:   MQTT311,
				KeepAlive: 1,
				ClientID:  []byte("ab"),
				Flags:     ConnectFlags(connectFlagUsernameFlag | connectFlagPasswordFlag),
				Username:  []byte("a"),
				Password:  []byte("b"),
			},
		},
		{
			name: "V5.0, no properties",
			data: []byte{
				0, 4, 'M', 'Q', 'T', 'T', // Protocol name
				5,        // Protocol version
				0,        // Packet flags
				255, 255, // Keep alive
				0,                   // Properties Length
				0, 3, 'a', 'b', 'c', // Client ID
			},
			packet: PacketConnect{
				Version:   MQTT50,
				KeepAlive: 65535,
				ClientID:  []byte("abc"),
			},
		},
		{
			name: "V5.0, no properties, password",
			data: []byte{
				0, 4, 'M', 'Q', 'T', 'T', // Protocol name
				5,        // Protocol version
				0x40,     // Packet flags
				255, 255, // Keep alive
				0,                   // Properties Length
				0, 3, 'a', 'b', 'c', // Client ID
				0, 1, 'd', // Password
			},
			packet: PacketConnect{
				Version:   MQTT50,
				KeepAlive: 65535,
				ClientID:  []byte("abc"),
				Flags:     connectFlagPasswordFlag,
				Password:  []byte("d"),
			},
		},
		{
			name: "V5.0, properties",
			data: []byte{
				0, 4, 'M', 'Q', 'T', 'T', // Protocol name
				5,        // Protocol version
				0x04,     // Packet flags
				255, 255, // Keep alive
				5,               // Property length
				17, 0, 0, 0, 10, // SessionExpiryInterval
				0, 3, 'a', 'b', 'c', // Client ID
				5,               // Will Property length
				24, 0, 0, 0, 15, // WillDelayInterval
				0, 1, 'a', // Will Topic
				0, 1, 'b', // Will Payload
			},
			packet: PacketConnect{
				Version:   MQTT50,
				KeepAlive: 65535,
				Flags:     connectFlagWillFlag,
				Properties: &PropertiesConnect{
					Flags:                 propertyFlags(0).set(PropertySessionExpiryInterval),
					SessionExpiryInterval: 10,
				},
				ClientID: []byte("abc"),
				WillProperties: &PropertiesWill{
					Flags:             propertyFlags(0).set(PropertyWillDelayInterval),
					WillDelayInterval: 15,
				},
				WillTopic:   []byte("a"),
				WillPayload: []byte("b"),
			},
		},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			var packet PacketConnect
			header := FixedHeader{PacketType: PacketTypeConnect, RemainingLength: len(test.data)}

			n, err := packet.Decode(test.data, header)
			s.Require().NoError(err)
			s.Assert().Equal(test.packet, packet)
			s.Assert().Equal(len(test.data), n)
		})
	}
}

func (s *PacketConnectTestSuite) TestDecodeError() {
	testCases := []struct {
		name string
		data []byte
		err  error
	}{
		{
			name: "Missing protocol name",
			data: []byte{0, 4},
			err:  ErrMalformedProtocolName,
		},
		{
			name: "Invalid protocol name",
			data: []byte{0, 4, 'M', 'Q', 'T', 'T', 3},
			err:  ErrMalformedProtocolName,
		},
		{
			name: "Missing protocol version",
			data: []byte{0, 4, 'M', 'Q', 'T', 'T'},
			err:  ErrMalformedProtocolVersion,
		},
		{
			name: "Invalid protocol version",
			data: []byte{0, 4, 'M', 'Q', 'T', 'T', 0},
			err:  ErrMalformedProtocolVersion,
		},
		{
			name: "Missing CONNECT flags",
			data: []byte{0, 4, 'M', 'Q', 'T', 'T', 4},
			err:  ErrMalformedConnectFlags,
		},
		{
			name: "Invalid CONNECT flags",
			data: []byte{0, 4, 'M', 'Q', 'T', 'T', 4, 1},
			err:  ErrMalformedConnectFlags,
		},
		{
			name: "Invalid WillFlag and WillQoS combination",
			data: []byte{0, 4, 'M', 'Q', 'T', 'T', 4, 0x10},
			err:  ErrMalformedConnectFlags,
		},
		{
			name: "Invalid WillQoS",
			data: []byte{0, 4, 'M', 'Q', 'T', 'T', 4, 0x1c},
			err:  ErrMalformedConnectFlags,
		},
		{
			name: "Invalid WillFlag and WillRetain combination",
			data: []byte{0, 4, 'M', 'Q', 'T', 'T', 4, 0x20},
			err:  ErrMalformedConnectFlags,
		},
		{
			name: "Invalid Username and Password flags combination",
			data: []byte{0, 4, 'M', 'Q', 'T', 'T', 4, 0x40},
			err:  ErrMalformedConnectFlags,
		},
		{
			name: "Missing keep alive",
			data: []byte{0, 4, 'M', 'Q', 'T', 'T', 4, 0},
			err:  ErrMalformedKeepAlive,
		},
		{
			name: "Missing property length",
			data: []byte{0, 4, 'M', 'Q', 'T', 'T', 5, 0, 0, 10},
			err:  ErrMalformedPropertyLength,
		},
		{
			name: "Invalid property",
			data: []byte{0, 4, 'M', 'Q', 'T', 'T', 5, 0, 0, 10, 1, 255},
			err:  ErrMalformedPropertyInvalid,
		},
		{
			name: "Missing client ID",
			data: []byte{0, 4, 'M', 'Q', 'T', 'T', 4, 0, 0, 10},
			err:  ErrMalformedClientID,
		},
		{
			name: "Zero-byte Client ID with Clean Session flag (V3.1)",
			data: []byte{0, 6, 'M', 'Q', 'I', 's', 'd', 'p', 3, 2, 0, 10, 0, 0},
			err:  ErrV3ClientIDRejected,
		},
		{
			name: "Zero-byte Client ID without Clean Session flag (V3.1.1)",
			data: []byte{0, 4, 'M', 'Q', 'T', 'T', 4, 0, 0, 10, 0, 0},
			err:  ErrV3ClientIDRejected,
		},
		{
			name: "Client ID more than 23 bytes (V3.1)",
			data: []byte{0, 6, 'M', 'Q', 'I', 's', 'd', 'p', 3, 0, 0, 10, 0, 24, 65, 65, 65, 65, 65, 65, 65, 65, 65,
				65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65},
			err: ErrV3ClientIDRejected,
		},
		{
			name: "Missing Will property length",
			data: []byte{0, 4, 'M', 'Q', 'T', 'T', 5, 4, 0, 10, 0, 0, 0},
			err:  ErrMalformedPropertyLength,
		},
		{
			name: "Invalid Will property",
			data: []byte{0, 4, 'M', 'Q', 'T', 'T', 5, 4, 0, 10, 0, 0, 0, 1, 255},
			err:  ErrMalformedPropertyInvalid,
		},
		{
			name: "Missing Will topic",
			data: []byte{0, 4, 'M', 'Q', 'T', 'T', 4, 6, 0, 10, 0, 0},
			err:  ErrMalformedWillTopic,
		},
		{
			name: "Empty Will topic",
			data: []byte{0, 4, 'M', 'Q', 'T', 'T', 4, 6, 0, 10, 0, 0, 0, 0},
			err:  ErrMalformedWillTopic,
		},
		{
			name: "Invalid Will Topic",
			data: []byte{0, 4, 'M', 'Q', 'T', 'T', 4, 6, 0, 10, 0, 0, 0, 1, '#'},
			err:  ErrMalformedWillTopic,
		},
		{
			name: "Missing Will Payload",
			data: []byte{0, 4, 'M', 'Q', 'T', 'T', 4, 6, 0, 10, 0, 0, 0, 1, 'a'},
			err:  ErrMalformedWillPayload,
		},
		{
			name: "Missing username",
			data: []byte{0, 4, 'M', 'Q', 'T', 'T', 4, 0x82, 0, 10, 0, 0},
			err:  ErrMalformedUsername,
		},
		{
			name: "Missing password",
			data: []byte{0, 4, 'M', 'Q', 'T', 'T', 5, 0x40, 0, 10, 0, 0, 0},
			err:  ErrMalformedPassword,
		},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			var packet PacketConnect
			header := FixedHeader{PacketType: PacketTypeConnect, RemainingLength: len(test.data)}

			_, err := packet.Decode(test.data, header)
			s.Require().ErrorIs(err, test.err)
			s.Assert().NotEmpty(err.Error())
		})
	}
}

func (s *PacketConnectTestSuite) TestDecodeErrorInvalidHeader() {
	testCases := []struct {
		name   string
		header FixedHeader
		err    error
	}{
		{name: "Invalid packet type", header: FixedHeader{PacketType: PacketTypeReserved}, err: ErrMalformedPacketType},
		{name: "Invalid flags", header: FixedHeader{PacketType: PacketTypeConnect, Flags: 1}, err: ErrMalformedFlags},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			var packet PacketConnect

			_, err := packet.Decode(nil, test.header)
			s.Require().ErrorIs(err, test.err)
		})
	}
}

func TestPacketConnectTestSuite(t *testing.T) {
	suite.Run(t, new(PacketConnectTestSuite))
}

func BenchmarkPacketConnectSize(b *testing.B) {
	testCases := []struct {
		name   string
		packet PacketConnect
	}{
		{
			name: "V3",
			packet: PacketConnect{
				Version:   MQTT311,
				KeepAlive: 30,
				ClientID:  []byte("ab"),
			},
		},
		{
			name: "V5",
			packet: PacketConnect{
				Version:   MQTT50,
				KeepAlive: 500,
				ClientID:  []byte("abc"),
			},
		},
	}

	for _, test := range testCases {
		b.Run(test.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = test.packet.Size()
			}
		})
	}
}

func BenchmarkPacketConnectDecode(b *testing.B) {
	testCases := []struct {
		name string
		data []byte
	}{
		{
			name: "V3",
			data: []byte{
				0, 4, 'M', 'Q', 'T', 'T', // Protocol name
				4,      // Protocol version
				0,      // Packet flags
				0, 255, // Keep alive
				0, 2, 'a', 'b', // Client ID
			},
		},
		{
			name: "V5",
			data: []byte{
				0, 4, 'M', 'Q', 'T', 'T', // Protocol name
				5,        // Protocol version
				0,        // Packet flags
				255, 255, // Keep alive
				0,                   // Properties Length
				0, 3, 'a', 'b', 'c', // Client ID
			},
		},
	}

	for _, test := range testCases {
		b.Run(test.name, func(b *testing.B) {
			header := FixedHeader{PacketType: PacketTypeConnect, RemainingLength: len(test.data)}
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				packet := PacketConnect{}

				_, err := packet.Decode(test.data, header)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
