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
	"bufio"
	"bytes"
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/gsalomao/akira/packet/testdata"
)

func TestConnectType(t *testing.T) {
	var p Connect
	tp := p.Type()
	if tp != TypeConnect {
		t.Fatalf("Unexpected type\nwant: %s\ngot:  %s", TypeConnect, tp)
	}
}

func TestConnectSize(t *testing.T) {
	fixtures, err := testdata.ReadFixtures("connect.json")
	if err != nil {
		t.Fatalf("Unexpected error while getting the fixture\n%s", err)
	}

	for i := range fixtures {
		fx := fixtures[i]
		t.Run(fx.Name, func(t *testing.T) {
			p := decodeConnectPacket(t, fx.Packet, nil)

			size := p.Size()
			if size != len(fx.Packet) {
				t.Fatalf("Unxpected packet size\nwant: %v\ngot:  %v", len(fx.Packet), size)
			}
		})
	}
}

func BenchmarkConnectSize(b *testing.B) {
	fixtures, err := testdata.ReadFixtures("connect.json")
	if err != nil {
		b.Fatalf("Unexpected error while getting the fixture\n%s", err)
	}

	for idx := range fixtures {
		fx := fixtures[idx]
		b.Run(fx.Name, func(b *testing.B) {
			var header FixedHeader
			var n int

			n, err = header.Read(bufio.NewReader(bytes.NewReader(fx.Packet)))
			if err != nil {
				b.Fatalf("Unexpected error while reading header\n%s", err)
			}

			var p Connect
			_, err = p.Decode(fx.Packet[n:], header)
			if err != nil {
				b.Fatalf("Unexpected error while decoding packet\n%s", err)
			}

			for i := 0; i < b.N; i++ {
				size := p.Size()
				if size != len(fx.Packet) {
					b.Fatalf("Unxpected packet size\nwant: %v\ngot:  %v", len(fx.Packet), size)
				}
			}
		})
	}
}

