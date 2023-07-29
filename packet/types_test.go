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
	"encoding/binary"
	"errors"
	"fmt"
	"testing"
	"unicode/utf8"
	"unsafe"
)

func TestMQTTVersionString(t *testing.T) {
	testCases := map[byte]string{
		0: "Unknown",
		3: "3.1",
		4: "3.1.1",
		5: "5.0",
	}

	for v, n := range testCases {
		t.Run(fmt.Sprint(v), func(t *testing.T) {
			version := Version(v)
			name := version.String()
			if name != n {
				t.Errorf("Unexpected name\nwant: %s\ngot:  %s", n, name)
			}
		})
	}
}

func TestPacketTypeString(t *testing.T) {
	testCases := map[Type]string{
		TypeReserved:    "RESERVED",
		TypeConnect:     "CONNECT",
		TypeConnAck:     "CONNACK",
		TypePublish:     "PUBLISH",
		TypePubAck:      "PUBACK",
		TypePubRec:      "PUBREC",
		TypePubRel:      "PUBREL",
		TypePubComp:     "PUBCOMP",
		TypeSubscribe:   "SUBSCRIBE",
		TypeSubAck:      "SUBACK",
		TypeUnsubscribe: "UNSUBSCRIBE",
		TypeUnsubAck:    "UNSUBACK",
		TypePingReq:     "PINGREQ",
		TypePingResp:    "PINGRESP",
		TypeDisconnect:  "DISCONNECT",
		TypeAuth:        "AUTH",
		TypeInvalid:     "INVALID",
	}

	for pt, n := range testCases {
		t.Run(pt.String(), func(t *testing.T) {
			name := pt.String()
			if name != n {
				t.Fatalf("Unexpected name\nwant: %s\ngot:  %s", n, name)
			}
		})
	}
}

func TestSizeVarInteger(t *testing.T) {
	testCases := []struct {
		val  int
		size int
	}{
		{0, 1},
		{127, 1},
		{128, 2},
		{16383, 2},
		{16384, 3},
		{2097151, 3},
		{2097152, 4},
		{268435455, 4},
		{268435456, 0},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprint(tc.val), func(t *testing.T) {
			size := sizeVarInteger(tc.val)
			if size != tc.size {
				t.Fatalf("Unexpected size\nwant: %v\ngot:  %v", tc.size, size)
			}
		})
	}
}

func TestReadVarInteger(t *testing.T) {
	testCases := []struct {
		data []byte
		val  int
	}{
		{[]byte{0x00}, 0},
		{[]byte{0x7f}, 127},
		{[]byte{0x80, 0x01}, 128},
		{[]byte{0xff, 0x7f}, 16383},
		{[]byte{0x80, 0x80, 0x01}, 16384},
		{[]byte{0xff, 0xff, 0x7f}, 2097151},
		{[]byte{0x80, 0x80, 0x80, 0x01}, 2097152},
		{[]byte{0xff, 0xff, 0xff, 0x7f}, 268435455},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprint(tc.val), func(t *testing.T) {
			var val int
			reader := bufio.NewReader(bytes.NewReader(tc.data))

			n, err := readVarInteger(reader, &val)
			if err != nil {
				t.Fatalf("Unexpected error\n%s", err)
			}
			if n != len(tc.data) {
				t.Fatalf("Unexpected number of bytes read\nwant: %v\ngot:  %v", len(tc.data), n)
			}
			if val != tc.val {
				t.Fatalf("Unexpected value\nwant: %v\ngot:  %v", tc.val, val)
			}
		})
	}
}

func TestReadVarIntegerError(t *testing.T) {
	var val int
	reader := bufio.NewReader(bytes.NewReader(nil))

	_, err := readVarInteger(reader, &val)
	if err == nil {
		t.Fatal("An error was expected")
	}
}

func TestReadVarIntegerInvalidValue(t *testing.T) {
	var val int
	data := []byte{0xff, 0xff, 0xff, 0x80}
	reader := bufio.NewReader(bytes.NewReader(data))

	_, err := readVarInteger(reader, &val)
	if !errors.Is(err, ErrMalformedPacket) {
		t.Fatalf("Unexpected error\n%s", err)
	}
}

func TestDecodeVarInteger(t *testing.T) {
	testCases := []struct {
		data []byte
		val  int
	}{
		{[]byte{0x00}, 0},
		{[]byte{0x7f}, 127},
		{[]byte{0x80, 0x01}, 128},
		{[]byte{0xff, 0x7f}, 16383},
		{[]byte{0x80, 0x80, 0x01}, 16384},
		{[]byte{0xff, 0xff, 0x7f}, 2097151},
		{[]byte{0x80, 0x80, 0x80, 0x01}, 2097152},
		{[]byte{0xff, 0xff, 0xff, 0x7f}, 268435455},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprint(tc.val), func(t *testing.T) {
			var val int

			n, err := decodeVarInteger(tc.data, &val)
			if err != nil {
				t.Fatalf("Unexpected error\n%s", err)
			}
			if n != len(tc.data) {
				t.Fatalf("Unexpected number of bytes decoded\nwant: %v\ngot:  %v", len(tc.data), n)
			}
			if val != tc.val {
				t.Fatalf("Unexpected value\nwant: %v\ngot:  %v", tc.val, val)
			}
		})
	}
}

