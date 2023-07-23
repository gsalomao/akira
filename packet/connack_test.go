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
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/gsalomao/akira/packet/testdata"
)

func TestConnAckType(t *testing.T) {
	var p ConnAck
	tp := p.Type()
	if tp != TypeConnAck {
		t.Fatalf("Unexpected type\nwant: %s\ngot:  %s", TypeConnAck, tp)
	}
}

func TestConnAckSize(t *testing.T) {
	testCases := []struct {
		name   string
		packet ConnAck
		size   int
	}{
		{name: "V3.1", packet: ConnAck{Version: MQTT31}, size: 4},
		{name: "V3.1.1", packet: ConnAck{Version: MQTT311}, size: 4},
		{name: "V5.0 No Properties", packet: ConnAck{Version: MQTT50}, size: 5},
		{name: "V5.0 Empty Properties", packet: ConnAck{Version: MQTT50, Properties: &ConnAckProperties{}}, size: 5},
		{
			name: "V5.0 With Properties",
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			size := tc.packet.Size()
			if size != tc.size {
				t.Fatalf("Unexpected size\nwant: %v\ngot:  %v", tc.size, size)
			}
		})
	}
}

func TestConnAckEncode(t *testing.T) {
	testCases := []struct {
		name    string
		fixture string
		connack ConnAck
	}{
		{"V3.1 Accepted", "V3 Accepted", ConnAck{Version: MQTT31}},
		{"V3.1 Ignores Session present", "V3 Accepted", ConnAck{Version: MQTT31, SessionPresent: true}},
		{
			"V3.1.1 Client ID rejected", "V3 Client ID rejected",
			ConnAck{Version: MQTT311, Code: ReasonCodeV3IdentifierRejected},
		},
		{"V3.1.1 Session present", "V3.1.1 Session Present", ConnAck{Version: MQTT311, SessionPresent: true}},
		{
			"V3.1.1 Client ID rejected ignores session present", "V3 Client ID rejected",
			ConnAck{Version: MQTT311, Code: ReasonCodeV3IdentifierRejected, SessionPresent: true},
		},
		{"V5.0 Success", "V5.0 Success", ConnAck{Version: MQTT50}},
		{"V5.0 Malformed", "V5.0 Malformed", ConnAck{Version: MQTT50, Code: ReasonCodeMalformedPacket}},
		{
			"V5.0 Properties", "V5.0 Properties",
			ConnAck{
				Version: MQTT50,
				Code:    ReasonCodeClientIDNotValid,
				Properties: &ConnAckProperties{
					Flags: PropertyFlags(0).
						Set(PropertySessionExpiryInterval).
						Set(PropertyReceiveMaximum).
						Set(PropertyMaximumQoS).
						Set(PropertyRetainAvailable).
						Set(PropertyMaximumPacketSize).
						Set(PropertyAssignedClientID).
						Set(PropertyTopicAliasMaximum).
						Set(PropertyReasonString).
						Set(PropertyUserProperty).
						Set(PropertyWildcardSubscriptionAvailable).
						Set(PropertySubscriptionIDAvailable).
						Set(PropertySharedSubscriptionAvailable).
						Set(PropertyServerKeepAlive).
						Set(PropertyResponseInfo).
						Set(PropertyServerReference).
						Set(PropertyAuthenticationMethod).
						Set(PropertyAuthenticationData),
					SessionExpiryInterval:         1,
					ReceiveMaximum:                2,
					MaximumQoS:                    byte(QoS1),
					RetainAvailable:               true,
					MaximumPacketSize:             200,
					AssignedClientID:              []byte("a"),
					TopicAliasMaximum:             1,
					ReasonString:                  []byte("a"),
					UserProperties:                []UserProperty{{[]byte("a"), []byte("b")}},
					WildcardSubscriptionAvailable: true,
					SubscriptionIDAvailable:       true,
					SharedSubscriptionAvailable:   true,
					ServerKeepAlive:               1,
					ResponseInfo:                  []byte("a"),
					ServerReference:               []byte("a"),
					AuthenticationMethod:          []byte("a"),
					AuthenticationData:            []byte("a"),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture, err := testdata.FixtureWithName("connack.json", tc.fixture)
			if err != nil {
				t.Fatalf("Unexpected error while getting fixture\n%s", err)
			}

			var n int
			pkt := make([]byte, tc.connack.Size())

			n, err = tc.connack.Encode(pkt)
			if err != nil {
				t.Fatalf("Unexpected error\n%s", err)
			}
			if n != tc.connack.Size() {
				t.Fatalf("Unexpected number of bytes encoded\nwant: %v\ngot:  %v", tc.connack.Size(), n)
			}
			if !bytes.Equal(pkt, fixture.Packet) {
				t.Fatalf("Unexpected encoded packet\nwant: %v\ngot:  %v %+v", fixture.Packet, pkt, tc.connack)
			}
		})
	}
}

func TestConnAckEncodeErrorBufferTooSmall(t *testing.T) {
	packet := ConnAck{Version: MQTT50, Code: ReasonCodeSuccess}

	n, err := packet.Encode(nil)
	if err == nil {
		t.Fatal("An error was expected")
	}
	if n != 0 {
		t.Fatalf("Unexpected number of bytes encoded\nwant: %v\ngot:  %v", 0, n)
	}
}

func TestConnAckEncodeMalformedVersion(t *testing.T) {
	connack := ConnAck{}
	data := make([]byte, connack.Size())

	n, err := connack.Encode(data)
	if !errors.Is(err, ErrMalformedPacket) {
		t.Fatalf("Unexpected error\nwant: %s\ngot:  %s", ErrMalformedPacket, err)
	}
	if n != 0 {
		t.Fatalf("Unexpected number of bytes encoded\nwant: %v\ngot:  %v", 0, n)
	}
}

func TestConnAckEncodeMalformedReasonCode(t *testing.T) {
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
		{MQTT311, ReasonCodeServerMoved},
		{MQTT311, ReasonCodeConnectionRateExceeded},
		{MQTT50, ReasonCodeV3UnacceptableProtocolVersion},
		{MQTT50, ReasonCodeV3IdentifierRejected},
		{MQTT50, ReasonCodeV3ServerUnavailable},
		{MQTT50, ReasonCodeV3BadUsernameOrPassword},
		{MQTT50, ReasonCodeV3NotAuthorized},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("V%s-0x%.2X", tc.version.String(), tc.code), func(t *testing.T) {
			connack := ConnAck{Version: tc.version, Code: tc.code}
			data := make([]byte, connack.Size())

			n, err := connack.Encode(data)
			if !errors.Is(err, ErrMalformedPacket) {
				t.Fatalf("Unexpected error\nwant: %s\ngot:  %s", ErrMalformedPacket, err)
			}
			if n != 0 {
				t.Fatalf("Unexpected number of bytes encoded\nwant: %v\ngot:  %v", 0, n)
			}
		})
	}
}

