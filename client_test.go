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
	"net"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ClientsTestSuite struct {
	suite.Suite
	server  *Server
	clients *clients
	client  *Client
	conn1   net.Conn
	conn2   net.Conn
}

func (s *ClientsTestSuite) SetupTest() {
	var err error

	s.server, err = NewServer(NewDefaultOptions())
	s.Require().NoError(err)

	s.clients = newClients()
	s.conn1, s.conn2 = net.Pipe()

	s.client = newClient(s.conn2, s.server, nil)
	s.Require().Equal(ClientPending, s.client.State())
}

func (s *ClientsTestSuite) TearDownTest() {
	s.server.Close()
	_ = s.conn1.Close()
	_ = s.conn2.Close()
}

func (s *ClientsTestSuite) TestNewClients() {
	c := newClients()
	s.Require().NotNil(c)
}

func (s *ClientsTestSuite) TestAddSuccess() {
	s.clients.add(s.client)
	s.Require().Equal(1, s.clients.pending.Len())
	s.Assert().Equal(s.client, s.clients.pending.Front().Value.(*Client))
}

func (s *ClientsTestSuite) TestStopAll() {
	s.clients.add(s.client)

	s.clients.stopAll()
	s.Require().Equal(1, s.clients.pending.Len())
	s.Assert().Equal(ClientStopped, s.client.State())
}

func TestClientsTestSuite(t *testing.T) {
	suite.Run(t, new(ClientsTestSuite))
}
