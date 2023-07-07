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
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPacketTypeString(t *testing.T) {
	types := map[PacketType]string{
		PacketTypeReserved:    "RESERVED",
		PacketTypeConnect:     "CONNECT",
		PacketTypeConnAck:     "CONNACK",
		PacketTypePublish:     "PUBLISH",
		PacketTypePubAck:      "PUBACK",
		PacketTypePubRec:      "PUBREC",
		PacketTypePubRel:      "PUBREL",
		PacketTypePubComp:     "PUBCOMP",
		PacketTypeSubscribe:   "SUBSCRIBE",
		PacketTypeSubAck:      "SUBACK",
		PacketTypeUnsubscribe: "UNSUBSCRIBE",
		PacketTypeUnsubAck:    "UNSUBACK",
		PacketTypePingReq:     "PINGREQ",
		PacketTypePingResp:    "PINGRESP",
		PacketTypeDisconnect:  "DISCONNECT",
		PacketTypeAuth:        "AUTH",
		PacketTypeInvalid:     "INVALID",
	}

	for pt, n := range types {
		t.Run(pt.String(), func(t *testing.T) {
			name := pt.String()
			require.Equal(t, n, name)
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

	for _, test := range testCases {
		t.Run(fmt.Sprint(test.val), func(t *testing.T) {
			size := sizeVarInteger(test.val)
			assert.Equal(t, test.size, size)
		})
	}
}

func TestReadVarIntegerSuccess(t *testing.T) {
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

	for _, test := range testCases {
		t.Run(fmt.Sprint(test.val), func(t *testing.T) {
			var val int
			reader := bufio.NewReader(bytes.NewReader(test.data))

			n, err := readVarInteger(reader, &val)
			require.NoError(t, err)
			assert.Equal(t, len(test.data), n)
			assert.Equal(t, test.val, val)
		})
	}
}

func TestReadVarIntegerReadError(t *testing.T) {
	var val int
	reader := bufio.NewReader(bytes.NewReader(nil))

	_, err := readVarInteger(reader, &val)
	require.Error(t, err)
}

func TestReadVarIntegerInvalidValue(t *testing.T) {
	var val int
	data := []byte{0xff, 0xff, 0xff, 0x80}
	reader := bufio.NewReader(bytes.NewReader(data))

	_, err := readVarInteger(reader, &val)
	require.ErrorIs(t, err, ErrMalformedVarInteger)
}

func TestDecodeVarIntegerSuccess(t *testing.T) {
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

	for _, test := range testCases {
		t.Run(fmt.Sprint(test.val), func(t *testing.T) {
			var val int

			n, err := decodeVarInteger(test.data, &val)
			require.NoError(t, err)
			assert.Equal(t, len(test.data), n)
			assert.Equal(t, test.val, val)
		})
	}
}

func TestDecodeVarIntegerReadError(t *testing.T) {
	var val int

	_, err := decodeVarInteger(nil, &val)
	require.Error(t, err)
}

func TestDecodeVarIntegerInvalidValue(t *testing.T) {
	var val int
	data := []byte{0xff, 0xff, 0xff, 0x80}

	_, err := decodeVarInteger(data, &val)
	require.ErrorIs(t, err, ErrMalformedVarInteger)
}

func TestDecodeStringSuccess(t *testing.T) {
	testCases := []struct {
		data []byte
		str  []byte
	}{
		{[]byte{0, 0}, []byte("")},
		{[]byte{0, 1, 'a'}, []byte("a")},
		{[]byte{0, 3, 'a', 'b', 'c'}, []byte("abc")},
	}

	for _, test := range testCases {
		t.Run(string(test.str), func(t *testing.T) {
			str, n, err := decodeString(test.data)
			require.NoError(t, err)
			assert.Equal(t, test.str, str)
			assert.Equal(t, len(test.data), n)
		})
	}
}

func TestDecodeStringInvalid(t *testing.T) {
	testCases := []rune{0x00, 0x1f, 0x7f, 0x9f, 0xd800, 0xdfff}

	for _, test := range testCases {
		code := make([]byte, 4)
		cpLenBuf := make([]byte, 2)

		codeLen := utf8.EncodeRune(code, test)
		code = code[:codeLen]
		binary.BigEndian.PutUint16(cpLenBuf, uint16(codeLen))

		data := make([]byte, 0, len(cpLenBuf)+len(code))
		data = append(data, cpLenBuf...)
		data = append(data, code...)

		t.Run(fmt.Sprint(test), func(t *testing.T) {
			_, _, err := decodeString(data)
			require.ErrorIs(t, err, ErrMalformedString)
		})
	}
}

func TestDecodeStringError(t *testing.T) {
	_, _, err := decodeString([]byte{})
	require.ErrorIs(t, err, ErrMalformedString)
}

func TestEncodeStringSuccess(t *testing.T) {
	testCases := []struct {
		val string
		str []byte
	}{
		{"", []byte{0, 0}},
		{"abc", []byte{0, 3, 'a', 'b', 'c'}},
	}

	for _, test := range testCases {
		t.Run(test.val, func(t *testing.T) {
			data := make([]byte, len(test.str))

			n, err := encodeString(data, []byte(test.val))
			require.NoError(t, err)
			assert.Equal(t, len(test.str), n)
			assert.Equal(t, test.str, data)
		})
	}
}

func TestEncodeStringError(t *testing.T) {
	_, err := encodeString(make([]byte, 10), []byte{0})
	require.Error(t, err)
}

func TestDecodeBinarySuccess(t *testing.T) {
	testCases := []struct {
		data []byte
		bin  []byte
	}{
		{[]byte{0, 0}, []byte("")},
		{[]byte{0, 1, 'a'}, []byte("a")},
		{[]byte{0, 3, 'a', 'b', 'c'}, []byte("abc")},
	}

	for _, test := range testCases {
		t.Run(string(test.bin), func(t *testing.T) {
			bin, n, err := decodeBinary(test.data)
			require.NoError(t, err)
			assert.Equal(t, test.bin, bin)
			assert.Equal(t, len(test.data), n)
		})
	}
}

func TestDecodeBinaryError(t *testing.T) {
	_, _, err := decodeBinary([]byte{})
	require.ErrorIs(t, err, ErrMalformedBinary)
}

func TestEncodeBinarySuccess(t *testing.T) {
	testCases := []struct {
		name string
		val  []byte
		bin  []byte
	}{
		{"Empty", []byte{}, []byte{0, 0}},
		{"Non-Empty", []byte{0, 1, 2}, []byte{0, 3, 0, 1, 2}},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			data := make([]byte, len(test.bin))

			n := encodeBinary(data, test.val)
			assert.Equal(t, len(test.bin), n)
			assert.Equal(t, test.bin, data)
		})
	}
}

func TestDecodeUintErrorNoData(t *testing.T) {
	err := decodeUint[uint8]([]byte{}, nil)
	require.ErrorIs(t, err, ErrMalformedInteger)
}

func TestDecodeUintErrorSize(t *testing.T) {
	err := decodeUint[uint64]([]byte{0, 0, 0, 0, 0, 0, 0, 0}, nil)
	require.Error(t, err)
}

func TestDecodeUint8(t *testing.T) {
	testCases := []struct {
		data []byte
		val  uint8
	}{
		{[]byte{0}, 0},
		{[]byte{255}, 255},
	}

	for _, test := range testCases {
		t.Run(fmt.Sprint(test.val), func(t *testing.T) {
			var val uint8

			err := decodeUint[uint8](test.data, &val)
			require.NoError(t, err)
			assert.Equal(t, test.val, val)
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

	for _, test := range testCases {
		t.Run(fmt.Sprint(test.val), func(t *testing.T) {
			var val uint16

			err := decodeUint[uint16](test.data, &val)
			require.NoError(t, err)
			assert.Equal(t, test.val, val)
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

	for _, test := range testCases {
		t.Run(fmt.Sprint(test.val), func(t *testing.T) {
			var val uint32

			err := decodeUint[uint32](test.data, &val)
			require.NoError(t, err)
			assert.Equal(t, test.val, val)
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

	for _, test := range testCases {
		t.Run(fmt.Sprint(test.val), func(t *testing.T) {
			data := make([]byte, len(test.data))

			n := encodeUint(data, test.val)
			assert.Equal(t, 1, n)
			assert.Equal(t, test.data, data)
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

	for _, test := range testCases {
		t.Run(fmt.Sprint(test.val), func(t *testing.T) {
			data := make([]byte, len(test.data))

			n := encodeUint(data, test.val)
			assert.Equal(t, 2, n)
			assert.Equal(t, test.data, data)
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

	for _, test := range testCases {
		t.Run(fmt.Sprint(test.val), func(t *testing.T) {
			data := make([]byte, len(test.data))

			n := encodeUint(data, test.val)
			assert.Equal(t, 4, n)
			assert.Equal(t, test.data, data)
		})
	}
}

func TestEncodeUintInvalidSize(t *testing.T) {
	n := encodeUint(make([]byte, 10), uint64(10))
	assert.Zero(t, n)
}
