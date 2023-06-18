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
	srv  *Server
	lsn  *mockListener
	hook *mockHook
}

func (s *ServerTestSuite) SetupTest() {
	s.srv = NewServer(NewDefaultOptions())
	s.lsn = newMockListener("mock", ":1883")
	s.hook = nil
	_ = s.srv.AddListener(s.lsn)
}

func (s *ServerTestSuite) TearDownTest() {
	if s.hook != nil {
		s.hook.AssertExpectations(s.T())
	}
}

func (s *ServerTestSuite) addHook() {
	s.hook = newMockHook()

	err := s.srv.AddHook(s.hook)
	s.Require().NoError(err)
}

func (s *ServerTestSuite) startServer() {
	if s.hook != nil {
		s.hook.On("OnServerStart", s.srv)
		s.hook.On("OnStart")
		s.hook.On("OnServerStarted", s.srv)
	}

	err := s.srv.Start(context.Background())
	s.Require().NoError(err)
}

func (s *ServerTestSuite) stopServer() {
	if s.hook != nil {
		s.hook.On("OnServerStop", s.srv)
		s.hook.On("OnStop")
		s.hook.On("OnServerStopped", s.srv)
	}

	err := s.srv.Stop(context.Background())
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

	err := s.srv.AddListener(l)
	s.Require().NoError(err)

	_, ok := s.srv.listeners.get(l.Name())
	s.Require().True(ok)
}

func (s *ServerTestSuite) TestAddListenerError() {
	l := newMockListener("mock", ":1883")

	err := s.srv.AddListener(l)
	s.Require().ErrorIs(err, ErrListenerAlreadyExists)
}

func (s *ServerTestSuite) TestAddListenerServerRunning() {
	_ = s.srv.Start(context.Background())
	l := newMockListener("mock2", ":1883")

	err := s.srv.AddListener(l)
	s.Require().NoError(err)
	s.Require().True(l.Listening())
}

func (s *ServerTestSuite) TestAddListenerServerRunningError() {
	_ = s.srv.Start(context.Background())
	l := newMockListener("mock2", "abc")

	err := s.srv.AddListener(l)
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

	err := s.srv.AddHook(hook)
	s.Assert().Error(err)
}

func (s *ServerTestSuite) TestAddHookOnStartError() {
	s.startServer()
	hook := newMockHook()
	hook.On("OnStart").Return(errors.New("failed"))

	err := s.srv.AddHook(hook)
	s.Assert().Error(err)
	hook.AssertExpectations(s.T())
}

func (s *ServerTestSuite) TestStartSuccess() {
	err := s.srv.Start(context.Background())
	s.Require().NoError(err)
	s.Assert().Equal(ServerRunning, s.srv.State())
	s.Assert().True(s.lsn.Listening())
}

func (s *ServerTestSuite) TestStartWithoutListeners() {
	srv := NewServer(NewDefaultOptions())

	err := srv.Start(context.Background())
	s.Require().NoError(err)
	s.Assert().Equal(ServerRunning, srv.State())
}

func (s *ServerTestSuite) TestStartWithHook() {
	s.addHook()
	s.hook.On("OnServerStart", s.srv)
	s.hook.On("OnStart")
	s.hook.On("OnServerStarted", s.srv)

	err := s.srv.Start(context.Background())
	s.Require().NoError(err)
	s.Assert().Equal(ServerRunning, s.srv.State())
}

func (s *ServerTestSuite) TestStartWhenClosedError() {
	s.addHook()
	s.startServer()
	s.stopServer()
	s.srv.Close()
	s.Require().Equal(ServerClosed, s.srv.State())

	err := s.srv.Start(context.Background())
	s.Require().Error(err, ErrInvalidServerState)
}

func (s *ServerTestSuite) TestStartOnServerStartError() {
	s.addHook()
	err := errors.New("failed")
	s.hook.On("OnServerStart", s.srv).Return(err)
	s.hook.On("OnServerStartFailed", s.srv, err)

	err = s.srv.Start(context.Background())
	s.Require().Error(err)
	s.Assert().Equal(ServerFailed, s.srv.State())
}

func (s *ServerTestSuite) TestStartOnStartError() {
	s.addHook()
	err := errors.New("failed")
	s.hook.On("OnServerStart", s.srv)
	s.hook.On("OnStart").Return(err)
	s.hook.On("OnServerStartFailed", s.srv, err)

	err = s.srv.Start(context.Background())
	s.Require().Error(err)
	s.Assert().Equal(ServerFailed, s.srv.State())
}

func (s *ServerTestSuite) TestStartListenerListenError() {
	s.addHook()
	s.hook.On("OnServerStart", s.srv)
	s.hook.On("OnStart")
	s.hook.On("OnServerStartFailed", s.srv, mock.Anything)
	l := newMockListener("mock2", "abc")
	_ = s.srv.AddListener(l)

	err := s.srv.Start(context.Background())
	s.Require().Error(err)
	s.Assert().Equal(ServerFailed, s.srv.State())
}

