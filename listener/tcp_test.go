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
	"testing"

	"github.com/gsalomao/akira"
	"github.com/stretchr/testify/suite"
)

const certificate = `-----BEGIN CERTIFICATE-----
MIIBhTCCASugAwIBAgIQIRi6zePL6mKjOipn+dNuaTAKBggqhkjOPQQDAjASMRAw
DgYDVQQKEwdBY21lIENvMB4XDTE3MTAyMDE5NDMwNloXDTE4MTAyMDE5NDMwNlow
EjEQMA4GA1UEChMHQWNtZSBDbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABD0d
7VNhbWvZLWPuj/RtHFjvtJBEwOkhbN/BnnE8rnZR8+sbwnc/KhCk3FhnpHZnQz7B
5aETbbIgmuvewdjvSBSjYzBhMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggr
BgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MCkGA1UdEQQiMCCCDmxvY2FsaG9zdDo1
NDUzgg4xMjcuMC4wLjE6NTQ1MzAKBggqhkjOPQQDAgNIADBFAiEA2zpJEPQyz6/l
Wf86aX6PepsntZv2GYlA5UpabfT2EZICICpJ5h/iI+i341gBmLiAFQOyTDT+/wQc
6MF9+Yw1Yy0t
-----END CERTIFICATE-----`

const key = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIIrYSSNQFaA2Hwf1duRSxKtLYX5CB04fSeQ6tF1aY/PuoAoGCCqGSM49
AwEHoUQDQgAEPR3tU2Fta9ktY+6P9G0cWO+0kETA6SFs38GecTyudlHz6xvCdz8q
EKTcWGekdmdDPsHloRNtsiCa697B2O9IFA==
-----END EC PRIVATE KEY-----`

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
	cert, err := tls.X509KeyPair([]byte(certificate), []byte(key))
	s.Require().NoError(err)

	tlsConfig := tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{cert},
	}
	l := NewTCPListener("tcp1", ":1883", &tlsConfig)
	defer func() { _ = l.Close() }()

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
