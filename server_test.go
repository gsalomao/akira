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
	"context"
	"errors"
	"io"
	"net"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServerTestSuite struct {
	suite.Suite
	server   *Server
	listener *listenerMock
	hook     *hookMock
}

func (s *ServerTestSuite) SetupTest() {
	var err error

	s.server, err = NewServer(NewDefaultOptions())
	s.Require().NoError(err)

	s.listener = newListenerMock("mock", ":1883")
	s.hook = nil
	_ = s.server.AddListener(s.listener)
}

func (s *ServerTestSuite) TearDownTest() {
	s.server.Close()
	if s.hook != nil {
		s.hook.AssertExpectations(s.T())
	}
}

func (s *ServerTestSuite) addHook() {
	s.hook = newHookMock()

	err := s.server.AddHook(s.hook)
	s.Require().NoError(err)

	err = s.server.AddHook(&hookSpy{})
	s.Require().NoError(err)
}

func (s *ServerTestSuite) startServer() {
	if s.hook != nil {
		s.hook.On("OnServerStart", s.server)
		s.hook.On("OnStart", s.server)
		s.hook.On("OnServerStarted", s.server)
	}

	err := s.server.Start(context.Background())
	s.Require().NoError(err)
}

func (s *ServerTestSuite) stopServer() {
	if s.hook != nil {
		s.hook.On("OnServerStop", s.server)
		s.hook.On("OnStop", s.server)
		s.hook.On("OnServerStopped", s.server)
	}

	err := s.server.Stop(context.Background())
	s.Require().NoError(err)
}

func (s *ServerTestSuite) TestNewServer() {
	srv, err := NewServer(nil)

	s.Require().NoError(err)
	s.Assert().Equal(ServerNotStarted, srv.State())
}

func (s *ServerTestSuite) TestNewServerDefaultConfig() {
	srv, err := NewServer(&Options{})

	s.Require().NoError(err)
	s.Assert().Equal(NewDefaultConfig(), &srv.config)
}

func (s *ServerTestSuite) TestNewServerWithListeners() {
	l := []Listener{newListenerMock("mock", ":1883")}
	srv, err := NewServer(&Options{Listeners: l})

	s.Require().NoError(err)
	_, ok := srv.listeners.get(l[0].Name())
	s.Require().True(ok)
}

func (s *ServerTestSuite) TestNewServerWithListenersError() {
	_, err := NewServer(&Options{
		Listeners: []Listener{
			newListenerMock("mock", ":1883"),
			newListenerMock("mock", ":1884"),
		},
	})

	s.Require().Error(err)
}

func (s *ServerTestSuite) TestNewServerWithHooks() {
	h := []Hook{newHookMock()}
	srv, err := NewServer(&Options{Hooks: h})

	s.Require().NoError(err)
	_, ok := srv.hooks.hookNames[onStartHook][h[0].Name()]
	s.Require().True(ok)
}

func (s *ServerTestSuite) TestNewServerWithHooksError() {
	_, err := NewServer(&Options{
		Hooks: []Hook{
			newHookMock(),
			newHookMock(),
		},
	})

	s.Require().Error(err)
}

func (s *ServerTestSuite) TestAddListenerSuccess() {
	l := newListenerMock("mock2", ":1883")

	err := s.server.AddListener(l)
	s.Require().NoError(err)

	_, ok := s.server.listeners.get(l.Name())
	s.Require().True(ok)
}

func (s *ServerTestSuite) TestAddListenerError() {
	l := newListenerMock("mock", ":1883")

	err := s.server.AddListener(l)
	s.Require().ErrorIs(err, ErrListenerAlreadyExists)
}

func (s *ServerTestSuite) TestAddListenerServerRunning() {
	_ = s.server.Start(context.Background())
	l := newListenerMock("mock2", ":1883")

	err := s.server.AddListener(l)
	s.Require().NoError(err)
	s.Require().True(l.Listening())
}

