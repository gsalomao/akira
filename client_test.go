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
	s.Require().Equal(ClientStatePending, s.client.State())
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
	s.Require().Len(s.clients.pending, 1)
	s.Assert().Equal(s.client, s.clients.pending[0])
}

func (s *ClientsTestSuite) TestRemove() {
	client1 := newClient(s.conn2, s.server, nil)
	client2 := newClient(s.conn2, s.server, nil)
	client3 := newClient(s.conn2, s.server, nil)
	s.clients.add(client1)
	s.clients.add(client2)
	s.clients.add(client3)

	s.clients.remove(client2)
	s.Require().Len(s.clients.pending, 2)

	s.clients.remove(client1)
	s.Require().Len(s.clients.pending, 1)

	s.clients.remove(client3)
	s.Require().Empty(s.clients.pending)
}

func (s *ClientsTestSuite) TestRemoveUnknownClient() {
	s.clients.add(s.client)
	client := newClient(s.conn2, s.server, nil)

	s.clients.remove(client)
	s.Require().Len(s.clients.pending, 1)

	s.clients.remove(s.client)
	s.Require().Empty(s.clients.pending)
}

func (s *ClientsTestSuite) TestCloseAll() {
	cls := []*Client{
		newClient(s.conn2, s.server, nil),
		newClient(s.conn2, s.server, nil),
		newClient(s.conn2, s.server, nil),
	}
	for i := range cls {
		s.clients.add(cls[i])
	}

	s.clients.closeAll()
	s.Require().Empty(s.clients.pending)

	for i := range cls {
		s.Assert().Equal(ClientStateClosed, cls[i].State())
	}
}

func TestClientsTestSuite(t *testing.T) {
	suite.Run(t, new(ClientsTestSuite))
}
