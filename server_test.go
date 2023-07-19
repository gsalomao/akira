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

package akira_test

import (
	"context"
	"io"
	"net"
	"os"
	"testing"
	"time"

	"github.com/gsalomao/akira"
	"github.com/gsalomao/akira/internal/mocks"
	"github.com/gsalomao/akira/packet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServerTestSuite struct {
	suite.Suite
	server *akira.Server
}

func (s *ServerTestSuite) SetupTest() {
	var err error

	options := akira.NewDefaultOptions()
	options.Config.ConnectTimeout = 1

	s.server, err = akira.NewServer(options)
	s.Require().NoError(err)
}

func (s *ServerTestSuite) TearDownTest() {
	s.server.Close()
}

func (s *ServerTestSuite) addListener(srv *akira.Server) (akira.Listener, <-chan akira.OnConnectionFunc) {
	onConnectionStream := make(chan akira.OnConnectionFunc, 1)

	listener := mocks.NewMockListener(s.T())
	listener.EXPECT().Listen(mock.Anything).RunAndReturn(func(cb akira.OnConnectionFunc) (<-chan bool, error) {
		onConnectionStream <- cb
		close(onConnectionStream)
		doneCh := make(chan bool)
		close(doneCh)
		return doneCh, nil
	})
	listener.EXPECT().Stop()
	err := srv.AddListener(listener)
	s.Require().NoError(err)

	return listener, onConnectionStream
}

func (s *ServerTestSuite) TestNewServer() {
	srv, err := akira.NewServer(nil)
	s.Require().NoError(err)
	s.Assert().Equal(akira.ServerNotStarted, srv.State())
}

func (s *ServerTestSuite) TestNewServerWithDefaultConfig() {
	srv, err := akira.NewServer(&akira.Options{})
	s.Require().NoError(err)
	s.Assert().Equal(akira.ServerNotStarted, srv.State())
}

func (s *ServerTestSuite) TestNewServerWithListeners() {
	listener := mocks.NewMockListener(s.T())

	srv, err := akira.NewServer(&akira.Options{Listeners: []akira.Listener{listener}})
	s.Require().NoError(err)
	s.Assert().Equal(akira.ServerNotStarted, srv.State())
}

func (s *ServerTestSuite) TestNewServerWithDuplicatedListeners() {
	listener := mocks.NewMockListener(s.T())

	srv, err := akira.NewServer(&akira.Options{Listeners: []akira.Listener{listener, listener}})
	s.Require().NoError(err)
	s.Assert().Equal(akira.ServerNotStarted, srv.State())
}

func (s *ServerTestSuite) TestNewServerWithHooks() {
	hook := mocks.NewMockOnServerStartHook(s.T())
	hook.EXPECT().Name().Return("mock")

	srv, err := akira.NewServer(&akira.Options{Hooks: []akira.Hook{hook}})
	s.Require().NoError(err)
	s.Assert().Equal(akira.ServerNotStarted, srv.State())
}

func (s *ServerTestSuite) TestNewServerWithDuplicatedHooksReturnsError() {
	hook := mocks.NewMockOnServerStartHook(s.T())
	hook.EXPECT().Name().Return("mock")

	srv, err := akira.NewServer(&akira.Options{Hooks: []akira.Hook{hook, hook}})
	s.Require().Error(err)
	s.Require().Nil(srv)
}

func (s *ServerTestSuite) TestAddHook() {
	hook := mocks.NewMockHook(s.T())
	hook.EXPECT().Name().Return("hook")

	err := s.server.AddHook(hook)
	s.Require().NoError(err)
}

func (s *ServerTestSuite) TestStart() {
	err := s.server.Start(context.Background())
	s.Require().NoError(err)
	s.Assert().Equal(akira.ServerRunning, s.server.State())
}

func (s *ServerTestSuite) TestStartWithHooks() {
	onServerStart := mocks.NewMockOnServerStartHook(s.T())
	onServerStart.EXPECT().Name().Return("onServerStart")
	onServerStart.EXPECT().OnServerStart(s.server).Return(nil)
	_ = s.server.AddHook(onServerStart)

	onStart := mocks.NewMockOnStartHook(s.T())
	onStart.EXPECT().Name().Return("onStart")
	onStart.EXPECT().OnStart(s.server).Return(nil)
	_ = s.server.AddHook(onStart)

	onServerStarted := mocks.NewMockOnServerStartedHook(s.T())
	onServerStarted.EXPECT().Name().Return("onServerStarted")
	onServerStarted.EXPECT().OnServerStarted(s.server)
	_ = s.server.AddHook(onServerStarted)

	onServerStop := mocks.NewMockOnServerStopHook(s.T())
	onServerStop.EXPECT().Name().Return("onServerStop")
	onServerStop.EXPECT().OnServerStop(s.server)
	_ = s.server.AddHook(onServerStop)

	onStop := mocks.NewMockOnStopHook(s.T())
	onStop.EXPECT().Name().Return("onStop")
	onStop.EXPECT().OnStop(s.server)
	_ = s.server.AddHook(onStop)

	onServerStopped := mocks.NewMockOnServerStoppedHook(s.T())
	onServerStopped.EXPECT().Name().Return("onServerStopped")
	onServerStopped.EXPECT().OnServerStopped(s.server)
	_ = s.server.AddHook(onServerStopped)

	err := s.server.Start(context.Background())
	s.Require().NoError(err)
	s.Assert().Equal(akira.ServerRunning, s.server.State())
}

func (s *ServerTestSuite) TestStartWithOnServerStartReturningError() {
	onServerStart := mocks.NewMockOnServerStartHook(s.T())
	onServerStart.EXPECT().Name().Return("onServerStart")
	onServerStart.EXPECT().OnServerStart(s.server).Return(assert.AnError)
	_ = s.server.AddHook(onServerStart)

	onServerStartFailed := mocks.NewMockOnServerStartFailedHook(s.T())
	onServerStartFailed.EXPECT().Name().Return("onServerStartFailed")
	onServerStartFailed.EXPECT().OnServerStartFailed(s.server, assert.AnError)
	_ = s.server.AddHook(onServerStartFailed)

	err := s.server.Start(context.Background())
	s.Require().ErrorIs(err, assert.AnError)
	s.Assert().Equal(akira.ServerFailed, s.server.State())
}

func (s *ServerTestSuite) TestStartWithOnStartReturningError() {
	onStart := mocks.NewMockOnStartHook(s.T())
	onStart.EXPECT().Name().Return("onStart")
	onStart.EXPECT().OnStart(s.server).Return(assert.AnError)
	_ = s.server.AddHook(onStart)

	onServerStartFailed := mocks.NewMockOnServerStartFailedHook(s.T())
	onServerStartFailed.EXPECT().Name().Return("onServerStartFailed")
	onServerStartFailed.EXPECT().OnServerStartFailed(s.server, assert.AnError)
	_ = s.server.AddHook(onServerStartFailed)

	err := s.server.Start(context.Background())
	s.Require().ErrorIs(err, assert.AnError)
	s.Assert().Equal(akira.ServerFailed, s.server.State())
}

func (s *ServerTestSuite) TestStartWhenServerClosedReturnsError() {
	_ = s.server.Start(context.Background())
	s.server.Close()
	s.Require().Equal(akira.ServerClosed, s.server.State())

	err := s.server.Start(context.Background())
	s.Require().Error(err, akira.ErrInvalidServerState)
}

func (s *ServerTestSuite) TestAddListener() {
	listener := mocks.NewMockListener(s.T())

	err := s.server.AddListener(listener)
	s.Require().NoError(err)
}

func (s *ServerTestSuite) TestAddListenerWhenServerRunning() {
	_ = s.server.Start(context.Background())
	listener := mocks.NewMockListener(s.T())
	listener.EXPECT().Listen(mock.Anything).RunAndReturn(func(_ akira.OnConnectionFunc) (<-chan bool, error) {
		doneCh := make(chan bool)
		close(doneCh)
		return doneCh, nil
	})
	listener.EXPECT().Stop()

	err := s.server.AddListener(listener)
	s.Require().NoError(err)
}

func (s *ServerTestSuite) TestAddDuplicatedListener() {
	listener1 := mocks.NewMockListener(s.T())
	_ = s.server.AddListener(listener1)
	listener2 := mocks.NewMockListener(s.T())

	err := s.server.AddListener(listener2)
	s.Require().NoError(err)
}

