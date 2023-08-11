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
	"fmt"
	"reflect"
	"testing"

	"github.com/gsalomao/akira/packet/testdata"
)

func TestAuthType(t *testing.T) {
	var p Auth
	tp := p.Type()
	if tp != TypeAuth {
		t.Fatalf("Unexpected type\nwant: %s\ngot:  %s", TypeAuth, tp)
	}
}

func TestAuthSize(t *testing.T) {
	fixtures, err := testdata.ReadFixtures("auth.json")
	if err != nil {
		t.Fatalf("Unexpected error while getting the fixture\n%s", err)
	}

	for i := range fixtures {
		fx := fixtures[i]
		t.Run(fx.Name, func(t *testing.T) {
			p := decodeAuthPacket(t, fx.Packet, nil)

			size := p.Size()
			if size != len(fx.Packet) {
				t.Fatalf("Unxpected packet size\nwant: %v\ngot:  %v", len(fx.Packet), size)
			}
		})
	}
}

func BenchmarkAuthSize(b *testing.B) {
	fixtures, err := testdata.ReadFixtures("auth.json")
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

			var p Auth
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

func TestAuthDecode(t *testing.T) {
	testCases := []struct {
		fixture string
		auth    Auth
	}{
		{"Success", Auth{Code: ReasonCodeSuccess}},
		{
			"Success With Auth Method",
			Auth{
				Code: ReasonCodeSuccess,
				Properties: &AuthProperties{
					Flags:                PropertyFlags(0).Set(PropertyAuthenticationMethod),
					AuthenticationMethod: []byte("a"),
				},
			},
		},
		{"Continue Authentication", Auth{Code: ReasonCodeContinueAuthentication}},
		{"Re-authenticate", Auth{Code: ReasonCodeReAuthenticate}},
		{
			"All Properties",
			Auth{
				Properties: &AuthProperties{
					Flags: PropertyFlags(0).
						Set(PropertyAuthenticationMethod).
						Set(PropertyAuthenticationData).
						Set(PropertyReasonString).
						Set(PropertyUserProperty),
					AuthenticationMethod: []byte("a"),
					AuthenticationData:   []byte{5},
					ReasonString:         []byte("b"),
					UserProperties: []UserProperty{
						{[]byte("c"), []byte("d")},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.fixture, func(t *testing.T) {
			fixture, err := testdata.ReadFixture("auth.json", tc.fixture)
			if err != nil {
				t.Fatalf("Unexpected error while getting the fixture\n%s", err)
			}

			p := decodeAuthPacket(t, fixture.Packet, nil)

			if !reflect.DeepEqual(tc.auth.Properties, p.Properties) {
				t.Fatalf("Unexpected AUTH Properties\nwant: %+v\ngot:  %+v",
					tc.auth.Properties, p.Properties)
			}
			if !reflect.DeepEqual(tc.auth, p) {
				t.Fatalf("Unexpected AUTH packet\nwant: %+v\ngot:  %+v", tc.auth, p)
			}
		})
	}
}

func BenchmarkAuthDecode(b *testing.B) {
	fixtures, err := testdata.ReadFixtures("auth.json")
	if err != nil {
		b.Fatalf("Unexpected error while getting the fixture\n%s", err)
	}
	b.ResetTimer()

	for idx := range fixtures {
		fx := fixtures[idx]
		b.Run(fx.Name, func(b *testing.B) {
			var p Auth
			var header FixedHeader
			var n int

			n, err = header.Read(bufio.NewReader(bytes.NewReader(fx.Packet)))
			if err != nil {
				b.Fatalf("Unexpected error while reading header\n%s", err)
			}

			for i := 0; i < b.N; i++ {
				_, err = p.Decode(fx.Packet[n:], header)
				if err != nil {
					b.Fatalf("Unexpected error while decoding packet\n%s", err)
				}
			}
		})
	}
}

func TestAuthDecodeMalformedPacket(t *testing.T) {
	fixtures, err := testdata.ReadFixtures("auth-malformed.json")
	if err != nil {
		t.Fatalf("Unexpected error while getting the fixture\n%s", err)
	}

	for idx := range fixtures {
		fx := fixtures[idx]
		t.Run(fx.Name, func(t *testing.T) {
			_ = decodeAuthPacket(t, fx.Packet, ErrMalformedPacket)
		})
	}
}

func TestAuthDecodeMalformedPacketReasonCode(t *testing.T) {
	t.Parallel()
	validCodes := []ReasonCode{
		ReasonCodeSuccess,
		ReasonCodeContinueAuthentication,
		ReasonCodeReAuthenticate,
	}

	isValid := func(code ReasonCode) bool {
		for _, c := range validCodes {
			if c == code {
				return true
			}
		}
		return false
	}

	for code := ReasonCode(0); code < 255; code++ {
		c := code
		if !isValid(c) {
			t.Run(fmt.Sprint(c), func(t *testing.T) {
				t.Parallel()
				packet := []byte{240, 2, byte(c), 0}
				_ = decodeAuthPacket(t, packet, ErrMalformedPacket)
			})
		}
	}
}

func TestAuthEncode(t *testing.T) {
	testCases := []struct {
		fixture string
		auth    Auth
	}{
		{"Success", Auth{}},
		{
			"Success With Auth Method",
			Auth{
				Code: ReasonCodeSuccess,
				Properties: &AuthProperties{
					Flags:                PropertyFlags(0).Set(PropertyAuthenticationMethod),
					AuthenticationMethod: []byte("a"),
				},
			},
		},
		{"Continue Authentication", Auth{Code: ReasonCodeContinueAuthentication}},
		{"Re-authenticate", Auth{Code: ReasonCodeReAuthenticate}},
		{
			"All Properties",
			Auth{
				Properties: &AuthProperties{
					Flags: PropertyFlags(0).
						Set(PropertyAuthenticationMethod).
						Set(PropertyAuthenticationData).
						Set(PropertyReasonString).
						Set(PropertyUserProperty),
					AuthenticationMethod: []byte("a"),
					AuthenticationData:   []byte{5},
					ReasonString:         []byte("b"),
					UserProperties: []UserProperty{
						{[]byte("c"), []byte("d")},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.fixture, func(t *testing.T) {
			fixture, err := testdata.ReadFixture("auth.json", tc.fixture)
			if err != nil {
				t.Fatalf("Unexpected error while getting fixture\n%s", err)
			}

			var n int
			pkt := make([]byte, tc.auth.Size())

			n, err = tc.auth.Encode(pkt)
			if err != nil {
				t.Fatalf("Unexpected error\n%s", err)
			}
			if n != tc.auth.Size() {
				t.Fatalf("Unexpected number of bytes encoded\nwant: %v\ngot:  %v", tc.auth.Size(), n)
			}
			if !bytes.Equal(pkt, fixture.Packet) {
				t.Fatalf("Unexpected encoded packet\nwant: %v\ngot:  %v %+v", fixture.Packet, pkt, tc.auth)
			}
		})
	}
}

func TestAuthEncodeErrorBufferTooSmall(t *testing.T) {
	auth := Auth{}

	n, err := auth.Encode(nil)
	if err == nil {
		t.Fatal("An error was expected")
	}
	if n != 0 {
		t.Fatalf("Unexpected number of bytes encoded\nwant: %v\ngot:  %v", 0, n)
	}
}

func TestAuthEncodeMalformedProperty(t *testing.T) {
	testCases := []struct {
		name string
		prop AuthProperties
	}{
		{"Invalid property", AuthProperties{Flags: PropertyFlags(0).Set(PropertyReasonString)}},
		{"Missing Reason String", AuthProperties{Flags: PropertyFlags(0).Set(PropertyReasonString)}},
		{"Missing User Property", AuthProperties{Flags: PropertyFlags(0).Set(PropertyUserProperty)}},
		{"Missing Authentication Method", AuthProperties{Flags: PropertyFlags(0).Set(PropertyAuthenticationMethod)}},
		{"Missing Authentication Data", AuthProperties{Flags: PropertyFlags(0).Set(PropertyAuthenticationData)}},
		{
			"Invalid Reason String",
			AuthProperties{Flags: PropertyFlags(0).Set(PropertyReasonString), ReasonString: []byte{0}},
		},
		{
			"Invalid User Property key",
			AuthProperties{
				Flags:          PropertyFlags(0).Set(PropertyUserProperty),
				UserProperties: []UserProperty{{[]byte{0}, []byte("b")}},
			},
		},
		{
			"Invalid User Property value",
			AuthProperties{
				Flags:          PropertyFlags(0).Set(PropertyUserProperty),
				UserProperties: []UserProperty{{[]byte("a"), []byte{0}}},
			},
		},
		{
			"Invalid Authentication Method",
			AuthProperties{
				Flags:                PropertyFlags(0).Set(PropertyAuthenticationMethod),
				AuthenticationMethod: []byte{0},
			},
		},
	}

	for _, tc := range testCases {
		props := tc.prop
		t.Run(tc.name, func(t *testing.T) {
			auth := Auth{Properties: &props}
			data := make([]byte, auth.Size())

			n, err := auth.Encode(data)
			if !errors.Is(err, ErrMalformedPacket) {
				t.Fatalf("Unexpected error\nwant: %s\ngot:  %s", ErrMalformedPacket, err)
			}
			if n != 0 {
				t.Fatalf("Unexpected number of bytes encoded\nwant: %v\ngot:  %v", 0, n)
			}
		})
	}
}

func BenchmarkAuthEncode(b *testing.B) {
	testCases := []struct {
		name string
		auth Auth
	}{
		{"Success", Auth{}},
		{
			"Success With Auth Method",
			Auth{
				Code: ReasonCodeSuccess,
				Properties: &AuthProperties{
					Flags:                PropertyFlags(0).Set(PropertyAuthenticationMethod),
					AuthenticationMethod: []byte("a"),
				},
			},
		},
		{"Continue Authentication", Auth{Code: ReasonCodeContinueAuthentication}},
		{"Re-authenticate", Auth{Code: ReasonCodeReAuthenticate}},
		{
			"All Properties",
			Auth{
				Properties: &AuthProperties{
					Flags: PropertyFlags(0).
						Set(PropertyAuthenticationMethod).
						Set(PropertyAuthenticationData).
						Set(PropertyReasonString).
						Set(PropertyUserProperty),
					AuthenticationMethod: []byte("a"),
					AuthenticationData:   []byte{5},
					ReasonString:         []byte("b"),
					UserProperties: []UserProperty{
						{[]byte("c"), []byte("d")},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		b.Run(test.name, func(b *testing.B) {
			data := make([]byte, test.auth.Size())
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, err := test.auth.Encode(data)
				if err != nil {
					b.Fatalf("Unexpected error\n%s", err)
				}
			}
		})
	}
}

func TestAuthPropertiesNil(t *testing.T) {
	var p *AuthProperties

	size := p.size()
	if size != 0 {
		t.Fatalf("Unexpected size\nwant: %v\ngot:  %v", 0, size)
	}

	p.Set(PropertyUserProperty)
	if p.Has(PropertyUserProperty) {
		t.Fatalf("Property should not be set on nil")
	}
}

func decodeAuthPacket(t *testing.T, packet []byte, expErr error) Auth {
	t.Helper()

	var header FixedHeader
	n, err := header.Read(bufio.NewReader(bytes.NewReader(packet)))
	if err != nil {
		t.Fatalf("Unexpected error while reading header\n%s", err)
	}

	var p Auth
	_, err = p.Decode(packet[n:], header)
	if expErr == nil && err != nil {
		t.Fatalf("Unexpected error while decoding packet\n%s", err)
	}
	if expErr != nil && !errors.Is(err, expErr) {
		t.Fatalf("Unexpexted error\nwant: %s\ngot:  %v", expErr, err)
	}

	return p
}
