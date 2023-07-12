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

func (s *ServerTestSuite) addListener() (akira.Listener, <-chan akira.OnConnectionFunc) {
	onConnectionStream := make(chan akira.OnConnectionFunc, 1)

	listener := mocks.NewMockListener(s.T())
	listener.EXPECT().Name().Return("mock")
	listener.EXPECT().Listen(mock.Anything).RunAndReturn(func(cb akira.OnConnectionFunc) (<-chan bool, error) {
		onConnectionStream <- cb
		close(onConnectionStream)
		done := make(chan bool)
		close(done)
		return done, nil
	})
	listener.EXPECT().Stop()
	err := s.server.AddListener(listener)
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
	listener.EXPECT().Name().Return("mock")

	srv, err := akira.NewServer(&akira.Options{Listeners: []akira.Listener{listener}})
	s.Require().NoError(err)
	s.Assert().Equal(akira.ServerNotStarted, srv.State())
}

func (s *ServerTestSuite) TestNewServerWithDuplicatedListenersReturnsError() {
	listener := mocks.NewMockListener(s.T())
	listener.EXPECT().Name().Return("mock")

	srv, err := akira.NewServer(&akira.Options{Listeners: []akira.Listener{listener, listener}})
	s.Require().Error(err)
	s.Require().Nil(srv)
}

func (s *ServerTestSuite) TestNewServerWithHooks() {
	hook := mocks.NewMockHookOnServerStart(s.T())
	hook.EXPECT().Name().Return("mock")

	srv, err := akira.NewServer(&akira.Options{Hooks: []akira.Hook{hook}})
	s.Require().NoError(err)
	s.Assert().Equal(akira.ServerNotStarted, srv.State())
}

func (s *ServerTestSuite) TestNewServerWithDuplicatedHooksReturnsError() {
	hook := mocks.NewMockHookOnServerStart(s.T())
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
	onServerStart := mocks.NewMockHookOnServerStart(s.T())
	onServerStart.EXPECT().Name().Return("onServerStart")
	onServerStart.EXPECT().OnServerStart(s.server).Return(nil)
	_ = s.server.AddHook(onServerStart)

	onStart := mocks.NewMockHookOnStart(s.T())
	onStart.EXPECT().Name().Return("onStart")
	onStart.EXPECT().OnStart(s.server).Return(nil)
	_ = s.server.AddHook(onStart)

	onServerStarted := mocks.NewMockHookOnServerStarted(s.T())
	onServerStarted.EXPECT().Name().Return("onServerStarted")
	onServerStarted.EXPECT().OnServerStarted(s.server)
	_ = s.server.AddHook(onServerStarted)

	onServerStop := mocks.NewMockHookOnServerStop(s.T())
	onServerStop.EXPECT().Name().Return("onServerStop")
	onServerStop.EXPECT().OnServerStop(s.server)
	_ = s.server.AddHook(onServerStop)

	onStop := mocks.NewMockHookOnStop(s.T())
	onStop.EXPECT().Name().Return("onStop")
	onStop.EXPECT().OnStop(s.server)
	_ = s.server.AddHook(onStop)

	onServerStopped := mocks.NewMockHookOnServerStopped(s.T())
	onServerStopped.EXPECT().Name().Return("onServerStopped")
	onServerStopped.EXPECT().OnServerStopped(s.server)
	_ = s.server.AddHook(onServerStopped)

	err := s.server.Start(context.Background())
	s.Require().NoError(err)
	s.Assert().Equal(akira.ServerRunning, s.server.State())
}

func (s *ServerTestSuite) TestStartWithOnServerStartReturningError() {
	onServerStart := mocks.NewMockHookOnServerStart(s.T())
	onServerStart.EXPECT().Name().Return("onServerStart")
	onServerStart.EXPECT().OnServerStart(s.server).Return(assert.AnError)
	_ = s.server.AddHook(onServerStart)

	onServerStartFailed := mocks.NewMockHookOnServerStartFailed(s.T())
	onServerStartFailed.EXPECT().Name().Return("onServerStartFailed")
	onServerStartFailed.EXPECT().OnServerStartFailed(s.server, assert.AnError)
	_ = s.server.AddHook(onServerStartFailed)

	err := s.server.Start(context.Background())
	s.Require().ErrorIs(err, assert.AnError)
	s.Assert().Equal(akira.ServerFailed, s.server.State())
}

