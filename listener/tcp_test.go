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

package listener

import (
	"crypto/tls"
	"net"
	"os"
	"testing"

	"github.com/gsalomao/akira"
	"github.com/stretchr/testify/suite"
)

type TCPListenerTestSuite struct {
	suite.Suite
}

func (s *TCPListenerTestSuite) TestNewTCPListenerSuccess() {
	lsn := NewTCPListener("tcp1", ":1883", nil)
	s.Require().NotNil(lsn)
	defer func() { _ = lsn.Close() }()

	s.Assert().Equal("tcp1", lsn.Name())
}

func (s *TCPListenerTestSuite) TestListenSuccess() {
	lsn := NewTCPListener("tcp1", ":1883", nil)
	defer func() { _ = lsn.Close() }()

	err := lsn.Listen(func(akira.Listener, net.Conn) {})
	s.Require().NoError(err)
}

func (s *TCPListenerTestSuite) TestListenWithTLSConfigSuccess() {
	cert, err := os.ReadFile("testdata/test.crt")
	s.Require().NoError(err)

	key, err := os.ReadFile("testdata/test.key")
	s.Require().NoError(err)

	x509, err := tls.X509KeyPair(cert, key)
	s.Require().NoError(err)

	tlsConfig := tls.Config{MinVersion: tls.VersionTLS12, Certificates: []tls.Certificate{x509}}
	lsn := NewTCPListener("tcp1", ":1883", &tlsConfig)
	defer func() { _ = lsn.Close() }()

	err = lsn.Listen(func(akira.Listener, net.Conn) {})
	s.Require().NoError(err)
}

func (s *TCPListenerTestSuite) TestListenError() {
	lsn := NewTCPListener("tcp1", ":abc", nil)
	defer func() { _ = lsn.Close() }()

	err := lsn.Listen(func(akira.Listener, net.Conn) {})
	s.Require().Error(err)
}

func (s *TCPListenerTestSuite) TestListenOnConnection() {
	lsn := NewTCPListener("tcp1", ":1883", nil)
	defer func() { _ = lsn.Close() }()

	var nc net.Conn
	doneCh := make(chan struct{})

	_ = lsn.Listen(func(lsn akira.Listener, c net.Conn) {
		s.Require().Equal(lsn, lsn)
		s.Require().NotNil(c)
		nc = c
		close(doneCh)
	})

	conn, err := net.Dial("tcp", ":1883")
	s.Require().NotNil(conn)
	s.Require().NoError(err)
	defer func() { _ = conn.Close() }()

	<-doneCh
	s.Require().NotNil(nc)
}

func (s *TCPListenerTestSuite) TestClose() {
	lsn := NewTCPListener("tcp1", ":1883", nil)
	_ = lsn.Listen(func(akira.Listener, net.Conn) {})

	err := lsn.Close()
	s.Require().NoError(err)
}

func (s *TCPListenerTestSuite) TestCloseWhenNotListening() {
	lsn := NewTCPListener("tcp1", ":1883", nil)

	err := lsn.Close()
	s.Require().NoError(err)
}

func TestTCPListenerTestSuite(t *testing.T) {
	suite.Run(t, new(TCPListenerTestSuite))
}
