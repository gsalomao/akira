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

type ListenersTestSuite struct {
	suite.Suite
	listeners *listeners
	listener  *mockListener
}

func (s *ListenersTestSuite) SetupTest() {
	s.listeners = newListeners()
	s.listener = newMockListener("mock", ":1883")
}

func (s *ListenersTestSuite) TestNewListeners() {
	l := newListeners()
	s.Require().NotNil(l)
}

func (s *ListenersTestSuite) TestAdd() {
	s.listeners.add(s.listener)

	l2, ok := s.listeners.internal[s.listener.Name()]
	s.Require().True(ok)
	s.Assert().Equal(s.listener, l2)
}

func (s *ListenersTestSuite) TestGet() {
	l, ok := s.listeners.get(s.listener.Name())
	s.Require().False(ok)
	s.Require().Nil(l)

	s.listeners.internal[s.listener.Name()] = s.listener
	l, ok = s.listeners.get(s.listener.Name())
	s.Assert().True(ok)
	s.Assert().Equal(s.listener, l)
}

func (s *ListenersTestSuite) TestDelete() {
	s.listeners.internal[s.listener.Name()] = s.listener

	s.listeners.delete(s.listener.Name())
	_, ok := s.listeners.internal[s.listener.Name()]
	s.Require().False(ok)
}

func (s *ListenersTestSuite) TestLen() {
	s.Require().Zero(s.listeners.len())

	s.listeners.internal[s.listener.Name()] = s.listener
	s.Require().Equal(1, s.listeners.len())
}

func (s *ListenersTestSuite) TestListen() {
	s.listeners.add(s.listener)

	err := s.listeners.listen(s.listener, func(_ Listener, _ net.Conn) {})
	s.Require().NoError(err)
	s.Assert().True(s.listener.Listening())
}

func (s *ListenersTestSuite) TestListenAll() {
	s.listeners.add(s.listener)
	l := newMockListener("mock2", ":1883")
	s.listeners.add(l)

	err := s.listeners.listenAll(func(_ Listener, _ net.Conn) {})
	s.Require().NoError(err)
	s.Assert().True(s.listener.Listening())
	s.Assert().True(l.Listening())
}

func (s *ListenersTestSuite) TestStopAll() {
	s.listeners.add(s.listener)
	l := newMockListener("mock2", ":1883")
	s.listeners.add(l)
	_ = s.listeners.listenAll(func(_ Listener, _ net.Conn) {})

	s.listeners.stopAll()
	s.listeners.wait()
	s.Assert().False(s.listener.Listening())
	s.Assert().False(l.Listening())
}

func TestListenersTestSuite(t *testing.T) {
	suite.Run(t, new(ListenersTestSuite))
}