func (s *ServerTestSuite) TestAddListenerServerRunningError() {
	_ = s.server.Start(context.Background())
	l := newListenerMock("mock2", "abc")

	err := s.server.AddListener(l)
	s.Require().Error(err)
	s.Require().False(l.Listening())
}

func (s *ServerTestSuite) TestAddHookSuccess() {
	srv, _ := NewServer(NewDefaultOptions())
	hook := newHookMock()

	err := srv.AddHook(hook)
	s.Require().NoError(err)
	hook.AssertExpectations(s.T())
}

func (s *ServerTestSuite) TestAddHookCallsOnStartWhenServerRunning() {
	srv, _ := NewServer(NewDefaultOptions())
	_ = srv.Start(context.Background())
	hook := newHookMock()
	hook.On("OnStart", srv)

	err := srv.AddHook(hook)
	s.Assert().NoError(err)
	hook.AssertExpectations(s.T())
}

func (s *ServerTestSuite) TestAddHookError() {
	s.addHook()
	hook := newHookMock()

	err := s.server.AddHook(hook)
	s.Assert().Error(err)
}

func (s *ServerTestSuite) TestAddHookOnStartError() {
	s.startServer()
	hook := newHookMock()
	hook.On("OnStart", s.server).Return(assert.AnError)

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
	srv, _ := NewServer(NewDefaultOptions())

	err := srv.Start(context.Background())
	s.Require().NoError(err)
	s.Assert().Equal(ServerRunning, srv.State())
}

func (s *ServerTestSuite) TestStartWithHook() {
	s.addHook()
	s.hook.On("OnServerStart", s.server)
	s.hook.On("OnStart", s.server)
	s.hook.On("OnServerStarted", s.server)
	s.hook.On("OnServerStop", s.server)
	s.hook.On("OnStop", s.server)
	s.hook.On("OnServerStopped", s.server)

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
	s.hook.On("OnServerStart", s.server).Return(assert.AnError)
	s.hook.On("OnServerStartFailed", s.server, assert.AnError)

	err := s.server.Start(context.Background())
	s.Require().Error(err)
	s.Assert().Equal(ServerFailed, s.server.State())
}

func (s *ServerTestSuite) TestStartOnStartError() {
	s.addHook()
	s.hook.On("OnServerStart", s.server)
	s.hook.On("OnStart", s.server).Return(assert.AnError)
	s.hook.On("OnServerStartFailed", s.server, assert.AnError)

	err := s.server.Start(context.Background())
	s.Require().Error(err)
	s.Assert().Equal(ServerFailed, s.server.State())
}

func (s *ServerTestSuite) TestStartListenerListenError() {
	s.addHook()
	s.hook.On("OnServerStart", s.server)
	s.hook.On("OnStart", s.server)
	s.hook.On("OnServerStartFailed", s.server, mock.Anything)
	l := newListenerMock("mock2", "abc")
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
	s.hook.On("OnServerStop", s.server)
	s.hook.On("OnStop", s.server)
	s.hook.On("OnServerStopped", s.server)
	s.startServer()

	err := s.server.Stop(context.Background())
	s.Require().NoError(err)
	s.Assert().Equal(ServerStopped, s.server.State())
	s.Assert().False(s.listener.Listening())
}

func (s *ServerTestSuite) TestStopClosesClients() {
	ready := make(chan struct{})
	s.addHook()
	s.hook.On("OnServerStop", s.server)
	s.hook.On("OnStop", s.server)
	s.hook.On("OnServerStopped", s.server)
	s.hook.On("OnConnectionOpen", s.server, s.listener)
	s.hook.On("OnConnectionOpened", s.server, s.listener)
	s.hook.On("OnPacketReceive", s.server, mock.Anything).Run(func(_ mock.Arguments) { close(ready) })
	s.hook.On("OnConnectionClose", s.server, s.listener, nil)
	s.hook.On("OnConnectionClosed", s.server, s.listener, nil)
	s.startServer()

	c1, c2 := net.Pipe()
	defer func() { _ = c1.Close() }()

	s.listener.handle(c2)
	<-ready

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
	s.hook.On("OnStop", s.server)
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

	s.listener.handle(c2)
	_ = c1.Close()
	s.stopServer()
	s.Require().Equal(ServerStopped, s.server.State())
}