func (s *ServerTestSuite) TestAddListenerWithListenErrorWhenServerRunning() {
	_ = s.server.Start(context.Background())
	listener := mocks.NewMockListener(s.T())
	listener.EXPECT().Listen(mock.Anything).Return(nil, assert.AnError)

	err := s.server.AddListener(listener)
	s.Require().ErrorIs(err, assert.AnError)
}

func (s *ServerTestSuite) TestStartWithListenerFailingToListen() {
	listener := mocks.NewMockListener(s.T())
	listener.EXPECT().Listen(mock.Anything).Return(nil, assert.AnError)
	_ = s.server.AddListener(listener)

	err := s.server.Start(context.Background())
	s.Require().ErrorIs(err, assert.AnError)
	s.Assert().Equal(akira.ServerFailed, s.server.State())
}

func (s *ServerTestSuite) TestAddHookCallsOnStartWhenServerRunning() {
	_ = s.server.Start(context.Background())
	hook := mocks.NewMockOnStartHook(s.T())
	hook.EXPECT().Name().Return("hook")
	hook.EXPECT().OnStart(s.server).Return(nil)

	err := s.server.AddHook(hook)
	s.Require().NoError(err)
}

func (s *ServerTestSuite) TestAddHookWithOnStartReturningError() {
	_ = s.server.Start(context.Background())
	hook := mocks.NewMockOnStartHook(s.T())
	hook.EXPECT().OnStart(s.server).Return(assert.AnError)

	err := s.server.AddHook(hook)
	s.Require().ErrorIs(err, assert.AnError)
}

func (s *ServerTestSuite) TestStop() {
	listener := mocks.NewMockListener(s.T())
	listener.EXPECT().Listen(mock.Anything).RunAndReturn(func(_ akira.OnConnectionFunc) (<-chan bool, error) {
		doneCh := make(chan bool)
		close(doneCh)
		return doneCh, nil
	})
	listener.EXPECT().Stop()
	_ = s.server.AddListener(listener)
	_ = s.server.Start(context.Background())

	err := s.server.Stop(context.Background())
	s.Require().NoError(err)
	s.Assert().Equal(akira.ServerStopped, s.server.State())
}

func (s *ServerTestSuite) TestStopWithHooks() {
	onServerStop := mocks.NewMockOnServerStopHook(s.T())
	onServerStop.EXPECT().Name().Return("onServerStop")
	onServerStop.EXPECT().OnServerStop(s.server)
	_ = s.server.AddHook(onServerStop)

	onStop := mocks.NewMockOnStopHook(s.T())
	onStop.EXPECT().Name().Return("onStop")
	onStop.EXPECT().OnStop(s.server)
	_ = s.server.AddHook(onStop)

	onServerStopped := mocks.NewMockOnServerStoppedHook(s.T())
	onServerStopped.EXPECT().Name().Return("onServerStopped")
	onServerStopped.EXPECT().OnServerStopped(s.server)
	_ = s.server.AddHook(onServerStopped)
	_ = s.server.Start(context.Background())

	err := s.server.Stop(context.Background())
	s.Require().NoError(err)
	s.Assert().Equal(akira.ServerStopped, s.server.State())
}

func (s *ServerTestSuite) TestStopClosesAllClients() {
	var onConnection akira.OnConnectionFunc
	listeningCh := make(chan struct{})

	listener := mocks.NewMockListener(s.T())
	listener.EXPECT().Listen(mock.Anything).RunAndReturn(func(cb akira.OnConnectionFunc) (<-chan bool, error) {
		onConnection = cb
		doneCh := make(chan bool)
		close(doneCh)
		close(listeningCh)
		return doneCh, nil
	})
	listener.EXPECT().Stop()
	_ = s.server.AddListener(listener)

	receivingCh := make(chan struct{})
	onPacketReceive := mocks.NewMockOnPacketReceiveHook(s.T())
	onPacketReceive.EXPECT().Name().Return("onPacketReceive")
	onPacketReceive.EXPECT().OnPacketReceive(mock.Anything).RunAndReturn(func(_ *akira.Client) error {
		close(receivingCh)
		return nil
	})
	_ = s.server.AddHook(onPacketReceive)
	_ = s.server.Start(context.Background())

	conn1, conn2 := net.Pipe()
	defer func() { _ = conn1.Close() }()

	<-listeningCh
	onConnection(listener, conn2)
	<-receivingCh

	err := s.server.Stop(context.Background())
	s.Require().NoError(err)
	s.Assert().Equal(akira.ServerStopped, s.server.State())
}

func (s *ServerTestSuite) TestStopReturnsErrorWhenCancelled() {
	_ = s.server.Start(context.Background())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := s.server.Stop(ctx)
	s.Require().Error(err)
	s.Assert().Equal(akira.ServerStopping, s.server.State())
}

func (s *ServerTestSuite) TestStopWhenServerNotRunning() {
	err := s.server.Stop(context.Background())
	s.Require().NoError(err)
}

func (s *ServerTestSuite) TestCloseWhenServerStopping() {
	_ = s.server.Start(context.Background())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_ = s.server.Stop(ctx)

	s.server.Close()
	s.Require().Equal(akira.ServerClosed, s.server.State())
}

func (s *ServerTestSuite) TestCloseWhenServerStopped() {
	_ = s.server.Start(context.Background())
	_ = s.server.Stop(context.Background())

	s.server.Close()
	s.Require().Equal(akira.ServerClosed, s.server.State())
}

func (s *ServerTestSuite) TestCloseWhenServerRunning() {
	_ = s.server.Start(context.Background())

	s.server.Close()
	s.Require().Equal(akira.ServerClosed, s.server.State())
}

func (s *ServerTestSuite) TestCloseWhenServerNotStarted() {
	s.server.Close()
	s.Require().Equal(akira.ServerClosed, s.server.State())
}

func (s *ServerTestSuite) TestHandleConnection() {
	listener, onConnectionStream := s.addListener(s.server)

	onConnOpen := mocks.NewMockOnConnectionOpenHook(s.T())
	onConnOpen.EXPECT().Name().Return("onConnOpen")
	onConnOpen.EXPECT().OnConnectionOpen(s.server, listener).Return(nil)
	_ = s.server.AddHook(onConnOpen)

	connOpenedCh := make(chan struct{})
	onConnOpened := mocks.NewMockOnConnectionOpenedHook(s.T())
	onConnOpened.EXPECT().Name().Return("onConnOpened")
	onConnOpened.EXPECT().OnConnectionOpened(s.server, listener).Run(func(_ *akira.Server, _ akira.Listener) {
		close(connOpenedCh)
	})
	_ = s.server.AddHook(onConnOpened)

	onPacketRcv := mocks.NewMockOnPacketReceiveHook(s.T())
	onPacketRcv.EXPECT().Name().Return("onPacketRcv")
	onPacketRcv.EXPECT().OnPacketReceive(mock.Anything).Return(nil)
	_ = s.server.AddHook(onPacketRcv)

	onConnClose := mocks.NewMockOnConnectionCloseHook(s.T())
	onConnClose.EXPECT().Name().Return("onConnClose")
	onConnClose.EXPECT().OnConnectionClose(s.server, listener, mock.Anything).
		Run(func(_ *akira.Server, _ akira.Listener, err error) { s.Assert().ErrorIs(err, io.EOF) })
	_ = s.server.AddHook(onConnClose)

	connClosedCh := make(chan struct{})
	onConnClosed := mocks.NewMockOnConnectionClosedHook(s.T())
	onConnClosed.EXPECT().Name().Return("onConnClosed")
	onConnClosed.EXPECT().OnConnectionClosed(s.server, listener, mock.Anything).
		Run(func(_ *akira.Server, _ akira.Listener, err error) {
			s.Assert().ErrorIs(err, io.EOF)
			close(connClosedCh)
		})
	_ = s.server.AddHook(onConnClosed)
	_ = s.server.Start(context.Background())

	conn1, conn2 := net.Pipe()
	onConnection := <-onConnectionStream
	onConnection(listener, conn2)
	<-connOpenedCh

	_ = conn1.Close()
	<-connClosedCh
}

