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
	"context"
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServerTestSuite struct {
	suite.Suite
	server   *Server
	listener *mockListener
	hook     *mockHook
}

func (s *ServerTestSuite) SetupTest() {
	s.server = NewServer(NewDefaultOptions())
	s.listener = newMockListener("mock", ":1883")
	s.hook = nil
	_ = s.server.AddListener(s.listener)
}

func (s *ServerTestSuite) TearDownTest() {
	if s.hook != nil {
		s.hook.AssertExpectations(s.T())
	}
}

func (s *ServerTestSuite) addHook() {
	s.hook = newMockHook()

	err := s.server.AddHook(s.hook)
	s.Require().NoError(err)
}

func (s *ServerTestSuite) startServer() {
	if s.hook != nil {
		s.hook.On("OnServerStart", s.server)
		s.hook.On("OnStart")
		s.hook.On("OnServerStarted", s.server)
	}

	err := s.server.Start(context.Background())
	s.Require().NoError(err)
}

func (s *ServerTestSuite) stopServer() {
	if s.hook != nil {
		s.hook.On("OnServerStop", s.server)
		s.hook.On("OnStop")
		s.hook.On("OnServerStopped", s.server)
	}

	err := s.server.Stop(context.Background())
	s.Require().NoError(err)
}

func (s *ServerTestSuite) TestNewServerDefaultOptions() {
	srv := NewServer(nil)

	s.Assert().Equal(NewDefaultOptions(), &srv.Options)
	s.Assert().Equal(ServerNotStarted, srv.State())
}

func (s *ServerTestSuite) TestNewServerDefaultConfig() {
	srv := NewServer(&Options{})

	s.Assert().Equal(NewDefaultConfig(), srv.Options.Config)
	s.Assert().Equal(ServerNotStarted, srv.State())
}

func (s *ServerTestSuite) TestAddListenerSuccess() {
	l := newMockListener("mock2", ":1883")

	err := s.server.AddListener(l)
	s.Require().NoError(err)

	_, ok := s.server.listeners.get(l.Name())
	s.Require().True(ok)
}

func (s *ServerTestSuite) TestAddListenerError() {
	l := newMockListener("mock", ":1883")

	err := s.server.AddListener(l)
	s.Require().ErrorIs(err, ErrListenerAlreadyExists)
}

func (s *ServerTestSuite) TestAddListenerServerRunning() {
	_ = s.server.Start(context.Background())
	l := newMockListener("mock2", ":1883")

	err := s.server.AddListener(l)
	s.Require().NoError(err)
	s.Require().True(l.Listening())
}

func (s *ServerTestSuite) TestAddListenerServerRunningError() {
	_ = s.server.Start(context.Background())
	l := newMockListener("mock2", "abc")

	err := s.server.AddListener(l)
	s.Require().Error(err)
	s.Require().False(l.Listening())
}

func (s *ServerTestSuite) TestAddHookSuccess() {
	srv := NewServer(NewDefaultOptions())
	hook := newMockHook()

	err := srv.AddHook(hook)
	s.Require().NoError(err)
	hook.AssertExpectations(s.T())
}

func (s *ServerTestSuite) TestAddHookCallsOnStartWhenServerRunning() {
	srv := NewServer(NewDefaultOptions())
	_ = srv.Start(context.Background())
	hook := newMockHook()
	hook.On("OnStart")

	err := srv.AddHook(hook)
	s.Assert().NoError(err)
	hook.AssertExpectations(s.T())
}

func (s *ServerTestSuite) TestAddHookError() {
	s.addHook()
	hook := newMockHook()

	err := s.server.AddHook(hook)
	s.Assert().Error(err)
}

func (s *ServerTestSuite) TestAddHookOnStartError() {
	s.startServer()
	hook := newMockHook()
	hook.On("OnStart").Return(errors.New("failed"))

	err := s.server.AddHook(hook)
	s.Assert().Error(err)
	hook.AssertExpectations(s.T())
}

func (s *ServerTestSuite) TestStartSuccess() {
	err := s.server.Start(context.Background())
	s.Require().NoError(err)
	s.Assert().Equal(ServerRunning, s.server.State())
	s.Assert().True(s.listener.Listening())
}

func (s *ServerTestSuite) TestStartWithoutListeners() {
	srv := NewServer(NewDefaultOptions())

	err := srv.Start(context.Background())
	s.Require().NoError(err)
	s.Assert().Equal(ServerRunning, srv.State())
}

func (s *ServerTestSuite) TestStartWithHook() {
	s.addHook()
	s.hook.On("OnServerStart", s.server)
	s.hook.On("OnStart")
	s.hook.On("OnServerStarted", s.server)

	err := s.server.Start(context.Background())
	s.Require().NoError(err)
	s.Assert().Equal(ServerRunning, s.server.State())
}

func (s *ServerTestSuite) TestStartWhenClosedError() {
	s.addHook()
	s.startServer()
	s.stopServer()
	s.server.Close()
	s.Require().Equal(ServerClosed, s.server.State())

	err := s.server.Start(context.Background())
	s.Require().Error(err, ErrInvalidServerState)
}

func (s *ServerTestSuite) TestStartOnServerStartError() {
	s.addHook()
	err := errors.New("failed")
	s.hook.On("OnServerStart", s.server).Return(err)
	s.hook.On("OnServerStartFailed", s.server, err)

	err = s.server.Start(context.Background())
	s.Require().Error(err)
	s.Assert().Equal(ServerFailed, s.server.State())
}