func (s *ServerTestSuite) TestHandleConnectionWithHook() {
	connOpened := make(chan struct{})
	connClosed := make(chan struct{})
	s.addHook()
	s.hook.On("OnConnectionOpen", s.server, s.listener)
	s.hook.On("OnConnectionOpened", s.server, s.listener).Run(func(_ mock.Arguments) {
		close(connOpened)
	})
	s.hook.On("OnPacketReceive", s.server, mock.Anything)
	s.hook.On("OnConnectionClose", s.server, s.listener, mock.MatchedBy(func(e error) bool {
		return errors.Is(e, io.EOF)
	}))
	s.hook.On("OnConnectionClosed", s.server, s.listener, mock.MatchedBy(func(e error) bool {
		return errors.Is(e, io.EOF)
	})).Run(func(_ mock.Arguments) {
		close(connClosed)
	})
	s.startServer()
	defer s.stopServer()

	c1, c2 := net.Pipe()

	s.listener.handle(c2)
	<-connOpened

	_ = c1.Close()
	<-connClosed
}

func (s *ServerTestSuite) TestHandleConnectionReadTimeout() {
	connOpened := make(chan struct{})
	connClosed := make(chan struct{})
	s.addHook()
	s.hook.On("OnConnectionOpen", s.server, s.listener)
	s.hook.On("OnConnectionOpened", s.server, s.listener).Run(func(_ mock.Arguments) {
		close(connOpened)
	})
	s.hook.On("OnPacketReceive", s.server, mock.Anything)
	s.hook.On("OnConnectionClose", s.server, s.listener, mock.MatchedBy(func(err error) bool {
		return errors.Is(err, os.ErrDeadlineExceeded)
	}))
	s.hook.On("OnConnectionClosed", s.server, s.listener, mock.MatchedBy(func(err error) bool {
		return errors.Is(err, os.ErrDeadlineExceeded)
	})).Run(func(_ mock.Arguments) {
		close(connClosed)
	})
	s.server.config.ConnectTimeout = 1
	s.startServer()
	defer s.stopServer()

	c1, c2 := net.Pipe()
	defer func() { _ = c1.Close() }()

	s.listener.handle(c2)
	<-connOpened
	<-connClosed
}

func (s *ServerTestSuite) TestOnConnectionOpenedError() {
	connClosed := make(chan struct{})
	s.addHook()
	s.hook.On("OnConnectionOpen", s.server, s.listener).Return(assert.AnError)
	s.hook.On("OnConnectionClose", s.server, s.listener, assert.AnError)
	s.hook.On("OnConnectionClosed", s.server, s.listener, assert.AnError).Run(func(_ mock.Arguments) {
		close(connClosed)
	})
	s.startServer()
	defer s.stopServer()

	c1, c2 := net.Pipe()
	defer func() { _ = c1.Close() }()

	s.listener.handle(c2)
	<-connClosed
}

func (s *ServerTestSuite) TestReceivePacket() {
	msg := []byte{
		byte(PacketTypeConnect) << 4, 14, // Fixed header
		0, 4, 'M', 'Q', 'T', 'T', // Protocol name
		4,      // Protocol version
		2,      // Packet flags (Clean Session)
		0, 255, // Keep alive
		0, 2, 'a', 'b', // Client ID
	}
	c1, c2 := net.Pipe()
	defer func() { _ = c1.Close() }()

	c := newClient(c2, s.server, s.listener)

	var wg sync.WaitGroup
	defer wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = c1.Write(msg)
	}()

	p, err := s.server.receivePacket(c)
	s.Require().NoError(err)
	s.Require().NotNil(p)

	packet := &PacketConnect{
		Version:   MQTT311,
		KeepAlive: 255,
		Flags:     connectFlagCleanSession,
		ClientID:  []byte("ab"),
	}
	s.Assert().Equal(packet, p)
}