func (s *ServerTestSuite) TestHandleConnectionWithReadTimeout() {
	listener, onConnectionStream := s.addListener(s.server)

	connClosedCh := make(chan struct{})
	onConnClosed := mocks.NewMockOnConnectionClosedHook(s.T())
	onConnClosed.EXPECT().Name().Return("onConnClosed")
	onConnClosed.EXPECT().OnConnectionClosed(s.server, listener, mock.Anything).
		Run(func(_ *akira.Server, _ akira.Listener, err error) {
			s.Assert().ErrorIs(err, os.ErrDeadlineExceeded)
			close(connClosedCh)
		})
	_ = s.server.AddHook(onConnClosed)
	_ = s.server.Start(context.Background())

	conn1, conn2 := net.Pipe()
	defer func() { _ = conn1.Close() }()

	onConnection := <-onConnectionStream
	onConnection(listener, conn2)
	<-connClosedCh
}

func (s *ServerTestSuite) TestHandleConnectionWithOnConnectionOpenReturningError() {
	listener, onConnectionStream := s.addListener(s.server)

	onConnOpen := mocks.NewMockOnConnectionOpenHook(s.T())
	onConnOpen.EXPECT().Name().Return("onConnOpen")
	onConnOpen.EXPECT().OnConnectionOpen(s.server, listener).Return(assert.AnError)
	_ = s.server.AddHook(onConnOpen)

	connClosedCh := make(chan struct{})
	onConnClosed := mocks.NewMockOnConnectionClosedHook(s.T())
	onConnClosed.EXPECT().Name().Return("onConnClosed")
	onConnClosed.EXPECT().OnConnectionClosed(s.server, listener, assert.AnError).
		Run(func(_ *akira.Server, _ akira.Listener, _ error) { close(connClosedCh) })
	_ = s.server.AddHook(onConnClosed)
	_ = s.server.Start(context.Background())

	conn1, conn2 := net.Pipe()
	defer func() { _ = conn1.Close() }()

	onConnection := <-onConnectionStream
	onConnection(listener, conn2)
	<-connClosedCh
}

func (s *ServerTestSuite) TestReceivePacket() {
	var connect *packet.Connect
	listener, onConnectionStream := s.addListener(s.server)
	receivedCh := make(chan struct{})

	onPacketReceived := mocks.NewMockOnPacketReceivedHook(s.T())
	onPacketReceived.EXPECT().Name().Return("onPacketReceived")
	onPacketReceived.EXPECT().OnPacketReceived(mock.Anything, mock.Anything).
		RunAndReturn(func(_ *akira.Client, p akira.Packet) error {
			s.Require().Equal(packet.TypeConnect, p.Type())
			connect = p.(*packet.Connect)
			close(receivedCh)
			return nil
		})
	_ = s.server.AddHook(onPacketReceived)
	_ = s.server.Start(context.Background())

	conn1, conn2 := net.Pipe()
	defer func() { _ = conn1.Close() }()

	onConnection := <-onConnectionStream
	onConnection(listener, conn2)

	msg := []byte{0x10, 14, 0, 4, 'M', 'Q', 'T', 'T', 4, 2, 0, 255, 0, 2, 'a', 'b'}
	_, _ = conn1.Write(msg)
	<-receivedCh

	expected := &packet.Connect{
		Version:   packet.MQTT311,
		KeepAlive: 255,
		Flags:     packet.ConnectFlags(0x02), // Clean session flag
		ClientID:  []byte("ab"),
	}
	s.Assert().Equal(expected, connect)
}

func (s *ServerTestSuite) TestCloseConnectionWhenReceiveInvalidPacket() {
	listener, onConnectionStream := s.addListener(s.server)

	connClosedCh := make(chan struct{})
	onConnClosed := mocks.NewMockOnConnectionClosedHook(s.T())
	onConnClosed.EXPECT().Name().Return("onConnClosed")
	onConnClosed.EXPECT().OnConnectionClosed(s.server, listener, mock.Anything).
		Run(func(_ *akira.Server, _ akira.Listener, _ error) { close(connClosedCh) })
	_ = s.server.AddHook(onConnClosed)
	_ = s.server.Start(context.Background())

	conn1, conn2 := net.Pipe()
	defer func() { _ = conn1.Close() }()

	onConnection := <-onConnectionStream
	onConnection(listener, conn2)

	msg := []byte{0x10, 7, 0, 4, 'M', 'Q', 'T', 'T', 0}
	_, _ = conn1.Write(msg)
	<-connClosedCh
}

func (s *ServerTestSuite) TestCloseConnectionIfOnPacketReceiveErrorReturnsError() {
	listener, onConnectionStream := s.addListener(s.server)

	onPacketRcvError := mocks.NewMockOnPacketReceiveErrorHook(s.T())
	onPacketRcvError.EXPECT().Name().Return("onPacketRcvError")
	onPacketRcvError.EXPECT().OnPacketReceiveError(mock.Anything, mock.Anything).Return(assert.AnError)
	_ = s.server.AddHook(onPacketRcvError)

	connClosedCh := make(chan struct{})
	onConnClosed := mocks.NewMockOnConnectionClosedHook(s.T())
	onConnClosed.EXPECT().Name().Return("onConnClosed")
	onConnClosed.EXPECT().OnConnectionClosed(s.server, listener, assert.AnError).
		Run(func(_ *akira.Server, _ akira.Listener, _ error) { close(connClosedCh) })
	_ = s.server.AddHook(onConnClosed)
	_ = s.server.Start(context.Background())

	conn1, conn2 := net.Pipe()
	defer func() { _ = conn1.Close() }()

	onConnection := <-onConnectionStream
	onConnection(listener, conn2)

	msg := []byte{0x10, 7, 0, 4, 'M', 'Q', 'T', 'T', 0}
	_, _ = conn1.Write(msg)
	<-connClosedCh
}

func (s *ServerTestSuite) TestKeepReceivingWhenOnPacketReceiveErrorDoesNotReturnError() {
	listener, onConnectionStream := s.addListener(s.server)

	onPacketRcv := mocks.NewMockOnPacketReceiveHook(s.T())
	onPacketRcv.EXPECT().Name().Return("onPacketRcv")
	onPacketRcv.EXPECT().OnPacketReceive(mock.Anything).Return(nil).Twice()
	_ = s.server.AddHook(onPacketRcv)

	receivedCh := make(chan struct{})
	onPacketRcvError := mocks.NewMockOnPacketReceiveErrorHook(s.T())
	onPacketRcvError.EXPECT().Name().Return("onPacketRcvError")
	onPacketRcvError.EXPECT().OnPacketReceiveError(mock.Anything, mock.Anything).
		RunAndReturn(func(_ *akira.Client, _ error) error {
			close(receivedCh)
			return nil
		})
	_ = s.server.AddHook(onPacketRcvError)

	connClosedCh := make(chan struct{})
	onConnClosed := mocks.NewMockOnConnectionClosedHook(s.T())
	onConnClosed.EXPECT().Name().Return("onConnClosed")
	onConnClosed.EXPECT().OnConnectionClosed(s.server, listener, mock.Anything).
		Run(func(_ *akira.Server, _ akira.Listener, _ error) { close(connClosedCh) })
	_ = s.server.AddHook(onConnClosed)
	_ = s.server.Start(context.Background())

	conn1, conn2 := net.Pipe()

	onConnection := <-onConnectionStream
	onConnection(listener, conn2)

	msg := []byte{0x10, 7, 0, 4, 'M', 'Q', 'T', 'T', 0}
	_, _ = conn1.Write(msg)
	<-receivedCh

	_ = conn1.Close()
	<-connClosedCh
}

func (s *ServerTestSuite) TestCloseConnectionWhenOnPacketReceiveReturnsError() {
	listener, onConnectionStream := s.addListener(s.server)

	onPacketRcv := mocks.NewMockOnPacketReceiveHook(s.T())
	onPacketRcv.EXPECT().Name().Return("onPacketRcv")
	onPacketRcv.EXPECT().OnPacketReceive(mock.Anything).Return(assert.AnError)
	_ = s.server.AddHook(onPacketRcv)

	connClosedCh := make(chan struct{})
	onConnClosed := mocks.NewMockOnConnectionClosedHook(s.T())
	onConnClosed.EXPECT().Name().Return("onConnClosed")
	onConnClosed.EXPECT().OnConnectionClosed(s.server, listener, assert.AnError).
		Run(func(_ *akira.Server, _ akira.Listener, err error) { close(connClosedCh) })
	_ = s.server.AddHook(onConnClosed)
	_ = s.server.Start(context.Background())

	conn1, conn2 := net.Pipe()
	defer func() { _ = conn1.Close() }()

	onConnection := <-onConnectionStream
	onConnection(listener, conn2)
	<-connClosedCh
}

