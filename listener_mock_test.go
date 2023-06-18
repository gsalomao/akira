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

type MockListenerTestSuite struct {
	suite.Suite
	listener *mockListener
}

func (s *MockListenerTestSuite) SetupTest() {
	s.listener = newMockListener("mock", ":1883")
}

func (s *MockListenerTestSuite) TestNewMockListener() {
	l := newMockListener("mock", ":1883")
	s.Assert().Equal("mock", l.Name())
	s.Assert().Equal(":1883", l.Address())
	s.Assert().Equal("mock", l.Protocol())
}

func (s *MockListenerTestSuite) TestListen() {
	listening, err := s.listener.Listen(nil)
	s.Require().NotNil(listening)
	s.Require().NoError(err)
	s.Require().True(<-listening)
	s.Require().True(s.listener.Listening())

	close(s.listener.done)
	<-listening
}

func (s *MockListenerTestSuite) TestStop() {
	listening, _ := s.listener.Listen(nil)
	<-listening

	s.listener.Stop()
	<-listening
	s.Require().False(s.listener.Listening())
}

func (s *MockListenerTestSuite) TestListening() {
	listening, _ := s.listener.Listen(nil)
	<-listening
	s.Require().True(s.listener.Listening())

	s.listener.Stop()
	<-listening
	s.Require().False(s.listener.Listening())
}

func TestMockListenerTestSuite(t *testing.T) {
	suite.Run(t, new(MockListenerTestSuite))
}
