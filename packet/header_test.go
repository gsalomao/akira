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
	"io"
	"reflect"
	"testing"
)

func TestFixedHeaderRead(t *testing.T) {
	testCases := []struct {
		data   []byte
		header FixedHeader
	}{
		{[]byte{0x10, 0x00}, FixedHeader{PacketType: TypeConnect}},
		{[]byte{0x20, 0x7F}, FixedHeader{PacketType: TypeConnAck, RemainingLength: 127}},
		{[]byte{0x30, 0x80, 0x01}, FixedHeader{PacketType: TypePublish, RemainingLength: 128}},
		{[]byte{0x40, 0xFF, 0x7F}, FixedHeader{PacketType: TypePubAck, RemainingLength: 16383}},
		{[]byte{0x50, 0x80, 0x80, 0x01}, FixedHeader{PacketType: TypePubRec, RemainingLength: 16384}},
		{[]byte{0x60, 0xFF, 0xFF, 0x7F}, FixedHeader{PacketType: TypePubRel, RemainingLength: 2097151}},
		{[]byte{0x70, 0x80, 0x80, 0x80, 0x01}, FixedHeader{PacketType: TypePubComp, RemainingLength: 2097152}},
		{[]byte{0x80, 0xFF, 0xFF, 0xFF, 0x7F}, FixedHeader{PacketType: TypeSubscribe, RemainingLength: 268435455}},
		{[]byte{0x91, 0x01}, FixedHeader{PacketType: TypeSubAck, RemainingLength: 1, Flags: 1}},
		{[]byte{0xA2, 0x02}, FixedHeader{PacketType: TypeUnsubscribe, RemainingLength: 2, Flags: 2}},
		{[]byte{0xB3, 0x03}, FixedHeader{PacketType: TypeUnsubAck, RemainingLength: 3, Flags: 3}},
		{[]byte{0xC4, 0x04}, FixedHeader{PacketType: TypePingReq, RemainingLength: 4, Flags: 4}},
		{[]byte{0xD5, 0x05}, FixedHeader{PacketType: TypePingResp, RemainingLength: 5, Flags: 5}},
		{[]byte{0xE6, 0x06}, FixedHeader{PacketType: TypeDisconnect, RemainingLength: 6, Flags: 6}},
		{[]byte{0xF7, 0x07}, FixedHeader{PacketType: TypeAuth, RemainingLength: 7, Flags: 7}},
	}

	for _, tc := range testCases {
		t.Run(tc.header.PacketType.String(), func(t *testing.T) {
			var header FixedHeader
			reader := bufio.NewReader(bytes.NewReader(tc.data))

			n, err := header.Read(reader)
			if err != nil {
				t.Fatalf("Unexpected error\n%s", err)
			}
			if n != len(tc.data) {
				t.Fatalf("Unexpexted number of bytes read\nwant: %v\ngot:  %v", len(tc.data), n)
			}

			if !reflect.DeepEqual(tc.header, header) {
				t.Fatalf("Unexpected header\nwant: %+v\ngot:  %+v", tc.header, header)
			}
		})
	}
}

func TestFixedHeaderReadError(t *testing.T) {
	testCases := []struct {
		name string
		data []byte
		err  error
	}{
		{"Missing packet type", []byte{}, io.EOF},
		{"Missing remaining length", []byte{0x00}, io.EOF},
		{"Invalid remaining length", []byte{0x10, 0xff, 0xff, 0xff, 0x80}, ErrMalformedPacket},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var header FixedHeader
			reader := bufio.NewReader(bytes.NewReader(tc.data))

			n, err := header.Read(reader)
			if !errors.Is(err, tc.err) {
				t.Fatalf("Unexpected error\nwant: %s\ngot:  %s", tc.err, err)
			}
			if n != len(tc.data) {
				t.Fatalf("Unexpexted number of bytes read\nwant: %v\ngot:  %v", len(tc.data), n)
			}
		})
	}
}

func TestFixedHeaderEncode(t *testing.T) {
	testCases := []struct {
		header FixedHeader
		data   []byte
	}{
		{FixedHeader{PacketType: TypeConnect}, []byte{0x10, 0x00}},
		{FixedHeader{PacketType: TypeConnAck, RemainingLength: 127}, []byte{0x20, 0x7F}},
		{FixedHeader{PacketType: TypePublish, RemainingLength: 128}, []byte{0x30, 0x80, 0x01}},
		{FixedHeader{PacketType: TypePubAck, RemainingLength: 16383}, []byte{0x40, 0xFF, 0x7F}},
		{FixedHeader{PacketType: TypePubRec, RemainingLength: 16384}, []byte{0x50, 0x80, 0x80, 0x01}},
		{FixedHeader{PacketType: TypePubRel, RemainingLength: 2097151}, []byte{0x60, 0xFF, 0xFF, 0x7F}},
		{FixedHeader{PacketType: TypePubComp, RemainingLength: 2097152}, []byte{0x70, 0x80, 0x80, 0x80, 0x01}},
		{FixedHeader{PacketType: TypeSubscribe, RemainingLength: 268435455}, []byte{0x80, 0xFF, 0xFF, 0xFF, 0x7F}},
		{FixedHeader{PacketType: TypeSubAck, RemainingLength: 1, Flags: 1}, []byte{0x91, 0x01}},
		{FixedHeader{PacketType: TypeUnsubscribe, RemainingLength: 2, Flags: 2}, []byte{0xA2, 0x02}},
		{FixedHeader{PacketType: TypeUnsubAck, RemainingLength: 3, Flags: 3}, []byte{0xB3, 0x03}},
		{FixedHeader{PacketType: TypePingReq, RemainingLength: 4, Flags: 4}, []byte{0xC4, 0x04}},
		{FixedHeader{PacketType: TypePingResp, RemainingLength: 5, Flags: 5}, []byte{0xD5, 0x05}},
		{FixedHeader{PacketType: TypeDisconnect, RemainingLength: 6, Flags: 6}, []byte{0xE6, 0x06}},
		{FixedHeader{PacketType: TypeAuth, RemainingLength: 7, Flags: 7}, []byte{0xF7, 0x07}},
	}

	for _, tc := range testCases {
		t.Run(tc.header.PacketType.String(), func(t *testing.T) {
			data := make([]byte, len(tc.data))

			n := tc.header.encode(data)
			if n != len(tc.data) {
				t.Fatalf("Unexpected number of bytes encoded\nwant: %v\ngot:  %v", len(tc.data), n)
			}
		})
	}
}