func (s *ServerTestSuite) TestCloseConnectionWhenOnPacketReceivedReturnsError() {
	listener, onConnectionStream := s.addListener(s.server)

	onPacketReceived := mocks.NewMockOnPacketReceivedHook(s.T())
	onPacketReceived.EXPECT().Name().Return("onPacketReceived")
	onPacketReceived.EXPECT().OnPacketReceived(mock.Anything, mock.Anything).
		RunAndReturn(func(_ *akira.Client, _ akira.Packet) error { return assert.AnError })
	_ = s.server.AddHook(onPacketReceived)

	connClosedCh := make(chan struct{})
	onConnClosed := mocks.NewMockOnConnectionClosedHook(s.T())
	onConnClosed.EXPECT().Name().Return("onConnClosed")
	onConnClosed.EXPECT().OnConnectionClosed(s.server, listener, assert.AnError).
		Run(func(_ *akira.Server, _ akira.Listener, err error) { close(connClosedCh) })
	_ = s.server.AddHook(onConnClosed)
	_ = s.server.Start(context.Background())

	conn1, conn2 := net.Pipe()
	defer func() { _ = conn1.Close() }()

	onConnection := <-onConnectionStream
	onConnection(listener, conn2)

	msg := []byte{0x10, 14, 0, 4, 'M', 'Q', 'T', 'T', 4, 2, 0, 255, 0, 2, 'a', 'b'}
	_, _ = conn1.Write(msg)
	<-connClosedCh
}

func (s *ServerTestSuite) TestCloseConnectionWhenOnPacketSendReturnsError() {
	opts := akira.NewDefaultOptions()
	srv, _ := akira.NewServer(opts)
	defer srv.Close()

	listener, onConnectionStream := s.addListener(srv)

	onPacketSend := mocks.NewMockOnPacketSendHook(s.T())
	onPacketSend.EXPECT().Name().Return("onPacketSend")
	onPacketSend.EXPECT().OnPacketSend(mock.Anything, mock.Anything).Return(assert.AnError)
	_ = srv.AddHook(onPacketSend)

	conn1, conn2 := net.Pipe()
	defer func() { _ = conn1.Close() }()
	_ = srv.Start(context.Background())

	onConnection := <-onConnectionStream
	onConnection(listener, conn2)

	msg := []byte{0x10, 14, 0, 4, 'M', 'Q', 'T', 'T', 4, 2, 0, 100, 0, 2, 'a', 'b'}
	_, _ = conn1.Write(msg)

	reply := make([]byte, 4)
	_, err := conn1.Read(reply)
	s.Require().ErrorIs(err, io.EOF)
}

func (s *ServerTestSuite) TestCloseConnectionWhenOnPacketSentReturnsError() {
	opts := akira.NewDefaultOptions()
	srv, _ := akira.NewServer(opts)
	defer srv.Close()

	listener, onConnectionStream := s.addListener(srv)

	onPacketSent := mocks.NewMockOnPacketSentHook(s.T())
	onPacketSent.EXPECT().Name().Return("onPacketSent")
	onPacketSent.EXPECT().OnPacketSent(mock.Anything, mock.Anything).Return(assert.AnError)
	_ = srv.AddHook(onPacketSent)

	conn1, conn2 := net.Pipe()
	defer func() { _ = conn1.Close() }()
	_ = srv.Start(context.Background())

	onConnection := <-onConnectionStream
	onConnection(listener, conn2)

	msg := []byte{0x10, 14, 0, 4, 'M', 'Q', 'T', 'T', 4, 2, 0, 100, 0, 2, 'a', 'b'}
	_, _ = conn1.Write(msg)

	reply := make([]byte, 4)
	_, err := conn1.Read(reply)
	s.Require().NoError(err)

	_, err = conn1.Read(reply)
	s.Require().ErrorIs(err, io.EOF)
}

func (s *ServerTestSuite) TestOnPacketSendErrorIsCalledWhenFailedToSendPacket() {
	opts := akira.NewDefaultOptions()
	srv, _ := akira.NewServer(opts)
	defer srv.Close()

	listener, onConnectionStream := s.addListener(srv)

	sendErrCh := make(chan struct{})
	onPacketSendErr := mocks.NewMockOnPacketSendErrorHook(s.T())
	onPacketSendErr.EXPECT().Name().Return("onPacketSendErr")
	onPacketSendErr.EXPECT().OnPacketSendError(mock.Anything, mock.Anything, mock.Anything).
		Run(func(_ *akira.Client, _ akira.Packet, err error) {
			s.Require().Error(err)
			close(sendErrCh)
		})
	_ = srv.AddHook(onPacketSendErr)

	conn1, conn2 := net.Pipe()
	_ = srv.Start(context.Background())

	onConnection := <-onConnectionStream
	onConnection(listener, conn2)

	msg := []byte{0x10, 14, 0, 4, 'M', 'Q', 'T', 'T', 4, 2, 0, 100, 0, 2, 'a', 'b'}
	_, _ = conn1.Write(msg)
	_ = conn1.Close()
	<-sendErrCh
}

func (s *ServerTestSuite) TestConnectPacket() {
	testCases := []struct {
		name    string
		connect []byte
		connack []byte
	}{
		{
			"V3.1",
			[]byte{0x10, 16, 0, 6, 'M', 'Q', 'I', 's', 'd', 'p', 3, 0, 0, 100, 0, 2, 'a', 'b'},
			[]byte{0x20, 2, 0, 0},
		},
		{
			"V3.1.1",
			[]byte{0x10, 14, 0, 4, 'M', 'Q', 'T', 'T', 4, 0, 0, 100, 0, 2, 'a', 'b'},
			[]byte{0x20, 2, 0, 0},
		},
		{
			"V5.0",
			[]byte{0x10, 15, 0, 4, 'M', 'Q', 'T', 'T', 5, 0, 0, 100, 0, 0, 2, 'a', 'b'},
			[]byte{0x20, 3, 0, 0, 0},
		},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			var session *akira.Session

			store := mocks.NewMockSessionStore(s.T())
			store.EXPECT().GetSession([]byte{'a', 'b'}).Return(nil, nil)
			store.EXPECT().SaveSession(mock.Anything, mock.Anything).
				RunAndReturn(func(id []byte, ss *akira.Session) error {
					s.Assert().Equal([]byte{'a', 'b'}, id)
					s.Assert().Equal(id, ss.ClientID)
					s.Assert().True(ss.Connected)
					session = ss
					return nil
				})

			opts := akira.NewDefaultOptions()
			opts.SessionStore = store
			srv, _ := akira.NewServer(opts)
			defer srv.Close()

			listener, onConnectionStream := s.addListener(srv)

			onPacketSend := mocks.NewMockOnPacketSendHook(s.T())
			onPacketSend.EXPECT().Name().Return("onPacketSend")
			onPacketSend.EXPECT().OnPacketSend(mock.Anything, mock.Anything).
				RunAndReturn(func(c *akira.Client, p akira.Packet) error {
					s.Require().Equal(session, c.Session)
					s.Require().Equal(packet.TypeConnAck, p.Type())
					return nil
				})
			_ = srv.AddHook(onPacketSend)

			onPacketSent := mocks.NewMockOnPacketSentHook(s.T())
			onPacketSent.EXPECT().Name().Return("onPacketSent")
			onPacketSent.EXPECT().OnPacketSent(mock.Anything, mock.Anything).
				RunAndReturn(func(c *akira.Client, p akira.Packet) error {
					s.Require().Equal(session, c.Session)
					s.Require().Equal(packet.TypeConnAck, p.Type())
					return nil
				})
			_ = srv.AddHook(onPacketSent)

			onConnect := mocks.NewMockOnConnectHook(s.T())
			onConnect.EXPECT().Name().Return("onConnect")
			onConnect.EXPECT().OnConnect(mock.Anything, mock.Anything).
				RunAndReturn(func(c *akira.Client, p *packet.Connect) error {
					s.Require().NotNil(c)
					s.Require().NotNil(p)
					s.Assert().Equal([]byte{'a', 'b'}, p.ClientID)
					return nil
				})
			_ = srv.AddHook(onConnect)

			connectedCh := make(chan struct{})
			onConnected := mocks.NewMockOnConnectedHook(s.T())
			onConnected.EXPECT().Name().Return("onConnected")
			onConnected.EXPECT().OnConnected(mock.Anything).Run(func(c *akira.Client) {
				s.Require().Equal(akira.ClientConnected, c.State())
				s.Require().Equal(session, c.Session)
				s.Require().NotNil(c.Connection)
				s.Assert().WithinDuration(time.Now(), c.Connection.ConnectedAt, time.Second)
				close(connectedCh)
			})
			_ = srv.AddHook(onConnected)

			conn1, conn2 := net.Pipe()
			defer func() { _ = conn1.Close() }()
			_ = srv.Start(context.Background())

			onConnection := <-onConnectionStream
			onConnection(listener, conn2)
			_, _ = conn1.Write(test.connect)

			connack := make([]byte, len(test.connack))
			_, err := conn1.Read(connack)
			s.Require().NoError(err)
			s.Assert().Equal(test.connack, connack)
			<-connectedCh
		})
	}
}

