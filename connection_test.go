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
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/gsalomao/akira/packet"
	"github.com/gsalomao/akira/testdata"
)

func newConnection(tb testing.TB) (net.Conn, *Connection) {
	tb.Helper()
	cConn, sConn := net.Pipe()
	err := cConn.SetDeadline(time.Now().Add(100 * time.Millisecond))
	if err != nil {
		tb.Fatalf("Unexpected error\n%v", err)
	}

	conn := NewConnection(&mockListener{}, sConn)
	return cConn, conn
}

func TestNewConnection(t *testing.T) {
	l := &mockListener{}
	cConn, sConn := net.Pipe()
	defer func() { _ = cConn.Close() }()

	c := NewConnection(l, sConn)
	if c == nil {
		t.Fatal("An connection was expected")
	}
	if c.Listener != l {
		t.Errorf("Unexpected listener\nwant: %p\ngot:  %p", l, c.Listener)
	}
	if c.netConn != sConn {
		t.Errorf("Unexpected net conn\nwant: %p\ngot:  %p", sConn, c.netConn)
	}
}

func TestConnectionReadPacket(t *testing.T) {
	testCases := []struct {
		path   string
		name   string
		packet Packet
	}{
		{"connect.json", "V3.1", &packet.Connect{Version: packet.MQTT31, ClientID: []byte("a")}},
		{"connect.json", "V3.1.1", &packet.Connect{Version: packet.MQTT311, ClientID: []byte("a")}},
		{"connect.json", "V5.0", &packet.Connect{Version: packet.MQTT50, ClientID: []byte("a")}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s-%s", tc.path, tc.name), func(t *testing.T) {
			fixture, err := testdata.ReadPacketFixture(tc.path, tc.name)
			if err != nil {
				t.Fatalf("Unexpected error\n%v", err)
			}

			nc, conn := newConnection(t)

			var (
				wg       sync.WaitGroup
				writeErr error
			)

			wg.Add(1)
			go func() {
				defer wg.Done()
				_, writeErr = nc.Write(fixture.Packet)
				_ = nc.Close()
			}()

			r := bufio.NewReader(nil)
			var h packet.FixedHeader

			hSize, readErr := conn.readFixedHeader(r, &h)
			if readErr != nil {
				t.Fatalf("Unexpected read error\n%v", readErr)
			}
			if hSize != 2 {
				t.Errorf("Unexpected number of bytes read\nwant: %v\ngot:  %v", 2, hSize)
			}

			p, n, readErr := conn.readPacket(r, h)
			if readErr != nil {
				t.Fatalf("Unexpected read error\n%v", readErr)
			}
			if (hSize + n) != len(fixture.Packet) {
				t.Errorf("Unexpected number of bytes read\nwant: %v\ngot:  %v", len(fixture.Packet), n)
			}
			if !reflect.DeepEqual(tc.packet, p) {
				t.Errorf("Unexpected packet\nwant: %+v\ngot:  %+v", tc.packet, p)
			}

			wg.Wait()
			if writeErr != nil {
				t.Fatalf("Unexpected write error\n%v", writeErr)
			}
		})
	}
}

func TestConnectionReadFixedHeaderOnNilConnection(t *testing.T) {
	var conn *Connection

	n, err := conn.readFixedHeader(bufio.NewReader(nil), &packet.FixedHeader{})
	if !errors.Is(err, io.EOF) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
	}
	if n != 0 {
		t.Errorf("Unexpected number of bytes read\nwant: %v\ngot:  %v", 0, n)
	}
}

func TestConnectionReadPacketOnNilConnection(t *testing.T) {
	var conn *Connection

	_, n, err := conn.readPacket(bufio.NewReader(nil), packet.FixedHeader{})
	if !errors.Is(err, io.EOF) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
	}
	if n != 0 {
		t.Errorf("Unexpected number of bytes read\nwant: %v\ngot:  %v", 0, n)
	}
}