func (s *ServerTestSuite) TestStartWithOnStartReturningError() {
	onStart := mocks.NewMockHookOnStart(s.T())
	onStart.EXPECT().Name().Return("onStart")
	onStart.EXPECT().OnStart(s.server).Return(assert.AnError)
	_ = s.server.AddHook(onStart)

	onServerStartFailed := mocks.NewMockHookOnServerStartFailed(s.T())
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
	listener.EXPECT().Name().Return("new-mock")

	err := s.server.AddListener(listener)
	s.Require().NoError(err)
}

func (s *ServerTestSuite) TestAddListenerWhenServerRunning() {
	_ = s.server.Start(context.Background())
	listener := mocks.NewMockListener(s.T())
	listener.EXPECT().Name().Return("mock")
	listener.EXPECT().Listen(mock.Anything).RunAndReturn(func(_ akira.OnConnectionFunc) (<-chan bool, error) {
		done := make(chan bool)
		close(done)
		return done, nil
	})
	listener.EXPECT().Stop()

	err := s.server.AddListener(listener)
	s.Require().NoError(err)
}

func (s *ServerTestSuite) TestAddDuplicatedListenerReturnsError() {
	listener1 := mocks.NewMockListener(s.T())
	listener1.EXPECT().Name().Return("mock")
	_ = s.server.AddListener(listener1)

	listener2 := mocks.NewMockListener(s.T())
	listener2.EXPECT().Name().Return("mock")

	err := s.server.AddListener(listener2)
	s.Require().ErrorIs(err, akira.ErrListenerAlreadyExists)
}

func (s *ServerTestSuite) TestAddDuplicatedListenerWhenServerRunningStopsListener() {
	listener1 := mocks.NewMockListener(s.T())
	listener1.EXPECT().Name().Return("mock")
	listener1.EXPECT().Listen(mock.Anything).RunAndReturn(func(_ akira.OnConnectionFunc) (<-chan bool, error) {
		doneCh := make(chan bool)
		close(doneCh)
		return doneCh, nil
	})
	listener1.EXPECT().Stop()
	_ = s.server.AddListener(listener1)
	_ = s.server.Start(context.Background())

	listener2 := mocks.NewMockListener(s.T())
	listener2.EXPECT().Listen(mock.Anything).RunAndReturn(func(_ akira.OnConnectionFunc) (<-chan bool, error) {
		doneCh := make(chan bool)
		close(doneCh)
		return doneCh, nil
	})
	listener2.EXPECT().Name().Return("mock")
	listener2.EXPECT().Stop()

	err := s.server.AddListener(listener2)
	s.Require().ErrorIs(err, akira.ErrListenerAlreadyExists)
}

func (s *ServerTestSuite) TestAddListenerWithListenErrorWhenServerRunning() {
	_ = s.server.Start(context.Background())
	listener := mocks.NewMockListener(s.T())
	listener.EXPECT().Listen(mock.Anything).Return(nil, assert.AnError)

	err := s.server.AddListener(listener)
	s.Require().ErrorIs(err, assert.AnError)
}

func (s *ServerTestSuite) TestStartWhenListenerFailsToListenReturnsError() {
	listener := mocks.NewMockListener(s.T())
	listener.EXPECT().Name().Return("mock")
	listener.EXPECT().Listen(mock.Anything).Return(nil, assert.AnError)
	_ = s.server.AddListener(listener)

	err := s.server.Start(context.Background())
	s.Require().ErrorIs(err, assert.AnError)
	s.Assert().Equal(akira.ServerFailed, s.server.State())
}

