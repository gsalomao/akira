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
	"encoding/binary"
	"strings"
	"unicode/utf8"
	"unsafe"

	"golang.org/x/exp/constraints"
)

const (
	// MQTT31 represents the MQTT version 3.1.
	MQTT31 MQTTVersion = iota + 3

	// MQTT311 represents the MQTT version 3.1.1.
	MQTT311

	// MQTT50 represents the MQTT version 5.0.
	MQTT50
)

const (
	// QoS0 indicates the quality of service 0.
	QoS0 QoS = iota

	// QoS1 indicates the quality of service 1.
	QoS1

	// QoS2 indicates the quality of service 2.
	QoS2
)

const (
	// TypeReserved is the reserved packet type.
	TypeReserved Type = iota

	// TypeConnect is the CONNECT packet type.
	TypeConnect

	// TypeConnAck is the CONNACK packet type.
	TypeConnAck

	// TypePublish is the PUBLISH packet type.
	TypePublish

	// TypePubAck is the PUBACK packet type.
	TypePubAck

	// TypePubRec is the PUBREC packet type.
	TypePubRec

	// TypePubRel is the PUBREL packet type.
	TypePubRel

	// TypePubComp is the PUBCOMP packet type.
	TypePubComp

	// TypeSubscribe is the SUBSCRIBE packet type.
	TypeSubscribe

	// TypeSubAck is the SUBACK packet type.
	TypeSubAck

	// TypeUnsubscribe is the UNSUBSCRIBE packet type.
	TypeUnsubscribe

	// TypeUnsubAck is the UNSUBSCRIBE packet type.
	TypeUnsubAck

	// TypePingReq is the PINGREQ packet type.
	TypePingReq

	// TypePingResp is the PINGRESP packet type.
	TypePingResp

	// TypeDisconnect is the DISCONNECT packet type.
	TypeDisconnect

	// TypeAuth is the AUTH packet type.
	TypeAuth

	// TypeInvalid indicates that the packet type is invalid.
	TypeInvalid
)

const (
	packetTypeBitsShift = 4
	packetFlagsBitsMask = 0x0f
)

var packetTypeToString = map[Type]string{
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
}

// MQTTVersion represents the MQTT version.
type MQTTVersion byte

// QoS represents the quality of service.
type QoS byte

// Type represents the packet type (e.g. CONNECT, CONNACK, etc.).
type Type byte

// String returns the packet type name.
func (t Type) String() string {
	name, ok := packetTypeToString[t]
	if !ok {
		return "INVALID"
	}

	return name
}

func sizeVarInteger(val int) int {
	if val < 128 {
		return 1
	}
	if val < 16384 {
		return 2
	}
	if val < 2097152 {
		return 3
	}
	if val < 268435456 {
		return 4
	}
	return 0
}

func sizeUint[T constraints.Unsigned](val T) int {
	return int(unsafe.Sizeof(val))
}

func sizeBinary(data []byte) int {
	// The size of the data, +2 bytes for the data length.
	return len(data) + 2
}

func sizeString(s string) int {
	// The size of the string, +2 bytes for the string length.
	return len(s) + 2
}

func readVarInteger(r *bufio.Reader, val *int) (int, error) {
	var n int
	multiplier := 1
	*val = 0

	for {
		b, err := r.ReadByte()
		if err != nil {
			return n, err
		}

		n++
		*val += int(b&127) * multiplier

		multiplier *= 128
		if b&128 == 0 {
			break
		}

		if multiplier > 128*128*128 || n == 4 {
			return n, ErrMalformedVarInteger
		}
	}

	return n, nil
}

func decodeVarInteger(buf []byte, val *int) (int, error) {
	var n int
	multiplier := 1
	*val = 0

	for {
		if n >= len(buf) {
			return n, ErrMalformedVarInteger
		}

		b := buf[n]
		n++
		*val += int(b&127) * multiplier

		multiplier *= 128
		if b&128 == 0 {
			break
		}

		if multiplier > 128*128*128 || n == 4 {
			return n, ErrMalformedVarInteger
		}
	}

	return n, nil
}

func decodeUint[T constraints.Unsigned](data []byte, val *T) (err error) {
	size := int(unsafe.Sizeof(*val))

	if size > len(data) {
		return ErrMalformedInteger
	}

	switch size {
	case 1:
		*val = T(data[0])
	case 2:
		*val = T(binary.BigEndian.Uint16(data[:size]))
	case 4:
		*val = T(binary.BigEndian.Uint32(data[:size]))
	default:
		return ErrMalformedInteger
	}

	return nil
}

func decodeBool(data []byte, val *bool) error {
	var b byte
	err := decodeUint(data, &b)

	switch b {
	case 0:
		*val = false
	case 1:
		*val = true
	default:
		err = ErrMalformedInteger
	}

	return err
}

func decodeBinary(data []byte) ([]byte, int, error) {
	var n int
	var length uint16

	err := decodeUint[uint16](data, &length)
	if err != nil {
		return nil, 0, ErrMalformedBinary
	}
	n += 2

	if int(length) > len(data)-n {
		return nil, n, ErrMalformedBinary
	}

	bin := data[n : n+int(length)]
	n += int(length)

	return bin, n, nil
}

func decodeString(data []byte) ([]byte, int, error) {
	str, n, err := decodeBinary(data)
	if err != nil {
		return nil, n, ErrMalformedString
	}

	if !isValidString(str) {
		return nil, n, ErrMalformedString
	}

	return str, n, nil
}

func encodeVarInteger(buf []byte, val int) int {
	var n int
	var data byte

	for {
		data = byte(val % 128)

		val /= 128
		if val > 0 {
			data |= 128
		}

		buf[n] = data
		n++
		if val == 0 {
			return n
		}
	}
}

func encodeUint[T constraints.Unsigned](buf []byte, val T) int {
	size := int(unsafe.Sizeof(val))

	switch size {
	case 1:
		buf[0] = byte(val)
	case 2:
		binary.BigEndian.PutUint16(buf, uint16(val))
	case 4:
		binary.BigEndian.PutUint32(buf, uint32(val))
	default:
		return 0
	}

	return size
}

func encodeBool(buf []byte, val bool) int {
	var b byte
	if val {
		b = 1
	}
	return encodeUint(buf, b)
}

func encodeBinary(buf []byte, bin []byte) int {
	n := encodeUint(buf, uint16(len(bin)))

	copy(buf[n:], bin)
	n += len(bin)

	return n
}

func encodeString(buf []byte, str []byte) (int, error) {
	if !isValidString(str) {
		return 0, ErrMalformedString
	}

	n := encodeBinary(buf, str)
	return n, nil
}

func isValidString(str []byte) bool {
	s := str
	for len(s) > 0 {
		rn, size := utf8.DecodeRune(s)

		if rn == utf8.RuneError || !utf8.ValidRune(rn) {
			return false
		}

		if '\u0000' <= rn && rn <= '\u001F' || '\u007F' <= rn && rn <= '\u009F' {
			return false
		}

		s = s[size:]
	}
	return true
}

func isValidTopicName(topic string) bool {
	if len(topic) == 0 {
		return false
	}

	words := strings.Split(topic, "/")

	for _, word := range words {
		if strings.Contains(word, "#") || strings.Contains(word, "+") {
			return false
		}
	}

	return true
}
