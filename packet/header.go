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
	"fmt"
)

// FixedHeader contains the values of the fixed header of the MQTT Packet.
type FixedHeader struct {
	// RemainingLength indicates the number of remaining bytes in the payload.
	RemainingLength int `json:"remaining_length"`

	// PacketType indicates the type of packet (e.g. CONNECT, PUBLISH, SUBSCRIBE, etc).
	PacketType Type `json:"packet_type"`

	// Flags contains the flags specific to each packet type.
	Flags byte `json:"flags"`
}

// Read reads the fixed header from r and returns the number of bytes read or error.
func (h *FixedHeader) Read(r *bufio.Reader) (n int, err error) {
	var b byte

	// Read the control packet type and the flags.
	b, err = r.ReadByte()
	if err != nil {
		return 0, err
	}
	n++

	// Get the packet type from the 4 most significant bits (MSB).
	h.PacketType = Type(b >> packetTypeBitsShift)

	// Get the flag from the 4 least significant bits (LSB).
	h.Flags = b & packetFlagsBitsMask

	var rLenSize int

	// Read the remaining length.
	rLenSize, err = readVarInteger(r, &h.RemainingLength)
	if err != nil {
		return n + rLenSize, fmt.Errorf("remaining length: %w (%v)", err, h.PacketType)
	}
	n += rLenSize

	return n, nil
}

// Size returns the size of the fixed header.
func (h *FixedHeader) Size() int {
	// The size of the remaining length + 1 byte for the control byte.
	return sizeVarInteger(h.RemainingLength) + 1
}

func (h *FixedHeader) encode(buf []byte) int {
	var n int

	// Store the packet type in the 4 most significant bits (MSB) and the flags in the 4 least significant bits (LSB).
	buf[n] = byte(h.PacketType<<packetTypeBitsShift) | (h.Flags & packetFlagsBitsMask)
	n++

	size := encodeVarInteger(buf[n:], h.RemainingLength)
	n += size
	return n
}