func (s *ServerTestSuite) TestAddHookCallsOnStartWhenServerRunning() {
	_ = s.server.Start(context.Background())
	hook := mocks.NewMockHookOnStart(s.T())
	hook.EXPECT().Name().Return("hook")
	hook.EXPECT().OnStart(s.server).Return(nil)

	err := s.server.AddHook(hook)
	s.Require().NoError(err)
}

func (s *ServerTestSuite) TestAddHookWithOnStartReturningError() {
	_ = s.server.Start(context.Background())
	hook := mocks.NewMockHookOnStart(s.T())
	hook.EXPECT().OnStart(s.server).Return(assert.AnError)

	err := s.server.AddHook(hook)
	s.Require().ErrorIs(err, assert.AnError)
}

func (s *ServerTestSuite) TestStop() {
	listener := mocks.NewMockListener(s.T())
	listener.EXPECT().Name().Return("mock")
	listener.EXPECT().Listen(mock.Anything).RunAndReturn(func(_ akira.OnConnectionFunc) (<-chan bool, error) {
		done := make(chan bool)
		close(done)
		return done, nil
	})
	listener.EXPECT().Stop()
	_ = s.server.AddListener(listener)
	_ = s.server.Start(context.Background())

	err := s.server.Stop(context.Background())
	s.Require().NoError(err)
	s.Assert().Equal(akira.ServerStopped, s.server.State())
}