func TestConnAckEncodeMalformedProperty(t *testing.T) {
	testCases := []struct {
		name string
		prop ConnAckProperties
	}{
		{"Invalid property", ConnAckProperties{Flags: PropertyFlags(0).Set(PropertyReasonString)}},
		{"Invalid Maximum QoS", ConnAckProperties{Flags: PropertyFlags(0).Set(PropertyMaximumQoS), MaximumQoS: 3}},
		{"Missing Assigned Client ID", ConnAckProperties{Flags: PropertyFlags(0).Set(PropertyAssignedClientID)}},
		{"Missing Reason String", ConnAckProperties{Flags: PropertyFlags(0).Set(PropertyReasonString)}},
		{"Missing User Property", ConnAckProperties{Flags: PropertyFlags(0).Set(PropertyUserProperty)}},
		{"Missing Response Info", ConnAckProperties{Flags: PropertyFlags(0).Set(PropertyResponseInfo)}},
		{"Missing Server Reference", ConnAckProperties{Flags: PropertyFlags(0).Set(PropertyServerReference)}},
		{"Missing Authentication Method", ConnAckProperties{Flags: PropertyFlags(0).Set(PropertyAuthenticationMethod)}},
		{"Missing Authentication Data", ConnAckProperties{Flags: PropertyFlags(0).Set(PropertyAuthenticationData)}},
		{
			"Invalid Assigned Client ID",
			ConnAckProperties{Flags: PropertyFlags(0).Set(PropertyAssignedClientID), AssignedClientID: []byte{0}},
		},
		{
			"Invalid Reason String",
			ConnAckProperties{Flags: PropertyFlags(0).Set(PropertyReasonString), ReasonString: []byte{0}},
		},
		{
			"Invalid User Property key",
			ConnAckProperties{
				Flags:          PropertyFlags(0).Set(PropertyUserProperty),
				UserProperties: []UserProperty{{[]byte{0}, []byte("b")}},
			},
		},
		{
			"Invalid User Property value",
			ConnAckProperties{
				Flags:          PropertyFlags(0).Set(PropertyUserProperty),
				UserProperties: []UserProperty{{[]byte("a"), []byte{0}}},
			},
		},
		{
			"Invalid Response Info",
			ConnAckProperties{Flags: PropertyFlags(0).Set(PropertyResponseInfo), ResponseInfo: []byte{0}},
		},
		{
			"Invalid Server Reference",
			ConnAckProperties{Flags: PropertyFlags(0).Set(PropertyServerReference), ServerReference: []byte{0}},
		},
		{
			"Invalid Authentication Method",
			ConnAckProperties{
				Flags:                PropertyFlags(0).Set(PropertyAuthenticationMethod),
				AuthenticationMethod: []byte{0},
			},
		},
		{
			"Invalid Authentication Data",
			ConnAckProperties{
				Flags:              PropertyFlags(0).Set(PropertyAuthenticationData),
				AuthenticationData: []byte{0},
			},
		},
	}

	for _, tc := range testCases {
		props := tc.prop
		t.Run(tc.name, func(t *testing.T) {
			connack := ConnAck{Version: MQTT50, Properties: &props}
			data := make([]byte, connack.Size())

			n, err := connack.Encode(data)
			if !errors.Is(err, ErrMalformedPacket) {
				t.Fatalf("Unexpected error\nwant: %s\ngot:  %s", ErrMalformedPacket, err)
			}
			if n != 0 {
				t.Fatalf("Unexpected number of bytes encoded\nwant: %v\ngot:  %v", 0, n)
			}
		})
	}
}

