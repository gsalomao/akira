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
		{name: "V5.0, empty properties", packet: ConnAck{Version: MQTT50, Properties: &ConnAckProperties{}}, size: 5},
		{
			name: "V5.0, with properties",
			packet: ConnAck{
				Version: MQTT50,
				Properties: &ConnAckProperties{
					Flags:                 PropertyFlags(0).Set(PropertySessionExpiryInterval),
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

func (s *ConnAckTestSuite) TestEncode() {
	testCases := []struct {
		name   string
		packet ConnAck
		data   []byte
	}{
		{
			name:   "V3.1, accepted",
			packet: ConnAck{Version: MQTT31, Code: ReasonCodeSuccess},
			data:   []byte{0x20, 2, 0, 0},
		},
		{
			name:   "V3.1, accepted and ignore session present",
			packet: ConnAck{Version: MQTT31, SessionPresent: true, Code: ReasonCodeSuccess},
			data:   []byte{0x20, 2, 0, 0},
		},
		{
			name:   "V3.1, rejected",
			packet: ConnAck{Version: MQTT31, Code: ReasonCodeV3IdentifierRejected},
			data:   []byte{0x20, 2, 0, 2},
		},
		{
			name:   "V3.1.1, accepted",
			packet: ConnAck{Version: MQTT311, Code: ReasonCodeSuccess},
			data:   []byte{0x20, 2, 0, 0},
		},
		{
			name:   "V3.1.1, accepted with session present",
			packet: ConnAck{Version: MQTT311, SessionPresent: true, Code: ReasonCodeSuccess},
			data:   []byte{0x20, 2, 1, 0},
		},
		{
			name:   "V3.1.1, rejected",
			packet: ConnAck{Version: MQTT311, Code: ReasonCodeV3IdentifierRejected},
			data:   []byte{0x20, 2, 0, 2},
		},
		{
			name:   "V3.1.1, rejected and ignore session present",
			packet: ConnAck{Version: MQTT311, Code: ReasonCodeV3IdentifierRejected},
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
				Properties: &ConnAckProperties{
					Flags:                 PropertyFlags(0).Set(PropertySessionExpiryInterval),
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

func (s *ConnAckTestSuite) TestEncodeErrorWhenPacketInvalid() {
	data := make([]byte, 10)
	packet := ConnAck{}

	n, err := packet.Encode(data)
	s.Require().Error(err)
	s.Assert().Zero(n)
}

func (s *ConnAckTestSuite) TestValidateWithValidVersion() {
	testCases := []Version{MQTT31, MQTT311, MQTT50}

	for _, test := range testCases {
		s.Run(test.String(), func() {
			connack := ConnAck{Version: test}
			err := connack.Validate()
			s.Require().NoError(err)
		})
	}
}

func (s *ConnAckTestSuite) TestValidateWithInvalidVersion() {
	connack := ConnAck{}
	err := connack.Validate()
	s.Require().Error(err)
}

func (s *ConnAckTestSuite) TestValidateWithValidReasonCode() {
	testCases := []struct {
		version Version
		code    ReasonCode
	}{
		{MQTT31, ReasonCodeSuccess},
		{MQTT31, ReasonCodeV3UnacceptableProtocolVersion},
		{MQTT31, ReasonCodeV3IdentifierRejected},
		{MQTT31, ReasonCodeV3ServerUnavailable},
		{MQTT31, ReasonCodeV3BadUsernameOrPassword},
		{MQTT31, ReasonCodeV3NotAuthorized},
		{MQTT311, ReasonCodeSuccess},
		{MQTT311, ReasonCodeV3UnacceptableProtocolVersion},
		{MQTT311, ReasonCodeV3IdentifierRejected},
		{MQTT311, ReasonCodeV3ServerUnavailable},
		{MQTT311, ReasonCodeV3BadUsernameOrPassword},
		{MQTT311, ReasonCodeV3NotAuthorized},
		{MQTT50, ReasonCodeSuccess},
		{MQTT50, ReasonCodeUnspecifiedError},
		{MQTT50, ReasonCodeMalformedPacket},
		{MQTT50, ReasonCodeProtocolError},
		{MQTT50, ReasonCodeImplementationSpecificError},
		{MQTT50, ReasonCodeUnsupportedProtocolVersion},
		{MQTT50, ReasonCodeClientIDNotValid},
		{MQTT50, ReasonCodeBadUsernameOrPassword},
		{MQTT50, ReasonCodeNotAuthorized},
		{MQTT50, ReasonCodeServerUnavailable},
		{MQTT50, ReasonCodeServerBusy},
		{MQTT50, ReasonCodeBanned},
		{MQTT50, ReasonCodeBadAuthenticationMethod},
		{MQTT50, ReasonCodeTopicNameInvalid},
		{MQTT50, ReasonCodePacketTooLarge},
		{MQTT50, ReasonCodeQuotaExceeded},
		{MQTT50, ReasonCodePayloadFormatInvalid},
		{MQTT50, ReasonCodeRetainNotSupported},
		{MQTT50, ReasonCodeQoSNotSupported},
		{MQTT50, ReasonCodeUseAnotherServer},
		{MQTT50, ReasonCodeServerMoved},
		{MQTT50, ReasonCodeConnectionRateExceeded},
	}

	for _, test := range testCases {
		s.Run(fmt.Sprintf("V%s-0x%.2X", test.version.String(), test.code), func() {
			connack := ConnAck{Version: test.version, Code: test.code}
			err := connack.Validate()
			s.Require().NoError(err)
		})
	}
}

func (s *ConnAckTestSuite) TestValidateWithInvalidReasonCode() {
	testCases := []struct {
		version Version
		code    ReasonCode
	}{
		{MQTT31, ReasonCodeUnspecifiedError},
		{MQTT31, ReasonCodeMalformedPacket},
		{MQTT31, ReasonCodeProtocolError},
		{MQTT31, ReasonCodeImplementationSpecificError},
		{MQTT31, ReasonCodeUnsupportedProtocolVersion},
		{MQTT31, ReasonCodeClientIDNotValid},
		{MQTT31, ReasonCodeBadUsernameOrPassword},
		{MQTT31, ReasonCodeNotAuthorized},
		{MQTT31, ReasonCodeServerUnavailable},
		{MQTT31, ReasonCodeServerBusy},
		{MQTT31, ReasonCodeBanned},
		{MQTT31, ReasonCodeBadAuthenticationMethod},
		{MQTT31, ReasonCodeTopicNameInvalid},
		{MQTT31, ReasonCodePacketTooLarge},
		{MQTT31, ReasonCodeQuotaExceeded},
		{MQTT31, ReasonCodePayloadFormatInvalid},
		{MQTT31, ReasonCodeRetainNotSupported},
		{MQTT31, ReasonCodeQoSNotSupported},
		{MQTT31, ReasonCodeUseAnotherServer},
		{MQTT31, ReasonCodeServerMoved},
		{MQTT31, ReasonCodeConnectionRateExceeded},
		{MQTT311, ReasonCodeUnspecifiedError},
		{MQTT311, ReasonCodeMalformedPacket},
		{MQTT311, ReasonCodeProtocolError},
		{MQTT311, ReasonCodeImplementationSpecificError},
		{MQTT311, ReasonCodeUnsupportedProtocolVersion},
		{MQTT311, ReasonCodeClientIDNotValid},
		{MQTT311, ReasonCodeBadUsernameOrPassword},
		{MQTT311, ReasonCodeNotAuthorized},
		{MQTT311, ReasonCodeServerUnavailable},
		{MQTT311, ReasonCodeServerBusy},
		{MQTT311, ReasonCodeBanned},
		{MQTT311, ReasonCodeBadAuthenticationMethod},
		{MQTT311, ReasonCodeTopicNameInvalid},
		{MQTT311, ReasonCodePacketTooLarge},
		{MQTT311, ReasonCodeQuotaExceeded},
		{MQTT311, ReasonCodePayloadFormatInvalid},
		{MQTT311, ReasonCodeRetainNotSupported},
		{MQTT311, ReasonCodeQoSNotSupported},
		{MQTT311, ReasonCodeUseAnotherServer},
		{MQTT311, ReasonCodeServerMoved},
		{MQTT311, ReasonCodeConnectionRateExceeded},
		{MQTT50, ReasonCodeV3UnacceptableProtocolVersion},
		{MQTT50, ReasonCodeV3IdentifierRejected},
		{MQTT50, ReasonCodeV3ServerUnavailable},
		{MQTT50, ReasonCodeV3BadUsernameOrPassword},
		{MQTT50, ReasonCodeV3NotAuthorized},
	}

	for _, test := range testCases {
		s.Run(fmt.Sprintf("V%s-0x%.2X", test.version.String(), test.code), func() {
			connack := ConnAck{Version: test.version, Code: test.code}
			err := connack.Validate()
			s.Require().Error(err)
		})
	}
}

func (s *ConnAckTestSuite) TestValidateWithInvalidProperties() {
	connack := ConnAck{
		Version:    MQTT50,
		Properties: &ConnAckProperties{Flags: PropertyFlags(0).Set(PropertyReasonString)},
	}
	err := connack.Validate()
	s.Require().Error(err)
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
			packet: ConnAck{Version: MQTT311, Code: ReasonCodeSuccess},
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

type ConnAckPropertiesTestSuite struct {
	suite.Suite
}

func (s *ConnAckPropertiesTestSuite) TestHas() {
	testCases := []struct {
		props  *ConnAckProperties
		id     PropertyID
		result bool
	}{
		{&ConnAckProperties{}, PropertyUserProperty, true},
		{&ConnAckProperties{}, PropertyAssignedClientID, true},
		{&ConnAckProperties{}, PropertyReasonString, true},
		{&ConnAckProperties{}, PropertyResponseInfo, true},
		{&ConnAckProperties{}, PropertyServerReference, true},
		{&ConnAckProperties{}, PropertyAuthenticationMethod, true},
		{&ConnAckProperties{}, PropertyAuthenticationData, true},
		{&ConnAckProperties{}, PropertySessionExpiryInterval, true},
		{&ConnAckProperties{}, PropertyMaximumPacketSize, true},
		{&ConnAckProperties{}, PropertyReceiveMaximum, true},
		{&ConnAckProperties{}, PropertyTopicAliasMaximum, true},
		{&ConnAckProperties{}, PropertyServerKeepAlive, true},
		{&ConnAckProperties{}, PropertyMaximumQoS, true},
		{&ConnAckProperties{}, PropertyRetainAvailable, true},
		{&ConnAckProperties{}, PropertyWildcardSubscriptionAvailable, true},
		{&ConnAckProperties{}, PropertySubscriptionIDAvailable, true},
		{&ConnAckProperties{}, PropertySharedSubscriptionAvailable, true},
		{nil, 0, false},
	}

	for _, test := range testCases {
		s.Run(fmt.Sprint(test.id), func() {
			test.props.Set(test.id)
			s.Require().Equal(test.result, test.props.Has(test.id))
		})
	}
}

func (s *ConnAckPropertiesTestSuite) TestSize() {
	var flags PropertyFlags
	flags = flags.Set(PropertyUserProperty)
	flags = flags.Set(PropertyAssignedClientID)
	flags = flags.Set(PropertyReasonString)
	flags = flags.Set(PropertyResponseInfo)
	flags = flags.Set(PropertyServerReference)
	flags = flags.Set(PropertyAuthenticationMethod)
	flags = flags.Set(PropertyAuthenticationData)
	flags = flags.Set(PropertySessionExpiryInterval)
	flags = flags.Set(PropertyMaximumPacketSize)
	flags = flags.Set(PropertyReceiveMaximum)
	flags = flags.Set(PropertyTopicAliasMaximum)
	flags = flags.Set(PropertyServerKeepAlive)
	flags = flags.Set(PropertyMaximumQoS)
	flags = flags.Set(PropertyRetainAvailable)
	flags = flags.Set(PropertyWildcardSubscriptionAvailable)
	flags = flags.Set(PropertySubscriptionIDAvailable)
	flags = flags.Set(PropertySharedSubscriptionAvailable)

	props := ConnAckProperties{
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

func (s *ConnAckPropertiesTestSuite) TestSizeOnNil() {
	var props *ConnAckProperties

	size := props.size()
	s.Assert().Equal(0, size)
}

func (s *ConnAckPropertiesTestSuite) TestEncodeSuccess() {
	testCases := []struct {
		name  string
		props *ConnAckProperties
		data  []byte
	}{
		{"Nil", nil, []byte{0}},
		{"Empty", &ConnAckProperties{}, []byte{0}},
		{
			"Session Expiry Interval",
			&ConnAckProperties{
				Flags:                 PropertyFlags(0).Set(PropertySessionExpiryInterval),
				SessionExpiryInterval: 256,
			},
			[]byte{5, 0x11, 0, 0, 1, 0},
		},
		{
			"Receive Maximum",
			&ConnAckProperties{
				Flags:          PropertyFlags(0).Set(PropertyReceiveMaximum),
				ReceiveMaximum: 256,
			},
			[]byte{3, 0x21, 1, 0},
		},
		{
			"Maximum QoS",
			&ConnAckProperties{Flags: PropertyFlags(0).Set(PropertyMaximumQoS), MaximumQoS: 1},
			[]byte{2, 0x24, 1},
		},
		{
			"Retain Available",
			&ConnAckProperties{
				Flags:           PropertyFlags(0).Set(PropertyRetainAvailable),
				RetainAvailable: true,
			},
			[]byte{2, 0x25, 1},
		},
		{
			"Maximum Packet Size",
			&ConnAckProperties{
				Flags:             PropertyFlags(0).Set(PropertyMaximumPacketSize),
				MaximumPacketSize: 4294967295,
			},
			[]byte{5, 0x27, 0xff, 0xff, 0xff, 0xff},
		},
		{
			"Assigned Client Identifier",
			&ConnAckProperties{
				Flags:            PropertyFlags(0).Set(PropertyAssignedClientID),
				AssignedClientID: []byte("abc"),
			},
			[]byte{6, 0x12, 0, 3, 'a', 'b', 'c'},
		},
		{
			"Topic Alias Maximum",
			&ConnAckProperties{
				Flags:             PropertyFlags(0).Set(PropertyTopicAliasMaximum),
				TopicAliasMaximum: 256,
			},
			[]byte{3, 0x22, 1, 0},
		},
		{
			"Reason String",
			&ConnAckProperties{
				Flags:        PropertyFlags(0).Set(PropertyReasonString),
				ReasonString: []byte("abc"),
			},
			[]byte{6, 0x1f, 0, 3, 'a', 'b', 'c'},
		},
		{
			"User PropertyID",
			&ConnAckProperties{
				Flags: PropertyFlags(0).Set(PropertyUserProperty),
				UserProperties: []UserProperty{
					{[]byte("a"), []byte("b")},
					{[]byte("c"), []byte("d")},
				},
			},
			[]byte{14, 0x26, 0, 1, 'a', 0, 1, 'b', 0x26, 0, 1, 'c', 0, 1, 'd'},
		},
		{
			"Wildcard Subscription Available",
			&ConnAckProperties{
				Flags:                         PropertyFlags(0).Set(PropertyWildcardSubscriptionAvailable),
				WildcardSubscriptionAvailable: true,
			},
			[]byte{2, 0x28, 1},
		},
		{
			"Subscription Identifiers Available",
			&ConnAckProperties{
				Flags:                   PropertyFlags(0).Set(PropertySubscriptionIDAvailable),
				SubscriptionIDAvailable: true,
			},
			[]byte{2, 0x29, 1},
		},
		{
			"Shared Subscription Available",
			&ConnAckProperties{
				Flags:                       PropertyFlags(0).Set(PropertySharedSubscriptionAvailable),
				SharedSubscriptionAvailable: true,
			},
			[]byte{2, 0x2a, 1},
		},
		{
			"Server Keep Alive",
			&ConnAckProperties{
				Flags:           PropertyFlags(0).Set(PropertyServerKeepAlive),
				ServerKeepAlive: 30,
			},
			[]byte{3, 0x13, 0, 30},
		},
		{
			"Response Information",
			&ConnAckProperties{
				Flags:        PropertyFlags(0).Set(PropertyResponseInfo),
				ResponseInfo: []byte("abc"),
			},
			[]byte{6, 0x1a, 0, 3, 'a', 'b', 'c'},
		},
		{
			"Server Reference",
			&ConnAckProperties{
				Flags:           PropertyFlags(0).Set(PropertyServerReference),
				ServerReference: []byte("abc"),
			},
			[]byte{6, 0x1c, 0, 3, 'a', 'b', 'c'},
		},
		{
			"Authentication Method",
			&ConnAckProperties{
				Flags:                PropertyFlags(0).Set(PropertyAuthenticationMethod),
				AuthenticationMethod: []byte("abc"),
			},
			[]byte{6, 0x15, 0, 3, 'a', 'b', 'c'},
		},
		{
			"Authentication Data",
			&ConnAckProperties{
				Flags:              PropertyFlags(0).Set(PropertyAuthenticationData),
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

func (s *ConnAckPropertiesTestSuite) TestEncodeError() {
	testCases := []struct {
		name  string
		props *ConnAckProperties
	}{
		{
			"Invalid Assigned Client ID",
			&ConnAckProperties{
				Flags:            PropertyFlags(0).Set(PropertyAssignedClientID),
				AssignedClientID: []byte{0},
			},
		},
		{
			"Invalid Reason String",
			&ConnAckProperties{
				Flags:        PropertyFlags(0).Set(PropertyReasonString),
				ReasonString: []byte{0},
			},
		},
		{
			"Invalid User Properties - Key",
			&ConnAckProperties{
				Flags:          PropertyFlags(0).Set(PropertyUserProperty),
				UserProperties: []UserProperty{{[]byte{0}, []byte("b")}},
			},
		},
		{
			"Invalid User Properties - Value",
			&ConnAckProperties{
				Flags:          PropertyFlags(0).Set(PropertyUserProperty),
				UserProperties: []UserProperty{{[]byte("a"), []byte{0}}},
			},
		},
		{
			"Invalid Response Info",
			&ConnAckProperties{
				Flags:        PropertyFlags(0).Set(PropertyResponseInfo),
				ResponseInfo: []byte{0},
			},
		},
		{
			"Invalid Server Reference",
			&ConnAckProperties{
				Flags:           PropertyFlags(0).Set(PropertyServerReference),
				ServerReference: []byte{0},
			},
		},
		{
			"Invalid Authentication Method",
			&ConnAckProperties{
				Flags:                PropertyFlags(0).Set(PropertyAuthenticationMethod),
				AuthenticationMethod: []byte{0},
			},
		},
		{
			"Invalid Authentication Data",
			&ConnAckProperties{
				Flags:              PropertyFlags(0).Set(PropertyAuthenticationData),
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

func (s *ConnAckPropertiesTestSuite) TestValidateWithMissingProperty() {
	testCases := []struct {
		name string
		id   PropertyID
	}{
		{"Assigned client ID", PropertyAssignedClientID},
		{"Reason string", PropertyReasonString},
		{"Reason info", PropertyResponseInfo},
		{"Server reference", PropertyServerReference},
		{"Authentication method", PropertyAuthenticationMethod},
		{"Authentication data", PropertyAuthenticationData},
		{"User property", PropertyUserProperty},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			props := ConnAckProperties{Flags: PropertyFlags(0).Set(test.id)}
			err := props.Validate()
			s.Require().Error(err)
		})
	}
}

func (s *ConnAckPropertiesTestSuite) TestValidateWithInvalidMaximumQoS() {
	props := ConnAckProperties{Flags: PropertyFlags(0).Set(PropertyMaximumQoS), MaximumQoS: 3}
	err := props.Validate()
	s.Require().Error(err)
}

func TestConnAckPropertiesTestSuite(t *testing.T) {
	suite.Run(t, new(ConnAckPropertiesTestSuite))
}
