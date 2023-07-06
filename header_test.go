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
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/suite"
)

type FixedHeaderTestSuite struct {
	suite.Suite
	header FixedHeader
}

func (s *FixedHeaderTestSuite) SetupTest() {
	s.header = FixedHeader{}
}

func (s *FixedHeaderTestSuite) TestReadPacketType() {
	testCases := []struct {
		data       []byte
		packetType PacketType
	}{
		{[]byte{0x10, 0x00}, PacketTypeConnect},
		{[]byte{0x20, 0x00}, PacketTypeConnAck},
		{[]byte{0x30, 0x00}, PacketTypePublish},
		{[]byte{0x40, 0x00}, PacketTypePubAck},
		{[]byte{0x50, 0x00}, PacketTypePubRec},
		{[]byte{0x60, 0x00}, PacketTypePubRel},
		{[]byte{0x70, 0x00}, PacketTypePubComp},
		{[]byte{0x80, 0x00}, PacketTypeSubscribe},
		{[]byte{0x90, 0x00}, PacketTypeSubAck},
		{[]byte{0xa0, 0x00}, PacketTypeUnsubscribe},
		{[]byte{0xb0, 0x00}, PacketTypeUnsubAck},
		{[]byte{0xc0, 0x00}, PacketTypePingReq},
		{[]byte{0xd0, 0x00}, PacketTypePingResp},
		{[]byte{0xe0, 0x00}, PacketTypeDisconnect},
		{[]byte{0xf0, 0x00}, PacketTypeAuth},
	}

	for _, test := range testCases {
		s.Run(test.packetType.String(), func() {
			var header FixedHeader
			reader := bufio.NewReader(bytes.NewReader(test.data))

			n, err := header.read(reader)
			s.Require().NoError(err)
			s.Assert().Equal(len(test.data), n)
			s.Assert().Equal(test.packetType, header.PacketType)
		})
	}
}

func (s *FixedHeaderTestSuite) TestReadFlags() {
	for i := 0; i < 0x0F; i++ {
		s.Run(fmt.Sprint(i), func() {
			var header FixedHeader
			data := []byte{byte(i), 0x00}
			reader := bufio.NewReader(bytes.NewReader(data))

			n, err := header.read(reader)
			s.Require().NoError(err)
			s.Assert().Equal(len(data), n)
			s.Assert().Equal(byte(i), header.Flags)
		})
	}
}

func (s *FixedHeaderTestSuite) TestReadRemainingLength() {
	testCases := []struct {
		data      []byte
		remaining int
	}{
		{[]byte{0x00, 0x7f}, 127},
		{[]byte{0x00, 0xff, 0x7f}, 16383},
		{[]byte{0x00, 0xff, 0xff, 0x7f}, 2097151},
		{[]byte{0x00, 0xff, 0xff, 0xff, 0x7f}, 268435455},
	}

	for _, test := range testCases {
		s.Run(fmt.Sprint(test.remaining), func() {
			var header FixedHeader
			reader := bufio.NewReader(bytes.NewReader(test.data))

			n, err := header.read(reader)
			s.Require().NoError(err)
			s.Assert().Equal(len(test.data), n)
			s.Assert().Equal(test.remaining, header.RemainingLength)
		})
	}
}

func (s *FixedHeaderTestSuite) TestReadError() {
	testCases := []struct {
		name string
		data []byte
		err  error
	}{
		{"PacketTypeError", []byte{}, io.EOF},
		{"RemainingLengthError", []byte{0x00}, io.EOF},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			var header FixedHeader
			reader := bufio.NewReader(bytes.NewReader(test.data))

			n, err := header.read(reader)
			s.Require().ErrorIs(err, test.err)
			s.Assert().Equal(len(test.data), n)
		})
	}
}

func (s *FixedHeaderTestSuite) TestReadMalformedPacket() {
	var header FixedHeader
	data := []byte{0x10, 0xff, 0xff, 0xff, 0x80}
	reader := bufio.NewReader(bytes.NewReader(data))

	n, err := header.read(reader)

	var code Error
	s.Require().ErrorIs(err, ErrMalformedVarInteger)
	s.Require().ErrorAs(err, &code)
	s.Assert().Equal(len(data), n)
	s.Assert().Equal(ReasonCodeMalformedPacket, code.Code)
}

func (s *FixedHeaderTestSuite) TestEncodePacketType() {
	testCases := []struct {
		packetType PacketType
		data       []byte
	}{
		{PacketTypeConnect, []byte{0x10, 0x00}},
		{PacketTypeConnAck, []byte{0x20, 0x00}},
		{PacketTypePublish, []byte{0x30, 0x00}},
		{PacketTypePubAck, []byte{0x40, 0x00}},
		{PacketTypePubRec, []byte{0x50, 0x00}},
		{PacketTypePubRel, []byte{0x60, 0x00}},
		{PacketTypePubComp, []byte{0x70, 0x00}},
		{PacketTypeSubscribe, []byte{0x80, 0x00}},
		{PacketTypeSubAck, []byte{0x90, 0x00}},
		{PacketTypeUnsubscribe, []byte{0xa0, 0x00}},
		{PacketTypeUnsubAck, []byte{0xb0, 0x00}},
		{PacketTypePingReq, []byte{0xc0, 0x00}},
		{PacketTypePingResp, []byte{0xd0, 0x00}},
		{PacketTypeDisconnect, []byte{0xe0, 0x00}},
		{PacketTypeAuth, []byte{0xf0, 0x00}},
	}

	for _, test := range testCases {
		s.Run(test.packetType.String(), func() {
			header := FixedHeader{PacketType: test.packetType}
			data := make([]byte, 2)

			n := header.encode(data)
			s.Assert().Equal(len(test.data), n)
			s.Assert().Equal(test.data, data)
		})
	}
}

func (s *FixedHeaderTestSuite) TestEncodeFlags() {
	for i := 0; i < 0x0F; i++ {
		s.Run(fmt.Sprint(i), func() {
			header := FixedHeader{Flags: byte(i)}
			data := make([]byte, 2)

			n := header.encode(data)
			s.Assert().Equal(len(data), n)
			s.Assert().Equal(data[0], header.Flags)
		})
	}
}

func (s *FixedHeaderTestSuite) TestEncodeRemainingLength() {
	testCases := []struct {
		remaining int
		data      []byte
	}{
		{127, []byte{0x00, 0x7f}},
		{16383, []byte{0x00, 0xff, 0x7f}},
		{2097151, []byte{0x00, 0xff, 0xff, 0x7f}},
		{268435455, []byte{0x00, 0xff, 0xff, 0xff, 0x7f}},
	}

	for _, test := range testCases {
		s.Run(fmt.Sprint(test.remaining), func() {
			header := FixedHeader{RemainingLength: test.remaining}
			data := make([]byte, len(test.data))

			n := header.encode(data)
			s.Assert().Equal(len(test.data), n)
			s.Assert().Equal(test.data, data)
		})
	}
}

func TestFixedHeaderTestSuite(t *testing.T) {
	suite.Run(t, new(FixedHeaderTestSuite))
}
