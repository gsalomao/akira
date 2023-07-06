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

type PacketConnAckTestSuite struct {
	suite.Suite
}

func (s *PacketConnAckTestSuite) TestType() {
	var p PacketConnAck
	s.Require().Equal(PacketTypeConnAck, p.Type())
}

func (s *PacketConnAckTestSuite) TestSize() {
	testCases := []struct {
		name   string
		packet PacketConnAck
		size   int
	}{
		{
			name:   "V3.1",
			packet: PacketConnAck{Version: MQTT31},
			size:   4,
		},
		{
			name:   "V3.1.1",
			packet: PacketConnAck{Version: MQTT311, Code: ReasonCodeIdentifierRejected, SessionPresent: true},
			size:   4,
		},
		{
			name:   "V3.1.1",
			packet: PacketConnAck{Version: MQTT311, Code: ReasonCodeIdentifierRejected, SessionPresent: true},
			size:   4,
		},
		{
			name:   "V5.0, no properties",
			packet: PacketConnAck{Version: MQTT50},
			size:   5,
		},
		{
			name: "V5.0, empty properties",
			packet: PacketConnAck{
				Version:    MQTT50,
				Properties: &PropertiesConnAck{},
			},
			size: 5,
		},
		{
			name: "V5.0, with properties",
			packet: PacketConnAck{
				Version: MQTT50,
				Properties: &PropertiesConnAck{
					Flags:                 propertyFlags(0).set(PropertySessionExpiryInterval),
					SessionExpiryInterval: 30,
				},
			},
			size: 10,
		},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			size := test.packet.Size()
			s.Assert().Equal(test.size, size)
		})
	}
}

func (s *PacketConnAckTestSuite) TestEncodeSuccess() {
	testCases := []struct {
		name   string
		packet PacketConnAck
		data   []byte
	}{
		{
			name:   "V3.1",
			packet: PacketConnAck{Version: MQTT31, SessionPresent: true, Code: ReasonCodeConnectionAccepted},
			data:   []byte{0x20, 2, 1, 0},
		},
		{
			name:   "V3.1.1",
			packet: PacketConnAck{Version: MQTT311, Code: ReasonCodeIdentifierRejected},
			data:   []byte{0x20, 2, 0, 2},
		},
		{
			name:   "V5.0, no properties",
			packet: PacketConnAck{Version: MQTT50, Code: ReasonCodeSuccess},
			data:   []byte{0x20, 3, 0, 0, 0},
		},
		{
			name: "V5.0, with properties",
			packet: PacketConnAck{
				Version: MQTT50,
				Code:    ReasonCodeMalformedPacket,
				Properties: &PropertiesConnAck{
					Flags:                 propertyFlags(0).set(PropertySessionExpiryInterval),
					SessionExpiryInterval: 10,
				},
			},
			data: []byte{0x20, 8, 0, 0x81, 5, 0x11, 0, 0, 0, 10},
		},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			data := make([]byte, test.packet.Size())

			n, err := test.packet.Encode(data)
			s.Require().NoError(err)
			s.Assert().Equal(len(test.data), n)
			s.Assert().Equal(test.data, data)
		})
	}
}

func (s *PacketConnAckTestSuite) TestEncodeError() {
	packet := PacketConnAck{Version: MQTT50, Code: ReasonCodeSuccess}

	n, err := packet.Encode(nil)
	s.Require().Error(err)
	s.Assert().Zero(n)
}

func TestPacketConnAckTestSuite(t *testing.T) {
	suite.Run(t, new(PacketConnAckTestSuite))
}

