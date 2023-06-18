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
	"net"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ClientTestSuite struct {
	suite.Suite
}

func (s *ClientTestSuite) TestNewClient() {
	srv := NewServer(NewDefaultOptions())
	defer srv.Close()
	c1, c2 := net.Pipe()
	defer func() { _ = c1.Close() }()
	defer func() { _ = c2.Close() }()

	c := newClient(c2, srv, nil)
	s.Require().NotNil(c)
	s.Assert().False(c.Closed())
}

func (s *ClientTestSuite) TestClose() {
	srv := NewServer(NewDefaultOptions())
	defer srv.Close()
	c1, c2 := net.Pipe()
	defer func() { _ = c1.Close() }()
	c := newClient(c2, srv, nil)

	c.Close()
	<-c.Done()
	s.Assert().True(c.Closed())
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
