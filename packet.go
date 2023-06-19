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
	"bufio"
	"io"
	"time"
)

const (
	// PacketTypeReserved is the reserved packet type.
	PacketTypeReserved PacketType = iota

	// PacketTypeConnect is the CONNECT packet type.
	PacketTypeConnect

	// PacketTypeConnAck is the CONNACK packet type.
	PacketTypeConnAck

	// PacketTypePublish is the PUBLISH packet type.
	PacketTypePublish

	// PacketTypePubAck is the PUBACK packet type.
	PacketTypePubAck

	// PacketTypePubRec is the PUBREC packet type.
	PacketTypePubRec

	// PacketTypePubRel is the PUBREL packet type.
	PacketTypePubRel

	// PacketTypePubComp is the PUBCOMP packet type.
	PacketTypePubComp

	// PacketTypeSubscribe is the SUBSCRIBE packet type.
	PacketTypeSubscribe

	// PacketTypeSubAck is the SUBACK packet type.
	PacketTypeSubAck

	// PacketTypeUnsubscribe is the UNSUBSCRIBE packet type.
	PacketTypeUnsubscribe

	// PacketTypeUnsubAck is the UNSUBSCRIBE packet type.
	PacketTypeUnsubAck

	// PacketTypePingReq is the PINGREQ packet type.
	PacketTypePingReq

	// PacketTypePingResp is the PINGRESP packet type.
	PacketTypePingResp

	// PacketTypeDisconnect is the DISCONNECT packet type.
	PacketTypeDisconnect

	// PacketTypeAuth is the AUTH packet type.
	PacketTypeAuth

	// PacketTypeInvalid indicates that the packet type is invalid.
	PacketTypeInvalid
)

var packetTypeToString = map[PacketType]string{
	PacketTypeReserved:    "Reserved",
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
}

// PacketType represents the packet type (e.g. CONNECT, CONNACK, etc.).
type PacketType byte

// String returns the packet type name.
func (t PacketType) String() string {
	name, ok := packetTypeToString[t]
	if !ok {
		return "Invalid"
	}

	return name
}

// Byte returns the packet type in byte type.
func (t PacketType) Byte() byte {
	return byte(t)
}

// Packet is the interface representing all MQTT packets.
type Packet interface {
	// Type returns the packet type.
	Type() PacketType

	// Size returns the packet size in bytes.
	Size() int

	// CreatedAt returns the timestamp of the moment which the packet was created.
	CreatedAt() time.Time
}

// PacketReader is the interface that wraps the ReadPacket method. The ReadPacket method reads the Packet, by
// deserializing it, from a bufio.Reader.
type PacketReader interface {
	ReadPacket(r *io.Reader) error
}

// PacketWriter is the interface that wraps the WritePacket method. The WritePacket method writes the Packet, by
// serializing it, into a bufio.Writer.
type PacketWriter interface {
	WritePacket(w *bufio.Writer)
}