func BenchmarkConnAckEncode(b *testing.B) {
	testCases := []struct {
		name   string
		packet ConnAck
	}{
		{"V3.1", ConnAck{Version: MQTT31}},
		{"V3.1.1", ConnAck{Version: MQTT311}},
		{"V5.0", ConnAck{Version: MQTT50}},
		{
			"V5.0 Properties",
			ConnAck{
				Version:    MQTT50,
				Properties: &ConnAckProperties{Flags: PropertyFlags(0).Set(PropertyMaximumQoS)},
			},
		},
	}

	for _, test := range testCases {
		b.Run(test.name, func(b *testing.B) {
			data := make([]byte, test.packet.Size())
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, err := test.packet.Encode(data)
				if err != nil {
					b.Fatalf("Unexpected error\n%s", err)
				}
			}
		})
	}
}

func TestConnAckPropertiesOnNil(t *testing.T) {
	var p *ConnAckProperties

	size := p.size()
	if size != 0 {
		t.Fatalf("Unexpected size\nwant: %v\ngot:  %v", 0, size)
	}

	p.Set(PropertyUserProperty)
	if p.Has(PropertyUserProperty) {
		t.Fatalf("Property should not be set on nil")
	}

	p = &ConnAckProperties{}
	p.Set(PropertyUserProperty)
	if !p.Has(PropertyUserProperty) {
		t.Fatalf("Property should be set on nil")
	}
}
