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
	s.hook.On("OnStart", s.server)
	s.addHook(s.hook)

	err := s.hooks.onStart(s.server)
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
	s.hook.On("OnServerStart", s.server)
	s.addHook(s.hook)

	err := s.hooks.onServerStart(s.server)
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
	l := newMockListener("mock", ":1883")
	s.hook.On("OnConnectionOpen", s.server, l)
	s.addHook(s.hook)

	err := s.hooks.onConnectionOpen(s.server, l)
	s.Require().NoError(err)
}

func (s *HooksTestSuite) TestOnConnectionOpenNoHookSuccess() {
	l := newMockListener("mock", ":1883")

	err := s.hooks.onConnectionOpen(s.server, l)
	s.Require().NoError(err)
}

func (s *HooksTestSuite) TestOnConnectionOpenError() {
	l := newMockListener("mock", ":1883")
	s.hook.On("OnConnectionOpen", s.server, l).Return(assert.AnError)
	s.addHook(s.hook)

	err := s.hooks.onConnectionOpen(s.server, l)
	s.Require().Error(err)
}

func (s *HooksTestSuite) TestOnConnectionOpened() {
	l := newMockListener("mock", ":1883")
	s.hook.On("OnConnectionOpened", s.server, l)
	s.addHook(s.hook)

	s.hooks.onConnectionOpened(s.server, l)
}

func (s *HooksTestSuite) TestOnConnectionClose() {
	l := newMockListener("mock", ":1883")
	s.hook.On("OnConnectionClose", s.server, l, nil)
	s.addHook(s.hook)

	s.hooks.onConnectionClose(s.server, l, nil)
}

func (s *HooksTestSuite) TestOnConnectionCloseWithError() {
	err := errors.New("failed")
	l := newMockListener("mock", ":1883")
	s.hook.On("OnConnectionClose", s.server, l, err)
	s.addHook(s.hook)

	s.hooks.onConnectionClose(s.server, l, err)
}

func (s *HooksTestSuite) TestOnConnectionClosed() {
	l := newMockListener("mock", ":1883")
	s.hook.On("OnConnectionClosed", s.server, l, nil)
	s.addHook(s.hook)

	s.hooks.onConnectionClosed(s.server, l, nil)
}

func (s *HooksTestSuite) TestOnConnectionClosedWithError() {
	err := errors.New("failed")
	l := newMockListener("mock", ":1883")
	s.hook.On("OnConnectionClosed", s.server, l, err)
	s.addHook(s.hook)

	s.hooks.onConnectionClosed(s.server, l, err)
}

func (s *HooksTestSuite) TestOnPacketReceiveSuccess() {
	l := newMockListener("mock", ":1883")
	c := newClient(nil, s.server, l)
	s.hook.On("OnPacketReceive", s.server, c)
	s.addHook(s.hook)

	err := s.hooks.onPacketReceive(s.server, c)
	s.Require().NoError(err)
}

func (s *HooksTestSuite) TestOnPacketReceiveWithError() {
	l := newMockListener("mock", ":1883")
	c := newClient(nil, s.server, l)
	s.hook.On("OnPacketReceive", s.server, c).Return(assert.AnError)
	s.addHook(s.hook)

	err := s.hooks.onPacketReceive(s.server, c)
	s.Require().Error(err)
}

func (s *HooksTestSuite) TestOnPacketReceiveError() {
	l := newMockListener("mock", ":1883")
	c := newClient(nil, s.server, l)
	s.hook.On("OnPacketReceiveError", s.server, c, assert.AnError)
	s.addHook(s.hook)

	err := s.hooks.onPacketReceiveError(s.server, c, assert.AnError)
	s.Require().NoError(err)
}

func (s *HooksTestSuite) TestOnPacketReceiveErrorWithError() {
	l := newMockListener("mock", ":1883")
	c := newClient(nil, s.server, l)
	err := errors.New("new error")
	s.hook.On("OnPacketReceiveError", s.server, c, assert.AnError).Return(err)
	s.addHook(s.hook)

	newErr := s.hooks.onPacketReceiveError(s.server, c, assert.AnError)
	s.Require().ErrorIs(newErr, err)
}

