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
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

type HooksTestSuite struct {
	suite.Suite
	srv   *Server
	hooks *hooks
	hook  *mockHook
}

func (s *HooksTestSuite) SetupTest() {
	s.srv = NewServer(NewDefaultOptions())
	s.hooks = newHooks()
	s.hook = newMockHook()
}

func (s *HooksTestSuite) TearDownTest() {
	s.hook.AssertExpectations(s.T())
}

func (s *HooksTestSuite) addHook(h Hook) {
	err := s.hooks.add(h)
	s.Require().NoError(err)
}

func (s *HooksTestSuite) TestNewHooks() {
	hks := newHooks()
	s.Require().NotNil(hks)
}

func (s *HooksTestSuite) TestAddSuccess() {
	err := s.hooks.add(s.hook)
	s.Require().NoError(err)

	for ht := hookType(0); ht < numHookTypes; ht++ {
		_, ok := s.hooks.hookNames[ht][s.hook.Name()]
		s.Assert().Truef(ok, "Missing hook type %v", ht)
	}
}

func (s *HooksTestSuite) TestAddError() {
	s.addHook(s.hook)

	err := s.hooks.add(s.hook)
	s.Assert().ErrorIs(err, ErrHookAlreadyExists)
}

func (s *HooksTestSuite) TestOnStartSuccess() {
	s.hook.On("OnStart")
	s.addHook(s.hook)

	err := s.hooks.onStart()
	s.Require().NoError(err)
}

func (s *HooksTestSuite) TestOnStartError() {
	s.hook.On("OnStart").Return(errors.New("failed"))
	s.addHook(s.hook)

	err := s.hooks.onStart()
	s.Require().Error(err)
}

func (s *HooksTestSuite) TestOnStop() {
	s.hook.On("OnStop")
	s.addHook(s.hook)

	s.hooks.onStop()
}

func (s *HooksTestSuite) TestOnServerStartSuccess() {
	s.hook.On("OnServerStart", s.srv)
	s.addHook(s.hook)

	err := s.hooks.onServerStart(s.srv)
	s.Require().NoError(err)
}

func (s *HooksTestSuite) TestOnServerStartError() {
	s.hook.On("OnServerStart", s.srv).Return(errors.New("failed"))
	s.addHook(s.hook)

	err := s.hooks.onServerStart(s.srv)
	s.Require().Error(err)
}

func (s *HooksTestSuite) TestOnServerStartFailed() {
	err := errors.New("failed")
	s.hook.On("OnServerStartFailed", s.srv, err)
	s.addHook(s.hook)

	s.hooks.onServerStartFailed(s.srv, err)
}

func (s *HooksTestSuite) TestOnServerStarted() {
	s.hook.On("OnServerStarted", s.srv)
	s.addHook(s.hook)

	s.hooks.onServerStarted(s.srv)
}

func (s *HooksTestSuite) TestOnServerStop() {
	s.hook.On("OnServerStop", s.srv)
	s.addHook(s.hook)

	s.hooks.onServerStop(s.srv)
}

func (s *HooksTestSuite) TestOnServerStopped() {
	s.hook.On("OnServerStopped", s.srv)
	s.addHook(s.hook)

	s.hooks.onServerStopped(s.srv)
}

func (s *HooksTestSuite) TestOnClientOpenSuccess() {
	l := newMockListener("mock")
	c := newClient(nil, NewDefaultConfig(), nil)
	s.hook.On("OnClientOpen", s.srv, l, c)
	s.addHook(s.hook)

	err := s.hooks.onClientOpen(s.srv, l, c)
	s.Require().NoError(err)
}

func (s *HooksTestSuite) TestOnClientOpenNoHookSuccess() {
	l := newMockListener("mock")
	c := newClient(nil, NewDefaultConfig(), nil)

	err := s.hooks.onClientOpen(s.srv, l, c)
	s.Require().NoError(err)
}

func (s *HooksTestSuite) TestOnClientOpenError() {
	l := newMockListener("mock")
	c := newClient(nil, NewDefaultConfig(), nil)
	s.hook.On("OnClientOpen", s.srv, l, c).Return(errors.New("failed"))
	s.addHook(s.hook)

	err := s.hooks.onClientOpen(s.srv, l, c)
	s.Require().Error(err)
}

func (s *HooksTestSuite) TestOnClientOpened() {
	l := newMockListener("mock")
	c := newClient(nil, NewDefaultConfig(), nil)
	s.hook.On("OnClientOpened", s.srv, l, c)
	s.addHook(s.hook)

	s.hooks.onClientOpened(s.srv, l, c)
}

func (s *HooksTestSuite) TestOnClientClose() {
	c := newClient(nil, NewDefaultConfig(), nil)
	s.hook.On("OnClientClose", c)
	s.addHook(s.hook)

	s.hooks.onClientClose(c)
}

func (s *HooksTestSuite) TestOnClientClosed() {
	c := newClient(nil, NewDefaultConfig(), nil)
	s.hook.On("OnClientClosed", c)
	s.addHook(s.hook)

	s.hooks.onClientClosed(c)
}

func TestHooksTestSuite(t *testing.T) {
	suite.Run(t, new(HooksTestSuite))
}