func (s *ServerTestSuite) TestConnectPacketWithSessionPresent() {
	testCases := []struct {
		name    string
		connect []byte
		connack []byte
	}{
		{
			"V3.1",
			[]byte{0x10, 16, 0, 6, 'M', 'Q', 'I', 's', 'd', 'p', 3, 0, 0, 255, 0, 2, 'a', 'b'},
			[]byte{0x20, 2, 0, 0},
		},
		{
			"V3.1.1",
			[]byte{0x10, 14, 0, 4, 'M', 'Q', 'T', 'T', 4, 0, 0, 255, 0, 2, 'a', 'b'},
			[]byte{0x20, 2, 1, 0},
		},
		{
			"V5.0",
			[]byte{0x10, 15, 0, 4, 'M', 'Q', 'T', 'T', 5, 0, 0, 255, 0, 0, 2, 'a', 'b'},
			[]byte{0x20, 3, 1, 0, 0},
		},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			session := &akira.Session{ClientID: []byte{'a', 'b'}, Connected: false}

			store := mocks.NewMockSessionStore(s.T())
			store.EXPECT().GetSession([]byte{'a', 'b'}).Return(session, nil)
			store.EXPECT().SaveSession([]byte{'a', 'b'}, session).
				RunAndReturn(func(_ []byte, session *akira.Session) error {
					s.Assert().True(session.Connected)
					return nil
				})

			opts := akira.NewDefaultOptions()
			opts.SessionStore = store
			srv, _ := akira.NewServer(opts)
			defer srv.Close()

			listener, onConnectionStream := s.addListener(srv)

			conn1, conn2 := net.Pipe()
			defer func() { _ = conn1.Close() }()
			_ = srv.Start(context.Background())

			onConnection := <-onConnectionStream
			onConnection(listener, conn2)
			_, _ = conn1.Write(test.connect)

			connack := make([]byte, len(test.connack))
			_, err := conn1.Read(connack)
			s.Require().NoError(err)
			s.Assert().Equal(test.connack, connack)
		})
	}
}

func (s *ServerTestSuite) TestConnectPacketWithCleanSession() {
	store := mocks.NewMockSessionStore(s.T())
	store.EXPECT().DeleteSession([]byte{'a', 'b'}).Return(nil)
	store.EXPECT().SaveSession(mock.Anything, mock.Anything).
		RunAndReturn(func(id []byte, session *akira.Session) error {
			s.Assert().Equal([]byte{'a', 'b'}, id)
			s.Assert().Equal(id, session.ClientID)
			s.Assert().True(session.Connected)
			return nil
		})

	opts := akira.NewDefaultOptions()
	opts.SessionStore = store
	srv, _ := akira.NewServer(opts)
	defer srv.Close()

	listener, onConnectionStream := s.addListener(srv)

	conn1, conn2 := net.Pipe()
	defer func() { _ = conn1.Close() }()
	_ = srv.Start(context.Background())

	onConnection := <-onConnectionStream
	onConnection(listener, conn2)

	connect := []byte{0x10, 15, 0, 4, 'M', 'Q', 'T', 'T', 5, 2, 0, 255, 0, 0, 2, 'a', 'b'}
	_, _ = conn1.Write(connect)

	expected := []byte{0x20, 3, 0, 0, 0}
	connack := make([]byte, len(expected))

	_, err := conn1.Read(connack)
	s.Require().NoError(err)
	s.Assert().Equal(expected, connack)
}

func (s *ServerTestSuite) TestConnectPacketWithConfig() {
	testCases := []struct {
		name    string
		config  akira.Config
		connect []byte
		connack []byte
	}{
		{
			"Session expiry interval",
			akira.Config{
				MaxSessionExpiryInterval:      150,
				MaxQoS:                        2,
				RetainAvailable:               true,
				WildcardSubscriptionAvailable: true,
				SubscriptionIDAvailable:       true,
				SharedSubscriptionAvailable:   true,
			},
			[]byte{0x10, 20, 0, 4, 'M', 'Q', 'T', 'T', 5, 0, 0, 100, 5, 0x11, 0, 0, 0, 200, 0, 2, 'a', 'b'},
			[]byte{0x20, 8, 0, 0, 5, 0x11, 0, 0, 0, 150},
		},
		{
			"Server keep alive",
			akira.Config{
				MaxKeepAlive:                  50,
				MaxQoS:                        2,
				RetainAvailable:               true,
				WildcardSubscriptionAvailable: true,
				SubscriptionIDAvailable:       true,
				SharedSubscriptionAvailable:   true,
			},
			[]byte{0x10, 15, 0, 4, 'M', 'Q', 'T', 'T', 5, 2, 0, 100, 0, 0, 2, 'a', 'b'},
			[]byte{0x20, 6, 0, 0, 3, 0x13, 0, 50},
		},
		{
			"Receive maximum",
			akira.Config{
				MaxInflightMessages:           100,
				MaxQoS:                        2,
				RetainAvailable:               true,
				WildcardSubscriptionAvailable: true,
				SubscriptionIDAvailable:       true,
				SharedSubscriptionAvailable:   true,
			},
			[]byte{0x10, 15, 0, 4, 'M', 'Q', 'T', 'T', 5, 2, 0, 100, 0, 0, 2, 'a', 'b'},
			[]byte{0x20, 6, 0, 0, 3, 0x21, 0, 100},
		},
		{
			"Topic alias maximum",
			akira.Config{
				TopicAliasMax:                 10,
				MaxQoS:                        2,
				RetainAvailable:               true,
				WildcardSubscriptionAvailable: true,
				SubscriptionIDAvailable:       true,
				SharedSubscriptionAvailable:   true,
			},
			[]byte{0x10, 15, 0, 4, 'M', 'Q', 'T', 'T', 5, 2, 0, 100, 0, 0, 2, 'a', 'b'},
			[]byte{0x20, 6, 0, 0, 3, 0x22, 0, 10},
		},
		{
			"Maximum QoS",
			akira.Config{
				MaxQoS:                        1,
				RetainAvailable:               true,
				WildcardSubscriptionAvailable: true,
				SubscriptionIDAvailable:       true,
				SharedSubscriptionAvailable:   true,
			},
			[]byte{0x10, 15, 0, 4, 'M', 'Q', 'T', 'T', 5, 2, 0, 100, 0, 0, 2, 'a', 'b'},
			[]byte{0x20, 5, 0, 0, 2, 0x24, 1},
		},
		{
			"Retain available",
			akira.Config{
				MaxQoS:                        2,
				RetainAvailable:               false,
				WildcardSubscriptionAvailable: true,
				SubscriptionIDAvailable:       true,
				SharedSubscriptionAvailable:   true,
			},
			[]byte{0x10, 15, 0, 4, 'M', 'Q', 'T', 'T', 5, 2, 0, 100, 0, 0, 2, 'a', 'b'},
			[]byte{0x20, 5, 0, 0, 2, 0x25, 0},
		},
		{
			"Maximum packet size",
			akira.Config{
				MaxPacketSize:                 200,
				MaxQoS:                        2,
				RetainAvailable:               true,
				WildcardSubscriptionAvailable: true,
				SubscriptionIDAvailable:       true,
				SharedSubscriptionAvailable:   true,
			},
			[]byte{0x10, 15, 0, 4, 'M', 'Q', 'T', 'T', 5, 2, 0, 100, 0, 0, 2, 'a', 'b'},
			[]byte{0x20, 8, 0, 0, 5, 0x27, 0, 0, 0, 200},
		},
		{
			"Wildcard subscription available",
			akira.Config{
				MaxQoS:                        2,
				RetainAvailable:               true,
				WildcardSubscriptionAvailable: false,
				SubscriptionIDAvailable:       true,
				SharedSubscriptionAvailable:   true,
			},
			[]byte{0x10, 15, 0, 4, 'M', 'Q', 'T', 'T', 5, 2, 0, 100, 0, 0, 2, 'a', 'b'},
			[]byte{0x20, 5, 0, 0, 2, 0x28, 0},
		},
		{
			"Subscription identifier available",
			akira.Config{
				MaxQoS:                        2,
				RetainAvailable:               true,
				WildcardSubscriptionAvailable: true,
				SubscriptionIDAvailable:       false,
				SharedSubscriptionAvailable:   true,
			},
			[]byte{0x10, 15, 0, 4, 'M', 'Q', 'T', 'T', 5, 2, 0, 100, 0, 0, 2, 'a', 'b'},
			[]byte{0x20, 5, 0, 0, 2, 0x29, 0},
		},
		{
			"Shared subscription available",
			akira.Config{
				MaxQoS:                        2,
				RetainAvailable:               true,
				WildcardSubscriptionAvailable: true,
				SubscriptionIDAvailable:       true,
				SharedSubscriptionAvailable:   false,
			},
			[]byte{0x10, 15, 0, 4, 'M', 'Q', 'T', 'T', 5, 2, 0, 100, 0, 0, 2, 'a', 'b'},
			[]byte{0x20, 5, 0, 0, 2, 0x2A, 0},
		},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			opts := akira.NewDefaultOptions()
			if test.config.MaxSessionExpiryInterval != opts.Config.MaxSessionExpiryInterval {
				opts.Config.MaxSessionExpiryInterval = test.config.MaxSessionExpiryInterval
			}
			if test.config.MaxKeepAlive != opts.Config.MaxKeepAlive {
				opts.Config.MaxKeepAlive = test.config.MaxKeepAlive
			}
			if test.config.MaxInflightMessages != opts.Config.MaxInflightMessages {
				opts.Config.MaxInflightMessages = test.config.MaxInflightMessages
			}
			if test.config.TopicAliasMax != opts.Config.TopicAliasMax {
				opts.Config.TopicAliasMax = test.config.TopicAliasMax
			}
			if test.config.MaxQoS != opts.Config.MaxQoS {
				opts.Config.MaxQoS = test.config.MaxQoS
			}
			if test.config.RetainAvailable != opts.Config.RetainAvailable {
				opts.Config.RetainAvailable = test.config.RetainAvailable
			}
			if test.config.MaxPacketSize != opts.Config.MaxPacketSize {
				opts.Config.MaxPacketSize = test.config.MaxPacketSize
			}
			if test.config.WildcardSubscriptionAvailable != opts.Config.WildcardSubscriptionAvailable {
				opts.Config.WildcardSubscriptionAvailable = test.config.WildcardSubscriptionAvailable
			}
			if test.config.SubscriptionIDAvailable != opts.Config.SubscriptionIDAvailable {
				opts.Config.SubscriptionIDAvailable = test.config.SubscriptionIDAvailable
			}
			if test.config.SharedSubscriptionAvailable != opts.Config.SharedSubscriptionAvailable {
				opts.Config.SharedSubscriptionAvailable = test.config.SharedSubscriptionAvailable
			}

			srv, _ := akira.NewServer(opts)
			defer srv.Close()

			listener, onConnectionStream := s.addListener(srv)

			conn1, conn2 := net.Pipe()
			defer func() { _ = conn1.Close() }()
			_ = srv.Start(context.Background())

			onConnection := <-onConnectionStream
			onConnection(listener, conn2)

			_, _ = conn1.Write(test.connect)
			connack := make([]byte, len(test.connack))

			_, err := conn1.Read(connack)
			s.Require().NoError(err)
			s.Assert().Equal(test.connack, connack)
		})
	}
}