func TestConnectionReadPacketError(t *testing.T) {
	testCases := []struct {
		name string
		data []byte
		err  error
	}{
		{name: "Invalid packet type", data: []byte{0, 0}, err: packet.ErrProtocolError},
		{name: "Invalid packet", data: []byte{16, 0}, err: packet.ErrMalformedPacket},
		{name: "Missing remaining length", data: []byte{16, 10}, err: io.EOF},
		{
			name: "Unexpected packet length",
			data: []byte{16, 17, 0, 4, 'M', 'Q', 'T', 'T', 4, 2, 0, 255, 0, 2, 'a', 'b', 0, 0, 0},
			err:  packet.ErrMalformedPacket,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nc, conn := newConnection(t)

			var (
				wg       sync.WaitGroup
				writeErr error
			)

			wg.Add(1)
			go func() {
				defer wg.Done()
				_, writeErr = nc.Write(tc.data)
				_ = nc.Close()
			}()

			r := bufio.NewReader(nil)
			var h packet.FixedHeader

			hSize, readErr := conn.readFixedHeader(r, &h)
			if readErr != nil {
				t.Fatalf("Unexpected read error\n%v", readErr)
			}
			if hSize != 2 {
				t.Errorf("Unexpected number of bytes read\nwant: %v\ngot:  %v", 2, hSize)
			}

			_, n, readErr := conn.readPacket(r, h)
			if !errors.Is(readErr, tc.err) {
				t.Errorf("Unexpected read error\nwant: %v\ngot:  %v", tc.err, readErr)
			}
			if (hSize + n) != len(tc.data) {
				t.Errorf("Unexpected number of bytes read\nwant: %v\ngot:  %v", len(tc.data), hSize)
			}

			wg.Wait()
			if writeErr != nil {
				t.Fatalf("Unexpected write error\n%v", writeErr)
			}
		})
	}
}

func TestConnectionWritePacket(t *testing.T) {
	fixture, err := testdata.ReadPacketFixture("connack.json", "V5.0 Success")
	if err != nil {
		t.Fatalf("Unexpected error\n%v", err)
	}

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	var (
		wg      sync.WaitGroup
		readErr error
		p       = &packet.ConnAck{Version: packet.MQTT50}
		buf     = make([]byte, len(fixture.Packet))
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, readErr = nc.Read(buf)
	}()

	n, writeErr := conn.writePacket(p)
	if writeErr != nil {
		t.Fatalf("Unexpected send error\n%v", writeErr)
	}
	if n != len(fixture.Packet) {
		t.Errorf("Unexpected number of bytes sent\nwant: %v\ngot:  %v", len(fixture.Packet), n)
	}

	wg.Wait()
	if readErr != nil {
		t.Fatalf("Unexpected read error\n%v", readErr)
	}
	if !bytes.Equal(fixture.Packet, buf) {
		t.Errorf("Unexpected packet\nwant: %v\ngot:  %v", fixture.Packet, buf)
	}
}

func TestConnectionWritePacketOnNilConnection(t *testing.T) {
	var conn *Connection

	n, err := conn.writePacket(nil)
	if !errors.Is(err, io.EOF) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
	}
	if n != 0 {
		t.Errorf("Unexpected number of bytes written\nwant: %v\ngot:  %v", 0, n)
	}
}

func TestConnectionWritePacketError(t *testing.T) {
	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	p := &packet.ConnAck{}
	n, err := conn.writePacket(p)
	if !errors.Is(err, packet.ErrMalformedPacket) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", packet.ErrMalformedPacket, err)
	}
	if n != 0 {
		t.Errorf("Unexpected number of bytes written\nwant: %v\ngot:  %v", 0, n)
	}
}

func TestConnectWritePacketTimeout(t *testing.T) {
	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()
	conn.sendTimeoutMs = 10

	p := &packet.ConnAck{Version: packet.MQTT50}
	_, err := conn.writePacket(p)
	if !errors.Is(err, os.ErrDeadlineExceeded) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", os.ErrDeadlineExceeded, err)
	}
}