func TestDecodeVarIntegerError(t *testing.T) {
	var val int

	_, err := decodeVarInteger(nil, &val)
	if err == nil {
		t.Fatal("An error was expected")
	}

	_, err = decodeVarInteger([]byte{0xff, 0xff, 0xff, 0x80}, &val)
	if err == nil {
		t.Fatal("An error was expected")
	}
}

func TestDecodeString(t *testing.T) {
	testCases := []struct {
		data []byte
		str  []byte
	}{
		{[]byte{0, 0}, []byte("")},
		{[]byte{0, 1, 'a'}, []byte("a")},
		{[]byte{0, 3, 'a', 'b', 'c'}, []byte("abc")},
	}

	for _, tc := range testCases {
		t.Run(string(tc.str), func(t *testing.T) {
			str, n, err := decodeString(tc.data)
			if err != nil {
				t.Fatalf("Unexpected error\n%s", err)
			}
			if n != len(tc.data) {
				t.Fatalf("Unexpected number of bytes decoded\nwant: %v\ngot:  %v", len(tc.data), n)
			}
			if !bytes.Equal(str, tc.str) {
				t.Fatalf("Unexpected string\nwant: %v\ngot:  %v", tc.str, str)
			}
		})
	}
}

func TestDecodeStringInvalid(t *testing.T) {
	testCases := []rune{0x00, 0x1f, 0x7f, 0x9f, 0xd800, 0xdfff}

	for _, tc := range testCases {
		code := make([]byte, 4)
		cpLenBuf := make([]byte, 2)

		codeLen := utf8.EncodeRune(code, tc)
		code = code[:codeLen]
		binary.BigEndian.PutUint16(cpLenBuf, uint16(codeLen))

		data := make([]byte, 0, len(cpLenBuf)+len(code))
		data = append(data, cpLenBuf...)
		data = append(data, code...)

		t.Run(fmt.Sprint(tc), func(t *testing.T) {
			_, _, err := decodeString(data)
			if err == nil {
				t.Fatal("An error was expected")
			}
		})
	}
}

func TestDecodeStringError(t *testing.T) {
	_, _, err := decodeString([]byte{})
	if err == nil {
		t.Fatal("An error was expected")
	}
}

func TestEncodeString(t *testing.T) {
	testCases := []struct {
		str  string
		data []byte
	}{
		{"", []byte{0, 0}},
		{"abc", []byte{0, 3, 'a', 'b', 'c'}},
	}

	for _, tc := range testCases {
		t.Run(tc.str, func(t *testing.T) {
			data := make([]byte, len(tc.data))

			n, err := encodeString(data, []byte(tc.str))
			if err != nil {
				t.Fatalf("Unexpected error\n%s", err)
			}
			if n != len(tc.data) {
				t.Fatalf("Unexpected number of bytes encoded\nwant: %v\ngot:  %v", len(tc.data), n)
			}
			if !bytes.Equal(data, tc.data) {
				t.Fatalf("Unexpected string\nwant: %v\ngot:  %v", tc.data, data)
			}
		})
	}
}

func TestEncodeStringError(t *testing.T) {
	_, err := encodeString(make([]byte, 10), []byte{0})
	if err == nil {
		t.Fatal("An error was expected")
	}
}

func TestDecodeBinary(t *testing.T) {
	testCases := []struct {
		data []byte
		bin  []byte
	}{
		{[]byte{0, 0}, []byte("")},
		{[]byte{0, 1, 'a'}, []byte("a")},
		{[]byte{0, 3, 'a', 'b', 'c'}, []byte("abc")},
	}

	for _, tc := range testCases {
		t.Run(string(tc.bin), func(t *testing.T) {
			bin, n, err := decodeBinary(tc.data)
			if err != nil {
				t.Fatalf("Unexpected error\n%s", err)
			}
			if n != len(tc.data) {
				t.Fatalf("Unexpected number of bytes decoded\nwant: %v\ngot:  %v", len(tc.data), n)
			}
			if !bytes.Equal(bin, tc.bin) {
				t.Fatalf("Unexpected bin\nwant: %v\ngot:  %v", tc.bin, bin)
			}
		})
	}
}

func TestDecodeBinaryError(t *testing.T) {
	_, _, err := decodeBinary([]byte{})
	if err == nil {
		t.Fatal("An error was expected")
	}
}