func (s *ServerTestSuite) TestReceivePacketWithHooks() {
	msg := []byte{
		byte(PacketTypeConnect) << 4, 14, // Fixed header
		0, 4, 'M', 'Q', 'T', 'T', // Protocol name
		4,      // Protocol version
		2,      // Packet flags (Clean Session)
		0, 255, // Keep alive
		0, 2, 'a', 'b', // Client ID
	}
	c1, c2 := net.Pipe()
	defer func() { _ = c1.Close() }()

	c := newClient(c2, s.server, s.listener)
	packet := &PacketConnect{
		Version:   MQTT311,
		KeepAlive: 255,
		Flags:     connectFlagCleanSession,
		ClientID:  []byte("ab"),
	}
	s.addHook()
	s.hook.On("OnPacketReceive", s.server, c)
	s.hook.On("OnPacketReceived", s.server, c, packet)

	var wg sync.WaitGroup
	defer wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = c1.Write(msg)
	}()

	p, err := s.server.receivePacket(c)
	s.Require().NoError(err)
	s.Require().NotNil(p)
	s.Assert().Equal(packet, p)
}

func (s *ServerTestSuite) TestReceivePacketOnPacketReceiveWithError() {
	s.addHook()
	s.hook.On("OnPacketReceive", s.server, mock.Anything).Return(assert.AnError)
	c1, c2 := net.Pipe()
	defer func() { _ = c1.Close() }()
	c := newClient(c2, s.server, s.listener)

	_, err := s.server.receivePacket(c)
	s.Require().ErrorIs(err, assert.AnError)
}

func (s *ServerTestSuite) TestReceivePacketOnPacketReceiveError() {
	msg := []byte{
		byte(PacketTypeConnect) << 4, 7, // Fixed header
		0, 4, 'M', 'Q', 'T', 'T', // Protocol name
		0, // Protocol version
	}
	c1, c2 := net.Pipe()
	defer func() { _ = c1.Close() }()

	c := newClient(c2, s.server, s.listener)
	s.addHook()
	s.hook.On("OnPacketReceive", s.server, c)
	s.hook.On("OnPacketReceiveError", s.server, c, mock.MatchedBy(func(err error) bool {
		return errors.Is(err, ErrMalformedProtocolVersion)
	})).Return(assert.AnError)

	var wg sync.WaitGroup
	defer wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = c1.Write(msg)
	}()

	p, err := s.server.receivePacket(c)
	s.Require().ErrorIs(err, assert.AnError)
	s.Require().Nil(p)
}

func (s *ServerTestSuite) TestReceivePacketOnPacketReceiveErrorNoError() {
	msg := []byte{
		byte(PacketTypeConnect) << 4, 7, // Fixed header
		0, 4, 'M', 'Q', 'T', 'T', // Protocol name
		0, // Protocol version
	}
	c1, c2 := net.Pipe()
	defer func() { _ = c1.Close() }()

	c := newClient(c2, s.server, s.listener)
	var wg sync.WaitGroup
	defer wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = c1.Write(msg)
	}()

	p, err := s.server.receivePacket(c)
	s.Require().NoError(err)
	s.Require().Nil(p)
}

