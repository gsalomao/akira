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

type PacketTypeTestSuite struct {
	suite.Suite
}

func (s *PacketTypeTestSuite) TestString() {
	types := map[PacketType]string{
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
		PacketTypeInvalid:     "Invalid",
	}

	for t, n := range types {
		s.Run(n, func() {
			name := t.String()
			s.Require().Equal(n, name)
		})
	}
}

func (s *PacketTypeTestSuite) TestByte() {
	types := map[PacketType]byte{
		PacketTypeReserved:    0,
		PacketTypeConnect:     1,
		PacketTypeConnAck:     2,
		PacketTypePublish:     3,
		PacketTypePubAck:      4,
		PacketTypePubRec:      5,
		PacketTypePubRel:      6,
		PacketTypePubComp:     7,
		PacketTypeSubscribe:   8,
		PacketTypeSubAck:      9,
		PacketTypeUnsubscribe: 10,
		PacketTypeUnsubAck:    11,
		PacketTypePingReq:     12,
		PacketTypePingResp:    13,
		PacketTypeDisconnect:  14,
		PacketTypeAuth:        15,
	}

	for t, v := range types {
		s.Run(t.String(), func() {
			b := t.Byte()
			s.Require().Equal(v, b)
		})
	}
}

func TestPacketTypeTestSuite(t *testing.T) {
	suite.Run(t, new(PacketTypeTestSuite))
}
