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
	"errors"
	"testing"

	"github.com/gsalomao/akira/packet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type HooksTestSuite struct {
	suite.Suite
	server *Server
	hooks  *hooks
	hook   *hookMock
}

func (s *HooksTestSuite) SetupTest() {
	var err error

	s.server, err = NewServer(NewDefaultOptions())
	s.Require().NoError(err)

	s.hooks = newHooks()
	s.hook = newHookMock()
}

func (s *HooksTestSuite) TearDownTest() {
	s.server.Close()
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

	for ht := hookType(0); ht < numOfHookTypes; ht++ {
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
	err := s.hooks.onStart(s.server)
	s.Require().NoError(err)

	s.hook.On("OnStart", s.server)
	s.addHook(s.hook)

	err = s.hooks.onStart(s.server)
	s.Require().NoError(err)
}

func (s *HooksTestSuite) TestOnStartError() {
	s.hook.On("OnStart", s.server).Return(assert.AnError)
	s.addHook(s.hook)

	err := s.hooks.onStart(s.server)
	s.Require().Error(err)
}

func (s *HooksTestSuite) TestOnStop() {
	s.hook.On("OnStop", s.server)
	s.addHook(s.hook)

	s.hooks.onStop(s.server)
}

func (s *HooksTestSuite) TestOnServerStartSuccess() {
	err := s.hooks.onServerStart(s.server)
	s.Require().NoError(err)

	s.hook.On("OnServerStart", s.server)
	s.addHook(s.hook)

	err = s.hooks.onServerStart(s.server)
	s.Require().NoError(err)
}

func (s *HooksTestSuite) TestOnServerStartError() {
	s.hook.On("OnServerStart", s.server).Return(assert.AnError)
	s.addHook(s.hook)

	err := s.hooks.onServerStart(s.server)
	s.Require().Error(err)
}

func (s *HooksTestSuite) TestOnServerStartFailed() {
	err := errors.New("failed")
	s.hook.On("OnServerStartFailed", s.server, err)
	s.addHook(s.hook)

	s.hooks.onServerStartFailed(s.server, err)
}

func (s *HooksTestSuite) TestOnServerStarted() {
	s.hook.On("OnServerStarted", s.server)
	s.addHook(s.hook)

	s.hooks.onServerStarted(s.server)
}

func (s *HooksTestSuite) TestOnServerStop() {
	s.hook.On("OnServerStop", s.server)
	s.addHook(s.hook)

	s.hooks.onServerStop(s.server)
}

func (s *HooksTestSuite) TestOnServerStopped() {
	s.hook.On("OnServerStopped", s.server)
	s.addHook(s.hook)

	s.hooks.onServerStopped(s.server)
}

func (s *HooksTestSuite) TestOnConnectionOpenSuccess() {
	err := s.hooks.onConnectionOpen(s.server, nil)
	s.Require().NoError(err)

	s.hook.On("OnConnectionOpen", s.server, nil)
	s.addHook(s.hook)

	err = s.hooks.onConnectionOpen(s.server, nil)
	s.Require().NoError(err)
}

func (s *HooksTestSuite) TestOnConnectionOpenError() {
	s.hook.On("OnConnectionOpen", s.server, nil).Return(assert.AnError)
	s.addHook(s.hook)

	err := s.hooks.onConnectionOpen(s.server, nil)
	s.Require().Error(err)
}

func (s *HooksTestSuite) TestOnConnectionOpened() {
	s.hook.On("OnConnectionOpened", s.server, nil)
	s.addHook(s.hook)

	s.hooks.onConnectionOpened(s.server, nil)
}

func (s *HooksTestSuite) TestOnConnectionClose() {
	s.hook.On("OnConnectionClose", s.server, nil, nil)
	s.addHook(s.hook)

	s.hooks.onConnectionClose(s.server, nil, nil)
}

func (s *HooksTestSuite) TestOnConnectionCloseWithError() {
	err := errors.New("failed")
	s.hook.On("OnConnectionClose", s.server, nil, err)
	s.addHook(s.hook)

	s.hooks.onConnectionClose(s.server, nil, err)
}

func (s *HooksTestSuite) TestOnConnectionClosed() {
	s.hook.On("OnConnectionClosed", s.server, nil, nil)
	s.addHook(s.hook)

	s.hooks.onConnectionClosed(s.server, nil, nil)
}

func (s *HooksTestSuite) TestOnConnectionClosedWithError() {
	err := errors.New("failed")
	s.hook.On("OnConnectionClosed", s.server, nil, err)
	s.addHook(s.hook)

	s.hooks.onConnectionClosed(s.server, nil, err)
}

func (s *HooksTestSuite) TestOnPacketReceiveSuccess() {
	c := newClient(nil, s.server, nil)

	err := s.hooks.onPacketReceive(c)
	s.Require().NoError(err)

	s.hook.On("OnPacketReceive", c)
	s.addHook(s.hook)

	err = s.hooks.onPacketReceive(c)
	s.Require().NoError(err)
}

func (s *HooksTestSuite) TestOnPacketReceiveWithError() {
	c := newClient(nil, s.server, nil)

	s.hook.On("OnPacketReceive", c).Return(assert.AnError)
	s.addHook(s.hook)

	err := s.hooks.onPacketReceive(c)
	s.Require().Error(err)
}

func (s *HooksTestSuite) TestOnPacketReceiveError() {
	c := newClient(nil, s.server, nil)

	err := s.hooks.onPacketReceiveError(c, assert.AnError)
	s.Require().NoError(err)

	s.hook.On("OnPacketReceiveError", c, assert.AnError)
	s.addHook(s.hook)

	err = s.hooks.onPacketReceiveError(c, assert.AnError)
	s.Require().NoError(err)
}

func (s *HooksTestSuite) TestOnPacketReceiveErrorWithError() {
	c := newClient(nil, s.server, nil)
	err := errors.New("new error")
	s.hook.On("OnPacketReceiveError", c, assert.AnError).Return(err)
	s.addHook(s.hook)

	newErr := s.hooks.onPacketReceiveError(c, assert.AnError)
	s.Require().ErrorIs(newErr, err)
}

func (s *HooksTestSuite) TestOnPacketReceivedSuccess() {
	c := newClient(nil, s.server, nil)
	p := &packet.Connect{}

	err := s.hooks.onPacketReceived(c, p)
	s.Require().NoError(err)

	s.hook.On("OnPacketReceived", c, p)
	s.addHook(s.hook)

	err = s.hooks.onPacketReceived(c, p)
	s.Require().NoError(err)
}

func (s *HooksTestSuite) TestOnPacketReceivedError() {
	c := newClient(nil, s.server, nil)
	p := &packet.Connect{}
	s.hook.On("OnPacketReceived", c, p).Return(assert.AnError)
	s.addHook(s.hook)

	err := s.hooks.onPacketReceived(c, p)
	s.Require().ErrorIs(err, assert.AnError)
}

func TestHooksTestSuite(t *testing.T) {
	suite.Run(t, new(HooksTestSuite))
}

func BenchmarkHooksOnConnectionOpen(b *testing.B) {
	b.Run("No Hook", func(b *testing.B) {
		h := newHooks()

		for i := 0; i < b.N; i++ {
			_ = h.onConnectionOpen(nil, nil)
		}
	})

	b.Run("With Hook", func(b *testing.B) {
		h := newHooks()
		_ = h.add(&hookSpy{})

		for i := 0; i < b.N; i++ {
			_ = h.onConnectionOpen(nil, nil)
		}
	})
}

func BenchmarkHooksOnConnectionOpened(b *testing.B) {
	b.Run("No Hook", func(b *testing.B) {
		h := newHooks()

		for i := 0; i < b.N; i++ {
			h.onConnectionOpened(nil, nil)
		}
	})

	b.Run("With Hook", func(b *testing.B) {
		h := newHooks()
		_ = h.add(&hookSpy{})

		for i := 0; i < b.N; i++ {
			h.onConnectionOpened(nil, nil)
		}
	})
}

func BenchmarkHooksOnConnectionClose(b *testing.B) {
	b.Run("No Hook", func(b *testing.B) {
		h := newHooks()

		for i := 0; i < b.N; i++ {
			h.onConnectionClose(nil, nil, nil)
		}
	})

	b.Run("With Hook", func(b *testing.B) {
		h := newHooks()
		_ = h.add(&hookSpy{})

		for i := 0; i < b.N; i++ {
			h.onConnectionClose(nil, nil, nil)
		}
	})
}

func BenchmarkHooksOnConnectionClosed(b *testing.B) {
	b.Run("No Hook", func(b *testing.B) {
		h := newHooks()

		for i := 0; i < b.N; i++ {
			h.onConnectionClosed(nil, nil, nil)
		}
	})

	b.Run("With Hook", func(b *testing.B) {
		h := newHooks()
		_ = h.add(&hookSpy{})

		for i := 0; i < b.N; i++ {
			h.onConnectionClosed(nil, nil, nil)
		}
	})
}

func BenchmarkHooksOnPacketReceive(b *testing.B) {
	b.Run("No Hook", func(b *testing.B) {
		h := newHooks()

		for i := 0; i < b.N; i++ {
			_ = h.onPacketReceive(nil)
		}
	})

	b.Run("With Hook", func(b *testing.B) {
		h := newHooks()
		_ = h.add(&hookSpy{})

		for i := 0; i < b.N; i++ {
			_ = h.onPacketReceive(nil)
		}
	})
}

func BenchmarkHooksOnPacketReceived(b *testing.B) {
	b.Run("No Hook", func(b *testing.B) {
		h := newHooks()

		for i := 0; i < b.N; i++ {
			_ = h.onPacketReceived(nil, nil)
		}
	})

	b.Run("With Hook", func(b *testing.B) {
		h := newHooks()
		_ = h.add(&hookSpy{})

		for i := 0; i < b.N; i++ {
			_ = h.onPacketReceived(nil, nil)
		}
	})
}
