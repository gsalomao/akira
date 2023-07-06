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

type PropertiesTestSuite struct {
	suite.Suite
}

func (s *PropertiesTestSuite) TestDecodePropertiesNoProperties() {
	data := []byte{0}

	p, n, err := decodeProperties[PropertiesConnect](data)
	s.Require().NoError(err)
	s.Assert().Nil(p)
	s.Assert().Equal(1, n)
}

func (s *PropertiesTestSuite) TestDecodePropertiesErrorPropertyLength() {
	data := []byte{0xFF, 0xFF, 0xFF, 0xFF}

	_, _, err := decodeProperties[PropertiesConnect](data)
	s.Require().Error(err)
}

func (s *PropertiesTestSuite) TestDecodePropertiesInvalidPropertyType() {
	data := []byte{1}

	p, n, err := decodeProperties[int](data)
	s.Require().Error(err)
	s.Assert().Nil(p)
	s.Assert().Equal(1, n)
}

func (s *PropertiesTestSuite) TestPropertiesConnAckHas() {
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

func (s *PropertiesTestSuite) TestPropertiesConnAckSize() {
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

func (s *PropertiesTestSuite) TestPropertiesConnAckSizeOnNil() {
	var props *PropertiesConnAck

	size := props.size()
	s.Assert().Equal(0, size)
}

func (s *PropertiesTestSuite) TestPropertiesConnAckEncodeSuccess() {
	testCases := []struct {
		name  string
		props *PropertiesConnAck
		data  []byte
	}{
		{name: "Nil",
			props: nil,
			data:  []byte{0},
		},
		{
			name:  "Empty",
			props: &PropertiesConnAck{},
			data:  []byte{0},
		},
		{
			name: "Session Expiry Interval",
			props: &PropertiesConnAck{
				Flags:                 propertyFlags(0).set(PropertySessionExpiryInterval),
				SessionExpiryInterval: 256,
			},
			data: []byte{5, 0x11, 0, 0, 1, 0},
		},
		{
			name: "Receive Maximum",
			props: &PropertiesConnAck{
				Flags:          propertyFlags(0).set(PropertyReceiveMaximum),
				ReceiveMaximum: 256,
			},
			data: []byte{3, 0x21, 1, 0},
		},
		{
			name: "Maximum QoS",
			props: &PropertiesConnAck{
				Flags:      propertyFlags(0).set(PropertyMaximumQoS),
				MaximumQoS: 1,
			},
			data: []byte{2, 0x24, 1},
		},
		{
			name: "Retain Available",
			props: &PropertiesConnAck{
				Flags:           propertyFlags(0).set(PropertyRetainAvailable),
				RetainAvailable: true,
			},
			data: []byte{2, 0x25, 1},
		},
		{
			name: "Maximum Packet Size",
			props: &PropertiesConnAck{
				Flags:             propertyFlags(0).set(PropertyMaximumPacketSize),
				MaximumPacketSize: 4294967295,
			},
			data: []byte{5, 0x27, 0xff, 0xff, 0xff, 0xff},
		},
		{
			name: "Assigned Client Identifier",
			props: &PropertiesConnAck{
				Flags:            propertyFlags(0).set(PropertyAssignedClientID),
				AssignedClientID: []byte("abc"),
			},
			data: []byte{6, 0x12, 0, 3, 'a', 'b', 'c'},
		},
		{
			name: "Topic Alias Maximum",
			props: &PropertiesConnAck{
				Flags:             propertyFlags(0).set(PropertyTopicAliasMaximum),
				TopicAliasMaximum: 256,
			},
			data: []byte{3, 0x22, 1, 0},
		},
		{
			name: "Reason String",
			props: &PropertiesConnAck{
				Flags:        propertyFlags(0).set(PropertyReasonString),
				ReasonString: []byte("abc"),
			},
			data: []byte{6, 0x1f, 0, 3, 'a', 'b', 'c'},
		},
		{
			name: "User Property",
			props: &PropertiesConnAck{
				Flags: propertyFlags(0).set(PropertyUserProperty),
				UserProperties: []UserProperty{
					{[]byte("a"), []byte("b")},
					{[]byte("c"), []byte("d")},
				},
			},
			data: []byte{14, 0x26, 0, 1, 'a', 0, 1, 'b', 0x26, 0, 1, 'c', 0, 1, 'd'},
		},
		{
			name: "Wildcard Subscription Available",
			props: &PropertiesConnAck{
				Flags:                         propertyFlags(0).set(PropertyWildcardSubscriptionAvailable),
				WildcardSubscriptionAvailable: true,
			},
			data: []byte{2, 0x28, 1},
		},
		{
			name: "Subscription Identifiers Available",
			props: &PropertiesConnAck{
				Flags:                   propertyFlags(0).set(PropertySubscriptionIDAvailable),
				SubscriptionIDAvailable: true,
			},
			data: []byte{2, 0x29, 1},
		},
		{
			name: "Shared Subscription Available",
			props: &PropertiesConnAck{
				Flags:                       propertyFlags(0).set(PropertySharedSubscriptionAvailable),
				SharedSubscriptionAvailable: true,
			},
			data: []byte{2, 0x2a, 1},
		},
		{
			name: "Server Keep Alive",
			props: &PropertiesConnAck{
				Flags:           propertyFlags(0).set(PropertyServerKeepAlive),
				ServerKeepAlive: 30,
			},
			data: []byte{3, 0x13, 0, 30},
		},
		{
			name: "Response Information",
			props: &PropertiesConnAck{
				Flags:        propertyFlags(0).set(PropertyResponseInfo),
				ResponseInfo: []byte("abc"),
			},
			data: []byte{6, 0x1a, 0, 3, 'a', 'b', 'c'},
		},
		{
			name: "Server Reference",
			props: &PropertiesConnAck{
				Flags:           propertyFlags(0).set(PropertyServerReference),
				ServerReference: []byte("abc"),
			},
			data: []byte{6, 0x1c, 0, 3, 'a', 'b', 'c'},
		},
		{
			name: "Authentication Method",
			props: &PropertiesConnAck{
				Flags:                propertyFlags(0).set(PropertyAuthenticationMethod),
				AuthenticationMethod: []byte("abc"),
			},
			data: []byte{6, 0x15, 0, 3, 'a', 'b', 'c'},
		},
		{
			name: "Authentication Data",
			props: &PropertiesConnAck{
				Flags:              propertyFlags(0).set(PropertyAuthenticationData),
				AuthenticationData: []byte("abc"),
			},
			data: []byte{6, 0x16, 0, 3, 'a', 'b', 'c'},
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

func (s *PropertiesTestSuite) TestPropertiesConnAckEncodeError() {
	testCases := []struct {
		name  string
		props *PropertiesConnAck
	}{
		{
			name: "Invalid Assigned Client ID",
			props: &PropertiesConnAck{
				Flags:            propertyFlags(0).set(PropertyAssignedClientID),
				AssignedClientID: []byte{0},
			},
		},
		{
			name: "Invalid Reason String",
			props: &PropertiesConnAck{
				Flags:        propertyFlags(0).set(PropertyReasonString),
				ReasonString: []byte{0},
			},
		},
		{
			name: "Invalid User Properties - Key",
			props: &PropertiesConnAck{
				Flags: propertyFlags(0).set(PropertyUserProperty),
				UserProperties: []UserProperty{
					{[]byte{0}, []byte("b")},
				},
			},
		},
		{
			name: "Invalid User Properties - Value",
			props: &PropertiesConnAck{
				Flags: propertyFlags(0).set(PropertyUserProperty),
				UserProperties: []UserProperty{
					{[]byte("a"), []byte{0}},
				},
			},
		},
		{
			name: "Invalid Response Info",
			props: &PropertiesConnAck{
				Flags:        propertyFlags(0).set(PropertyResponseInfo),
				ResponseInfo: []byte{0},
			},
		},
		{
			name: "Invalid Server Reference",
			props: &PropertiesConnAck{
				Flags:           propertyFlags(0).set(PropertyServerReference),
				ServerReference: []byte{0},
			},
		},
		{
			name: "Invalid Authentication Method",
			props: &PropertiesConnAck{
				Flags:                propertyFlags(0).set(PropertyAuthenticationMethod),
				AuthenticationMethod: []byte{0},
			},
		},
		{
			name: "Invalid Authentication Data",
			props: &PropertiesConnAck{
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

func TestPropertiesTestSuite(t *testing.T) {
	suite.Run(t, new(PropertiesTestSuite))
}
