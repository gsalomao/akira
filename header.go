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
	"fmt"
)

// FixedHeader contains the values of the fixed header of the MQTT Packet.
type FixedHeader struct {
	// RemainingLength indicates the number of remaining bytes in the payload.
	RemainingLength int `json:"remaining_length"`

	// PacketType indicates the type of packet (e.g. CONNECT, PUBLISH, SUBSCRIBE, etc).
	PacketType PacketType `json:"packet_type"`

	// Flags contains the flags specific to each packet type.
	Flags byte `json:"flags"`
}

func (h *FixedHeader) size() int {
	// The size of the remaining length + 1 byte for the control byte.
	return sizeVarInteger(h.RemainingLength) + 1
}

func (h *FixedHeader) read(r *bufio.Reader) (n int, err error) {
	var b byte

	// Read the control packet type and the flags
	b, err = r.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("error control byte: %w", err)
	}
	n++

	// Get the packet type from the 4 most significant bits (MSB).
	h.PacketType = PacketType(b >> packetTypeBitsShift)

	// Get the flag from the 4 least significant bits (LSB).
	h.Flags = b & packetFlagsBitsMask

	var rLenSize int

	// Read the remaining length
	rLenSize, err = readVarInteger(r, &h.RemainingLength)
	if err != nil {
		return n + rLenSize, fmt.Errorf("error remaining length: %w", err)
	}
	n += rLenSize

	return n, nil
}