func (s *ServerTestSuite) TestStop() {
	s.startServer()

	err := s.srv.Stop(context.Background())
	s.Require().NoError(err)
	s.Assert().Equal(ServerStopped, s.srv.State())
	s.Assert().False(s.lsn.Listening())
}

func (s *ServerTestSuite) TestStopWithHook() {
	s.addHook()
	s.hook.On("OnServerStart", s.srv)
	s.hook.On("OnStart")
	s.hook.On("OnServerStarted", s.srv)
	s.hook.On("OnServerStop", s.srv)
	s.hook.On("OnStop")
	s.hook.On("OnServerStopped", s.srv)
	s.startServer()

	err := s.srv.Stop(context.Background())
	s.Require().NoError(err)
	s.Assert().Equal(ServerStopped, s.srv.State())
	s.Assert().False(s.lsn.Listening())
}

func (s *ServerTestSuite) TestStopCancelled() {
	s.startServer()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := s.srv.Stop(ctx)
	s.Require().Error(err)
}

func (s *ServerTestSuite) TestStopWhenNotStartedSuccess() {
	s.addHook()

	err := s.srv.Stop(context.Background())
	s.Require().NoError(err)
}

func (s *ServerTestSuite) TestStopWhenStoppedSuccess() {
	s.addHook()
	s.startServer()
	s.stopServer()

	err := s.srv.Stop(context.Background())
	s.Require().NoError(err)
}

func (s *ServerTestSuite) TestCloseWhenStopping() {
	s.startServer()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_ = s.srv.Stop(ctx)

	s.srv.Close()
	s.Require().Equal(ServerClosed, s.srv.State())
}

func (s *ServerTestSuite) TestCloseWhenStopped() {
	s.startServer()
	_ = s.srv.Stop(context.Background())

	s.srv.Close()
	s.Require().Equal(ServerClosed, s.srv.State())
}

func (s *ServerTestSuite) TestCloseWhenRunning() {
	s.addHook()
	s.startServer()
	s.hook.On("OnServerStop", s.srv)
	s.hook.On("OnStop")
	s.hook.On("OnServerStopped", s.srv)

	s.srv.Close()
	s.Require().Equal(ServerClosed, s.srv.State())
}

func (s *ServerTestSuite) TestCloseWhenNotStartedNoAction() {
	s.addHook()

	s.srv.Close()
	s.Require().Equal(ServerClosed, s.srv.State())
}

func (s *ServerTestSuite) TestHandleConnection() {
	s.startServer()
	c1, c2 := net.Pipe()

	s.lsn.onConnection(c2)
	_ = c1.Close()
	s.stopServer()
	s.Require().Equal(ServerStopped, s.srv.State())
}

func (s *ServerTestSuite) TestHandleConnectionWithHook() {
	var c *Client
	s.addHook()
	s.hook.On("OnServerStart", s.srv)
	s.hook.On("OnStart")
	s.hook.On("OnServerStarted", s.srv)
	s.hook.On("OnClientOpen", s.srv, s.lsn, mock.MatchedBy(func(cl *Client) bool {
		c = cl
		return true
	}))
	s.hook.On("OnClientOpened", s.srv, s.lsn, mock.MatchedBy(func(cl *Client) bool { return cl == c }))
	s.hook.On("OnClientClose", mock.MatchedBy(func(cl *Client) bool { return cl == c }))
	s.hook.On("OnClientClosed", mock.MatchedBy(func(cl *Client) bool { return cl == c }))
	s.hook.On("OnServerStop", s.srv)
	s.hook.On("OnStop")
	s.hook.On("OnServerStopped", s.srv)
	s.startServer()
	c1, c2 := net.Pipe()

	s.lsn.onConnection(c2)
	_ = c1.Close()
	s.stopServer()
	s.Require().NotNil(c)
}

func (s *ServerTestSuite) TestOnConnectionOpenedError() {
	var c *Client
	s.addHook()
	s.hook.On("OnServerStart", s.srv)
	s.hook.On("OnStart")
	s.hook.On("OnServerStarted", s.srv)
	s.hook.On("OnClientOpen", s.srv, s.lsn, mock.Anything).Return(errors.New("failed"))
	s.hook.On("OnClientClose", mock.MatchedBy(func(cl *Client) bool {
		c = cl
		return true
	}))
	s.hook.On("OnClientClosed", mock.MatchedBy(func(cl *Client) bool {
		return cl == c
	}))
	s.startServer()
	c1, c2 := net.Pipe()
	defer func() { _ = c1.Close() }()

	s.lsn.onConnection(c2)
	s.Require().NotNil(c)
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}