func (s *ServerTestSuite) TestStopWithHooks() {
	onServerStop := mocks.NewMockHookOnServerStop(s.T())
	onServerStop.EXPECT().Name().Return("onServerStop")
	onServerStop.EXPECT().OnServerStop(s.server)
	_ = s.server.AddHook(onServerStop)

	onStop := mocks.NewMockHookOnStop(s.T())
	onStop.EXPECT().Name().Return("onStop")
	onStop.EXPECT().OnStop(s.server)
	_ = s.server.AddHook(onStop)

	onServerStopped := mocks.NewMockHookOnServerStopped(s.T())
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
	listener.EXPECT().Name().Return("mock")
	listener.EXPECT().Listen(mock.Anything).RunAndReturn(func(cb akira.OnConnectionFunc) (<-chan bool, error) {
		onConnection = cb
		done := make(chan bool)
		close(done)
		close(listeningCh)
		return done, nil
	})
	listener.EXPECT().Stop()
	_ = s.server.AddListener(listener)

	receivingCh := make(chan struct{})
	onPacketReceive := mocks.NewMockHookOnPacketReceive(s.T())
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
	listener, onConnectionStream := s.addListener()

	onConnOpen := mocks.NewMockHookOnConnectionOpen(s.T())
	onConnOpen.EXPECT().Name().Return("onConnOpen")
	onConnOpen.EXPECT().OnConnectionOpen(s.server, listener).Return(nil)
	_ = s.server.AddHook(onConnOpen)

	connOpenedCh := make(chan struct{})
	onConnOpened := mocks.NewMockHookOnConnectionOpened(s.T())
	onConnOpened.EXPECT().Name().Return("onConnOpened")
	onConnOpened.EXPECT().OnConnectionOpened(s.server, listener).Run(func(_ *akira.Server, _ akira.Listener) {
		close(connOpenedCh)
	})
	_ = s.server.AddHook(onConnOpened)

	onPacketRcv := mocks.NewMockHookOnPacketReceive(s.T())
	onPacketRcv.EXPECT().Name().Return("onPacketRcv")
	onPacketRcv.EXPECT().OnPacketReceive(mock.Anything).Return(nil)
	_ = s.server.AddHook(onPacketRcv)

	onConnClose := mocks.NewMockHookOnConnectionClose(s.T())
	onConnClose.EXPECT().Name().Return("onConnClose")
	onConnClose.EXPECT().OnConnectionClose(s.server, listener, mock.Anything).
		Run(func(_ *akira.Server, _ akira.Listener, err error) { s.Assert().ErrorIs(err, io.EOF) })
	_ = s.server.AddHook(onConnClose)

	connClosedCh := make(chan struct{})
	onConnClosed := mocks.NewMockHookOnConnectionClosed(s.T())
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
	listener, onConnectionStream := s.addListener()

	connClosedCh := make(chan struct{})
	onConnClosed := mocks.NewMockHookOnConnectionClosed(s.T())
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
	listener, onConnectionStream := s.addListener()

	onConnOpen := mocks.NewMockHookOnConnectionOpen(s.T())
	onConnOpen.EXPECT().Name().Return("onConnOpen")
	onConnOpen.EXPECT().OnConnectionOpen(s.server, listener).Return(assert.AnError)
	_ = s.server.AddHook(onConnOpen)

	connClosedCh := make(chan struct{})
	onConnClosed := mocks.NewMockHookOnConnectionClosed(s.T())
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

func (s *ServerTestSuite) TestHandlePacketConnect() {
	var connect *packet.Connect
	listener, onConnectionStream := s.addListener()
	receivedCh := make(chan struct{})

	onPacketReceived := mocks.NewMockHookOnPacketReceived(s.T())
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

	msg := []byte{
		byte(packet.TypeConnect) << 4, 14, // Fixed header
		0, 4, 'M', 'Q', 'T', 'T', // Protocol name
		4,      // Protocol version
		2,      // Packet flags (Clean session)
		0, 255, // Keep alive
		0, 2, 'a', 'b', // Client ID
	}
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
	listener, onConnectionStream := s.addListener()

	connClosedCh := make(chan struct{})
	onConnClosed := mocks.NewMockHookOnConnectionClosed(s.T())
	onConnClosed.EXPECT().Name().Return("onConnClosed")
	onConnClosed.EXPECT().OnConnectionClosed(s.server, listener, mock.Anything).
		Run(func(_ *akira.Server, _ akira.Listener, _ error) { close(connClosedCh) })
	_ = s.server.AddHook(onConnClosed)
	_ = s.server.Start(context.Background())

	conn1, conn2 := net.Pipe()
	defer func() { _ = conn1.Close() }()

	onConnection := <-onConnectionStream
	onConnection(listener, conn2)

	msg := []byte{
		byte(packet.TypeConnect) << 4, 7, // Fixed header
		0, 4, 'M', 'Q', 'T', 'T', // Protocol name
		0, // Protocol version
	}
	_, _ = conn1.Write(msg)
	<-connClosedCh
}

func (s *ServerTestSuite) TestCloseConnectionIfOnPacketReceiveErrorReturnsError() {
	listener, onConnectionStream := s.addListener()

	onPacketRcvError := mocks.NewMockHookOnPacketReceiveError(s.T())
	onPacketRcvError.EXPECT().Name().Return("onPacketRcvError")
	onPacketRcvError.EXPECT().OnPacketReceiveError(mock.Anything, mock.Anything).Return(assert.AnError)
	_ = s.server.AddHook(onPacketRcvError)

	connClosedCh := make(chan struct{})
	onConnClosed := mocks.NewMockHookOnConnectionClosed(s.T())
	onConnClosed.EXPECT().Name().Return("onConnClosed")
	onConnClosed.EXPECT().OnConnectionClosed(s.server, listener, assert.AnError).
		Run(func(_ *akira.Server, _ akira.Listener, _ error) { close(connClosedCh) })
	_ = s.server.AddHook(onConnClosed)
	_ = s.server.Start(context.Background())

	conn1, conn2 := net.Pipe()
	defer func() { _ = conn1.Close() }()

	onConnection := <-onConnectionStream
	onConnection(listener, conn2)

	msg := []byte{
		byte(packet.TypeConnect) << 4, 7, // Fixed header
		0, 4, 'M', 'Q', 'T', 'T', // Protocol name
		0, // Protocol version
	}
	_, _ = conn1.Write(msg)
	<-connClosedCh
}

func (s *ServerTestSuite) TestKeepReceivingWhenOnPacketReceiveErrorDoesNotReturnError() {
	listener, onConnectionStream := s.addListener()

	onPacketRcv := mocks.NewMockHookOnPacketReceive(s.T())
	onPacketRcv.EXPECT().Name().Return("onPacketRcv")
	onPacketRcv.EXPECT().OnPacketReceive(mock.Anything).Return(nil).Twice()
	_ = s.server.AddHook(onPacketRcv)

	receivedCh := make(chan struct{})
	onPacketRcvError := mocks.NewMockHookOnPacketReceiveError(s.T())
	onPacketRcvError.EXPECT().Name().Return("onPacketRcvError")
	onPacketRcvError.EXPECT().OnPacketReceiveError(mock.Anything, mock.Anything).
		RunAndReturn(func(_ *akira.Client, _ error) error {
			close(receivedCh)
			return nil
		})
	_ = s.server.AddHook(onPacketRcvError)

	connClosedCh := make(chan struct{})
	onConnClosed := mocks.NewMockHookOnConnectionClosed(s.T())
	onConnClosed.EXPECT().Name().Return("onConnClosed")
	onConnClosed.EXPECT().OnConnectionClosed(s.server, listener, mock.Anything).
		Run(func(_ *akira.Server, _ akira.Listener, _ error) { close(connClosedCh) })
	_ = s.server.AddHook(onConnClosed)
	_ = s.server.Start(context.Background())

	conn1, conn2 := net.Pipe()

	onConnection := <-onConnectionStream
	onConnection(listener, conn2)

	msg := []byte{
		byte(packet.TypeConnect) << 4, 7, // Fixed header
		0, 4, 'M', 'Q', 'T', 'T', // Protocol name
		0, // Protocol version
	}
	_, _ = conn1.Write(msg)
	<-receivedCh

	_ = conn1.Close()
	<-connClosedCh
}

func (s *ServerTestSuite) TestCloseConnectionWhenOnPacketReceiveReturnsError() {
	listener, onConnectionStream := s.addListener()

	onPacketRcv := mocks.NewMockHookOnPacketReceive(s.T())
	onPacketRcv.EXPECT().Name().Return("onPacketRcv")
	onPacketRcv.EXPECT().OnPacketReceive(mock.Anything).Return(assert.AnError)
	_ = s.server.AddHook(onPacketRcv)

	connClosedCh := make(chan struct{})
	onConnClosed := mocks.NewMockHookOnConnectionClosed(s.T())
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
	listener, onConnectionStream := s.addListener()

	onPacketReceived := mocks.NewMockHookOnPacketReceived(s.T())
	onPacketReceived.EXPECT().Name().Return("onPacketReceived")
	onPacketReceived.EXPECT().OnPacketReceived(mock.Anything, mock.Anything).
		RunAndReturn(func(_ *akira.Client, _ akira.Packet) error { return assert.AnError })
	_ = s.server.AddHook(onPacketReceived)

	connClosedCh := make(chan struct{})
	onConnClosed := mocks.NewMockHookOnConnectionClosed(s.T())
	onConnClosed.EXPECT().Name().Return("onConnClosed")
	onConnClosed.EXPECT().OnConnectionClosed(s.server, listener, assert.AnError).
		Run(func(_ *akira.Server, _ akira.Listener, err error) { close(connClosedCh) })
	_ = s.server.AddHook(onConnClosed)
	_ = s.server.Start(context.Background())

	conn1, conn2 := net.Pipe()
	defer func() { _ = conn1.Close() }()

	onConnection := <-onConnectionStream
	onConnection(listener, conn2)

	msg := []byte{
		byte(packet.TypeConnect) << 4, 14, // Fixed header
		0, 4, 'M', 'Q', 'T', 'T', // Protocol name
		4,      // Protocol version
		2,      // Packet flags (Clean session)
		0, 255, // Keep alive
		0, 2, 'a', 'b', // Client ID
	}
	_, _ = conn1.Write(msg)
	<-connClosedCh
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}
