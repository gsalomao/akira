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
	"errors"
)

// PacketConnAck represents the CONNACK Packet from MQTT specifications.
type PacketConnAck struct {
	// Properties contains the properties of the CONNACK packet.
	Properties *PropertiesConnAck `json:"properties"`

	// Version represents the MQTT version.
	Version MQTTVersion `json:"version"`

	// Code represents the reason code based on the MQTT specifications.
	Code ReasonCode `json:"code"`

	// SessionPresent indicates if there is already a session associated with the Client ID.
	SessionPresent bool `json:"session_present"`
}

// Type returns the packet type.
func (p *PacketConnAck) Type() PacketType {
	return PacketTypeConnAck
}

// Size returns the CONNACK packet size.
func (p *PacketConnAck) Size() int {
	size := p.remainingLength()
	header := FixedHeader{PacketType: PacketTypeConnAck, RemainingLength: size}

	size += header.size()
	return size
}

// Encode encodes the CONNACK Packet into buf and returns the number of bytes encoded. The buffer must have the
// length greater than or equals to the packet size, otherwise this method returns an error.
func (p *PacketConnAck) Encode(buf []byte) (n int, err error) {
	if len(buf) < p.Size() {
		return 0, errors.New("buffer too small")
	}

	header := FixedHeader{PacketType: PacketTypeConnAck, RemainingLength: p.remainingLength()}
	n = header.encode(buf)

	var flags byte
	if p.SessionPresent {
		flags = 1
	}

	buf[n] = flags
	n++

	buf[n] = byte(p.Code)
	n++

	if p.Version == MQTT50 {
		var size int

		size, err = p.Properties.encode(buf[n:])
		n += size
	}

	return n, err
}

func (p *PacketConnAck) remainingLength() int {
	// Start with 2 bytes for the ConnAck flags and Reason Code
	remainingLength := 2

	if p.Version == MQTT50 {
		n := p.Properties.size()
		n += sizeVarInteger(n)
		remainingLength += n
	}

	return remainingLength
}
