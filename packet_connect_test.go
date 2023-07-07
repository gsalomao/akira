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
	"fmt"
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
			"V3.1, no will, no username, no password",
			PacketConnect{Version: MQTT31, KeepAlive: 0, ClientID: []byte("a")},
			17,
		},
		{
			"V3.1.1, no will, no username, no password",
			PacketConnect{Version: MQTT311, KeepAlive: 30, ClientID: []byte("ab")},
			16,
		},
		{
			"V5.0, no will, no username, no password, no properties",
			PacketConnect{Version: MQTT50, KeepAlive: 500, ClientID: []byte("abc")},
			18,
		},
		{
			"V5.0, no will, no username, no password, empty properties",
			PacketConnect{
				Version:        MQTT50,
				KeepAlive:      500,
				ClientID:       []byte("abc"),
				Properties:     &PropertiesConnect{},
				WillProperties: &PropertiesWill{},
			},
			18,
		},
		{
			"V5.0, with will, no username, no password, and empty properties",
			PacketConnect{
				Version:        MQTT50,
				ClientID:       []byte("abc"),
				Flags:          connectFlagWillFlag,
				Properties:     &PropertiesConnect{},
				WillTopic:      []byte("a"),
				WillPayload:    []byte("b"),
				WillProperties: &PropertiesWill{},
			},
			25,
		},
		{
			"V5.0, with will, username, password, and properties",
			PacketConnect{
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
			47,
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
			"V3.1",
			[]byte{
				0, 6, 'M', 'Q', 'I', 's', 'd', 'p', // Protocol name
				3,     // Protocol version
				0,     // Packet flags
				0, 30, // Keep alive
				0, 1, 'a', // Client ID
			},
			PacketConnect{Version: MQTT31, KeepAlive: 30, ClientID: []byte("a")},
		},
		{
			"V3.1.1",
			[]byte{
				0, 4, 'M', 'Q', 'T', 'T', // Protocol name
				4,      // Protocol version
				0,      // Packet flags
				0, 255, // Keep alive
				0, 2, 'a', 'b', // Client ID
			},
			PacketConnect{Version: MQTT311, KeepAlive: 255, ClientID: []byte("ab")},
		},
		{
			"V3.1.1, clean session",
			[]byte{
				0, 4, 'M', 'Q', 'T', 'T', // Protocol name
				4,      // Protocol version
				2,      // Packet flags (Clean Session)
				0, 255, // Keep alive
				0, 2, 'a', 'b', // Client ID
			},
			PacketConnect{
				Version: MQTT311, KeepAlive: 255, Flags: connectFlagCleanSession, ClientID: []byte("ab"),
			},
		},
		{
			"V3.1.1, clean session + no client ID",
			[]byte{
				0, 4, 'M', 'Q', 'T', 'T', // Protocol name
				4,      // Protocol version
				2,      // Packet flags (Clean Session)
				0, 255, // Keep alive
				0, 0, // Client ID
			},
			PacketConnect{
				Version: MQTT311, KeepAlive: 255, Flags: connectFlagCleanSession, ClientID: []byte{},
			},
		},
		{
			"V3.1.1, will flags",
			[]byte{
				0, 4, 'M', 'Q', 'T', 'T', // Protocol name
				4,    // Protocol version
				0x2c, // Packet flags (Will Retain + Will QoS + Will Flag)
				1, 0, // Keep alive
				0, 2, 'a', 'b', // Client ID
				0, 1, 'a', // Will Topic
				0, 1, 'b', // Will Payload
			},
			PacketConnect{
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
			"V3.1.1, will flags + no will payload",
			[]byte{
				0, 4, 'M', 'Q', 'T', 'T', // Protocol name
				4,    // Protocol version
				0x2c, // Packet flags (Will Retain + Will QoS + Will Flag)
				1, 0, // Keep alive
				0, 2, 'a', 'b', // Client ID
				0, 1, 'a', // Will Topic
				0, 0, // Will Payload
			},
			PacketConnect{
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
			"V3.1.1, username/password",
			[]byte{
				0, 4, 'M', 'Q', 'T', 'T', // Protocol name
				4,    // Protocol version
				0xc0, // Packet flags (Username + Password)
				0, 1, // Keep alive
				0, 2, 'a', 'b', // Client ID
				0, 1, 'a', // Username
				0, 1, 'b', // Password
			},
			PacketConnect{
				Version:   MQTT311,
				KeepAlive: 1,
				ClientID:  []byte("ab"),
				Flags:     ConnectFlags(connectFlagUsernameFlag | connectFlagPasswordFlag),
				Username:  []byte("a"),
				Password:  []byte("b"),
			},
		},
		{
			"V5.0, no properties",
			[]byte{
				0, 4, 'M', 'Q', 'T', 'T', // Protocol name
				5,        // Protocol version
				0,        // Packet flags
				255, 255, // Keep alive
				0,                   // Properties Length
				0, 3, 'a', 'b', 'c', // Client ID
			},
			PacketConnect{Version: MQTT50, KeepAlive: 65535, ClientID: []byte("abc")},
		},
		{
			"V5.0, no properties, password",
			[]byte{
				0, 4, 'M', 'Q', 'T', 'T', // Protocol name
				5,        // Protocol version
				0x40,     // Packet flags
				255, 255, // Keep alive
				0,                   // Properties Length
				0, 3, 'a', 'b', 'c', // Client ID
				0, 1, 'd', // Password
			},
			PacketConnect{
				Version: MQTT50, KeepAlive: 65535, ClientID: []byte("abc"), Flags: connectFlagPasswordFlag,
				Password: []byte("d"),
			},
		},
		{
			"V5.0, properties",
			[]byte{
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
			PacketConnect{
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
		{"Missing protocol name", []byte{0, 4}, ErrMalformedProtocolName},
		{"Invalid protocol name", []byte{0, 4, 'M', 'Q', 'T', 'T', 3}, ErrMalformedProtocolName},
		{"Missing protocol version", []byte{0, 4, 'M', 'Q', 'T', 'T'}, ErrMalformedProtocolVersion},
		{"Invalid protocol version", []byte{0, 4, 'M', 'Q', 'T', 'T', 0}, ErrMalformedProtocolVersion},
		{"Missing CONNECT flags", []byte{0, 4, 'M', 'Q', 'T', 'T', 4}, ErrMalformedConnectFlags},
		{"Invalid CONNECT flags", []byte{0, 4, 'M', 'Q', 'T', 'T', 4, 1}, ErrMalformedConnectFlags},
		{
			"Invalid WillFlag and WillQoS combination",
			[]byte{0, 4, 'M', 'Q', 'T', 'T', 4, 0x10},
			ErrMalformedConnectFlags,
		},
		{"Invalid WillQoS", []byte{0, 4, 'M', 'Q', 'T', 'T', 4, 0x1c}, ErrMalformedConnectFlags},
		{
			"Invalid WillFlag and WillRetain combination",
			[]byte{0, 4, 'M', 'Q', 'T', 'T', 4, 0x20},
			ErrMalformedConnectFlags,
		},
		{
			"Invalid Username and Password flags combination",
			[]byte{0, 4, 'M', 'Q', 'T', 'T', 4, 0x40},
			ErrMalformedConnectFlags,
		},
		{"Missing keep alive", []byte{0, 4, 'M', 'Q', 'T', 'T', 4, 0}, ErrMalformedKeepAlive},
		{"Missing property length", []byte{0, 4, 'M', 'Q', 'T', 'T', 5, 0, 0, 10}, ErrMalformedPropertyLength},
		{"Invalid property", []byte{0, 4, 'M', 'Q', 'T', 'T', 5, 0, 0, 10, 1, 255}, ErrMalformedPropertyInvalid},
		{"Missing client ID", []byte{0, 4, 'M', 'Q', 'T', 'T', 4, 0, 0, 10}, ErrMalformedClientID},
		{
			"Zero-byte Client ID with Clean Session flag (V3.1)",
			[]byte{0, 6, 'M', 'Q', 'I', 's', 'd', 'p', 3, 2, 0, 10, 0, 0},
			ErrV3ClientIDRejected,
		},
		{
			"Zero-byte Client ID without Clean Session flag (V3.1.1)",
			[]byte{0, 4, 'M', 'Q', 'T', 'T', 4, 0, 0, 10, 0, 0},
			ErrV3ClientIDRejected,
		},
		{
			"Client ID more than 23 bytes (V3.1)",
			[]byte{
				0, 6, 'M', 'Q', 'I', 's', 'd', 'p', 3, 0, 0, 10, 0, 24, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65,
				65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65,
			},
			ErrV3ClientIDRejected,
		},
		{
			"Missing Will property length",
			[]byte{0, 4, 'M', 'Q', 'T', 'T', 5, 4, 0, 10, 0, 0, 0},
			ErrMalformedPropertyLength,
		},
		{
			"Invalid Will property",
			[]byte{0, 4, 'M', 'Q', 'T', 'T', 5, 4, 0, 10, 0, 0, 0, 1, 255},
			ErrMalformedPropertyInvalid,
		},
		{"Missing Will topic", []byte{0, 4, 'M', 'Q', 'T', 'T', 4, 6, 0, 10, 0, 0}, ErrMalformedWillTopic},
		{"Empty Will topic", []byte{0, 4, 'M', 'Q', 'T', 'T', 4, 6, 0, 10, 0, 0, 0, 0}, ErrMalformedWillTopic},
		{"Invalid Will Topic", []byte{0, 4, 'M', 'Q', 'T', 'T', 4, 6, 0, 10, 0, 0, 0, 1, '#'}, ErrMalformedWillTopic},
		{
			"Missing Will Payload",
			[]byte{0, 4, 'M', 'Q', 'T', 'T', 4, 6, 0, 10, 0, 0, 0, 1, 'a'},
			ErrMalformedWillPayload,
		},
		{"Missing username", []byte{0, 4, 'M', 'Q', 'T', 'T', 4, 0x82, 0, 10, 0, 0}, ErrMalformedUsername},
		{"Missing password", []byte{0, 4, 'M', 'Q', 'T', 'T', 5, 0x40, 0, 10, 0, 0, 0}, ErrMalformedPassword},
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
		{"Invalid packet type", FixedHeader{PacketType: PacketTypeReserved}, ErrMalformedPacketType},
		{"Invalid flags", FixedHeader{PacketType: PacketTypeConnect, Flags: 1}, ErrMalformedFlags},
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
		{"V3", PacketConnect{Version: MQTT311, KeepAlive: 30, ClientID: []byte("ab")}},
		{"V5", PacketConnect{Version: MQTT50, KeepAlive: 500, ClientID: []byte("abc")}},
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
		{"V3", []byte{0, 4, 'M', 'Q', 'T', 'T', 4, 0, 0, 255, 0, 2, 'a', 'b'}},
		{"V5", []byte{0, 4, 'M', 'Q', 'T', 'T', 5, 0, 255, 255, 0, 0, 3, 'a', 'b', 'c'}},
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

type PropertiesConnectTestSuite struct {
	suite.Suite
}

func (s *PropertiesConnectTestSuite) TestHas() {
	testCases := []struct {
		props  *PropertiesConnect
		prop   Property
		result bool
	}{
		{&PropertiesConnect{}, PropertyUserProperty, true},
		{&PropertiesConnect{}, PropertyAuthenticationMethod, true},
		{&PropertiesConnect{}, PropertyAuthenticationData, true},
		{&PropertiesConnect{}, PropertySessionExpiryInterval, true},
		{&PropertiesConnect{}, PropertyMaximumPacketSize, true},
		{&PropertiesConnect{}, PropertyReceiveMaximum, true},
		{&PropertiesConnect{}, PropertyTopicAliasMaximum, true},
		{&PropertiesConnect{}, PropertyRequestResponseInfo, true},
		{&PropertiesConnect{}, PropertyRequestProblemInfo, true},
		{nil, 0, false},
	}

	for _, test := range testCases {
		s.Run(fmt.Sprint(test.prop), func() {
			test.props.Set(test.prop)
			s.Require().Equal(test.result, test.props.Has(test.prop))
		})
	}
}

func (s *PropertiesConnectTestSuite) TestSize() {
	var flags propertyFlags

	flags = flags.set(PropertyUserProperty)
	flags = flags.set(PropertyAuthenticationMethod)
	flags = flags.set(PropertyAuthenticationData)
	flags = flags.set(PropertySessionExpiryInterval)
	flags = flags.set(PropertyMaximumPacketSize)
	flags = flags.set(PropertyReceiveMaximum)
	flags = flags.set(PropertyTopicAliasMaximum)
	flags = flags.set(PropertyRequestResponseInfo)
	flags = flags.set(PropertyRequestProblemInfo)

	props := PropertiesConnect{
		Flags: flags,
		UserProperties: []UserProperty{
			{[]byte("a"), []byte("b")},
			{[]byte("c"), []byte("d")},
		},
		AuthenticationMethod: []byte("auth"),
		AuthenticationData:   []byte("data"),
	}

	size := props.size()
	s.Assert().Equal(48, size)
}

func (s *PropertiesConnectTestSuite) TestSizeOnNil() {
	var props *PropertiesConnect

	size := props.size()
	s.Assert().Equal(0, size)
}

func (s *PropertiesConnectTestSuite) TestDecodeSuccess() {
	data := []byte{
		0,               // Property Length
		17, 0, 0, 0, 10, // Session Expiry Interval
		33, 0, 50, // Receive Maximum
		39, 0, 0, 0, 200, // Maximum Packet Size
		34, 0, 50, // Topic Alias Maximum
		25, 1, // Request Response Info
		23, 0, // Request Problem Info
		38, 0, 1, 'a', 0, 1, 'b', // User Property
		38, 0, 1, 'c', 0, 1, 'd', // User Property
		21, 0, 2, 'e', 'f', // Authentication Method
		22, 0, 1, 10, // Authentication Data
	}
	data[0] = byte(len(data) - 1)

	props, n, err := decodeProperties[PropertiesConnect](data)
	s.Require().NoError(err)
	s.Assert().Equal(len(data), n)
	s.Assert().True(props.Has(PropertySessionExpiryInterval))
	s.Assert().True(props.Has(PropertyReceiveMaximum))
	s.Assert().True(props.Has(PropertyMaximumPacketSize))
	s.Assert().True(props.Has(PropertyTopicAliasMaximum))
	s.Assert().True(props.Has(PropertyRequestResponseInfo))
	s.Assert().True(props.Has(PropertyRequestProblemInfo))
	s.Assert().True(props.Has(PropertyAuthenticationMethod))
	s.Assert().True(props.Has(PropertyAuthenticationData))
	s.Assert().True(props.Has(PropertyUserProperty))
	s.Assert().Equal(10, int(props.SessionExpiryInterval))
	s.Assert().Equal(50, int(props.ReceiveMaximum))
	s.Assert().Equal(200, int(props.MaximumPacketSize))
	s.Assert().Equal(50, int(props.TopicAliasMaximum))
	s.Assert().True(props.RequestResponseInfo)
	s.Assert().False(props.RequestProblemInfo)
	s.Assert().Equal([]byte("a"), props.UserProperties[0].Key)
	s.Assert().Equal([]byte("b"), props.UserProperties[0].Value)
	s.Assert().Equal([]byte("c"), props.UserProperties[1].Key)
	s.Assert().Equal([]byte("d"), props.UserProperties[1].Value)
	s.Assert().Equal([]byte("ef"), props.AuthenticationMethod)
	s.Assert().Equal([]byte{10}, props.AuthenticationData)
}

func (s *PropertiesTestSuite) TestDecodeError() {
	testCases := []struct {
		name string
		data []byte
		err  error
	}{
		{"No property", []byte{1}, ErrMalformedPropertyConnect},
		{"Missing Session Expiry Interval", []byte{1, 17}, ErrMalformedPropertySessionExpiryInterval},
		{"Session Expiry Interval - uint", []byte{2, 17, 0}, ErrMalformedPropertySessionExpiryInterval},
		{
			"Duplicated Session Expiry Interval",
			[]byte{10, 17, 0, 0, 0, 10, 17, 0, 0, 0, 11},
			ErrMalformedPropertySessionExpiryInterval,
		},
		{"Missing Receive Max", []byte{1, 33}, ErrMalformedPropertyReceiveMaximum},
		{"Invalid Receive Max", []byte{3, 33, 0, 0}, ErrMalformedPropertyReceiveMaximum},
		{"Duplicated Receive Max", []byte{6, 33, 0, 50, 33, 0, 51}, ErrMalformedPropertyReceiveMaximum},
		{"Missing Maximum Packet Size", []byte{1, 39}, ErrMalformedPropertyMaxPacketSize},
		{"Invalid Maximum Packet Size", []byte{5, 39, 0, 0, 0, 0}, ErrMalformedPropertyMaxPacketSize},
		{
			"Duplicated Maximum Packet Size",
			[]byte{10, 39, 0, 0, 0, 200, 39, 0, 0, 0, 201},
			ErrMalformedPropertyMaxPacketSize,
		},
		{"Missing Topic Alias Max", []byte{1, 34}, ErrMalformedPropertyTopicAliasMaximum},
		{"Duplicated Topic Alias Max", []byte{6, 34, 0, 50, 34, 0, 51}, ErrMalformedPropertyTopicAliasMaximum},
		{"Missing Request Response Info", []byte{1, 25}, ErrMalformedPropertyRequestResponseInfo},
		{"Invalid Request Response Info", []byte{2, 25, 2}, ErrMalformedPropertyRequestResponseInfo},
		{"Duplicated Request Response Info", []byte{4, 25, 0, 25, 1}, ErrMalformedPropertyRequestResponseInfo},
		{"Missing Request Problem Info", []byte{1, 23}, ErrMalformedPropertyRequestProblemInfo},
		{"Invalid Request Problem Info", []byte{2, 23, 2}, ErrMalformedPropertyRequestProblemInfo},
		{"Duplicated Request Problem Info", []byte{4, 23, 0, 23, 1}, ErrMalformedPropertyRequestProblemInfo},
		{"Missing User Property", []byte{1, 38}, ErrMalformedPropertyUserProperty},
		{"Missing User Property Value", []byte{4, 38, 0, 1, 'a'}, ErrMalformedPropertyUserProperty},
		{"User Prop Value - Incomplete str", []byte{4, 38, 0, 5, 'a'}, ErrMalformedPropertyUserProperty},
		{"Missing Authentication Method", []byte{1, 21}, ErrMalformedPropertyAuthenticationMethod},
		{
			"Duplicated Auth Method",
			[]byte{10, 21, 0, 2, 'a', 'b', 21, 0, 2, 'c', 'd'},
			ErrMalformedPropertyAuthenticationMethod,
		},
		{"Missing Auth Data", []byte{1, 22}, ErrMalformedPropertyAuthenticationData},
		{"Duplicated Auth Data", []byte{8, 22, 0, 1, 10, 22, 0, 1, 11}, ErrMalformedPropertyAuthenticationData},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			_, _, err := decodeProperties[PropertiesConnect](test.data)
			s.Require().ErrorIs(err, test.err)
		})
	}
}

func TestPropertiesConnectTestSuite(t *testing.T) {
	suite.Run(t, new(PropertiesConnectTestSuite))
}

type PropertiesWillTestSuite struct {
	suite.Suite
}

func (s *PropertiesWillTestSuite) TestHas() {
	testCases := []struct {
		props  *PropertiesWill
		prop   Property
		result bool
	}{
		{&PropertiesWill{}, PropertyUserProperty, true},
		{&PropertiesWill{}, PropertyCorrelationData, true},
		{&PropertiesWill{}, PropertyContentType, true},
		{&PropertiesWill{}, PropertyResponseTopic, true},
		{&PropertiesWill{}, PropertyWillDelayInterval, true},
		{&PropertiesWill{}, PropertyMessageExpiryInterval, true},
		{&PropertiesWill{}, PropertyPayloadFormatIndicator, true},
		{nil, 0, false},
	}

	for _, test := range testCases {
		s.Run(fmt.Sprint(test.prop), func() {
			test.props.Set(test.prop)
			s.Require().Equal(test.result, test.props.Has(test.prop))
		})
	}
}

func (s *PropertiesWillTestSuite) TestSize() {
	var flags propertyFlags
	flags = flags.set(PropertyUserProperty)
	flags = flags.set(PropertyCorrelationData)
	flags = flags.set(PropertyContentType)
	flags = flags.set(PropertyResponseTopic)
	flags = flags.set(PropertyWillDelayInterval)
	flags = flags.set(PropertyMessageExpiryInterval)
	flags = flags.set(PropertyPayloadFormatIndicator)

	props := PropertiesWill{
		Flags:                  flags,
		UserProperties:         []UserProperty{{[]byte("a"), []byte("b")}},
		CorrelationData:        []byte{20, 1},
		ContentType:            []byte("json"),
		ResponseTopic:          []byte("b"),
		WillDelayInterval:      10,
		MessageExpiryInterval:  100,
		PayloadFormatIndicator: true,
	}

	size := props.size()
	s.Assert().Equal(35, size)
}

func (s *PropertiesWillTestSuite) TestSizeOnNil() {
	var props *PropertiesWill

	size := props.size()
	s.Assert().Equal(0, size)
}

func (s *PropertiesWillTestSuite) TestDecodeSuccess() {
	data := []byte{
		0,               // Property Length
		24, 0, 0, 0, 15, // Will Delay Interval
		1, 1, // Payload Format Indicator
		2, 0, 0, 0, 10, // Message Expiry Interval
		3, 0, 4, 'j', 's', 'o', 'n', // Content Type
		8, 0, 1, 'b', // Response Topic
		9, 0, 2, 20, 1, // Correlation Data
		38, 0, 1, 'a', 0, 1, 'b', // User Property
	}
	data[0] = byte(len(data) - 1)

	props, n, err := decodeProperties[PropertiesWill](data)
	s.Require().NoError(err)
	s.Assert().Equal(len(data), n)
	s.Assert().True(props.Has(PropertyWillDelayInterval))
	s.Assert().True(props.Has(PropertyPayloadFormatIndicator))
	s.Assert().True(props.Has(PropertyMessageExpiryInterval))
	s.Assert().True(props.Has(PropertyContentType))
	s.Assert().True(props.Has(PropertyResponseTopic))
	s.Assert().True(props.Has(PropertyCorrelationData))
	s.Assert().True(props.Has(PropertyUserProperty))
	s.Assert().Equal(15, int(props.WillDelayInterval))
	s.Assert().True(props.PayloadFormatIndicator)
	s.Assert().Equal(10, int(props.MessageExpiryInterval))
	s.Assert().Equal([]byte("json"), props.ContentType)
	s.Assert().Equal([]byte("b"), props.ResponseTopic)
	s.Assert().Equal([]byte{20, 1}, props.CorrelationData)
	s.Assert().Equal([]byte("a"), props.UserProperties[0].Key)
	s.Assert().Equal([]byte("b"), props.UserProperties[0].Value)
}

func (s *PropertiesWillTestSuite) TestDecodeError() {
	testCases := []struct {
		name string
		data []byte
		err  error
	}{
		{"No property", []byte{1}, ErrMalformedPropertyWill},
		{"Missing Will Delay Interval", []byte{1, 24}, ErrMalformedPropertyWillDelayInterval},
		{
			"Duplicated Will Delay Interval",
			[]byte{10, 24, 0, 0, 0, 15, 24, 0, 0, 0, 16},
			ErrMalformedPropertyWillDelayInterval,
		},
		{"Missing Payload Format Indicator", []byte{1, 1}, ErrMalformedPropertyPayloadFormatIndicator},
		{"Invalid Payload Format Indicator", []byte{2, 1, 2}, ErrMalformedPropertyPayloadFormatIndicator},
		{"Duplicated Payload Format Indicator", []byte{4, 1, 0, 1, 1}, ErrMalformedPropertyPayloadFormatIndicator},
		{"Missing Message Expiry Interval", []byte{1, 2}, ErrMalformedPropertyMessageExpiryInterval},
		{
			"Duplicated Message Expiry Interval",
			[]byte{10, 2, 0, 0, 0, 10, 2, 0, 0, 0, 11},
			ErrMalformedPropertyMessageExpiryInterval,
		},
		{"Missing Content Type", []byte{1, 3}, ErrMalformedPropertyContentType},
		{"Content Type - Missing string", []byte{3, 3, 0, 4}, ErrMalformedPropertyContentType},
		{
			"Duplicated Content Type",
			[]byte{13, 3, 0, 4, 'j', 's', 'o', 'n', 3, 0, 3, 'x', 'm', 'l'},
			ErrMalformedPropertyContentType,
		},
		{"Missing Response Topic", []byte{1, 8}, ErrMalformedPropertyResponseTopic},
		{"Response Topic - Missing string", []byte{3, 8, 0, 0}, ErrMalformedPropertyResponseTopic},
		{"Response Topic - Incomplete string", []byte{3, 8, 0, 1}, ErrMalformedPropertyResponseTopic},
		{"Invalid Response Topic", []byte{4, 8, 0, 1, '#'}, ErrMalformedPropertyResponseTopic},
		{"Duplicated Response Topic", []byte{8, 8, 0, 1, 'b', 8, 0, 1, 'c'}, ErrMalformedPropertyResponseTopic},
		{"Missing Correlation Data", []byte{1, 9}, ErrMalformedPropertyCorrelationData},
		{"Correlation Data - Missing data", []byte{3, 9, 0, 2}, ErrMalformedPropertyCorrelationData},
		{
			"Duplicated Correlation Data",
			[]byte{10, 9, 0, 2, 20, 1, 9, 0, 2, 20, 2},
			ErrMalformedPropertyCorrelationData,
		},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			_, _, err := decodeProperties[PropertiesWill](test.data)
			s.Require().ErrorIs(err, test.err)
		})
	}
}

func TestPropertiesWillTestSuite(t *testing.T) {
	suite.Run(t, new(PropertiesWillTestSuite))
}
