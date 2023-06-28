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
	"io"
)

// Packet is the interface representing all MQTT packets.
type Packet interface {
	// Type returns the packet type.
	Type() PacketType

	// Size returns the size of the packet.
	Size() int
}

// PacketDecoder is the interface for all MQTT packets which implement the Decode method.
type PacketDecoder interface {
	Packet

	// Decode decodes the Packet from buf and header. This method returns the number of bytes read
	// from buf and the error, if it fails to decode the packet correctly.
	Decode(buf []byte, header FixedHeader) (n int, err error)
}

func readPacket(r *bufio.Reader) (Packet, int, error) {
	var err error
	var hSize int
	var header FixedHeader

	if hSize, err = header.read(r); err != nil {
		return nil, hSize, fmt.Errorf("failed to read packet: %w", err)
	}

	var p PacketDecoder
	var pSize int

	switch header.PacketType {
	case PacketTypeConnect:
		p = &PacketConnect{}
	default:
		return nil, hSize, fmt.Errorf("failed to read packet: %w: %v", ErrMalformedPacketType, header.PacketType)
	}

	// Allocate the slice which will be the backing data for the packet.
	data := make([]byte, header.RemainingLength)

	if _, err = io.ReadFull(r, data); err != nil {
		return nil, hSize, fmt.Errorf("failed to read remaining bytes: %w", err)
	}

	pSize, err = p.Decode(data, header)
	n := hSize + pSize
	if err != nil {
		return nil, n, fmt.Errorf("failed to read %s packet: %w", p.Type(), err)
	}
	if pSize != header.RemainingLength {
		return nil, n, fmt.Errorf("failed to read %s packet: %w", p.Type(), ErrMalformedPacketLength)
	}

	return p, n, nil
}
