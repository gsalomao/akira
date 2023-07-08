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

package packet

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ConnAckTestSuite struct {
	suite.Suite
}

func (s *ConnAckTestSuite) TestType() {
	var p ConnAck
	s.Require().Equal(TypeConnAck, p.Type())
}

func (s *ConnAckTestSuite) TestSize() {
	testCases := []struct {
		name   string
		packet ConnAck
		size   int
	}{
		{name: "V3.1", packet: ConnAck{Version: MQTT31}, size: 4},
		{name: "V3.1.1", packet: ConnAck{Version: MQTT311}, size: 4},
		{name: "V5.0, no properties", packet: ConnAck{Version: MQTT50}, size: 5},
		{name: "V5.0, empty properties", packet: ConnAck{Version: MQTT50, Properties: &PropertiesConnAck{}}, size: 5},
		{
			name: "V5.0, with properties",
			packet: ConnAck{
				Version: MQTT50,
				Properties: &PropertiesConnAck{
					Flags:                 PropertyFlags(0).Set(PropertyIDSessionExpiryInterval),
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

func (s *ConnAckTestSuite) TestEncodeSuccess() {
	testCases := []struct {
		name   string
		packet ConnAck
		data   []byte
	}{
		{
			name:   "V3.1",
			packet: ConnAck{Version: MQTT31, SessionPresent: true, Code: ReasonCodeConnectionAccepted},
			data:   []byte{0x20, 2, 1, 0},
		},
		{
			name:   "V3.1.1",
			packet: ConnAck{Version: MQTT311, Code: ReasonCodeIdentifierRejected},
			data:   []byte{0x20, 2, 0, 2},
		},
		{
			name:   "V5.0, no properties",
			packet: ConnAck{Version: MQTT50, Code: ReasonCodeSuccess},
			data:   []byte{0x20, 3, 0, 0, 0},
		},
		{
			name: "V5.0, with properties",
			packet: ConnAck{
				Version: MQTT50,
				Code:    ReasonCodeMalformedPacket,
				Properties: &PropertiesConnAck{
					Flags:                 PropertyFlags(0).Set(PropertyIDSessionExpiryInterval),
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

func (s *ConnAckTestSuite) TestEncodeError() {
	packet := ConnAck{Version: MQTT50, Code: ReasonCodeSuccess}

	n, err := packet.Encode(nil)
	s.Require().Error(err)
	s.Assert().Zero(n)
}

func TestConnAckTestSuite(t *testing.T) {
	suite.Run(t, new(ConnAckTestSuite))
}

func BenchmarkConnAckEncode(b *testing.B) {
	testCases := []struct {
		name   string
		packet ConnAck
	}{
		{
			name:   "V3",
			packet: ConnAck{Version: MQTT311, Code: ReasonCodeConnectionAccepted},
		},
		{
			name:   "V5",
			packet: ConnAck{Version: MQTT50, Code: ReasonCodeSuccess},
		},
	}

	for _, test := range testCases {
		b.Run(test.name, func(b *testing.B) {
			data := make([]byte, test.packet.Size())
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
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
		id     PropertyID
		result bool
	}{
		{&PropertiesConnAck{}, PropertyIDUserProperty, true},
		{&PropertiesConnAck{}, PropertyIDAssignedClientID, true},
		{&PropertiesConnAck{}, PropertyIDReasonString, true},
		{&PropertiesConnAck{}, PropertyIDResponseInfo, true},
		{&PropertiesConnAck{}, PropertyIDServerReference, true},
		{&PropertiesConnAck{}, PropertyIDAuthenticationMethod, true},
		{&PropertiesConnAck{}, PropertyIDAuthenticationData, true},
		{&PropertiesConnAck{}, PropertyIDSessionExpiryInterval, true},
		{&PropertiesConnAck{}, PropertyIDMaximumPacketSize, true},
		{&PropertiesConnAck{}, PropertyIDReceiveMaximum, true},
		{&PropertiesConnAck{}, PropertyIDTopicAliasMaximum, true},
		{&PropertiesConnAck{}, PropertyIDServerKeepAlive, true},
		{&PropertiesConnAck{}, PropertyIDMaximumQoS, true},
		{&PropertiesConnAck{}, PropertyIDRetainAvailable, true},
		{&PropertiesConnAck{}, PropertyIDWildcardSubscriptionAvailable, true},
		{&PropertiesConnAck{}, PropertyIDSubscriptionIDAvailable, true},
		{&PropertiesConnAck{}, PropertyIDSharedSubscriptionAvailable, true},
		{nil, 0, false},
	}

	for _, test := range testCases {
		s.Run(fmt.Sprint(test.id), func() {
			test.props.Set(test.id)
			s.Require().Equal(test.result, test.props.Has(test.id))
		})
	}
}

func (s *PropertiesConnAckTestSuite) TestSize() {
	var flags PropertyFlags
	flags = flags.Set(PropertyIDUserProperty)
	flags = flags.Set(PropertyIDAssignedClientID)
	flags = flags.Set(PropertyIDReasonString)
	flags = flags.Set(PropertyIDResponseInfo)
	flags = flags.Set(PropertyIDServerReference)
	flags = flags.Set(PropertyIDAuthenticationMethod)
	flags = flags.Set(PropertyIDAuthenticationData)
	flags = flags.Set(PropertyIDSessionExpiryInterval)
	flags = flags.Set(PropertyIDMaximumPacketSize)
	flags = flags.Set(PropertyIDReceiveMaximum)
	flags = flags.Set(PropertyIDTopicAliasMaximum)
	flags = flags.Set(PropertyIDServerKeepAlive)
	flags = flags.Set(PropertyIDMaximumQoS)
	flags = flags.Set(PropertyIDRetainAvailable)
	flags = flags.Set(PropertyIDWildcardSubscriptionAvailable)
	flags = flags.Set(PropertyIDSubscriptionIDAvailable)
	flags = flags.Set(PropertyIDSharedSubscriptionAvailable)

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
				Flags:                 PropertyFlags(0).Set(PropertyIDSessionExpiryInterval),
				SessionExpiryInterval: 256,
			},
			[]byte{5, 0x11, 0, 0, 1, 0},
		},
		{
			"Receive Maximum",
			&PropertiesConnAck{
				Flags:          PropertyFlags(0).Set(PropertyIDReceiveMaximum),
				ReceiveMaximum: 256,
			},
			[]byte{3, 0x21, 1, 0},
		},
		{
			"Maximum QoS",
			&PropertiesConnAck{Flags: PropertyFlags(0).Set(PropertyIDMaximumQoS), MaximumQoS: 1},
			[]byte{2, 0x24, 1},
		},
		{
			"Retain Available",
			&PropertiesConnAck{
				Flags:           PropertyFlags(0).Set(PropertyIDRetainAvailable),
				RetainAvailable: true,
			},
			[]byte{2, 0x25, 1},
		},
		{
			"Maximum Packet Size",
			&PropertiesConnAck{
				Flags:             PropertyFlags(0).Set(PropertyIDMaximumPacketSize),
				MaximumPacketSize: 4294967295,
			},
			[]byte{5, 0x27, 0xff, 0xff, 0xff, 0xff},
		},
		{
			"Assigned Client Identifier",
			&PropertiesConnAck{
				Flags:            PropertyFlags(0).Set(PropertyIDAssignedClientID),
				AssignedClientID: []byte("abc"),
			},
			[]byte{6, 0x12, 0, 3, 'a', 'b', 'c'},
		},
		{
			"Topic Alias Maximum",
			&PropertiesConnAck{
				Flags:             PropertyFlags(0).Set(PropertyIDTopicAliasMaximum),
				TopicAliasMaximum: 256,
			},
			[]byte{3, 0x22, 1, 0},
		},
		{
			"Reason String",
			&PropertiesConnAck{
				Flags:        PropertyFlags(0).Set(PropertyIDReasonString),
				ReasonString: []byte("abc"),
			},
			[]byte{6, 0x1f, 0, 3, 'a', 'b', 'c'},
		},
		{
			"User Property",
			&PropertiesConnAck{
				Flags: PropertyFlags(0).Set(PropertyIDUserProperty),
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
				Flags:                         PropertyFlags(0).Set(PropertyIDWildcardSubscriptionAvailable),
				WildcardSubscriptionAvailable: true,
			},
			[]byte{2, 0x28, 1},
		},
		{
			"Subscription Identifiers Available",
			&PropertiesConnAck{
				Flags:                   PropertyFlags(0).Set(PropertyIDSubscriptionIDAvailable),
				SubscriptionIDAvailable: true,
			},
			[]byte{2, 0x29, 1},
		},
		{
			"Shared Subscription Available",
			&PropertiesConnAck{
				Flags:                       PropertyFlags(0).Set(PropertyIDSharedSubscriptionAvailable),
				SharedSubscriptionAvailable: true,
			},
			[]byte{2, 0x2a, 1},
		},
		{
			"Server Keep Alive",
			&PropertiesConnAck{
				Flags:           PropertyFlags(0).Set(PropertyIDServerKeepAlive),
				ServerKeepAlive: 30,
			},
			[]byte{3, 0x13, 0, 30},
		},
		{
			"Response Information",
			&PropertiesConnAck{
				Flags:        PropertyFlags(0).Set(PropertyIDResponseInfo),
				ResponseInfo: []byte("abc"),
			},
			[]byte{6, 0x1a, 0, 3, 'a', 'b', 'c'},
		},
		{
			"Server Reference",
			&PropertiesConnAck{
				Flags:           PropertyFlags(0).Set(PropertyIDServerReference),
				ServerReference: []byte("abc"),
			},
			[]byte{6, 0x1c, 0, 3, 'a', 'b', 'c'},
		},
		{
			"Authentication Method",
			&PropertiesConnAck{
				Flags:                PropertyFlags(0).Set(PropertyIDAuthenticationMethod),
				AuthenticationMethod: []byte("abc"),
			},
			[]byte{6, 0x15, 0, 3, 'a', 'b', 'c'},
		},
		{
			"Authentication Data",
			&PropertiesConnAck{
				Flags:              PropertyFlags(0).Set(PropertyIDAuthenticationData),
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
				Flags:            PropertyFlags(0).Set(PropertyIDAssignedClientID),
				AssignedClientID: []byte{0},
			},
		},
		{
			"Invalid Reason String",
			&PropertiesConnAck{
				Flags:        PropertyFlags(0).Set(PropertyIDReasonString),
				ReasonString: []byte{0},
			},
		},
		{
			"Invalid User Properties - Key",
			&PropertiesConnAck{
				Flags:          PropertyFlags(0).Set(PropertyIDUserProperty),
				UserProperties: []UserProperty{{[]byte{0}, []byte("b")}},
			},
		},
		{
			"Invalid User Properties - Value",
			&PropertiesConnAck{
				Flags:          PropertyFlags(0).Set(PropertyIDUserProperty),
				UserProperties: []UserProperty{{[]byte("a"), []byte{0}}},
			},
		},
		{
			"Invalid Response Info",
			&PropertiesConnAck{
				Flags:        PropertyFlags(0).Set(PropertyIDResponseInfo),
				ResponseInfo: []byte{0},
			},
		},
		{
			"Invalid Server Reference",
			&PropertiesConnAck{
				Flags:           PropertyFlags(0).Set(PropertyIDServerReference),
				ServerReference: []byte{0},
			},
		},
		{
			"Invalid Authentication Method",
			&PropertiesConnAck{
				Flags:                PropertyFlags(0).Set(PropertyIDAuthenticationMethod),
				AuthenticationMethod: []byte{0},
			},
		},
		{
			"Invalid Authentication Data",
			&PropertiesConnAck{
				Flags:              PropertyFlags(0).Set(PropertyIDAuthenticationData),
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
