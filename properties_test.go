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

package melitte

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

func (s *PropertiesTestSuite) TestPropertiesConnectHas() {
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

func (s *PropertiesTestSuite) TestPropertiesConnectSize() {
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

func (s *PropertiesTestSuite) TestPropertiesConnectSizeOnNil() {
	var props *PropertiesConnect

	size := props.size()
	s.Assert().Equal(1, size)
}

func (s *PropertiesTestSuite) TestDecodePropertiesConnectSuccess() {
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

func (s *PropertiesTestSuite) TestDecodePropertiesConnectError() {
	testCases := []struct {
		name string
		data []byte
		err  error
	}{
		{
			name: "No property",
			data: []byte{1},
			err:  ErrMalformedPropertyConnect,
		},
		{
			name: "Missing Session Expiry Interval",
			data: []byte{1, 17},
			err:  ErrMalformedPropertySessionExpiryInterval,
		},
		{
			name: "Session Expiry Interval - Incomplete uint",
			data: []byte{2, 17, 0},
			err:  ErrMalformedPropertySessionExpiryInterval,
		},
		{
			name: "Duplicated Session Expiry Interval",
			data: []byte{10, 17, 0, 0, 0, 10, 17, 0, 0, 0, 11},
			err:  ErrMalformedPropertySessionExpiryInterval,
		},
		{
			name: "Missing Receive Maximum",
			data: []byte{1, 33},
			err:  ErrMalformedPropertyReceiveMaximum,
		},
		{
			name: "Invalid Receive Maximum",
			data: []byte{3, 33, 0, 0},
			err:  ErrMalformedPropertyReceiveMaximum,
		},
		{
			name: "Duplicated Receive Maximum",
			data: []byte{6, 33, 0, 50, 33, 0, 51},
			err:  ErrMalformedPropertyReceiveMaximum,
		},
		{
			name: "Missing Maximum Packet Size",
			data: []byte{1, 39},
			err:  ErrMalformedPropertyMaxPacketSize,
		},
		{
			name: "Invalid Maximum Packet Size",
			data: []byte{5, 39, 0, 0, 0, 0},
			err:  ErrMalformedPropertyMaxPacketSize,
		},
		{
			name: "Duplicated Maximum Packet Size",
			data: []byte{10, 39, 0, 0, 0, 200, 39, 0, 0, 0, 201},
			err:  ErrMalformedPropertyMaxPacketSize,
		},
		{
			name: "Missing Topic Alias Maximum",
			data: []byte{1, 34},
			err:  ErrMalformedPropertyTopicAliasMaximum,
		},
		{
			name: "Duplicated Topic Alias Maximum",
			data: []byte{6, 34, 0, 50, 34, 0, 51},
			err:  ErrMalformedPropertyTopicAliasMaximum,
		},
		{
			name: "Missing Request Response Info",
			data: []byte{1, 25},
			err:  ErrMalformedPropertyRequestResponseInfo,
		},
		{
			name: "Invalid Request Response Info",
			data: []byte{2, 25, 2},
			err:  ErrMalformedPropertyRequestResponseInfo,
		},
		{
			name: "Duplicated Request Response Info",
			data: []byte{4, 25, 0, 25, 1},
			err:  ErrMalformedPropertyRequestResponseInfo,
		},
		{
			name: "Missing Request Problem Info",
			data: []byte{1, 23},
			err:  ErrMalformedPropertyRequestProblemInfo,
		},
		{
			name: "Invalid Request Problem Info",
			data: []byte{2, 23, 2},
			err:  ErrMalformedPropertyRequestProblemInfo,
		},
		{
			name: "Duplicated Request Problem Info",
			data: []byte{4, 23, 0, 23, 1},
			err:  ErrMalformedPropertyRequestProblemInfo,
		},
		{
			name: "Missing User Property",
			data: []byte{1, 38},
			err:  ErrMalformedPropertyUserProperty,
		},
		{
			name: "Missing User Property Value",
			data: []byte{4, 38, 0, 1, 'a'},
			err:  ErrMalformedPropertyUserProperty,
		},
		{
			name: "User Property Value - Incomplete string",
			data: []byte{4, 38, 0, 5, 'a'},
			err:  ErrMalformedPropertyUserProperty,
		},
		{
			name: "Missing Authentication Method",
			data: []byte{1, 21},
			err:  ErrMalformedPropertyAuthenticationMethod,
		},
		{
			name: "Duplicated Authentication Method",
			data: []byte{10, 21, 0, 2, 'a', 'b', 21, 0, 2, 'c', 'd'},
			err:  ErrMalformedPropertyAuthenticationMethod,
		},
		{
			name: "Missing Authentication Data",
			data: []byte{1, 22},
			err:  ErrMalformedPropertyAuthenticationData,
		},
		{
			name: "Duplicated Authentication Data",
			data: []byte{8, 22, 0, 1, 10, 22, 0, 1, 11},
			err:  ErrMalformedPropertyAuthenticationData,
		},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			_, _, err := decodeProperties[PropertiesConnect](test.data)
			s.Require().ErrorIs(err, test.err)
		})
	}
}

func (s *PropertiesTestSuite) TestPropertiesWillHas() {
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

func (s *PropertiesTestSuite) TestPropertiesWillSize() {
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
	s.Assert().Equal(36, size)
}

func (s *PropertiesTestSuite) TestPropertiesWillSizeOnNil() {
	var props *PropertiesWill

	size := props.size()
	s.Assert().Equal(1, size)
}

func (s *PropertiesTestSuite) TestDecodePropertiesWillSuccess() {
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

func (s *PropertiesTestSuite) TestDecodePropertiesWillError() {
	testCases := []struct {
		name string
		data []byte
		err  error
	}{
		{
			name: "No property",
			data: []byte{1},
			err:  ErrMalformedPropertyWill,
		},
		{
			name: "Missing Will Delay Interval",
			data: []byte{1, 24},
			err:  ErrMalformedPropertyWillDelayInterval,
		},
		{
			name: "Duplicated Will Delay Interval",
			data: []byte{10, 24, 0, 0, 0, 15, 24, 0, 0, 0, 16},
			err:  ErrMalformedPropertyWillDelayInterval,
		},
		{
			name: "Missing Payload Format Indicator",
			data: []byte{1, 1},
			err:  ErrMalformedPropertyPayloadFormatIndicator,
		},
		{
			name: "Invalid Payload Format Indicator",
			data: []byte{2, 1, 2},
			err:  ErrMalformedPropertyPayloadFormatIndicator,
		},
		{
			name: "Duplicated Payload Format Indicator",
			data: []byte{4, 1, 0, 1, 1},
			err:  ErrMalformedPropertyPayloadFormatIndicator,
		},
		{
			name: "Missing Message Expiry Interval",
			data: []byte{1, 2},
			err:  ErrMalformedPropertyMessageExpiryInterval,
		},
		{
			name: "Duplicated Message Expiry Interval",
			data: []byte{10, 2, 0, 0, 0, 10, 2, 0, 0, 0, 11},
			err:  ErrMalformedPropertyMessageExpiryInterval,
		},
		{
			name: "Missing Content Type",
			data: []byte{1, 3},
			err:  ErrMalformedPropertyContentType,
		},
		{
			name: "Content Type - Missing string",
			data: []byte{3, 3, 0, 4},
			err:  ErrMalformedPropertyContentType,
		},
		{
			name: "Duplicated Content Type",
			data: []byte{13, 3, 0, 4, 'j', 's', 'o', 'n', 3, 0, 3, 'x', 'm', 'l'},
			err:  ErrMalformedPropertyContentType,
		},
		{
			name: "Missing Response Topic",
			data: []byte{1, 8},
			err:  ErrMalformedPropertyResponseTopic,
		},
		{
			name: "Response Topic - Missing string",
			data: []byte{3, 8, 0, 0},
			err:  ErrMalformedPropertyResponseTopic,
		},
		{
			name: "Response Topic - Incomplete string",
			data: []byte{3, 8, 0, 1},
			err:  ErrMalformedPropertyResponseTopic,
		},
		{
			name: "Invalid Response Topic",
			data: []byte{4, 8, 0, 1, '#'},
			err:  ErrMalformedPropertyResponseTopic,
		},
		{
			name: "Duplicated Response Topic",
			data: []byte{8, 8, 0, 1, 'b', 8, 0, 1, 'c'},
			err:  ErrMalformedPropertyResponseTopic,
		},
		{
			name: "Missing Correlation Data",
			data: []byte{1, 9},
			err:  ErrMalformedPropertyCorrelationData,
		},
		{
			name: "Correlation Data - Missing data",
			data: []byte{3, 9, 0, 2},
			err:  ErrMalformedPropertyCorrelationData,
		},
		{
			name: "Duplicated Correlation Data",
			data: []byte{10, 9, 0, 2, 20, 1, 9, 0, 2, 20, 2},
			err:  ErrMalformedPropertyCorrelationData,
		},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			_, _, err := decodeProperties[PropertiesWill](test.data)
			s.Require().ErrorIs(err, test.err)
		})
	}
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