func (s *HooksTestSuite) TestOnPacketReceivedSuccess() {
	l := newMockListener("mock", ":1883")
	c := newClient(nil, s.server, l)
	p := &PacketConnect{}
	s.hook.On("OnPacketReceived", s.server, c, p)
	s.addHook(s.hook)

	err := s.hooks.onPacketReceived(s.server, c, p)
	s.Require().NoError(err)
}

func (s *HooksTestSuite) TestOnPacketReceivedError() {
	l := newMockListener("mock", ":1883")
	c := newClient(nil, s.server, l)
	p := &PacketConnect{}
	s.hook.On("OnPacketReceived", s.server, c, p).Return(assert.AnError)
	s.addHook(s.hook)

	err := s.hooks.onPacketReceived(s.server, c, p)
	s.Require().ErrorIs(err, assert.AnError)
}

func TestHooksTestSuite(t *testing.T) {
	suite.Run(t, new(HooksTestSuite))
}

func BenchmarkHooks(b *testing.B) {
	b.Run("onConnectionOpen", func(b *testing.B) {
		b.Run("No Hook", func(b *testing.B) {
			h := newHooks()

			for i := 0; i < b.N; i++ {
				_ = h.onConnectionOpen(nil, nil)
			}
		})

		b.Run("With Hook", func(b *testing.B) {
			h := newHooks()
			_ = h.add(newHookSpy())

			for i := 0; i < b.N; i++ {
				_ = h.onConnectionOpen(nil, nil)
			}
		})
	})

	b.Run("onConnectionOpened", func(b *testing.B) {
		b.Run("No Hook", func(b *testing.B) {
			h := newHooks()

			for i := 0; i < b.N; i++ {
				h.onConnectionOpened(nil, nil)
			}
		})

		b.Run("With Hook", func(b *testing.B) {
			h := newHooks()
			_ = h.add(newHookSpy())

			for i := 0; i < b.N; i++ {
				h.onConnectionOpened(nil, nil)
			}
		})
	})

	b.Run("onConnectionClose", func(b *testing.B) {
		b.Run("No Hook", func(b *testing.B) {
			h := newHooks()

			for i := 0; i < b.N; i++ {
				h.onConnectionClose(nil, nil, nil)
			}
		})

		b.Run("With Hook", func(b *testing.B) {
			h := newHooks()
			_ = h.add(newHookSpy())

			for i := 0; i < b.N; i++ {
				h.onConnectionClose(nil, nil, nil)
			}
		})
	})

	b.Run("onConnectionClosed", func(b *testing.B) {
		b.Run("No Hook", func(b *testing.B) {
			h := newHooks()

			for i := 0; i < b.N; i++ {
				h.onConnectionClosed(nil, nil, nil)
			}
		})

		b.Run("With Hook", func(b *testing.B) {
			h := newHooks()
			_ = h.add(newHookSpy())

			for i := 0; i < b.N; i++ {
				h.onConnectionClosed(nil, nil, nil)
			}
		})
	})

	b.Run("onPacketReceive", func(b *testing.B) {
		b.Run("No Hook", func(b *testing.B) {
			h := newHooks()

			for i := 0; i < b.N; i++ {
				_ = h.onPacketReceive(nil, nil)
			}
		})

		b.Run("With Hook", func(b *testing.B) {
			h := newHooks()
			_ = h.add(newHookSpy())

			for i := 0; i < b.N; i++ {
				_ = h.onPacketReceive(nil, nil)
			}
		})
	})

	b.Run("onPacketReceived", func(b *testing.B) {
		b.Run("No Hook", func(b *testing.B) {
			h := newHooks()

			for i := 0; i < b.N; i++ {
				_ = h.onPacketReceived(nil, nil, nil)
			}
		})

		b.Run("With Hook", func(b *testing.B) {
			h := newHooks()
			_ = h.add(newHookSpy())

			for i := 0; i < b.N; i++ {
				_ = h.onPacketReceived(nil, nil, nil)
			}
		})
	})
}