func TestConnectDecode(t *testing.T) {
	testCases := []struct {
		fixture string
		connect Connect
	}{
		{"V3.1", Connect{Version: MQTT31, KeepAlive: 30, ClientID: []byte("a")}},
		{"V3.1.1", Connect{Version: MQTT311, KeepAlive: 50, ClientID: []byte("ab")}},
		{
			"V3.1.1 No Client ID", Connect{
				Version: MQTT311, Flags: ConnectFlags(connectFlagCleanSession), ClientID: []byte{},
			},
		},
		{
			"V5.0", Connect{
				Version: MQTT50, KeepAlive: 80, Flags: ConnectFlags(connectFlagCleanSession), ClientID: []byte("abc"),
			},
		},
		{
			"V5.0 Username/Password", Connect{
				Version: MQTT50, Flags: ConnectFlags(connectFlagUsernameFlag | connectFlagPasswordFlag),
				ClientID: []byte("abc"), Username: []byte("user"), Password: []byte("pass"),
			},
		},
		{
			"V5.0 Connect Properties", Connect{
				Version: MQTT50, ClientID: []byte("abc"),
				Properties: &ConnectProperties{
					Flags: PropertyFlags(0).
						Set(PropertySessionExpiryInterval).
						Set(PropertyReceiveMaximum).
						Set(PropertyMaximumPacketSize).
						Set(PropertyTopicAliasMaximum).
						Set(PropertyRequestResponseInfo).
						Set(PropertyRequestProblemInfo).
						Set(PropertyAuthenticationMethod).
						Set(PropertyAuthenticationData).
						Set(PropertyUserProperty),
					SessionExpiryInterval: 10,
					ReceiveMaximum:        50,
					MaximumPacketSize:     200,
					TopicAliasMaximum:     50,
					RequestResponseInfo:   true,
					RequestProblemInfo:    true,
					UserProperties: []UserProperty{
						{[]byte("a"), []byte("b")},
						{[]byte("c"), []byte("d")},
					},
					AuthenticationMethod: []byte("ef"),
					AuthenticationData:   []byte{10},
				},
			},
		},
		{
			"V5.0 Will", Connect{
				Version:     MQTT50,
				Flags:       ConnectFlags(connectFlagWillFlag | connectFlagWillRetain | (1 << connectFlagShiftWillQoS)),
				ClientID:    []byte("abc"),
				WillTopic:   []byte("a"),
				WillPayload: []byte("b"),
				WillProperties: &WillProperties{
					Flags: PropertyFlags(0).
						Set(PropertyWillDelayInterval).
						Set(PropertyPayloadFormatIndicator).
						Set(PropertyMessageExpiryInterval).
						Set(PropertyContentType).
						Set(PropertyResponseTopic).
						Set(PropertyCorrelationData).
						Set(PropertyUserProperty),
					WillDelayInterval:      15,
					PayloadFormatIndicator: true,
					MessageExpiryInterval:  10,
					ContentType:            []byte("json"),
					ResponseTopic:          []byte("b"),
					CorrelationData:        []byte{20, 1},
					UserProperties:         []UserProperty{{[]byte("c"), []byte("d")}},
				},
			},
		},
		{
			"V5.0 Connect and Will Properties", Connect{
				Version:  MQTT50,
				Flags:    ConnectFlags(connectFlagWillFlag | (2 << connectFlagShiftWillQoS)),
				ClientID: []byte("abc"),
				Properties: &ConnectProperties{
					Flags:                 PropertyFlags(0).Set(PropertySessionExpiryInterval),
					SessionExpiryInterval: 10,
				},
				WillTopic:   []byte("a"),
				WillPayload: []byte("b"),
				WillProperties: &WillProperties{
					Flags:             PropertyFlags(0).Set(PropertyWillDelayInterval),
					WillDelayInterval: 15,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.fixture, func(t *testing.T) {
			fixture, err := testdata.FixtureWithName("connect.json", tc.fixture)
			if err != nil {
				t.Fatalf("Unexpected error while getting the fixture\n%s", err)
			}

			p := decodeConnectPacket(t, fixture.Packet, nil)

			if !reflect.DeepEqual(tc.connect.Properties, p.Properties) {
				t.Fatalf("Unexpected CONNECT Properties\nwant: %+v\ngot:  %+v",
					tc.connect.Properties, p.Properties)
			}
			if !reflect.DeepEqual(tc.connect.WillProperties, p.WillProperties) {
				t.Fatalf("Unexpected Will Properties\nwant: %+v\ngot:  %+v",
					tc.connect.WillProperties, p.WillProperties)
			}
			if !reflect.DeepEqual(tc.connect, p) {
				t.Fatalf("Unexpected CONNECT packet\nwant: %+v\ngot:  %+v", tc.connect, p)
			}

			flags := connectFlags(t, &p)
			sort.Strings(fixture.Flags)
			if !reflect.DeepEqual(fixture.Flags, flags) {
				t.Fatalf("Unexpected flags\nwant: %v\ngot:  %v", fixture.Flags, flags)
			}
		})
	}
}

func TestConnectDecodeErrMalformedPacket(t *testing.T) {
	fixtures, err := testdata.ReadFixtures("connect-malformed.json")
	if err != nil {
		t.Fatalf("Unexpected error while getting the fixture\n%s", err)
	}

	for idx := range fixtures {
		fx := fixtures[idx]
		t.Run(fx.Name, func(t *testing.T) {
			_ = decodeConnectPacket(t, fx.Packet, ErrMalformedPacket)
		})
	}
}

func BenchmarkConnectDecode(b *testing.B) {
	fixtures, err := testdata.ReadFixtures("connect.json")
	if err != nil {
		b.Fatalf("Unexpected error while getting the fixture\n%s", err)
	}

	for idx := range fixtures {
		fx := fixtures[idx]
		b.Run(fx.Name, func(b *testing.B) {
			var header FixedHeader
			var n int

			n, err = header.Read(bufio.NewReader(bytes.NewReader(fx.Packet)))
			if err != nil {
				b.Fatalf("Unexpected error while reading header\n%s", err)
			}

			for i := 0; i < b.N; i++ {
				var p Connect
				_, err = p.Decode(fx.Packet[n:], header)
				if err != nil {
					b.Fatalf("Unexpected error while decoding packet\n%s", err)
				}
			}
		})
	}
}

func TestConnectPropertiesNil(t *testing.T) {
	var p *ConnectProperties

	size := p.size()
	if size != 0 {
		t.Fatalf("Unexpected size\nwant: %v\ngot:  %v", 0, size)
	}

	p.Set(PropertyUserProperty)
	if p.Has(PropertyUserProperty) {
		t.Fatalf("Property should not be set on nil")
	}
}

func TestWillPropertiesNil(t *testing.T) {
	var p *WillProperties

	size := p.size()
	if size != 0 {
		t.Fatalf("Unexpected size\nwant: %v\ngot:  %v", 0, size)
	}

	p.Set(PropertyUserProperty)
	if p.Has(PropertyUserProperty) {
		t.Fatalf("Property should not be set on nil")
	}
}

func decodeConnectPacket(t *testing.T, packet []byte, expErr error) Connect {
	t.Helper()

	var header FixedHeader
	n, err := header.Read(bufio.NewReader(bytes.NewReader(packet)))
	if err != nil {
		t.Fatalf("Unexpected error while reading header\n%s", err)
	}

	var p Connect
	_, err = p.Decode(packet[n:], header)
	if expErr == nil && err != nil {
		t.Fatalf("Unexpected error while decoding packet\n%s", err)
	}
	if expErr != nil && !errors.Is(err, expErr) {
		t.Fatalf("Unexpexted error\nwant: %s\ngot:  %v", expErr, err)
	}

	return p
}

func connectFlags(t *testing.T, p *Connect) []string {
	t.Helper()

	var flags []string
	if p.Flags.Reserved() {
		flags = append(flags, testdata.ConnectFlagReserved)
	}
	if p.Flags.CleanSession() {
		flags = append(flags, testdata.ConnectFlagCleanSession)
	}
	if p.Flags.WillFlag() {
		flags = append(flags, testdata.ConnectFlagWillFlag)
		switch p.Flags.WillQoS() {
		case QoS0:
			flags = append(flags, testdata.ConnectFlagWillQoS0)
		case QoS1:
			flags = append(flags, testdata.ConnectFlagWillQoS1)
		case QoS2:
			flags = append(flags, testdata.ConnectFlagWillQoS2)
		default:
			t.Fatalf("Invalid QoS: %v", p.Flags.WillQoS())
		}
	}
	if p.Flags.WillRetain() {
		flags = append(flags, testdata.ConnectFlagWillRetain)
	}
	if p.Flags.Password() {
		flags = append(flags, testdata.ConnectFlagPassword)
	}
	if p.Flags.Username() {
		flags = append(flags, testdata.ConnectFlagUsername)
	}

	sort.Strings(flags)
	return flags
}