func TestEncodeBinary(t *testing.T) {
	testCases := []struct {
		name string
		bin  []byte
		data []byte
	}{
		{"Empty", []byte{}, []byte{0, 0}},
		{"Non-Empty", []byte{0, 1, 2}, []byte{0, 3, 0, 1, 2}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := make([]byte, len(tc.data))

			n := encodeBinary(data, tc.bin)
			if n != len(tc.data) {
				t.Fatalf("Unexpected number of bytes encoded\nwant: %v\ngot:  %v", len(tc.data), n)
			}
			if !bytes.Equal(data, tc.data) {
				t.Fatalf("Unexpected bin\nwant: %v\ngot:  %v", tc.data, data)
			}
		})
	}
}

func TestDecodeUintErrorNoData(t *testing.T) {
	_, err := decodeUint[uint8](nil)
	if err == nil {
		t.Fatal("An error was expected")
	}
}

func TestDecodeUint8(t *testing.T) {
	testCases := []struct {
		data []byte
		val  uint8
	}{
		{[]byte{0}, 0},
		{[]byte{255}, 255},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprint(tc.val), func(t *testing.T) {
			val, err := decodeUint[uint8](tc.data)
			if err != nil {
				t.Fatalf("Unexpected error\n%s", err)
			}
			if val != tc.val {
				t.Fatalf("Unexpected value\nwant: %v\ngot:  %v", tc.val, val)
			}
		})
	}
}

func TestDecodeUint16(t *testing.T) {
	testCases := []struct {
		data []byte
		val  uint16
	}{
		{[]byte{0, 0}, 0},
		{[]byte{0, 255}, 255},
		{[]byte{1, 0}, 256},
		{[]byte{255, 255}, 65535},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprint(tc.val), func(t *testing.T) {
			val, err := decodeUint[uint16](tc.data)
			if err != nil {
				t.Fatalf("Unexpected error\n%s", err)
			}
			if val != tc.val {
				t.Fatalf("Unexpected value\nwant: %v\ngot:  %v", tc.val, val)
			}
		})
	}
}

func TestDecodeUint32(t *testing.T) {
	testCases := []struct {
		data []byte
		val  uint32
	}{
		{[]byte{0, 1, 0, 0}, 65536},
		{[]byte{255, 255, 255, 255}, 4294967295},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprint(tc.val), func(t *testing.T) {
			val, err := decodeUint[uint32](tc.data)
			if err != nil {
				t.Fatalf("Unexpected error\n%s", err)
			}
			if val != tc.val {
				t.Fatalf("Unexpected value\nwant: %v\ngot:  %v", tc.val, val)
			}
		})
	}
}

func TestEncodeUint8(t *testing.T) {
	testCases := []struct {
		val  uint8
		data []byte
	}{
		{0, []byte{0}},
		{255, []byte{255}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprint(tc.val), func(t *testing.T) {
			data := make([]byte, len(tc.data))

			n := encodeUint(data, tc.val)
			if n != int(unsafe.Sizeof(tc.val)) {
				t.Fatalf("Unexpected number of bytes encoded\nwant: %v\ngot:  %v", int(unsafe.Sizeof(tc.val)), n)
			}
			if !bytes.Equal(data, tc.data) {
				t.Fatalf("Unexpected data\nwant: %v\ngot:  %v", tc.data, data)
			}
		})
	}
}

func TestEncodeUint16(t *testing.T) {
	testCases := []struct {
		val  uint16
		data []byte
	}{
		{0, []byte{0, 0}},
		{255, []byte{0, 255}},
		{256, []byte{1, 0}},
		{65535, []byte{255, 255}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprint(tc.val), func(t *testing.T) {
			data := make([]byte, len(tc.data))

			n := encodeUint(data, tc.val)
			if n != int(unsafe.Sizeof(tc.val)) {
				t.Fatalf("Unexpected number of bytes encoded\nwant: %v\ngot:  %v", int(unsafe.Sizeof(tc.val)), n)
			}
			if !bytes.Equal(data, tc.data) {
				t.Fatalf("Unexpected data\nwant: %v\ngot:  %v", tc.data, data)
			}
		})
	}
}

func TestEncodeUint32(t *testing.T) {
	testCases := []struct {
		val  uint32
		data []byte
	}{
		{65536, []byte{0, 1, 0, 0}},
		{4294967295, []byte{255, 255, 255, 255}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprint(tc.val), func(t *testing.T) {
			data := make([]byte, len(tc.data))

			n := encodeUint(data, tc.val)
			if n != int(unsafe.Sizeof(tc.val)) {
				t.Fatalf("Unexpected number of bytes encoded\nwant: %v\ngot:  %v", int(unsafe.Sizeof(tc.val)), n)
			}
			if !bytes.Equal(data, tc.data) {
				t.Fatalf("Unexpected data\nwant: %v\ngot:  %v", tc.data, data)
			}
		})
	}
}