func (s *ServerTestSuite) TestConnectPacketWithSessionProperties() {
	srv, _ := akira.NewServer(akira.NewDefaultOptions())
	defer srv.Close()

	var props *akira.SessionProperties
	connectCh := make(chan struct{})

	onConnected := mocks.NewMockOnConnectedHook(s.T())
	onConnected.EXPECT().Name().Return("onConnected")
	onConnected.EXPECT().OnConnected(mock.Anything).Run(func(c *akira.Client) {
		props = c.Session.Properties
		close(connectCh)
	})
	_ = srv.AddHook(onConnected)

	listener, onConnectionStream := s.addListener(srv)

	conn1, conn2 := net.Pipe()
	defer func() { _ = conn1.Close() }()
	_ = srv.Start(context.Background())

	onConnection := <-onConnectionStream
	onConnection(listener, conn2)

	connect := []byte{0x10, 51, 0, 4, 'M', 'Q', 'T', 'T', 5, 0, 0, 100, 36,
		17, 0, 0, 0, 100, // Session Expiry Interval
		33, 0, 150, // Receive Maximum
		39, 0, 0, 0, 200, // Maximum Packet Size
		34, 0, 250, // Topic Alias Maximum
		25, 1, // Request Response Info
		23, 1, // Request Problem Info
		38, 0, 1, 'a', 0, 1, 'b', // User Property
		21, 0, 2, 'e', 'f', // Authentication Method
		22, 0, 1, 10, // Authentication Data
		0, 2, 'a', 'b'}
	_, _ = conn1.Write(connect)

	connack := make([]byte, 10)
	_, err := conn1.Read(connack)
	s.Require().NoError(err)

	<-connectCh
	s.Require().NotNil(props)
	s.Require().True(props.Has(packet.PropertySessionExpiryInterval))
	s.Require().True(props.Has(packet.PropertyMaximumPacketSize))
	s.Require().True(props.Has(packet.PropertyReceiveMaximum))
	s.Require().True(props.Has(packet.PropertyTopicAliasMaximum))
	s.Require().True(props.Has(packet.PropertyRequestResponseInfo))
	s.Require().True(props.Has(packet.PropertyRequestProblemInfo))
	s.Require().True(props.Has(packet.PropertyUserProperty))

	s.Assert().Equal(100, int(props.SessionExpiryInterval))
	s.Assert().Equal(150, int(props.ReceiveMaximum))
	s.Assert().Equal(200, int(props.MaximumPacketSize))
	s.Assert().Equal(250, int(props.TopicAliasMaximum))
	s.Assert().True(props.RequestResponseInfo)
	s.Assert().True(props.RequestProblemInfo)
	s.Assert().Equal([]packet.UserProperty{{Key: []byte("a"), Value: []byte("b")}}, props.UserProperties)
}

func (s *ServerTestSuite) TestConnectPacketWithLastWill() {
	srv, _ := akira.NewServer(akira.NewDefaultOptions())
	defer srv.Close()

	var will *akira.LastWill
	connectCh := make(chan struct{})

	onConnected := mocks.NewMockOnConnectedHook(s.T())
	onConnected.EXPECT().Name().Return("onConnected")
	onConnected.EXPECT().OnConnected(mock.Anything).Run(func(c *akira.Client) {
		will = c.Session.LastWill
		close(connectCh)
	})
	_ = srv.AddHook(onConnected)

	listener, onConnectionStream := s.addListener(srv)

	conn1, conn2 := net.Pipe()
	defer func() { _ = conn1.Close() }()
	_ = srv.Start(context.Background())

	onConnection := <-onConnectionStream
	onConnection(listener, conn2)

	connect := []byte{0x10, 57, 0, 4, 'M', 'Q', 'T', 'T', 5, 0x34, 0, 100, 0, 0, 2, 'a', 'b',
		35,              // Property Length
		24, 0, 0, 0, 10, // Will Delay Interval
		1, 1, // Payload Format Indicator
		2, 0, 0, 0, 20, // Message Expiry Interval
		3, 0, 4, 'j', 's', 'o', 'n', // Content Type
		8, 0, 1, 'b', // Response Topic
		9, 0, 2, 20, 1, // Correlation Data
		38, 0, 1, 'a', 0, 1, 'b', // User Property
		0, 1, 'a', // Will Topic
		0, 1, 'b', // Will Payload
	}
	_, _ = conn1.Write(connect)

	connack := make([]byte, 10)
	_, err := conn1.Read(connack)
	s.Require().NoError(err)

	<-connectCh
	s.Require().NotNil(will)

	expected := akira.LastWill{
		Topic:   []byte("a"),
		Payload: []byte("b"),
		QoS:     packet.QoS2,
		Retain:  true,
		Properties: &packet.WillProperties{
			Flags: packet.PropertyFlags(0).
				Set(packet.PropertyWillDelayInterval).
				Set(packet.PropertyPayloadFormatIndicator).
				Set(packet.PropertyMessageExpiryInterval).
				Set(packet.PropertyContentType).
				Set(packet.PropertyResponseTopic).
				Set(packet.PropertyCorrelationData).
				Set(packet.PropertyUserProperty),
			WillDelayInterval:      10,
			PayloadFormatIndicator: true,
			MessageExpiryInterval:  20,
			ContentType:            []byte("json"),
			ResponseTopic:          []byte("b"),
			CorrelationData:        []byte{20, 1},
			UserProperties:         []packet.UserProperty{{[]byte("a"), []byte("b")}},
		},
	}
	s.Assert().Equal(expected, *will)
}