func (s *ServerTestSuite) TestStartOnStartError() {
	s.addHook()
	err := errors.New("failed")
	s.hook.On("OnServerStart", s.server)
	s.hook.On("OnStart").Return(err)
	s.hook.On("OnServerStartFailed", s.server, err)

	err = s.server.Start(context.Background())
	s.Require().Error(err)
	s.Assert().Equal(ServerFailed, s.server.State())
}

func (s *ServerTestSuite) TestStartListenerListenError() {
	s.addHook()
	s.hook.On("OnServerStart", s.server)
	s.hook.On("OnStart")
	s.hook.On("OnServerStartFailed", s.server, mock.Anything)
	l := newMockListener("mock2", "abc")
	_ = s.server.AddListener(l)

	err := s.server.Start(context.Background())
	s.Require().Error(err)
	s.Assert().Equal(ServerFailed, s.server.State())
}

func (s *ServerTestSuite) TestStop() {
	s.startServer()

	err := s.server.Stop(context.Background())
	s.Require().NoError(err)
	s.Assert().Equal(ServerStopped, s.server.State())
	s.Assert().False(s.listener.Listening())
}

func (s *ServerTestSuite) TestStopWithHook() {
	s.addHook()
	s.hook.On("OnServerStart", s.server)
	s.hook.On("OnStart")
	s.hook.On("OnServerStarted", s.server)
	s.hook.On("OnServerStop", s.server)
	s.hook.On("OnStop")
	s.hook.On("OnServerStopped", s.server)
	s.startServer()

	err := s.server.Stop(context.Background())
	s.Require().NoError(err)
	s.Assert().Equal(ServerStopped, s.server.State())
	s.Assert().False(s.listener.Listening())
}

func (s *ServerTestSuite) TestStopCancelled() {
	s.startServer()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := s.server.Stop(ctx)
	s.Require().Error(err)
}

func (s *ServerTestSuite) TestStopWhenNotStartedSuccess() {
	s.addHook()

	err := s.server.Stop(context.Background())
	s.Require().NoError(err)
}

func (s *ServerTestSuite) TestStopWhenStoppedSuccess() {
	s.addHook()
	s.startServer()
	s.stopServer()

	err := s.server.Stop(context.Background())
	s.Require().NoError(err)
}

func (s *ServerTestSuite) TestCloseWhenStopping() {
	s.startServer()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_ = s.server.Stop(ctx)

	s.server.Close()
	s.Require().Equal(ServerClosed, s.server.State())
}

func (s *ServerTestSuite) TestCloseWhenStopped() {
	s.startServer()
	_ = s.server.Stop(context.Background())

	s.server.Close()
	s.Require().Equal(ServerClosed, s.server.State())
}

func (s *ServerTestSuite) TestCloseWhenRunning() {
	s.addHook()
	s.startServer()
	s.hook.On("OnServerStop", s.server)
	s.hook.On("OnStop")
	s.hook.On("OnServerStopped", s.server)

	s.server.Close()
	s.Require().Equal(ServerClosed, s.server.State())
}

func (s *ServerTestSuite) TestCloseWhenNotStartedNoAction() {
	s.addHook()

	s.server.Close()
	s.Require().Equal(ServerClosed, s.server.State())
}

func (s *ServerTestSuite) TestHandleConnection() {
	s.startServer()
	c1, c2 := net.Pipe()

	s.listener.onConnection(c2)
	_ = c1.Close()
	s.stopServer()
	s.Require().Equal(ServerStopped, s.server.State())
}

func (s *ServerTestSuite) TestHandleConnectionWithHook() {
	connOpened := make(chan struct{})
	connClosed := make(chan struct{})
	s.addHook()
	s.hook.On("OnServerStart", s.server)
	s.hook.On("OnStart")
	s.hook.On("OnServerStarted", s.server)
	s.hook.On("OnConnectionOpen", s.server, s.listener)
	s.hook.On("OnConnectionOpened", s.server, s.listener).Run(func(_ mock.Arguments) {
		close(connOpened)
	})
	s.hook.On("OnConnectionClose", s.server, s.listener)
	s.hook.On("OnConnectionClosed", s.server, s.listener).Run(func(_ mock.Arguments) {
		close(connClosed)
	})
	s.hook.On("OnServerStop", s.server)
	s.hook.On("OnStop")
	s.hook.On("OnServerStopped", s.server)
	s.startServer()
	c1, c2 := net.Pipe()

	s.listener.onConnection(c2)
	<-connOpened
	_ = c1.Close()
	<-connClosed
	s.stopServer()
}

func (s *ServerTestSuite) TestOnConnectionOpenedError() {
	s.addHook()
	s.hook.On("OnServerStart", s.server)
	s.hook.On("OnStart")
	s.hook.On("OnServerStarted", s.server)
	s.hook.On("OnConnectionOpen", s.server, s.listener).Return(errors.New("failed"))
	s.hook.On("OnConnectionClose", s.server, s.listener)
	s.hook.On("OnConnectionClosed", s.server, s.listener)
	s.startServer()
	c1, c2 := net.Pipe()
	defer func() { _ = c1.Close() }()

	s.listener.onConnection(c2)
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}