func (s *ServerTestSuite) TestReceivePacketOnPacketReceivedWithError() {
	msg := []byte{
		byte(PacketTypeConnect) << 4, 14, // Fixed header
		0, 4, 'M', 'Q', 'T', 'T', // Protocol name
		4,      // Protocol version
		2,      // Packet flags (Clean Session)
		0, 255, // Keep alive
		0, 2, 'a', 'b', // Client ID
	}
	c1, c2 := net.Pipe()
	defer func() { _ = c1.Close() }()

	c := newClient(c2, s.server, s.listener)
	packet := &PacketConnect{
		Version:   MQTT311,
		KeepAlive: 255,
		Flags:     connectFlagCleanSession,
		ClientID:  []byte("ab"),
	}
	s.addHook()
	s.hook.On("OnPacketReceive", s.server, c)
	s.hook.On("OnPacketReceived", s.server, c, packet).Return(assert.AnError)

	var wg sync.WaitGroup
	defer wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = c1.Write(msg)
	}()

	p, err := s.server.receivePacket(c)
	s.Require().ErrorIs(err, assert.AnError)
	s.Require().Nil(p)
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}

func BenchmarkReceivePacket(b *testing.B) {
	testCases := []struct {
		name string
		data []byte
	}{
		{
			name: "CONNECT-V3",
			data: []byte{
				byte(PacketTypeConnect) << 4, 14, // Fixed header
				0, 4, 'M', 'Q', 'T', 'T', // Protocol name
				4,      // Protocol version
				2,      // Packet flags (Clean Session)
				0, 255, // Keep alive
				0, 2, 'a', 'b', // Client ID
			},
		},
		{
			name: "CONNECT-V5",
			data: []byte{
				byte(PacketTypeConnect) << 4, 15, // Fixed header
				0, 4, 'M', 'Q', 'T', 'T', // Protocol name
				5,      // Protocol version
				2,      // Packet flags (Clean Session)
				0, 255, // Keep alive
				0,              // Property length
				0, 2, 'a', 'b', // Client ID
			},
		},
	}

	for _, test := range testCases {
		b.Run(test.name, func(b *testing.B) {
			b.Run("No Hooks", func(b *testing.B) {
				server, err := NewServer(NewDefaultOptions())
				if err != nil {
					b.Fatal(err)
				}
				defer server.Close()

				benchmarkServerReceivePacket(b, server, test.data)
			})

			b.Run("With Hooks", func(b *testing.B) {
				server, err := NewServer(NewDefaultOptions())
				if err != nil {
					b.Fatal(err)
				}
				defer server.Close()

				err = server.AddHook(&hookSpy{})
				if err != nil {
					b.Fatal(err)
				}

				benchmarkServerReceivePacket(b, server, test.data)
			})
		})
	}
}

func benchmarkServerReceivePacket(b *testing.B, server *Server, data []byte) {
	b.Helper()

	l := newListenerMock("mock", ":1883")
	err := server.AddListener(l)
	if err != nil {
		b.Fatal(err)
	}

	var sConn net.Conn
	var sErr error

	ctx, cancel := context.WithCancel(context.Background())
	lsn, _ := net.Listen("tcp", "127.0.0.1:1883")
	defer func() { _ = lsn.Close() }()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		cancel()
		sConn, sErr = lsn.Accept()
	}()

	<-ctx.Done()
	cConn, cErr := net.Dial("tcp", ":1883")
	if cErr != nil {
		b.Fatal(cErr)
	}

	wg.Wait()
	if sErr != nil {
		b.Fatal(sErr)
	}
	defer func() { _ = sConn.Close() }()

	c := newClient(sConn, server, l)

	ctx, cancel = context.WithCancel(context.Background())

	wg.Add(1)
	go func() {
		defer wg.Done()
		cancel()

		for {
			_, sErr = server.receivePacket(c)
			if sErr != nil {
				break
			}
		}
	}()

	<-ctx.Done()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, cErr = cConn.Write(data)
		if cErr != nil {
			b.Fatal(cErr)
		}
	}

	_ = cConn.Close()
	wg.Wait()

	if sErr != nil && !errors.Is(sErr, io.EOF) {
		b.Fatal(sErr)
	}
}