func (s *ServerTestSuite) TestConnectPacketV3WithMaxKeepAliveExceeded() {
	testCases := []struct {
		name    string
		connect []byte
		connack []byte
	}{
		{
			"V3.1",
			[]byte{0x10, 16, 0, 6, 'M', 'Q', 'I', 's', 'd', 'p', 3, 0, 0, 200, 0, 2, 'a', 'b'},
			[]byte{0x20, 2, 0, 2},
		},
		{
			"V3.1.1",
			[]byte{0x10, 14, 0, 4, 'M', 'Q', 'T', 'T', 4, 0, 0, 200, 0, 2, 'a', 'b'},
			[]byte{0x20, 2, 0, 2},
		},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			opts := akira.NewDefaultOptions()
			opts.Config.MaxKeepAlive = 100

			srv, _ := akira.NewServer(opts)
			defer srv.Close()

			listener, onConnectionStream := s.addListener(srv)

			conn1, conn2 := net.Pipe()
			defer func() { _ = conn1.Close() }()
			_ = srv.Start(context.Background())

			onConnection := <-onConnectionStream
			onConnection(listener, conn2)
			_, _ = conn1.Write(test.connect)

			connack := make([]byte, len(test.connack))
			_, err := conn1.Read(connack)
			s.Require().NoError(err)
			s.Assert().Equal(test.connack, connack)

			_, err = conn1.Read(connack)
			s.Require().ErrorIs(err, io.EOF)
		})
	}
}

func (s *ServerTestSuite) TestConnectPacketWithClientIDTooBig() {
	testCases := []struct {
		name    string
		connect []byte
		connack []byte
	}{
		{
			"V3.1",
			[]byte{0x10, 38, 0, 6, 'M', 'Q', 'I', 's', 'd', 'p', 3, 0, 0, 100, 0, 24,
				'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
				'0', '1', '2', '3'},
			[]byte{0x20, 2, 0, 2},
		},
		{
			"V3.1.1",
			[]byte{0x10, 36, 0, 4, 'M', 'Q', 'T', 'T', 4, 0, 0, 100, 0, 24,
				'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
				'0', '1', '2', '3'},
			[]byte{0x20, 2, 0, 2},
		},
		{
			"V5.0",
			[]byte{0x10, 37, 0, 4, 'M', 'Q', 'T', 'T', 5, 0, 0, 100, 0, 0, 24,
				'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
				'0', '1', '2', '3'},
			[]byte{0x20, 3, 0, 0x85, 0},
		},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			opts := akira.NewDefaultOptions()
			opts.Config.MaxClientIDSize = 23
			srv, _ := akira.NewServer(opts)
			defer srv.Close()

			listener, onConnectionStream := s.addListener(srv)

			conn1, conn2 := net.Pipe()
			defer func() { _ = conn1.Close() }()
			_ = srv.Start(context.Background())

			onConnection := <-onConnectionStream
			onConnection(listener, conn2)
			_, _ = conn1.Write(test.connect)

			connack := make([]byte, len(test.connack))
			_, err := conn1.Read(connack)
			s.Require().NoError(err)
			s.Assert().Equal(test.connack, connack)

			_, err = conn1.Read(connack)
			s.Require().ErrorIs(err, io.EOF)
		})
	}
}

func (s *ServerTestSuite) TestConnectPacketRejectedWhenEmptyClientID() {
	testCases := []struct {
		name    string
		connect []byte
		connack []byte
	}{
		{
			"V3.1.1, no clean session",
			[]byte{0x10, 12, 0, 4, 'M', 'Q', 'T', 'T', 4, 0, 0, 100, 0, 0},
			[]byte{0x20, 2, 0, 0x02},
		},
		{
			"V3.1.1, clean session",
			[]byte{0x10, 12, 0, 4, 'M', 'Q', 'T', 'T', 4, 2, 0, 100, 0, 0},
			[]byte{0x20, 2, 0, 0x02},
		},
		{
			"V5.0",
			[]byte{0x10, 13, 0, 4, 'M', 'Q', 'T', 'T', 5, 2, 0, 100, 0, 0, 0},
			[]byte{0x20, 3, 0, 0x85, 0},
		},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			srv, _ := akira.NewServer(akira.NewDefaultOptions())
			defer srv.Close()

			listener, onConnectionStream := s.addListener(srv)

			conn1, conn2 := net.Pipe()
			defer func() { _ = conn1.Close() }()
			_ = srv.Start(context.Background())

			onConnection := <-onConnectionStream
			onConnection(listener, conn2)
			_, _ = conn1.Write(test.connect)

			connack := make([]byte, len(test.connack))
			_, err := conn1.Read(connack)
			s.Require().NoError(err)
			s.Assert().Equal(test.connack, connack)

			_, err = conn1.Read(connack)
			s.Require().ErrorIs(err, io.EOF)
		})
	}
}

func (s *ServerTestSuite) TestConnectPacketWithGetSessionReturningError() {
	testCases := []struct {
		name    string
		connect []byte
		connack []byte
	}{
		{
			"V3.1",
			[]byte{0x10, 16, 0, 6, 'M', 'Q', 'I', 's', 'd', 'p', 3, 0, 0, 100, 0, 2, 'a', 'b'},
			[]byte{0x20, 2, 0, 3},
		},
		{
			"V3.1.1",
			[]byte{0x10, 14, 0, 4, 'M', 'Q', 'T', 'T', 4, 0, 0, 100, 0, 2, 'a', 'b'},
			[]byte{0x20, 2, 0, 3},
		},
		{
			"V5.0",
			[]byte{0x10, 15, 0, 4, 'M', 'Q', 'T', 'T', 5, 0, 0, 100, 0, 0, 2, 'a', 'b'},
			[]byte{0x20, 3, 0, 0x88, 0},
		},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			store := mocks.NewMockSessionStore(s.T())
			store.EXPECT().GetSession([]byte{'a', 'b'}).Return(nil, assert.AnError)

			opts := akira.NewDefaultOptions()
			opts.SessionStore = store
			srv, _ := akira.NewServer(opts)
			defer srv.Close()

			onConnectErr := mocks.NewMockOnConnectErrorHook(s.T())
			onConnectErr.EXPECT().Name().Return("onConnectErr")
			onConnectErr.EXPECT().OnConnectError(mock.Anything, mock.Anything, mock.Anything).
				Run(func(c *akira.Client, p *packet.Connect, err error) {
					s.Require().NotNil(c)
					s.Require().NotNil(p)
					s.Require().Error(err)
					s.Assert().Equal([]byte{'a', 'b'}, p.ClientID)
				})
			_ = srv.AddHook(onConnectErr)

			listener, onConnectionStream := s.addListener(srv)

			conn1, conn2 := net.Pipe()
			defer func() { _ = conn1.Close() }()
			_ = srv.Start(context.Background())

			onConnection := <-onConnectionStream
			onConnection(listener, conn2)
			_, _ = conn1.Write(test.connect)

			connack := make([]byte, len(test.connack))
			_, err := conn1.Read(connack)
			s.Require().NoError(err)
			s.Assert().Equal(test.connack, connack)

			_, err = conn1.Read(connack)
			s.Require().ErrorIs(err, io.EOF)
		})
	}
}

