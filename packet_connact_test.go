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
	"testing"

	"github.com/stretchr/testify/suite"
)

type PacketConnAckTestSuite struct {
	suite.Suite
}

func (s *PacketConnAckTestSuite) TestType() {
	var p PacketConnAck
	s.Require().Equal(PacketTypeConnAck, p.Type())
}

func (s *PacketConnAckTestSuite) TestSize() {
	testCases := []struct {
		name   string
		packet PacketConnAck
		size   int
	}{
		{
			name:   "V3.1",
			packet: PacketConnAck{Version: MQTT31},
			size:   4,
		},
		{
			name:   "V3.1.1",
			packet: PacketConnAck{Version: MQTT311, Code: ReasonCodeIdentifierRejected, SessionPresent: true},
			size:   4,
		},
		{
			name:   "V3.1.1",
			packet: PacketConnAck{Version: MQTT311, Code: ReasonCodeIdentifierRejected, SessionPresent: true},
			size:   4,
		},
		{
			name:   "V5.0, no properties",
			packet: PacketConnAck{Version: MQTT50},
			size:   5,
		},
		{
			name: "V5.0, empty properties",
			packet: PacketConnAck{
				Version:    MQTT50,
				Properties: &PropertiesConnAck{},
			},
			size: 5,
		},
		{
			name: "V5.0, with properties",
			packet: PacketConnAck{
				Version: MQTT50,
				Properties: &PropertiesConnAck{
					Flags:                 propertyFlags(0).set(PropertySessionExpiryInterval),
					SessionExpiryInterval: 30,
				},
			},
			size: 10,
		},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			size := test.packet.Size()
			s.Assert().Equal(test.size, size)
		})
	}
}

func (s *PacketConnAckTestSuite) TestEncodeSuccess() {
	testCases := []struct {
		name   string
		packet PacketConnAck
		data   []byte
	}{
		{
			name:   "V3.1",
			packet: PacketConnAck{Version: MQTT31, SessionPresent: true, Code: ReasonCodeConnectionAccepted},
			data:   []byte{0x20, 2, 1, 0},
		},
		{
			name:   "V3.1.1",
			packet: PacketConnAck{Version: MQTT311, Code: ReasonCodeIdentifierRejected},
			data:   []byte{0x20, 2, 0, 2},
		},
		{
			name:   "V5.0, no properties",
			packet: PacketConnAck{Version: MQTT50, Code: ReasonCodeSuccess},
			data:   []byte{0x20, 3, 0, 0, 0},
		},
		{
			name: "V5.0, with properties",
			packet: PacketConnAck{
				Version: MQTT50,
				Code:    ReasonCodeMalformedPacket,
				Properties: &PropertiesConnAck{
					Flags:                 propertyFlags(0).set(PropertySessionExpiryInterval),
					SessionExpiryInterval: 10,
				},
			},
			data: []byte{0x20, 8, 0, 0x81, 5, 0x11, 0, 0, 0, 10},
		},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			data := make([]byte, test.packet.Size())

			n, err := test.packet.Encode(data)
			s.Require().NoError(err)
			s.Assert().Equal(len(test.data), n)
			s.Assert().Equal(test.data, data)
		})
	}
}

func (s *PacketConnAckTestSuite) TestEncodeError() {
	packet := PacketConnAck{Version: MQTT50, Code: ReasonCodeSuccess}

	n, err := packet.Encode(nil)
	s.Require().Error(err)
	s.Assert().Zero(n)
}

func TestPacketConnAckTestSuite(t *testing.T) {
	suite.Run(t, new(PacketConnAckTestSuite))
}

func BenchmarkPacketConnAckEncode(b *testing.B) {
	testCases := []struct {
		name   string
		packet PacketConnAck
	}{
		{
			name:   "V3",
			packet: PacketConnAck{Version: MQTT311, Code: ReasonCodeConnectionAccepted},
		},
		{
			name:   "V5",
			packet: PacketConnAck{Version: MQTT50, Code: ReasonCodeSuccess},
		},
	}

	for _, test := range testCases {
		b.Run(test.name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				data := make([]byte, test.packet.Size())

				_, err := test.packet.Encode(data)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