func BenchmarkPacketConnAckEncode(b *testing.B) {
	testCases := []struct {
		name   string
		packet PacketConnAck
	}{
		{
			name:   "V3",
			packet: PacketConnAck{Version: MQTT311, Code: ReasonCodeConnectionAccepted},
		},
		{
			name:   "V5",
			packet: PacketConnAck{Version: MQTT50, Code: ReasonCodeSuccess},
		},
	}

	for _, test := range testCases {
		b.Run(test.name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				data := make([]byte, test.packet.Size())

				_, err := test.packet.Encode(data)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

type PropertiesConnAckTestSuite struct {
	suite.Suite
}

func (s *PropertiesConnAckTestSuite) TestHas() {
	testCases := []struct {
		props  *PropertiesConnAck
		prop   Property
		result bool
	}{
		{&PropertiesConnAck{}, PropertyUserProperty, true},
		{&PropertiesConnAck{}, PropertyAssignedClientID, true},
		{&PropertiesConnAck{}, PropertyReasonString, true},
		{&PropertiesConnAck{}, PropertyResponseInfo, true},
		{&PropertiesConnAck{}, PropertyServerReference, true},
		{&PropertiesConnAck{}, PropertyAuthenticationMethod, true},
		{&PropertiesConnAck{}, PropertyAuthenticationData, true},
		{&PropertiesConnAck{}, PropertySessionExpiryInterval, true},
		{&PropertiesConnAck{}, PropertyMaximumPacketSize, true},
		{&PropertiesConnAck{}, PropertyReceiveMaximum, true},
		{&PropertiesConnAck{}, PropertyTopicAliasMaximum, true},
		{&PropertiesConnAck{}, PropertyServerKeepAlive, true},
		{&PropertiesConnAck{}, PropertyMaximumQoS, true},
		{&PropertiesConnAck{}, PropertyRetainAvailable, true},
		{&PropertiesConnAck{}, PropertyWildcardSubscriptionAvailable, true},
		{&PropertiesConnAck{}, PropertySubscriptionIDAvailable, true},
		{&PropertiesConnAck{}, PropertySharedSubscriptionAvailable, true},
		{nil, 0, false},
	}

	for _, test := range testCases {
		s.Run(fmt.Sprint(test.prop), func() {
			test.props.Set(test.prop)
			s.Require().Equal(test.result, test.props.Has(test.prop))
		})
	}
}

func (s *PropertiesConnAckTestSuite) TestSize() {
	var flags propertyFlags
	flags = flags.set(PropertyUserProperty)
	flags = flags.set(PropertyAssignedClientID)
	flags = flags.set(PropertyReasonString)
	flags = flags.set(PropertyResponseInfo)
	flags = flags.set(PropertyServerReference)
	flags = flags.set(PropertyAuthenticationMethod)
	flags = flags.set(PropertyAuthenticationData)
	flags = flags.set(PropertySessionExpiryInterval)
	flags = flags.set(PropertyMaximumPacketSize)
	flags = flags.set(PropertyReceiveMaximum)
	flags = flags.set(PropertyTopicAliasMaximum)
	flags = flags.set(PropertyServerKeepAlive)
	flags = flags.set(PropertyMaximumQoS)
	flags = flags.set(PropertyRetainAvailable)
	flags = flags.set(PropertyWildcardSubscriptionAvailable)
	flags = flags.set(PropertySubscriptionIDAvailable)
	flags = flags.set(PropertySharedSubscriptionAvailable)

	props := PropertiesConnAck{
		Flags:                flags,
		UserProperties:       []UserProperty{{[]byte("a"), []byte("b")}},
		AssignedClientID:     []byte("c"),
		ReasonString:         []byte("d"),
		ResponseInfo:         []byte("e"),
		ServerReference:      []byte("f"),
		AuthenticationMethod: []byte("auth"),
		AuthenticationData:   []byte("data"),
	}

	size := props.size()
	s.Assert().Equal(66, size)
}

func (s *PropertiesConnAckTestSuite) TestSizeOnNil() {
	var props *PropertiesConnAck

	size := props.size()
	s.Assert().Equal(0, size)
}

func (s *PropertiesConnAckTestSuite) TestEncodeSuccess() {
	testCases := []struct {
		name  string
		props *PropertiesConnAck
		data  []byte
	}{
		{"Nil", nil, []byte{0}},
		{"Empty", &PropertiesConnAck{}, []byte{0}},
		{
			"Session Expiry Interval",
			&PropertiesConnAck{
				Flags:                 propertyFlags(0).set(PropertySessionExpiryInterval),
				SessionExpiryInterval: 256,
			},
			[]byte{5, 0x11, 0, 0, 1, 0},
		},
		{
			"Receive Maximum",
			&PropertiesConnAck{
				Flags:          propertyFlags(0).set(PropertyReceiveMaximum),
				ReceiveMaximum: 256,
			},
			[]byte{3, 0x21, 1, 0},
		},
		{
			"Maximum QoS",
			&PropertiesConnAck{Flags: propertyFlags(0).set(PropertyMaximumQoS), MaximumQoS: 1},
			[]byte{2, 0x24, 1},
		},
		{
			"Retain Available",
			&PropertiesConnAck{
				Flags:           propertyFlags(0).set(PropertyRetainAvailable),
				RetainAvailable: true,
			},
			[]byte{2, 0x25, 1},
		},
		{
			"Maximum Packet Size",
			&PropertiesConnAck{
				Flags:             propertyFlags(0).set(PropertyMaximumPacketSize),
				MaximumPacketSize: 4294967295,
			},
			[]byte{5, 0x27, 0xff, 0xff, 0xff, 0xff},
		},
		{
			"Assigned Client Identifier",
			&PropertiesConnAck{
				Flags:            propertyFlags(0).set(PropertyAssignedClientID),
				AssignedClientID: []byte("abc"),
			},
			[]byte{6, 0x12, 0, 3, 'a', 'b', 'c'},
		},
		{
			"Topic Alias Maximum",
			&PropertiesConnAck{
				Flags:             propertyFlags(0).set(PropertyTopicAliasMaximum),
				TopicAliasMaximum: 256,
			},
			[]byte{3, 0x22, 1, 0},
		},
		{
			"Reason String",
			&PropertiesConnAck{
				Flags:        propertyFlags(0).set(PropertyReasonString),
				ReasonString: []byte("abc"),
			},
			[]byte{6, 0x1f, 0, 3, 'a', 'b', 'c'},
		},
		{
			"User Property",
			&PropertiesConnAck{
				Flags: propertyFlags(0).set(PropertyUserProperty),
				UserProperties: []UserProperty{
					{[]byte("a"), []byte("b")},
					{[]byte("c"), []byte("d")},
				},
			},
			[]byte{14, 0x26, 0, 1, 'a', 0, 1, 'b', 0x26, 0, 1, 'c', 0, 1, 'd'},
		},
		{
			"Wildcard Subscription Available",
			&PropertiesConnAck{
				Flags:                         propertyFlags(0).set(PropertyWildcardSubscriptionAvailable),
				WildcardSubscriptionAvailable: true,
			},
			[]byte{2, 0x28, 1},
		},
		{
			"Subscription Identifiers Available",
			&PropertiesConnAck{
				Flags:                   propertyFlags(0).set(PropertySubscriptionIDAvailable),
				SubscriptionIDAvailable: true,
			},
			[]byte{2, 0x29, 1},
		},
		{
			"Shared Subscription Available",
			&PropertiesConnAck{
				Flags:                       propertyFlags(0).set(PropertySharedSubscriptionAvailable),
				SharedSubscriptionAvailable: true,
			},
			[]byte{2, 0x2a, 1},
		},
		{
			"Server Keep Alive",
			&PropertiesConnAck{
				Flags:           propertyFlags(0).set(PropertyServerKeepAlive),
				ServerKeepAlive: 30,
			},
			[]byte{3, 0x13, 0, 30},
		},
		{
			"Response Information",
			&PropertiesConnAck{
				Flags:        propertyFlags(0).set(PropertyResponseInfo),
				ResponseInfo: []byte("abc"),
			},
			[]byte{6, 0x1a, 0, 3, 'a', 'b', 'c'},
		},
		{
			"Server Reference",
			&PropertiesConnAck{
				Flags:           propertyFlags(0).set(PropertyServerReference),
				ServerReference: []byte("abc"),
			},
			[]byte{6, 0x1c, 0, 3, 'a', 'b', 'c'},
		},
		{
			"Authentication Method",
			&PropertiesConnAck{
				Flags:                propertyFlags(0).set(PropertyAuthenticationMethod),
				AuthenticationMethod: []byte("abc"),
			},
			[]byte{6, 0x15, 0, 3, 'a', 'b', 'c'},
		},
		{
			"Authentication Data",
			&PropertiesConnAck{
				Flags:              propertyFlags(0).set(PropertyAuthenticationData),
				AuthenticationData: []byte("abc"),
			},
			[]byte{6, 0x16, 0, 3, 'a', 'b', 'c'},
		},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			data := make([]byte, len(test.data))

			n, err := test.props.encode(data)
			s.Require().NoError(err)
			s.Assert().Equal(len(test.data), n)
			s.Assert().Equal(test.data, data)
		})
	}
}

func (s *PropertiesConnAckTestSuite) TestEncodeError() {
	testCases := []struct {
		name  string
		props *PropertiesConnAck
	}{
		{
			"Invalid Assigned Client ID",
			&PropertiesConnAck{
				Flags:            propertyFlags(0).set(PropertyAssignedClientID),
				AssignedClientID: []byte{0},
			},
		},
		{
			"Invalid Reason String",
			&PropertiesConnAck{
				Flags:        propertyFlags(0).set(PropertyReasonString),
				ReasonString: []byte{0},
			},
		},
		{
			"Invalid User Properties - Key",
			&PropertiesConnAck{
				Flags: propertyFlags(0).set(PropertyUserProperty),
				UserProperties: []UserProperty{
					{[]byte{0}, []byte("b")},
				},
			},
		},
		{
			"Invalid User Properties - Value",
			&PropertiesConnAck{
				Flags: propertyFlags(0).set(PropertyUserProperty),
				UserProperties: []UserProperty{
					{[]byte("a"), []byte{0}},
				},
			},
		},
		{
			"Invalid Response Info",
			&PropertiesConnAck{
				Flags:        propertyFlags(0).set(PropertyResponseInfo),
				ResponseInfo: []byte{0},
			},
		},
		{
			"Invalid Server Reference",
			&PropertiesConnAck{
				Flags:           propertyFlags(0).set(PropertyServerReference),
				ServerReference: []byte{0},
			},
		},
		{
			"Invalid Authentication Method",
			&PropertiesConnAck{
				Flags:                propertyFlags(0).set(PropertyAuthenticationMethod),
				AuthenticationMethod: []byte{0},
			},
		},
		{
			"Invalid Authentication Data",
			&PropertiesConnAck{
				Flags:              propertyFlags(0).set(PropertyAuthenticationData),
				AuthenticationData: []byte{0},
			},
		},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			data := make([]byte, 10)

			_, err := test.props.encode(data)
			s.Require().Error(err)
		})
	}
}

func TestPropertiesConnAckTestSuite(t *testing.T) {
	suite.Run(t, new(PropertiesConnAckTestSuite))
}