func (s *ServerTestSuite) TestConnectPacketWithDeleteSessionReturningError() {
	testCases := []struct {
		name    string
		connect []byte
		connack []byte
	}{
		{
			"V3.1",
			[]byte{0x10, 16, 0, 6, 'M', 'Q', 'I', 's', 'd', 'p', 3, 2, 0, 100, 0, 2, 'a', 'b'},
			[]byte{0x20, 2, 0, 3},
		},
		{
			"V3.1.1",
			[]byte{0x10, 14, 0, 4, 'M', 'Q', 'T', 'T', 4, 2, 0, 100, 0, 2, 'a', 'b'},
			[]byte{0x20, 2, 0, 3},
		},
		{
			"V5.0",
			[]byte{0x10, 15, 0, 4, 'M', 'Q', 'T', 'T', 5, 2, 0, 100, 0, 0, 2, 'a', 'b'},
			[]byte{0x20, 3, 0, 0x88, 0},
		},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			store := mocks.NewMockSessionStore(s.T())
			store.EXPECT().DeleteSession([]byte{'a', 'b'}).Return(assert.AnError)

			opts := akira.NewDefaultOptions()
			opts.SessionStore = store
			srv, _ := akira.NewServer(opts)
			defer srv.Close()

			onConnectErr := mocks.NewMockOnConnectErrorHook(s.T())
			onConnectErr.EXPECT().Name().Return("onConnectErr")
			onConnectErr.EXPECT().OnConnectError(mock.Anything, mock.Anything, mock.Anything).
				Run(func(c *akira.Client, p *packet.Connect, err error) {
					s.Require().NotNil(c)
					s.Require().NotNil(p)
					s.Require().Error(err)
				})
			_ = srv.AddHook(onConnectErr)

			listener, onConnectionStream := s.addListener(srv)

			conn1, conn2 := net.Pipe()
			defer func() { _ = conn1.Close() }()
			_ = srv.Start(context.Background())

			onConnection := <-onConnectionStream
			onConnection(listener, conn2)
			_, _ = conn1.Write(test.connect)

			connack := make([]byte, len(test.connack))
			_, err := conn1.Read(connack)
			s.Require().NoError(err)
			s.Assert().Equal(test.connack, connack)

			_, err = conn1.Read(connack)
			s.Require().ErrorIs(err, io.EOF)
		})
	}
}

func (s *ServerTestSuite) TestConnectPacketWithSaveSessionReturningError() {
	testCases := []struct {
		name    string
		connect []byte
		connack []byte
	}{
		{
			"V3.1",
			[]byte{0x10, 16, 0, 6, 'M', 'Q', 'I', 's', 'd', 'p', 3, 2, 0, 100, 0, 2, 'a', 'b'},
			[]byte{0x20, 2, 0, 3},
		},
		{
			"V3.1.1",
			[]byte{0x10, 14, 0, 4, 'M', 'Q', 'T', 'T', 4, 2, 0, 100, 0, 2, 'a', 'b'},
			[]byte{0x20, 2, 0, 3},
		},
		{
			"V5.0",
			[]byte{0x10, 15, 0, 4, 'M', 'Q', 'T', 'T', 5, 2, 0, 100, 0, 0, 2, 'a', 'b'},
			[]byte{0x20, 3, 0, 0x88, 0},
		},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			store := mocks.NewMockSessionStore(s.T())
			store.EXPECT().DeleteSession([]byte{'a', 'b'}).Return(nil)
			store.EXPECT().SaveSession(mock.Anything, mock.Anything).Return(assert.AnError)

			opts := akira.NewDefaultOptions()
			opts.SessionStore = store
			srv, _ := akira.NewServer(opts)
			defer srv.Close()

			onConnectErr := mocks.NewMockOnConnectErrorHook(s.T())
			onConnectErr.EXPECT().Name().Return("onConnectErr")
			onConnectErr.EXPECT().OnConnectError(mock.Anything, mock.Anything, mock.Anything).
				Run(func(c *akira.Client, p *packet.Connect, err error) {
					s.Require().NotNil(c)
					s.Require().NotNil(p)
					s.Require().Error(err)
				})
			_ = srv.AddHook(onConnectErr)

			listener, onConnectionStream := s.addListener(srv)

			conn1, conn2 := net.Pipe()
			defer func() { _ = conn1.Close() }()
			_ = srv.Start(context.Background())

			onConnection := <-onConnectionStream
			onConnection(listener, conn2)
			_, _ = conn1.Write(test.connect)

			connack := make([]byte, len(test.connack))
			_, err := conn1.Read(connack)
			s.Require().NoError(err)
			s.Assert().Equal(test.connack, connack)

			_, err = conn1.Read(connack)
			s.Require().ErrorIs(err, io.EOF)
		})
	}
}

func (s *ServerTestSuite) TestConnectPacketWithOnConnectReturningError() {
	srv, _ := akira.NewServer(akira.NewDefaultOptions())
	defer srv.Close()

	onConnect := mocks.NewMockOnConnectHook(s.T())
	onConnect.EXPECT().Name().Return("onConnect")
	onConnect.EXPECT().OnConnect(mock.Anything, mock.Anything).Return(assert.AnError)
	_ = srv.AddHook(onConnect)

	listener, onConnectionStream := s.addListener(srv)

	conn1, conn2 := net.Pipe()
	defer func() { _ = conn1.Close() }()
	_ = srv.Start(context.Background())

	onConnection := <-onConnectionStream
	onConnection(listener, conn2)

	connect := []byte{0x10, 14, 0, 4, 'M', 'Q', 'T', 'T', 4, 2, 0, 100, 0, 2, 'a', 'b'}
	_, _ = conn1.Write(connect)

	connack := make([]byte, 4)
	_, err := conn1.Read(connack)
	s.Require().ErrorIs(err, io.EOF)
}

func (s *ServerTestSuite) TestConnectPacketSendConnAckWhenOnConnectReturnPacketError() {
	srv, _ := akira.NewServer(akira.NewDefaultOptions())
	defer srv.Close()

	onConnect := mocks.NewMockOnConnectHook(s.T())
	onConnect.EXPECT().Name().Return("onConnect")
	onConnect.EXPECT().OnConnect(mock.Anything, mock.Anything).Return(packet.ErrNotAuthorized)
	_ = srv.AddHook(onConnect)

	listener, onConnectionStream := s.addListener(srv)

	conn1, conn2 := net.Pipe()
	defer func() { _ = conn1.Close() }()
	_ = srv.Start(context.Background())

	onConnection := <-onConnectionStream
	onConnection(listener, conn2)

	connect := []byte{0x10, 15, 0, 4, 'M', 'Q', 'T', 'T', 5, 2, 0, 100, 0, 0, 2, 'a', 'b'}
	_, _ = conn1.Write(connect)

	expected := []byte{0x20, 3, 0, 0x87, 0}
	connack := make([]byte, len(expected))

	_, err := conn1.Read(connack)
	s.Require().NoError(err)
	s.Assert().Equal(expected, connack)

	_, err = conn1.Read(connack)
	s.Require().ErrorIs(err, io.EOF)
}

func (s *ServerTestSuite) TestConnectPacketDontSendConnAckWhenOnConnectReturnReasonCodes() {
	testCases := []struct {
		name string
		code packet.ReasonCode
	}{
		{"Success", packet.ReasonCodeSuccess},
		{"Malformed packet", packet.ReasonCodeMalformedPacket},
	}

	for _, test := range testCases {
		s.Run(test.name, func() {
			srv, _ := akira.NewServer(akira.NewDefaultOptions())
			defer srv.Close()

			onConnect := mocks.NewMockOnConnectHook(s.T())
			onConnect.EXPECT().Name().Return("onConnect")
			onConnect.EXPECT().OnConnect(mock.Anything, mock.Anything).Return(packet.Error{Code: test.code})
			_ = srv.AddHook(onConnect)

			listener, onConnectionStream := s.addListener(srv)

			conn1, conn2 := net.Pipe()
			defer func() { _ = conn1.Close() }()
			_ = srv.Start(context.Background())

			onConnection := <-onConnectionStream
			onConnection(listener, conn2)

			connect := []byte{0x10, 15, 0, 4, 'M', 'Q', 'T', 'T', 5, 2, 0, 100, 0, 0, 2, 'a', 'b'}
			_, _ = conn1.Write(connect)

			connack := make([]byte, 4)
			_, err := conn1.Read(connack)
			s.Require().ErrorIs(err, io.EOF)
		})
	}
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}
