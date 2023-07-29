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
	"fmt"
	"io"

	"github.com/gsalomao/akira/packet"
)

// Packet is the interface representing all MQTT packets.
type Packet interface {
	// Type returns the packet type.
	Type() packet.Type

	// Size returns the size of the packet.
	Size() int
}

// PacketDecodable is the interface for all MQTT packets which can decode itself by implementing the Decode method.
type PacketDecodable interface {
	Packet

	// Decode decodes itself from buf and header. This method returns the number of bytes read from buf and the error,
	// if it fails to decode the packet correctly.
	Decode(buf []byte, header packet.FixedHeader) (n int, err error)
}

// PacketEncodable is the interface for all MQTT packets which can encode itself by implementing the Encode method.
type PacketEncodable interface {
	Packet

	// Encode encodes itself into buf and returns the number of bytes encoded. The buffer must have the length greater
	// than or equals to the packet size, otherwise this method returns an error.
	Encode(buf []byte) (n int, err error)
}

func readPacket(r *bufio.Reader) (p Packet, n int, err error) {
	var (
		header packet.FixedHeader
		pd     PacketDecodable
	)

	if n, err = header.Read(r); err != nil {
		return nil, n, err
	}

	switch header.PacketType {
	case packet.TypeConnect:
		pd = &packet.Connect{}
	default:
		return nil, n, fmt.Errorf("%w: %v: unsupported packet", packet.ErrProtocolError, header.PacketType)
	}

	// Read the remaining bytes.
	buf := make([]byte, header.RemainingLength)
	if _, err = io.ReadFull(r, buf); err != nil {
		return nil, n, err
	}

	p = pd
	n += header.RemainingLength

	var dSize int

	dSize, err = pd.Decode(buf, header)
	if err != nil {
		return nil, n, fmt.Errorf("decode error: %s: %w", p.Type(), err)
	}
	if dSize != header.RemainingLength {
		return nil, n, fmt.Errorf("%w: %s: packet size mismatch", packet.ErrMalformedPacket, p.Type())
	}

	return p, n, nil
}

func writePacket(w io.Writer, p PacketEncodable) (n int, err error) {
	buf := make([]byte, p.Size())

	_, err = p.Encode(buf)
	if err != nil {
		return n, err
	}

	return w.Write(buf)
}
