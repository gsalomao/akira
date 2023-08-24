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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/gsalomao/akira/packet"
	"github.com/gsalomao/akira/testdata"
)

func TestServerStateString(t *testing.T) {
	testCases := []struct {
		state ServerState
		str   string
	}{
		{ServerNotStarted, "Not Started"},
		{ServerStarting, "Starting"},
		{ServerRunning, "Running"},
		{ServerStopping, "Stopping"},
		{ServerStopped, "Stopped"},
		{ServerFailed, "Failed"},
		{ServerClosed, "Closed"},
		{ServerState(100), "Invalid"},
	}

	for _, tc := range testCases {
		t.Run(tc.str, func(t *testing.T) {
			str := tc.state.String()
			if str != tc.str {
				t.Errorf("Unexpected string\nwant: %s\ngot:  %s", tc.str, str)
			}
		})
	}
}

func TestNewServer(t *testing.T) {
	s, err := NewServer()
	if err != nil {
		t.Fatalf("Unexpected error\n%v", err)
	}
	if s == nil {
		t.Fatal("A server was expected")
	}
	defer s.Close()

	if s.State() != ServerNotStarted {
		t.Errorf("Unexpected state\nwant: %v\ngot:  %v", ServerNotStarted, s.State())
	}
	if !reflect.DeepEqual(NewDefaultConfig(), &s.config) {
		t.Errorf("Unexpected config\nwant: %+v\ngot:  %+v", NewDefaultConfig(), &s.config)
	}
}

func TestNewServerWithConfig(t *testing.T) {
	c := &Config{}

	s, err := NewServer(WithConfig(c))
	if err != nil {
		t.Fatalf("Unexpected error\n%v", err)
	}
	if s == nil {
		t.Fatal("A server was expected")
	}
	defer s.Close()

	if !reflect.DeepEqual(c, &s.config) {
		t.Errorf("Unexpected config\nwant: %+v\ngot:  %+v", c, &s.config)
	}
}

func TestNewServerWithListeners(t *testing.T) {
	l := &mockListener{}

	s, err := NewServer(WithListeners([]Listener{l}))
	if err != nil {
		t.Fatalf("Unexpected error\n%v", err)
	}
	if s == nil {
		t.Fatal("A server was expected")
	}
	defer s.Close()

	if l.listenCalls() != 0 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 0, l.listenCalls())
	}
	if l.closeCalls() != 0 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 0, l.closeCalls())
	}
}

func TestNewServerWithHooks(t *testing.T) {
	h := &mockOnStartHook{}

	s, err := NewServer(WithHooks([]Hook{h}))
	if err != nil {
		t.Fatalf("Unexpected error\n%v", err)
	}
	if s == nil {
		t.Fatal("A server was expected")
	}
	defer s.Close()

	if h.calls() != 0 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 0, h.calls())
	}
}

func TestNewServerWithHooksErrorDuplicatedHook(t *testing.T) {
	h := &mockOnStartHook{}

	s, err := NewServer(WithHooks([]Hook{h, h}))
	if err == nil {
		t.Error("An error was expected")
	}
	if s != nil {
		t.Error("No server should be created")
	}
}

func TestNewServerWithEnhancedAuths(t *testing.T) {
	e := &mockEnhancedAuth{}

	s, err := NewServer(WithEnhancedAuths([]EnhancedAuth{e}))
	if err != nil {
		t.Fatalf("Unexpected error\n%v", err)
	}
	if s == nil {
		t.Fatal("A server was expected")
	}
	s.Close()
}

func TestNewServerWithEnhancedAuthsErrorDuplicated(t *testing.T) {
	e := &mockEnhancedAuth{}

	s, err := NewServer(WithEnhancedAuths([]EnhancedAuth{e, e}))
	if err == nil {
		t.Error("An error was expected")
	}
	if s != nil {
		t.Error("No server should be created")
	}
}

func TestNewServerWithLogger(t *testing.T) {
	l := &noOpLogger{}

	s, err := NewServer(WithLogger(l))
	if err != nil {
		t.Fatalf("Unexpected error\n%v", err)
	}
	if s == nil {
		t.Fatal("A server was expected")
	}
	defer s.Close()

	if s.logger != l {
		t.Errorf("Unexpected log\nwant: %+v\ngot:  %+v", l, s.logger)
	}
}

func TestNewServerWithOptions(t *testing.T) {
	s, err := NewServerWithOptions(NewDefaultOptions())
	if err != nil {
		t.Fatalf("Unexpected error\n%v", err)
	}
	if s == nil {
		t.Fatal("A server was expected")
	}
	defer s.Close()

	if s.State() != ServerNotStarted {
		t.Errorf("Unexpected state\nwant: %v\ngot:  %v", ServerNotStarted, s.State())
	}
	if !reflect.DeepEqual(NewDefaultConfig(), &s.config) {
		t.Errorf("Unexpected config\nwant: %+v\ngot:  %+v", NewDefaultConfig(), &s.config)
	}
}

func TestNewServerWithOptionsNilIsTheSameAsDefaultConfig(t *testing.T) {
	s, err := NewServerWithOptions(nil)
	if err != nil {
		t.Fatalf("Unexpected error\n%v", err)
	}
	if s == nil {
		t.Fatal("A server was expected")
	}
	defer s.Close()

	if s.State() != ServerNotStarted {
		t.Errorf("Unexpected state\nwant: %v\ngot:  %v", ServerNotStarted, s.State())
	}
	if !reflect.DeepEqual(NewDefaultConfig(), &s.config) {
		t.Errorf("Unexpected config\nwant: %+v\ngot:  %+v", NewDefaultConfig(), &s.config)
	}
}

func TestNewServerWithOptionsWithoutConfig(t *testing.T) {
	s, err := NewServerWithOptions(&Options{})
	if err != nil {
		t.Fatalf("Unexpected error\n%v", err)
	}
	if s == nil {
		t.Fatal("A server was expected")
	}
	defer s.Close()

	if !reflect.DeepEqual(NewDefaultConfig(), &s.config) {
		t.Errorf("Unexpected config\nwant: %+v\ngot:  %+v", NewDefaultConfig(), &s.config)
	}
}

func TestServerAddHook(t *testing.T) {
	h := &mockOnStartHook{}
	s := newServer(t)
	defer s.Close()

	err := s.AddHook(h)
	if err != nil {
		t.Errorf("Unexpected error\n%v", err)
	}
	if h.calls() != 0 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 0, h.calls())
	}
}

func TestServerAddEnhancedAuth(t *testing.T) {
	e := &mockEnhancedAuth{}
	s := newServer(t)
	defer s.Close()

	err := s.AddEnhancedAuth(e)
	if err != nil {
		t.Errorf("Unexpected error\n%v", err)
	}
}

func TestServerStart(t *testing.T) {
	s := newServer(t)
	defer s.Close()

	err := s.Start()
	if err != nil {
		t.Errorf("Unexpected error\n%v", err)
	}
	if s.State() != ServerRunning {
		t.Errorf("Unexpected state\nwant: %v\ngot:  %v", ServerRunning, s.State())
	}
}

func TestServerStartWithHooks(t *testing.T) {
	var (
		onStart         = &mockOnStartHook{}
		onServerStart   = &mockOnServerStartHook{}
		onServerStarted = &mockOnServerStartedHook{}
	)

	s := newServer(t, WithHooks([]Hook{onStart, onServerStart, onServerStarted}))
	defer s.Close()

	err := s.Start()
	if err != nil {
		t.Errorf("Unexpected error\n%v", err)
	}
	if onStart.calls() != 1 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, onStart.calls())
	}
	if onServerStart.calls() != 1 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, onServerStart.calls())
	}
	if onServerStarted.calls() != 1 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, onServerStarted.calls())
	}
}

func TestServerStartFailsWhenOnServerStartReturnsError(t *testing.T) {
	startErr := make(chan error, 1)
	onServerStart := &mockOnServerStartHook{cb: func() error { return errHookFailed }}
	onServerStartFailed := &mockOnServerStartFailedHook{cb: func(err error) { startErr <- err }}

	s := newServer(t, WithHooks([]Hook{onServerStart, onServerStartFailed}))
	defer s.Close()

	err := s.Start()
	if !errors.Is(err, errHookFailed) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", errHookFailed, err)
	}
	if s.State() != ServerFailed {
		t.Errorf("Unexpected state\nwant: %v\ngot:  %v", ServerFailed, s.State())
	}

	err = <-startErr
	if !errors.Is(err, errHookFailed) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", errHookFailed, err)
	}
}

func TestServerStartFailsWhenOnStartReturnsError(t *testing.T) {
	startErr := make(chan error, 1)
	onStart := &mockOnStartHook{cb: func() error { return errHookFailed }}
	onServerStartFailed := &mockOnServerStartFailedHook{cb: func(err error) { startErr <- err }}

	s := newServer(t, WithHooks([]Hook{onStart, onServerStartFailed}))
	defer s.Close()

	err := s.Start()
	if !errors.Is(err, errHookFailed) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", errHookFailed, err)
	}
	if s.State() != ServerFailed {
		t.Errorf("Unexpected state\nwant: %v\ngot:  %v", ServerFailed, s.State())
	}

	err = <-startErr
	if !errors.Is(err, errHookFailed) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", errHookFailed, err)
	}
}

func TestServerStartFailsWhenServerClosed(t *testing.T) {
	s := newServer(t)
	s.Close()

	err := s.Start()
	if !errors.Is(err, ErrInvalidServerState) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", ErrInvalidServerState, err)
	}
	if s.State() != ServerClosed {
		t.Errorf("Unexpected state\nwant: %v\ngot:  %v", ServerClosed, s.State())
	}
}

func TestServerAddListener(t *testing.T) {
	s := newServer(t)
	defer s.Close()

	l := &mockListener{}
	err := s.AddListener(l)
	if err != nil {
		t.Errorf("Unexpected error\n%v", err)
	}
	if l.listenCalls() != 0 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 0, l.listenCalls())
	}
}

func TestServerAddListenerWhenServerRunning(t *testing.T) {
	s := newServer(t)
	defer s.Close()
	startServer(t, s)

	l := &mockListener{}
	err := s.AddListener(l)
	if err != nil {
		t.Errorf("Unexpected error\n%v", err)
	}
	if l.listenCalls() != 1 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, l.listenCalls())
	}
}

func TestServerAddListenerFailsWhenListenReturnsError(t *testing.T) {
	s := newServer(t)
	defer s.Close()
	startServer(t, s)

	l := &mockListener{listenCB: func(_ Handler) error { return errListenerFailed }}
	err := s.AddListener(l)
	if !errors.Is(err, errListenerFailed) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", errListenerFailed, err)
	}
}

func TestServerStartFailsWhenListenReturnsError(t *testing.T) {
	startErr := make(chan error, 1)
	l := &mockListener{listenCB: func(_ Handler) error { return errListenerFailed }}
	h := &mockOnServerStartFailedHook{cb: func(err error) { startErr <- err }}

	s := newServer(t, WithListeners([]Listener{l}), WithHooks([]Hook{h}))
	defer s.Close()

	err := s.Start()
	if !errors.Is(err, errListenerFailed) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", errListenerFailed, err)
	}
	if s.State() != ServerFailed {
		t.Errorf("Unexpected state\nwant: %v\ngot:  %v", ServerFailed, s.State())
	}
	if h.calls() != 1 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, h.calls())
	}

	err = <-startErr
	if !errors.Is(err, errListenerFailed) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", errListenerFailed, err)
	}
}

func TestServerAddHookStartsHookWhenServerRunning(t *testing.T) {
	s := newServer(t)
	defer s.Close()
	startServer(t, s)

	h := &mockOnStartHook{}
	err := s.AddHook(h)
	if err != nil {
		t.Errorf("Unexpected error\n%v", err)
	}
	if h.calls() != 1 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, h.calls())
	}
}

func TestServerAddHookFailsWhenOnStartReturnsError(t *testing.T) {
	s := newServer(t)
	defer s.Close()
	startServer(t, s)

	h := &mockOnStartHook{cb: func() error { return errHookFailed }}
	err := s.AddHook(h)
	if !errors.Is(err, errHookFailed) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", errHookFailed, err)
	}
}

func TestServerAddHookFailsWhenAlreadyExists(t *testing.T) {
	h := &mockOnStartStopHook{}
	s := newServer(t, WithHooks([]Hook{h}))
	defer s.Close()
	startServer(t, s)

	stopped := make(chan struct{}, 1)
	h.stopCB = func() { stopped <- struct{}{} }
	h._calls.Store(0)

	err := s.AddHook(h)
	if !errors.Is(err, ErrHookAlreadyExists) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", ErrHookAlreadyExists, err)
	}
	<-stopped
	if h.calls() != 2 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 2, h.calls())
	}
}

func TestServerAddEnhancedAuthFailsWhenAlreadyExists(t *testing.T) {
	e := &mockEnhancedAuth{}
	s := newServer(t, WithEnhancedAuths([]EnhancedAuth{e}))
	defer s.Close()

	err := s.AddEnhancedAuth(e)
	if !errors.Is(err, ErrEnhancedAuthAlreadyExists) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", ErrEnhancedAuthAlreadyExists, err)
	}
}

func TestServerStop(t *testing.T) {
	s := newServer(t)
	defer s.Close()
	startServer(t, s)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := s.Stop(ctx)
	if err != nil {
		t.Errorf("Unexpected error\n%v", err)
	}
	if s.State() != ServerStopped {
		t.Errorf("Unexpected state\nwant: %v\ngot:  %v", ServerStopped, s.State())
	}
}

func TestServerStopWithHooks(t *testing.T) {
	var (
		onStop          = &mockOnStopHook{}
		onServerStop    = &mockOnServerStopHook{}
		onServerStopped = &mockOnServerStoppedHook{}
	)

	s := newServer(t, WithHooks([]Hook{onStop, onServerStop, onServerStopped}))
	defer s.Close()
	startServer(t, s)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := s.Stop(ctx)
	if err != nil {
		t.Errorf("Unexpected error\n%v", err)
	}
	if onStop.calls() != 1 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, onStop.calls())
	}
	if onServerStop.calls() != 1 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, onServerStop.calls())
	}
	if onServerStopped.calls() != 1 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, onServerStopped.calls())
	}
}

func TestServerStopWhenServerNotStarted(t *testing.T) {
	s := newServer(t)
	defer s.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := s.Stop(ctx)
	if err != nil {
		t.Errorf("Unexpected error\n%v", err)
	}
}

func TestServerStopFailsWhenCancelled(t *testing.T) {
	s := newServer(t)
	defer s.Close()
	startServer(t, s)

	// To prevent the server stops before the stop method evaluates the cancelled context.
	s.wg.Add(1)
	defer s.wg.Done()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := s.Stop(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", context.Canceled, err)
	}
	if s.State() != ServerStopping {
		t.Errorf("Unexpected state\nwant: %v\ngot:  %v", ServerStopping, s.State())
	}
}

func TestServerCloseWhenServerNotStarted(t *testing.T) {
	s := newServer(t)

	s.Close()
	if s.State() != ServerClosed {
		t.Errorf("Unexpected state\nwant: %v\ngot:  %v", ServerClosed, s.State())
	}
}

func TestServerCloseWhenServerRunning(t *testing.T) {
	s := newServer(t)
	startServer(t, s)

	s.Close()
	if s.State() != ServerClosed {
		t.Errorf("Unexpected state\nwant: %v\ngot:  %v", ServerClosed, s.State())
	}
}

func TestServerCloseWhenServerStopping(t *testing.T) {
	s := newServer(t)
	startServer(t, s)

	// To prevent the server stops before the stop method evaluates the cancelled context.
	s.wg.Add(1)
	defer s.wg.Done()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_ = s.Stop(ctx)
	if s.State() != ServerStopping {
		t.Errorf("Unexpected state\nwant: %v\ngot:  %v", ServerStopping, s.State())
	}

	s.Close()
	if s.State() != ServerClosed {
		t.Errorf("Unexpected state\nwant: %v\ngot:  %v", ServerClosed, s.State())
	}
}

func TestServerCloseWhenServerStopped(t *testing.T) {
	s := newServer(t)
	startServer(t, s)
	stopServer(t, s)

	s.Close()
	if s.State() != ServerClosed {
		t.Errorf("Unexpected state\nwant: %v\ngot:  %v", ServerClosed, s.State())
	}
}

func TestServerCloseWhenServerClosed(t *testing.T) {
	s := newServer(t)
	s.Close()

	s.Close()
	if s.State() != ServerClosed {
		t.Errorf("Unexpected state\nwant: %v\ngot:  %v", ServerClosed, s.State())
	}
}

func TestServerServe(t *testing.T) {
	s := newServer(t)
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	err := s.Serve(conn)
	if err != nil {
		t.Errorf("Unexpected error\n%v", err)
	}
}

func TestServerServeWithHooks(t *testing.T) {
	var (
		onConnectionOpen = &mockOnConnectionOpenHook{}
		onClientOpened   = &mockOnClientOpenedHook{}
		onReceivePacket  = &mockOnReceivePacketHook{}
	)

	s := newServer(t, WithHooks([]Hook{onConnectionOpen, onClientOpened, onReceivePacket}))
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	ready := make(chan struct{})
	onReceivePacket.cb = func(_ *Client) error {
		close(ready)
		return nil
	}

	err := s.Serve(conn)
	if err != nil {
		t.Errorf("Unexpected error\n%v", err)
	}
	if onConnectionOpen.calls() != 1 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, onConnectionOpen.calls())
	}
	if onClientOpened.calls() != 1 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, onConnectionOpen.calls())
	}

	<-ready
	if onReceivePacket.calls() != 1 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, onReceivePacket.calls())
	}
}

func TestServerServeFailsWhenServerNotRunning(t *testing.T) {
	s := newServer(t)
	defer s.Close()

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	err := s.Serve(conn)
	if !errors.Is(err, ErrServerNotRunning) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", ErrServerNotRunning, err)
	}
}

func TestServerServeFailsWhenInvalidConnection(t *testing.T) {
	s := newServer(t)
	defer s.Close()
	startServer(t, s)

	err := s.Serve(nil)
	if !errors.Is(err, ErrInvalidConnection) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", ErrInvalidConnection, err)
	}

	err = s.Serve(&Connection{})
	if !errors.Is(err, ErrInvalidConnection) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", ErrInvalidConnection, err)
	}

	err = s.Serve(&Connection{Listener: &mockListener{}})
	if !errors.Is(err, ErrInvalidConnection) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", ErrInvalidConnection, err)
	}
}

func TestServerServeFailsWhenOnConnectionOpenReturnsError(t *testing.T) {
	var (
		onConnectionOpen   = &mockOnConnectionOpenHook{cb: func(_ *Connection) error { return errHookFailed }}
		onConnectionClosed = &mockOnConnectionClosedHook{}
	)

	s := newServer(t, WithHooks([]Hook{onConnectionOpen, onConnectionClosed}))
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	err := s.Serve(conn)
	if !errors.Is(err, errHookFailed) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", errHookFailed, err)
	}

	_, err = nc.Read(nil)
	if !errors.Is(err, io.EOF) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
	}
	if onConnectionClosed.calls() != 1 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, onConnectionClosed.calls())
	}
}

func TestServerConnectionClosesWhenConnectTimeout(t *testing.T) {
	c := NewDefaultConfig()
	c.ConnectTimeoutMs = 10

	s := newServer(t, WithConfig(c))
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()
	serveConnection(t, s, conn)

	_, err := nc.Read(nil)
	if !errors.Is(err, io.EOF) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
	}
}

func TestServerCallsHooksWhenConnectionCloses(t *testing.T) {
	var (
		onClientClose      = &mockOnClientCloseHook{}
		onConnectionClosed = &mockOnConnectionClosedHook{}
	)

	c := NewDefaultConfig()
	c.ConnectTimeoutMs = 10

	s := newServer(t, WithConfig(c), WithHooks([]Hook{onClientClose, onConnectionClosed}))
	defer s.Close()
	startServer(t, s)

	closedErr := make(chan error)
	onConnectionClosed.cb = func(_ *Connection, err error) { closedErr <- err }

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()
	serveConnection(t, s, conn)

	err := <-closedErr
	if !errors.Is(err, os.ErrDeadlineExceeded) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", os.ErrDeadlineExceeded, err)
	}
	if onClientClose.calls() != 1 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, onClientClose.calls())
	}
	if onConnectionClosed.calls() != 1 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, onConnectionClosed.calls())
	}
}

func TestServerStopClosesAllClients(t *testing.T) {
	s := newServer(t)
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()
	serveConnection(t, s, conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := s.Stop(ctx)
	if err != nil {
		t.Errorf("Unexpected error\n%v", err)
	}

	_, err = nc.Read(nil)
	if !errors.Is(err, io.EOF) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
	}
}

func TestServerClosesConnectionWhenReceivesInvalidPacket(t *testing.T) {
	s := newServer(t)
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()
	serveConnection(t, s, conn)

	msg := []byte{0x10, 7, 0, 4, 'M', 'Q', 'T', 'T', 0}
	sendPacket(t, nc, msg)

	_, err := nc.Read(nil)
	if !errors.Is(err, io.EOF) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
	}
}

func TestServerClosesConnectionWhenMaxPacketSizeExceeded(t *testing.T) {
	fixture := readPacketFixture(t, "connect.json", "V5.0")
	onPacketReceive := &mockOnPacketReceiveHook{}
	onPacketReceiveFailed := &mockOnPacketReceiveFailedHook{}

	c := NewDefaultConfig()
	c.MaxPacketSize = 1

	s := newServer(t, WithConfig(c), WithHooks([]Hook{onPacketReceive, onPacketReceiveFailed}))
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	receiveErr := make(chan error)
	onPacketReceiveFailed.cb = func(_ *Client, err error) { receiveErr <- err }

	serveConnection(t, s, conn)
	sendPacket(t, nc, fixture.Packet)

	err := <-receiveErr
	if !errors.Is(err, ErrPacketSizeExceeded) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", ErrPacketSizeExceeded, err)
	}
	if onPacketReceive.calls() != 0 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 0, onPacketReceive.calls())
	}
	if onPacketReceiveFailed.calls() != 1 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, onPacketReceiveFailed.calls())
	}

	_, err = nc.Read(nil)
	if !errors.Is(err, io.EOF) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
	}
}

func TestServerClosesConnectionWhenOnReceivePacketReturnsError(t *testing.T) {
	h := &mockOnReceivePacketHook{cb: func(_ *Client) error { return errHookFailed }}

	s := newServer(t, WithHooks([]Hook{h}))
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()
	serveConnection(t, s, conn)

	_, err := nc.Read(nil)
	if !errors.Is(err, io.EOF) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
	}
	if h.calls() != 1 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, h.calls())
	}
}

func TestServerClosesConnectionWhenOnPacketReceiveReturnsError(t *testing.T) {
	h := &mockOnPacketReceiveHook{cb: func(_ *Client, _ packet.FixedHeader) error { return errHookFailed }}
	fixture := readPacketFixture(t, "connect.json", "V5.0")

	s := newServer(t, WithHooks([]Hook{h}))
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	serveConnection(t, s, conn)
	sendPacket(t, nc, fixture.Packet)

	_, err := nc.Read(nil)
	if !errors.Is(err, io.EOF) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
	}
	if h.calls() != 1 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, h.calls())
	}
}

func TestServerClosesConnectionWhenOnPacketReceivedReturnsError(t *testing.T) {
	h := &mockOnPacketReceivedHook{cb: func(_ *Client, _ Packet) error { return errHookFailed }}
	fixture := readPacketFixture(t, "connect.json", "V5.0")

	s := newServer(t, WithHooks([]Hook{h}))
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	serveConnection(t, s, conn)
	sendPacket(t, nc, fixture.Packet)

	_, err := nc.Read(nil)
	if !errors.Is(err, io.EOF) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
	}
	if h.calls() != 1 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, h.calls())
	}
}

func TestServerClosesConnectionWhenOnPacketSendReturnsError(t *testing.T) {
	h := &mockOnPacketSendHook{cb: func(_ *Client, _ Packet) error { return errHookFailed }}
	fixture := readPacketFixture(t, "connect.json", "V5.0")

	s := newServer(t, WithHooks([]Hook{h}))
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	serveConnection(t, s, conn)
	sendPacket(t, nc, fixture.Packet)

	_, err := nc.Read(nil)
	if !errors.Is(err, io.EOF) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
	}
	if h.calls() != 1 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, h.calls())
	}
}

func TestServerCallsHookWhenReceiveInvalidPacket(t *testing.T) {
	var (
		onReceivePacket       = &mockOnReceivePacketHook{}
		onPacketReceiveFailed = &mockOnPacketReceiveFailedHook{}
	)

	s := newServer(t, WithHooks([]Hook{onReceivePacket, onPacketReceiveFailed}))
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()
	serveConnection(t, s, conn)

	receiveErr := make(chan error)
	onPacketReceiveFailed.cb = func(_ *Client, err error) { receiveErr <- err }

	msg := []byte{0x10, 7, 0, 4, 'M', 'Q', 'T', 'T', 0}
	sendPacket(t, nc, msg)

	err := <-receiveErr
	if !errors.Is(err, packet.ErrMalformedPacket) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", packet.ErrMalformedPacket, err)
	}
	if onReceivePacket.calls() != 1 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, onReceivePacket.calls())
	}
	if onPacketReceiveFailed.calls() != 1 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, onPacketReceiveFailed.calls())
	}
}

func TestServerCallsOnPacketSendErrorWhenFailsToSendPacket(t *testing.T) {
	var (
		onPacketReceived  = &mockOnPacketReceivedHook{}
		onPacketSendError = &mockOnPacketSendFailedHook{}
	)

	fixture := readPacketFixture(t, "connect.json", "V5.0")

	s := newServer(t, WithHooks([]Hook{onPacketReceived, onPacketSendError}))
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)

	ready := make(chan struct{})
	onPacketReceived.cb = func(_ *Client, _ Packet) error {
		<-ready
		return nil
	}

	sendErr := make(chan error)
	onPacketSendError.cb = func(_ *Client, _ Packet, err error) { sendErr <- err }

	serveConnection(t, s, conn)
	sendPacket(t, nc, fixture.Packet)

	_ = nc.Close()
	close(ready)

	err := <-sendErr
	if !errors.Is(err, io.ErrClosedPipe) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.ErrClosedPipe, err)
	}
	if onPacketSendError.calls() != 1 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, onPacketSendError.calls())
	}
}

func TestServerReceivePacketConnect(t *testing.T) {
	testCases := []struct {
		fixture string
		packet  Packet
	}{
		{"V3.1", &packet.Connect{Version: packet.MQTT31, ClientID: []byte("a")}},
		{"V3.1.1", &packet.Connect{Version: packet.MQTT311, ClientID: []byte("a")}},
		{"V5.0", &packet.Connect{Version: packet.MQTT50, ClientID: []byte("a")}},
	}

	for _, tc := range testCases {
		t.Run(tc.fixture, func(t *testing.T) {
			fixture := readPacketFixture(t, "connect.json", tc.fixture)
			onPacketReceive := &mockOnPacketReceiveHook{}

			received := make(chan Packet)
			onPacketReceived := &mockOnPacketReceivedHook{cb: func(_ *Client, p Packet) error {
				received <- p
				return nil
			}}

			s := newServer(t, WithHooks([]Hook{onPacketReceive, onPacketReceived}))
			defer s.Close()
			startServer(t, s)

			nc, conn := newConnection(t)
			defer func() { _ = nc.Close() }()

			serveConnection(t, s, conn)
			sendPacket(t, nc, fixture.Packet)

			p := <-received
			if !reflect.DeepEqual(tc.packet, p) {
				t.Errorf("Unexpected packet\nwant: %+v\ngot:  %+v", tc.packet, p)
			}
			if onPacketReceive.calls() != 1 {
				t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, onPacketReceive.calls())
			}
			if onPacketReceived.calls() != 1 {
				t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, onPacketReceived.calls())
			}
		})
	}
}

func TestServerConnectAccepted(t *testing.T) {
	testCases := []struct {
		connect string
		connack string
	}{
		{"V3.1", "V3 Accepted"},
		{"V3.1.1", "V3 Accepted"},
		{"V5.0", "V5.0 Success"},
	}

	for _, tc := range testCases {
		t.Run(tc.connect, func(t *testing.T) {
			connect := readPacketFixture(t, "connect.json", tc.connect)
			connack := readPacketFixture(t, "connack.json", tc.connack)

			s := newServer(t)
			defer s.Close()
			startServer(t, s)

			nc, conn := newConnection(t)
			defer func() { _ = nc.Close() }()

			serveConnection(t, s, conn)
			sendPacket(t, nc, connect.Packet)
			_ = receivePacket(t, nc, connack.Packet)
		})
	}
}

func TestServerConnectAcceptedWithHooks(t *testing.T) {
	testCases := []struct {
		connect string
		connack string
	}{
		{"V3.1", "V3 Accepted"},
		{"V3.1.1", "V3 Accepted"},
		{"V5.0", "V5.0 Success"},
	}

	for _, tc := range testCases {
		t.Run(tc.connect, func(t *testing.T) {
			connect := readPacketFixture(t, "connect.json", tc.connect)
			connack := readPacketFixture(t, "connack.json", tc.connack)

			onConnect := &mockOnConnectHook{}
			onConnected := &mockOnConnectedHook{}
			onPacketSend := &mockOnPacketSendHook{}
			onPacketSent := &mockOnPacketSentHook{}

			s := newServer(t, WithHooks([]Hook{onConnect, onConnected, onPacketSend, onPacketSent}))
			defer s.Close()
			startServer(t, s)

			nc, conn := newConnection(t)
			defer func() { _ = nc.Close() }()

			clients := make(chan *Client)
			onConnected.cb = func(c *Client) { clients <- c }

			serveConnection(t, s, conn)
			sendPacket(t, nc, connect.Packet)
			_ = receivePacket(t, nc, connack.Packet)

			client := <-clients
			if client.State() != ClientConnected {
				t.Error("It was expected the client to be connected")
			}
			if onConnect.calls() != 1 {
				t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, onConnect.calls())
			}
			if onConnected.calls() != 1 {
				t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, onConnected.calls())
			}
			if onPacketSend.calls() != 1 {
				t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, onPacketSend.calls())
			}
			if onPacketSent.calls() != 1 {
				t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, onPacketSent.calls())
			}
		})
	}
}

func TestServerConnectAcceptedMetrics(t *testing.T) {
	testCases := []struct {
		connect string
		connack string
	}{
		{"V3.1", "V3 Accepted"},
		{"V3.1.1", "V3 Accepted"},
		{"V5.0", "V5.0 Success"},
	}

	for _, tc := range testCases {
		t.Run(tc.connect, func(t *testing.T) {
			connect := readPacketFixture(t, "connect.json", tc.connect)
			connack := readPacketFixture(t, "connack.json", tc.connack)

			onConnect := &mockOnConnectHook{}
			onConnected := &mockOnConnectedHook{}
			onPacketSend := &mockOnPacketSendHook{}
			onPacketSent := &mockOnPacketSentHook{}

			s := newServer(t, WithHooks([]Hook{onConnect, onConnected, onPacketSend, onPacketSent}))
			defer s.Close()
			startServer(t, s)

			nc, conn := newConnection(t)
			defer func() { _ = nc.Close() }()

			connected := make(chan struct{})
			onConnected.cb = func(_ *Client) { close(connected) }

			serveConnection(t, s, conn)
			sendPacket(t, nc, connect.Packet)
			_ = receivePacket(t, nc, connack.Packet)

			<-connected
			if s.Metrics.ClientsConnected.Value() != 1 {
				t.Errorf("Unexpected metric value\nwant: %v\ngot:  %v",
					1, s.Metrics.ClientsConnected.Value())
			}
			if s.Metrics.PacketReceived.Value() != 1 {
				t.Errorf("Unexpected metric value\nwant: %v\ngot:  %v",
					1, s.Metrics.PacketReceived.Value())
			}
			if s.Metrics.PacketSent.Value() != 1 {
				t.Errorf("Unexpected metric value\nwant: %v\ngot:  %v", 1, s.Metrics.PacketSent.Value())
			}
			if s.Metrics.BytesReceived.Value() != uint64(len(connect.Packet)) {
				t.Errorf("Unexpected metric value\nwant: %v\ngot:  %v",
					len(connect.Packet), s.Metrics.BytesReceived.Value())
			}
			if s.Metrics.BytesSent.Value() != uint64(len(connack.Packet)) {
				t.Errorf("Unexpected metric value\nwant: %v\ngot:  %v",
					len(connack.Packet), s.Metrics.BytesSent.Value())
			}
		})
	}
}

func TestServerConnectRejectedDueToConfig(t *testing.T) {
	testCases := []struct {
		version packet.Version
		config  string
		value   any
		connect string
		connack string
	}{
		{
			packet.MQTT31,
			"max_keep_alive_sec",
			50,
			"V3.1 Keep Alive",
			"V3 Identifier Rejected",
		},
		{
			packet.MQTT311,
			"max_keep_alive_sec",
			50,
			"V3.1 Keep Alive",
			"V3 Identifier Rejected",
		},
		{
			packet.MQTT31,
			"max_client_id_size",
			50,
			"V3.1 Big Client ID",
			"V3 Identifier Rejected",
		},
		{
			packet.MQTT311,
			"max_client_id_size",
			23,
			"V3.1.1 Big Client ID",
			"V3 Identifier Rejected",
		},
		{
			packet.MQTT50,
			"max_client_id_size",
			23,
			"V5.0 Big Client ID",
			"V5.0 Client ID Not Valid",
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s-%s", tc.version.String(), tc.config), func(t *testing.T) {
			connect := readPacketFixture(t, "connect.json", tc.connect)
			connack := readPacketFixture(t, "connack.json", tc.connack)

			js, err := json.Marshal(map[string]any{tc.config: tc.value})
			if err != nil {
				t.Fatalf("Unexpected error\n%v", err)
			}

			c := NewDefaultConfig()
			err = json.Unmarshal(js, &c)
			if err != nil {
				t.Fatalf("Unexpected error\n%v", err)
			}

			s := newServer(t, WithConfig(c))
			defer s.Close()
			startServer(t, s)

			nc, conn := newConnection(t)
			defer func() { _ = nc.Close() }()

			serveConnection(t, s, conn)
			sendPacket(t, nc, connect.Packet)
			_ = receivePacket(t, nc, connack.Packet)
		})
	}
}

func TestServerConnectRejectedDueToEmptyClientID(t *testing.T) {
	testCases := []struct {
		version packet.Version
		connect string
		connack string
	}{
		{packet.MQTT311, "V3.1.1 Empty Client ID", "V3 Identifier Rejected"},
		{packet.MQTT50, "V5.0 Empty Client ID", "V5.0 Client ID Not Valid"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s-%s", tc.version.String(), tc.connect), func(t *testing.T) {
			connect := readPacketFixture(t, "connect.json", tc.connect)
			connack := readPacketFixture(t, "connack.json", tc.connack)

			s := newServer(t)
			defer s.Close()
			startServer(t, s)

			nc, conn := newConnection(t)
			defer func() { _ = nc.Close() }()

			serveConnection(t, s, conn)
			sendPacket(t, nc, connect.Packet)
			_ = receivePacket(t, nc, connack.Packet)
		})
	}
}

func TestServerConnectConnectionClosedWhenRejected(t *testing.T) {
	connect := readPacketFixture(t, "connect.json", "V5.0 Empty Client ID")
	connack := readPacketFixture(t, "connack.json", "V5.0 Client ID Not Valid")

	onConnect := &mockOnConnectHook{}
	onConnectionClosed := &mockOnConnectionClosedHook{}

	s := newServer(t, WithHooks([]Hook{onConnect, onConnectionClosed}))
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	clientStream := make(chan *Client, 1)
	onConnect.cb = func(c *Client, _ *packet.Connect) error {
		clientStream <- c
		return nil
	}

	closed := make(chan struct{})
	onConnectionClosed.cb = func(_ *Connection, _ error) { close(closed) }

	serveConnection(t, s, conn)
	sendPacket(t, nc, connect.Packet)
	_ = receivePacket(t, nc, connack.Packet)

	<-closed
	_, err := nc.Read(nil)
	if !errors.Is(err, io.EOF) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
	}

	c := <-clientStream
	if c.State() != ClientDisconnected {
		t.Error("The connection should be disconnected")
	}
	if c.Connection != nil {
		t.Errorf("Unexpected connection pointer\nwant: %v\ngot:  %+v", nil, c.Connection)
	}
}

func TestServerConnectPersistentSession(t *testing.T) {
	testCases := []struct {
		connect string
		connack string
	}{
		{"V3.1", "V3 Accepted"},
		{"V3.1.1", "V3 Accepted"},
		{"V5.0 Session Expiry Interval", "V5.0 Success"},
	}

	for _, tc := range testCases {
		t.Run(tc.connect, func(t *testing.T) {
			connect := readPacketFixture(t, "connect.json", tc.connect)
			connack := readPacketFixture(t, "connack.json", tc.connack)
			ss := &mockSessionStore{}
			h := &mockOnConnectedHook{}

			s := newServer(t, WithSessionStore(ss), WithHooks([]Hook{h}))
			defer s.Close()
			startServer(t, s)

			nc, conn := newConnection(t)
			defer func() { _ = nc.Close() }()

			connected := make(chan struct{})
			h.cb = func(_ *Client) { close(connected) }

			serveConnection(t, s, conn)
			sendPacket(t, nc, connect.Packet)
			_ = receivePacket(t, nc, connack.Packet)

			<-connected
			if ss.saveCalls() != 1 {
				t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, ss.saveCalls())
			}
			if s.Metrics.PersistentSessions.Value() != 1 {
				t.Errorf("Unexpected metric value\nwant: %v\ngot:  %v",
					1, s.Metrics.PersistentSessions.Value())
			}
		})
	}
}

func TestServerConnectNoSessionPersistence(t *testing.T) {
	testCases := []struct {
		connect string
		connack string
	}{
		{"V3.1 Clean Session", "V3 Accepted"},
		{"V3.1.1 Clean Session", "V3 Accepted"},
		{"V5.0", "V5.0 Success"},
	}

	for _, tc := range testCases {
		t.Run(tc.connect, func(t *testing.T) {
			connect := readPacketFixture(t, "connect.json", tc.connect)
			connack := readPacketFixture(t, "connack.json", tc.connack)
			ss := &mockSessionStore{}
			h := &mockOnConnectedHook{}

			s := newServer(t, WithSessionStore(ss), WithHooks([]Hook{h}))
			defer s.Close()
			startServer(t, s)

			nc, conn := newConnection(t)
			defer func() { _ = nc.Close() }()

			connected := make(chan struct{})
			h.cb = func(_ *Client) { close(connected) }

			serveConnection(t, s, conn)
			sendPacket(t, nc, connect.Packet)
			_ = receivePacket(t, nc, connack.Packet)

			<-connected
			if ss.saveCalls() != 0 {
				t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 0, ss.saveCalls())
			}
			if s.Metrics.PersistentSessions.Value() != 0 {
				t.Errorf("Unexpected metric value\nwant: %v\ngot:  %v",
					0, s.Metrics.PersistentSessions.Value())
			}
		})
	}
}

func TestServerConnectCleanStart(t *testing.T) {
	testCases := []struct {
		connect string
		connack string
	}{
		{"V3.1 Clean Session", "V3 Accepted"},
		{"V3.1.1 Clean Session", "V3 Accepted"},
		{"V5.0 Clean Start", "V5.0 Success"},
	}

	for _, tc := range testCases {
		t.Run(tc.connect, func(t *testing.T) {
			connect := readPacketFixture(t, "connect.json", tc.connect)
			connack := readPacketFixture(t, "connack.json", tc.connack)

			s := newServer(t)
			defer s.Close()
			startServer(t, s)

			nc, conn := newConnection(t)
			defer func() { _ = nc.Close() }()

			serveConnection(t, s, conn)
			sendPacket(t, nc, connect.Packet)
			_ = receivePacket(t, nc, connack.Packet)
		})
	}
}

func TestServerConnectNoCleanStart(t *testing.T) {
	testCases := []struct {
		connect string
		connack string
	}{
		{"V3.1", "V3 Accepted"},
		{"V3.1.1", "V3 Accepted"},
		{"V5.0", "V5.0 Success"},
	}

	for _, tc := range testCases {
		t.Run(tc.connect, func(t *testing.T) {
			connect := readPacketFixture(t, "connect.json", tc.connect)
			connack := readPacketFixture(t, "connack.json", tc.connack)
			ss := &mockSessionStore{}

			s := newServer(t, WithSessionStore(ss))
			defer s.Close()
			startServer(t, s)

			nc, conn := newConnection(t)
			defer func() { _ = nc.Close() }()

			serveConnection(t, s, conn)
			sendPacket(t, nc, connect.Packet)
			_ = receivePacket(t, nc, connack.Packet)

			if ss.deleteCalls() != 0 {
				t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 0, ss.deleteCalls())
			}
		})
	}
}

func TestServerConnectWithSessionPresent(t *testing.T) {
	testCases := []struct {
		connect string
		connack string
	}{
		{"V3.1", "V3 Accepted"},
		{"V3.1.1", "V3 Accepted + Session Present"},
		{"V5.0", "V5.0 Success + Session Present"},
	}

	for _, tc := range testCases {
		t.Run(tc.connect, func(t *testing.T) {
			connect := readPacketFixture(t, "connect.json", tc.connect)
			connack := readPacketFixture(t, "connack.json", tc.connack)

			ss := &mockSessionStore{}
			ss.getCB = func(_ []byte, s *Session) error { return nil }

			s := newServer(t, WithSessionStore(ss))
			defer s.Close()
			startServer(t, s)

			nc, conn := newConnection(t)
			defer func() { _ = nc.Close() }()

			serveConnection(t, s, conn)
			sendPacket(t, nc, connect.Packet)
			_ = receivePacket(t, nc, connack.Packet)

			if ss.getCalls() != 1 {
				t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, ss.getCalls())
			}
		})
	}
}

func TestServerConnectWithExpiredSession(t *testing.T) {
	testCases := []struct {
		connect string
		connack string
	}{
		{"V3.1", "V3 Accepted"},
		{"V3.1.1", "V3 Accepted"},
		{"V5.0", "V5.0 Success"},
	}

	for _, tc := range testCases {
		t.Run(tc.connect, func(t *testing.T) {
			connect := readPacketFixture(t, "connect.json", tc.connect)
			connack := readPacketFixture(t, "connack.json", tc.connack)

			ss := &mockSessionStore{}
			ss.getCB = func(_ []byte, s *Session) error {
				s.ConnectedAt = time.Now().Add(-2 * time.Second).UnixMilli()
				s.Properties = &SessionProperties{
					Flags:                 packet.PropertyFlags(0).Set(packet.PropertySessionExpiryInterval),
					SessionExpiryInterval: 1,
				}
				return nil
			}

			s := newServer(t, WithSessionStore(ss))
			defer s.Close()
			startServer(t, s)

			nc, conn := newConnection(t)
			defer func() { _ = nc.Close() }()

			serveConnection(t, s, conn)
			sendPacket(t, nc, connect.Packet)
			_ = receivePacket(t, nc, connack.Packet)

			if ss.deleteCalls() != 1 {
				t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, ss.deleteCalls())
			}
		})
	}
}

func TestServerConnectWithConfig(t *testing.T) {
	testCases := []struct {
		config  string
		value   any
		connect string
		connack string
	}{
		{
			"max_keep_alive_sec",
			100,
			"V3.1 Keep Alive",
			"V3 Accepted",
		},
		{
			"max_keep_alive_sec",
			100,
			"V3.1.1 Keep Alive",
			"V3 Accepted",
		},
		{
			"max_keep_alive_sec",
			100,
			"V5.0 Keep Alive",
			"V5.0 Success",
		},
		{
			"max_keep_alive_sec",
			50,
			"V5.0 Keep Alive",
			"V5.0 Success + Server Keep Alive",
		},
		{
			"max_session_expiry_interval_sec",
			30,
			"V5.0 Session Expiry Interval",
			"V5.0 Success + Session Expiry Interval",
		},
		{
			"max_inflight_messages",
			100,
			"V5.0",
			"V5.0 Success + Receive Maximum",
		},
		{
			"topic_alias_max",
			10,
			"V5.0",
			"V5.0 Success + Topic Alias Maximum",
		},
		{
			"max_qos",
			1,
			"V5.0",
			"V5.0 Success + Maximum QoS",
		},
		{
			"retain_available",
			false,
			"V5.0",
			"V5.0 Success + Retain Available",
		},
		{
			"max_packet_size",
			200,
			"V5.0",
			"V5.0 Success + Maximum Packet Size",
		},
		{
			"wildcard_subscription_available",
			false,
			"V5.0",
			"V5.0 Success + Wildcard Subscription Available",
		},
		{
			"subscription_id_available",
			false,
			"V5.0",
			"V5.0 Success + Subscription ID Available",
		},
		{
			"shared_subscription_available",
			false,
			"V5.0",
			"V5.0 Success + Shared Subscription Available",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.config, func(t *testing.T) {
			connect := readPacketFixture(t, "connect.json", tc.connect)
			connack := readPacketFixture(t, "connack.json", tc.connack)

			js, err := json.Marshal(map[string]any{tc.config: tc.value})
			if err != nil {
				t.Fatalf("Unexpected error\n%v", err)
			}

			c := NewDefaultConfig()
			err = json.Unmarshal(js, &c)
			if err != nil {
				t.Fatalf("Unexpected error\n%v", err)
			}

			s := newServer(t, WithConfig(c))
			defer s.Close()
			startServer(t, s)

			nc, conn := newConnection(t)
			defer func() { _ = nc.Close() }()

			serveConnection(t, s, conn)
			sendPacket(t, nc, connect.Packet)
			_ = receivePacket(t, nc, connack.Packet)
		})
	}
}

func TestServerConnectWithAllProperties(t *testing.T) {
	connect := readPacketFixture(t, "connect.json", "V5.0 Connect Properties")
	connack := readPacketFixture(t, "connack.json", "V5.0 Success + All Properties")
	h := &mockOnConnectedHook{}
	e := &mockEnhancedAuth{name: "cd"}

	c := NewDefaultConfig()
	c.MaxKeepAliveSec = 50
	c.MaxSessionExpiryIntervalSec = 30
	c.MaxInflightMessages = 100
	c.TopicAliasMax = 10
	c.MaxQoS = 1
	c.MaxPacketSize = 200
	c.RetainAvailable = false
	c.WildcardSubscriptionAvailable = false
	c.SubscriptionIDAvailable = false
	c.SharedSubscriptionAvailable = false

	s := newServer(t, WithConfig(c), WithHooks([]Hook{h}), WithEnhancedAuths([]EnhancedAuth{e}))
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	clients := make(chan *Client)
	h.cb = func(c *Client) { clients <- c }

	serveConnection(t, s, conn)
	sendPacket(t, nc, connect.Packet)
	_ = receivePacket(t, nc, connack.Packet)

	client := <-clients
	if client == nil {
		t.Fatal("A client was expected")
	}

	expectedSessionProps := &SessionProperties{
		Flags: packet.PropertyFlags(0).
			Set(packet.PropertySessionExpiryInterval).
			Set(packet.PropertyReceiveMaximum).
			Set(packet.PropertyMaximumPacketSize).
			Set(packet.PropertyTopicAliasMaximum).
			Set(packet.PropertyRequestResponseInfo).
			Set(packet.PropertyRequestProblemInfo).
			Set(packet.PropertyUserProperty),
		SessionExpiryInterval: c.MaxSessionExpiryIntervalSec,
		ReceiveMaximum:        c.MaxInflightMessages,
		MaximumPacketSize:     c.MaxPacketSize,
		TopicAliasMaximum:     c.TopicAliasMax,
		RequestResponseInfo:   true,
		RequestProblemInfo:    true,
		UserProperties:        []packet.UserProperty{{Key: []byte("a"), Value: []byte("b")}},
	}
	if !reflect.DeepEqual(expectedSessionProps, client.Session.Properties) {
		t.Errorf("Unexpected session properties\nwant: %+v\ngot:  %+v",
			expectedSessionProps, client.Session.Properties)
	}
}

func TestServerConnectWithLastWill(t *testing.T) {
	connect := readPacketFixture(t, "connect.json", "V5.0 Last Will")
	connack := readPacketFixture(t, "connack.json", "V5.0 Success")
	h := &mockOnConnectedHook{}

	s := newServer(t, WithHooks([]Hook{h}))
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	clients := make(chan *Client)
	h.cb = func(c *Client) { clients <- c }

	serveConnection(t, s, conn)
	sendPacket(t, nc, connect.Packet)
	_ = receivePacket(t, nc, connack.Packet)

	client := <-clients
	if client == nil {
		t.Fatal("A client was expected")
	}

	expectedLastWill := &LastWill{
		Topic:   []byte("e"),
		Payload: []byte("f"),
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
			ContentType:            []byte("a"),
			ResponseTopic:          []byte("b"),
			CorrelationData:        []byte{20, 1},
			UserProperties:         []packet.UserProperty{{Key: []byte("c"), Value: []byte("d")}},
		},
	}
	if !reflect.DeepEqual(expectedLastWill, client.Session.LastWill) {
		t.Errorf("Unexpected last will\nwant: %+v\ngot:  %+v",
			expectedLastWill, client.Session.LastWill)
	}
	if !reflect.DeepEqual(expectedLastWill.Properties, client.Session.LastWill.Properties) {
		t.Errorf("Unexpected last will properties\nwant: %+v\ngot:  %+v",
			expectedLastWill.Properties, client.Session.LastWill.Properties)
	}
}

func TestServerConnectWithEnhancedAuth(t *testing.T) {
	connect := readPacketFixture(t, "connect.json", "V5.0 Enhanced Authentication")
	auth := readPacketFixture(t, "auth.json", "Continue With Auth Method and Data")
	e := &mockEnhancedAuth{}
	onConnected := &mockOnConnectedHook{}
	onConnectFailed := &mockOnConnectFailedHook{}

	s := newServer(t, WithHooks([]Hook{onConnected, onConnectFailed}), WithEnhancedAuths([]EnhancedAuth{e}))
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	var client *Client
	packets := make(chan Packet, 1)
	e.cb = func(c *Client, p Packet) (PacketEncodable, error) {
		client = c
		packets <- p
		pkt := &packet.Auth{Code: packet.ReasonCodeContinueAuthentication, Properties: &packet.AuthProperties{}}

		pkt.Properties.Set(packet.PropertyAuthenticationMethod)
		pkt.Properties.AuthenticationMethod = []byte(e.Name())

		pkt.Properties.Set(packet.PropertyAuthenticationData)
		pkt.Properties.AuthenticationData = []byte{2}
		return pkt, nil
	}

	serveConnection(t, s, conn)
	sendPacket(t, nc, connect.Packet)
	_ = receivePacket(t, nc, auth.Packet)

	pkt := <-packets
	if pkt == nil {
		t.Fatalf("A packet was expected")
	}

	props := pkt.(*packet.Connect).Properties
	if !props.Has(packet.PropertyAuthenticationMethod) && e.Name() != string(props.AuthenticationMethod) {
		t.Errorf("Unexpected authentication method\nwant: %s\ngot:  %s", e.Name(),
			string(props.AuthenticationMethod))
	}
	if !props.Has(packet.PropertyAuthenticationData) && !bytes.Equal([]byte{1}, props.AuthenticationData) {
		t.Errorf("Unexpected authentication data\nwant: %v\ngot:  %v", []byte{1}, props.AuthenticationData)
	}
	if onConnected.calls() != 0 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 0, onConnected.calls())
	}
	if onConnectFailed.calls() != 0 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 0, onConnectFailed.calls())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	for ctx.Err() == nil {
		if client.State() == ClientAuthenticating {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if client.State() != ClientAuthenticating {
		t.Fatalf("Unexpected state\nwant: %v\ngot:  %v", ClientAuthenticating, client.State())
	}
}

func TestServerConnectWithEnhancedAuthReturningConnAck(t *testing.T) {
	connect := readPacketFixture(t, "connect.json", "V5.0 Enhanced Authentication")
	connack := readPacketFixture(t, "connack.json", "V5.0 Success + Enhanced Auth")
	e := &mockEnhancedAuth{}

	s := newServer(t, WithEnhancedAuths([]EnhancedAuth{e}))
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	e.cb = func(_ *Client, p Packet) (PacketEncodable, error) {
		pkt := &packet.ConnAck{Properties: &packet.ConnAckProperties{}}

		pkt.Properties.Set(packet.PropertyAuthenticationMethod)
		pkt.Properties.AuthenticationMethod = []byte(e.Name())

		pkt.Properties.Set(packet.PropertyAuthenticationData)
		pkt.Properties.AuthenticationData = []byte{2}
		return pkt, nil
	}

	serveConnection(t, s, conn)
	sendPacket(t, nc, connect.Packet)
	_ = receivePacket(t, nc, connack.Packet)
}

func TestServerConnectWithEnhancedAuthNotAuthorized(t *testing.T) {
	connect := readPacketFixture(t, "connect.json", "V5.0 Enhanced Authentication")
	connack := readPacketFixture(t, "connack.json", "V5.0 Not Authorized + Auth Method")
	e := &mockEnhancedAuth{}

	s := newServer(t, WithEnhancedAuths([]EnhancedAuth{e}))
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	e.cb = func(_ *Client, p Packet) (PacketEncodable, error) { return nil, packet.ErrNotAuthorized }

	serveConnection(t, s, conn)
	sendPacket(t, nc, connect.Packet)
	_ = receivePacket(t, nc, connack.Packet)

	_, err := nc.Read(nil)
	if !errors.Is(err, io.EOF) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
	}
}

func TestServerConnectWithEnhancedAuthReturningConnAckNotAuthorized(t *testing.T) {
	connect := readPacketFixture(t, "connect.json", "V5.0 Enhanced Authentication")
	connack := readPacketFixture(t, "connack.json", "V5.0 Not Authorized + Auth Method and Reason String")
	e := &mockEnhancedAuth{}
	onConnected := &mockOnConnectedHook{}
	onConnectFailed := &mockOnConnectFailedHook{}

	s := newServer(t, WithHooks([]Hook{onConnected, onConnectFailed}), WithEnhancedAuths([]EnhancedAuth{e}))
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	e.cb = func(_ *Client, p Packet) (PacketEncodable, error) {
		pkt := &packet.ConnAck{Code: packet.ReasonCodeNotAuthorized, Properties: &packet.ConnAckProperties{}}

		pkt.Properties.Set(packet.PropertyAuthenticationMethod)
		pkt.Properties.AuthenticationMethod = []byte(e.Name())

		pkt.Properties.Set(packet.PropertyReasonString)
		pkt.Properties.ReasonString = []byte("a")
		return pkt, nil
	}

	serveConnection(t, s, conn)
	sendPacket(t, nc, connect.Packet)
	_ = receivePacket(t, nc, connack.Packet)

	_, err := nc.Read(nil)
	if !errors.Is(err, io.EOF) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
	}

	if onConnected.calls() != 0 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 0, onConnected.calls())
	}
	if onConnectFailed.calls() != 1 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, onConnectFailed.calls())
	}
}

func TestServerConnectWithEnhancedAuthReturningAnyError(t *testing.T) {
	connect := readPacketFixture(t, "connect.json", "V5.0 Enhanced Authentication")
	e := &mockEnhancedAuth{}

	s := newServer(t, WithEnhancedAuths([]EnhancedAuth{e}))
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	e.cb = func(_ *Client, p Packet) (PacketEncodable, error) { return nil, errors.New("failed") }

	serveConnection(t, s, conn)
	sendPacket(t, nc, connect.Packet)

	_, err := nc.Read(nil)
	if !errors.Is(err, io.EOF) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
	}
}

func TestServerConnectWithUnknownEnhancedAuthMethod(t *testing.T) {
	connect := readPacketFixture(t, "connect.json", "V5.0 Enhanced Authentication")
	connack := readPacketFixture(t, "connack.json", "V5.0 Bad Authentication Method")

	s := newServer(t)
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	serveConnection(t, s, conn)
	sendPacket(t, nc, connect.Packet)
	_ = receivePacket(t, nc, connack.Packet)

	_, err := nc.Read(nil)
	if !errors.Is(err, io.EOF) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
	}
}

func TestServerAuthPacket(t *testing.T) {
	connect := readPacketFixture(t, "connect.json", "V5.0 Enhanced Authentication")
	auth := readPacketFixture(t, "auth.json", "Continue With Auth Method and Data")
	connack := readPacketFixture(t, "connack.json", "V5.0 Success + Auth Method")
	e := &mockEnhancedAuth{}
	onAuth := &mockOnAuthPacketHook{}
	onConnected := &mockOnConnectedHook{}

	s := newServer(t, WithHooks([]Hook{onAuth, onConnected}), WithEnhancedAuths([]EnhancedAuth{e}))
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	e.cb = func(_ *Client, p Packet) (PacketEncodable, error) {
		pkt := &packet.Auth{Code: packet.ReasonCodeContinueAuthentication, Properties: &packet.AuthProperties{}}

		pkt.Properties.Set(packet.PropertyAuthenticationMethod)
		pkt.Properties.AuthenticationMethod = []byte(e.Name())

		pkt.Properties.Set(packet.PropertyAuthenticationData)
		pkt.Properties.AuthenticationData = []byte{2}
		return pkt, nil
	}

	serveConnection(t, s, conn)
	sendPacket(t, nc, connect.Packet)
	_ = receivePacket(t, nc, auth.Packet)

	var client *Client
	packets := make(chan Packet, 1)
	e.cb = func(c *Client, p Packet) (PacketEncodable, error) {
		client = c
		packets <- p
		return nil, nil
	}

	connected := make(chan struct{})
	onConnected.cb = func(_ *Client) { close(connected) }

	sendPacket(t, nc, auth.Packet)
	_ = receivePacket(t, nc, connack.Packet)

	pkt := <-packets
	if pkt == nil {
		t.Fatalf("A packet was expected")
	}

	props := pkt.(*packet.Auth).Properties
	if !props.Has(packet.PropertyAuthenticationMethod) && e.Name() != string(props.AuthenticationMethod) {
		t.Errorf("Unexpected authentication method\nwant: %s\ngot:  %s", e.Name(),
			string(props.AuthenticationMethod))
	}
	if !props.Has(packet.PropertyAuthenticationData) && !bytes.Equal([]byte{1}, props.AuthenticationData) {
		t.Errorf("Unexpected authentication data\nwant: %v\ngot:  %v", []byte{1}, props.AuthenticationData)
	}
	if onAuth.calls() != 1 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, onAuth.calls())
	}

	<-connected
	if onConnected.calls() != 1 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, onConnected.calls())
	}
	if client.State() != ClientConnected {
		t.Errorf("Unexpected client state\nwant: %v\ngot:  %v", ClientConnected, client.State())
	}
}

func TestServerAuthPacketHookError(t *testing.T) {
	connect := readPacketFixture(t, "connect.json", "V5.0 Enhanced Authentication")
	auth := readPacketFixture(t, "auth.json", "Continue With Auth Method and Data")
	e := &mockEnhancedAuth{}
	h := &mockOnAuthPacketHook{}

	s := newServer(t, WithHooks([]Hook{h}), WithEnhancedAuths([]EnhancedAuth{e}))
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	e.cb = func(_ *Client, p Packet) (PacketEncodable, error) {
		pkt := &packet.Auth{Code: packet.ReasonCodeContinueAuthentication, Properties: &packet.AuthProperties{}}

		pkt.Properties.Set(packet.PropertyAuthenticationMethod)
		pkt.Properties.AuthenticationMethod = []byte(e.Name())

		pkt.Properties.Set(packet.PropertyAuthenticationData)
		pkt.Properties.AuthenticationData = []byte{2}
		return pkt, nil
	}

	serveConnection(t, s, conn)
	sendPacket(t, nc, connect.Packet)
	_ = receivePacket(t, nc, auth.Packet)

	h.cb = func(_ *Client, _ *packet.Auth) error { return errHookFailed }
	sendPacket(t, nc, auth.Packet)

	_, err := nc.Read(nil)
	if !errors.Is(err, io.EOF) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
	}
}

func TestServerAuthPacketEnhancedAuthReturningAuth(t *testing.T) {
	connect := readPacketFixture(t, "connect.json", "V5.0 Enhanced Authentication")
	auth := readPacketFixture(t, "auth.json", "Continue With Auth Method and Data")
	e := &mockEnhancedAuth{}

	s := newServer(t, WithEnhancedAuths([]EnhancedAuth{e}))
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	var client *Client
	packets := make(chan Packet, 1)
	e.cb = func(c *Client, p Packet) (PacketEncodable, error) {
		client = c
		packets <- p
		pkt := &packet.Auth{Code: packet.ReasonCodeContinueAuthentication, Properties: &packet.AuthProperties{}}

		pkt.Properties.Set(packet.PropertyAuthenticationMethod)
		pkt.Properties.AuthenticationMethod = []byte(e.Name())

		pkt.Properties.Set(packet.PropertyAuthenticationData)
		pkt.Properties.AuthenticationData = []byte{2}
		return pkt, nil
	}

	serveConnection(t, s, conn)
	sendPacket(t, nc, connect.Packet)
	_ = receivePacket(t, nc, auth.Packet)
	<-packets

	sendPacket(t, nc, auth.Packet)
	_ = receivePacket(t, nc, auth.Packet)

	pkt := <-packets
	if pkt == nil {
		t.Fatalf("A packet was expected")
	}

	props := pkt.(*packet.Auth).Properties
	if !props.Has(packet.PropertyAuthenticationMethod) && e.Name() != string(props.AuthenticationMethod) {
		t.Errorf("Unexpected authentication method\nwant: %s\ngot:  %s", e.Name(),
			string(props.AuthenticationMethod))
	}
	if !props.Has(packet.PropertyAuthenticationData) && !bytes.Equal([]byte{1}, props.AuthenticationData) {
		t.Errorf("Unexpected authentication data\nwant: %v\ngot:  %v", []byte{1}, props.AuthenticationData)
	}

	if client == nil {
		t.Fatal("A client was expected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	for ctx.Err() == nil {
		if client.State() == ClientAuthenticating {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if client.State() != ClientAuthenticating {
		t.Fatalf("Unexpected state\nwant: %v\ngot:  %v", ClientAuthenticating, client.State())
	}
}

func TestServerAuthPacketEnhancedAuthReturningConnAck(t *testing.T) {
	connect := readPacketFixture(t, "connect.json", "V5.0 Enhanced Authentication")
	auth := readPacketFixture(t, "auth.json", "Continue With Auth Method and Data")
	connack := readPacketFixture(t, "connack.json", "V5.0 Success + Enhanced Auth")
	e := &mockEnhancedAuth{}

	s := newServer(t, WithEnhancedAuths([]EnhancedAuth{e}))
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	e.cb = func(_ *Client, p Packet) (PacketEncodable, error) {
		pkt := &packet.Auth{Code: packet.ReasonCodeContinueAuthentication, Properties: &packet.AuthProperties{}}

		pkt.Properties.Set(packet.PropertyAuthenticationMethod)
		pkt.Properties.AuthenticationMethod = []byte(e.Name())

		pkt.Properties.Set(packet.PropertyAuthenticationData)
		pkt.Properties.AuthenticationData = []byte{2}
		return pkt, nil
	}

	serveConnection(t, s, conn)
	sendPacket(t, nc, connect.Packet)
	_ = receivePacket(t, nc, auth.Packet)

	var client *Client
	e.cb = func(c *Client, p Packet) (PacketEncodable, error) {
		client = c
		pkt := &packet.ConnAck{Properties: &packet.ConnAckProperties{}}

		pkt.Properties.Set(packet.PropertyAuthenticationMethod)
		pkt.Properties.AuthenticationMethod = []byte(e.Name())

		pkt.Properties.Set(packet.PropertyAuthenticationData)
		pkt.Properties.AuthenticationData = []byte{2}
		return pkt, nil
	}

	sendPacket(t, nc, auth.Packet)
	_ = receivePacket(t, nc, connack.Packet)

	if client == nil {
		t.Fatal("A client was expected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	for ctx.Err() == nil {
		if client.State() == ClientConnected {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if client.State() != ClientConnected {
		t.Fatalf("Unexpected state\nwant: %v\ngot:  %v", ClientConnected, client.State())
	}
}

func TestServerAuthPacketEnhancedAuthReturningNotAuthorized(t *testing.T) {
	connect := readPacketFixture(t, "connect.json", "V5.0 Enhanced Authentication")
	auth := readPacketFixture(t, "auth.json", "Continue With Auth Method and Data")
	connack := readPacketFixture(t, "connack.json", "V5.0 Not Authorized + Auth Method")
	e := &mockEnhancedAuth{}

	s := newServer(t, WithEnhancedAuths([]EnhancedAuth{e}))
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	var client *Client
	e.cb = func(c *Client, p Packet) (PacketEncodable, error) {
		client = c
		pkt := &packet.Auth{Code: packet.ReasonCodeContinueAuthentication, Properties: &packet.AuthProperties{}}

		pkt.Properties.Set(packet.PropertyAuthenticationMethod)
		pkt.Properties.AuthenticationMethod = []byte(e.Name())

		pkt.Properties.Set(packet.PropertyAuthenticationData)
		pkt.Properties.AuthenticationData = []byte{2}
		return pkt, nil
	}

	serveConnection(t, s, conn)
	sendPacket(t, nc, connect.Packet)
	_ = receivePacket(t, nc, auth.Packet)

	e.cb = func(_ *Client, p Packet) (PacketEncodable, error) { return nil, packet.ErrNotAuthorized }
	sendPacket(t, nc, auth.Packet)
	_ = receivePacket(t, nc, connack.Packet)

	_, err := nc.Read(nil)
	if !errors.Is(err, io.EOF) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	for ctx.Err() == nil {
		if client.State() == ClientDisconnected {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if client.State() != ClientDisconnected {
		t.Fatalf("Unexpected state\nwant: %v\ngot:  %v", ClientDisconnected, client.State())
	}
}

func TestServerAuthPacketEnhancedAuthReturningAnyError(t *testing.T) {
	connect := readPacketFixture(t, "connect.json", "V5.0 Enhanced Authentication")
	auth := readPacketFixture(t, "auth.json", "Continue With Auth Method and Data")
	e := &mockEnhancedAuth{}

	s := newServer(t, WithEnhancedAuths([]EnhancedAuth{e}))
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	e.cb = func(_ *Client, p Packet) (PacketEncodable, error) {
		pkt := &packet.Auth{Code: packet.ReasonCodeContinueAuthentication, Properties: &packet.AuthProperties{}}

		pkt.Properties.Set(packet.PropertyAuthenticationMethod)
		pkt.Properties.AuthenticationMethod = []byte(e.Name())

		pkt.Properties.Set(packet.PropertyAuthenticationData)
		pkt.Properties.AuthenticationData = []byte{2}
		return pkt, nil
	}

	serveConnection(t, s, conn)
	sendPacket(t, nc, connect.Packet)
	_ = receivePacket(t, nc, auth.Packet)

	e.cb = func(_ *Client, p Packet) (PacketEncodable, error) { return nil, errors.New("failed") }
	sendPacket(t, nc, auth.Packet)

	_, err := nc.Read(nil)
	if !errors.Is(err, io.EOF) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
	}
}

func TestServerAuthPacketEnhancedAuthReturningConnAckNotAuthorized(t *testing.T) {
	connect := readPacketFixture(t, "connect.json", "V5.0 Enhanced Authentication")
	auth := readPacketFixture(t, "auth.json", "Continue With Auth Method and Data")
	connack := readPacketFixture(t, "connack.json", "V5.0 Not Authorized + Auth Method and Reason String")
	e := &mockEnhancedAuth{}

	s := newServer(t, WithEnhancedAuths([]EnhancedAuth{e}))
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	e.cb = func(_ *Client, p Packet) (PacketEncodable, error) {
		pkt := &packet.Auth{Code: packet.ReasonCodeContinueAuthentication, Properties: &packet.AuthProperties{}}

		pkt.Properties.Set(packet.PropertyAuthenticationMethod)
		pkt.Properties.AuthenticationMethod = []byte(e.Name())

		pkt.Properties.Set(packet.PropertyAuthenticationData)
		pkt.Properties.AuthenticationData = []byte{2}
		return pkt, nil
	}

	serveConnection(t, s, conn)
	sendPacket(t, nc, connect.Packet)
	_ = receivePacket(t, nc, auth.Packet)

	e.cb = func(_ *Client, p Packet) (PacketEncodable, error) {
		pkt := &packet.ConnAck{Code: packet.ReasonCodeNotAuthorized, Properties: &packet.ConnAckProperties{}}

		pkt.Properties.Set(packet.PropertyAuthenticationMethod)
		pkt.Properties.AuthenticationMethod = []byte(e.Name())

		pkt.Properties.Set(packet.PropertyReasonString)
		pkt.Properties.ReasonString = []byte("a")
		return pkt, nil
	}

	sendPacket(t, nc, auth.Packet)
	_ = receivePacket(t, nc, connack.Packet)

	_, err := nc.Read(nil)
	if !errors.Is(err, io.EOF) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
	}
}

func TestServerAuthPacketEnhancedAuthMethodNotMatching(t *testing.T) {
	connect := readPacketFixture(t, "connect.json", "V5.0 Enhanced Authentication")
	auth1 := readPacketFixture(t, "auth.json", "Continue With Auth Method and Data")
	auth2 := readPacketFixture(t, "auth.json", "Continue With Auth Method")
	connack := readPacketFixture(t, "connack.json", "V5.0 Bad Authentication Method")
	e := &mockEnhancedAuth{}

	s := newServer(t, WithEnhancedAuths([]EnhancedAuth{e}))
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	e.cb = func(_ *Client, p Packet) (PacketEncodable, error) {
		pkt := &packet.Auth{Code: packet.ReasonCodeContinueAuthentication, Properties: &packet.AuthProperties{}}

		pkt.Properties.Set(packet.PropertyAuthenticationMethod)
		pkt.Properties.AuthenticationMethod = []byte(e.Name())

		pkt.Properties.Set(packet.PropertyAuthenticationData)
		pkt.Properties.AuthenticationData = []byte{2}
		return pkt, nil
	}

	serveConnection(t, s, conn)
	sendPacket(t, nc, connect.Packet)
	_ = receivePacket(t, nc, auth1.Packet)

	sendPacket(t, nc, auth2.Packet)
	_ = receivePacket(t, nc, connack.Packet)

	_, err := nc.Read(nil)
	if !errors.Is(err, io.EOF) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
	}
}

func TestServerAuthPacketReAuthenticate(t *testing.T) {
	connect := readPacketFixture(t, "connect.json", "V5.0 Enhanced Authentication")
	connack := readPacketFixture(t, "connack.json", "V5.0 Success + Auth Method")
	auth1 := readPacketFixture(t, "auth.json", "Re-authenticate With Auth Method and Data")
	auth2 := readPacketFixture(t, "auth.json", "Success With Auth Method")
	e := &mockEnhancedAuth{}

	s := newServer(t, WithEnhancedAuths([]EnhancedAuth{e}))
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	serveConnection(t, s, conn)
	sendPacket(t, nc, connect.Packet)
	_ = receivePacket(t, nc, connack.Packet)

	var client *Client
	e.cb = func(c *Client, p Packet) (PacketEncodable, error) {
		client = c
		return nil, nil
	}

	sendPacket(t, nc, auth1.Packet)
	_ = receivePacket(t, nc, auth2.Packet)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	for ctx.Err() == nil {
		if client.State() == ClientConnected {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if client.State() != ClientConnected {
		t.Fatalf("Unexpected state\nwant: %v\ngot:  %v", ClientConnected, client.State())
	}
}

func TestServerAuthPacketReAuthenticating(t *testing.T) {
	connect := readPacketFixture(t, "connect.json", "V5.0 Enhanced Authentication")
	connack := readPacketFixture(t, "connack.json", "V5.0 Success + Auth Method")
	auth1 := readPacketFixture(t, "auth.json", "Re-authenticate With Auth Method and Data")
	auth2 := readPacketFixture(t, "auth.json", "Continue With Auth Method and Data")
	e := &mockEnhancedAuth{}

	s := newServer(t, WithEnhancedAuths([]EnhancedAuth{e}))
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	serveConnection(t, s, conn)
	sendPacket(t, nc, connect.Packet)
	_ = receivePacket(t, nc, connack.Packet)

	var client *Client
	e.cb = func(c *Client, p Packet) (PacketEncodable, error) {
		client = c
		pkt := &packet.Auth{Code: packet.ReasonCodeContinueAuthentication, Properties: &packet.AuthProperties{}}

		pkt.Properties.Set(packet.PropertyAuthenticationMethod)
		pkt.Properties.AuthenticationMethod = []byte(e.Name())

		pkt.Properties.Set(packet.PropertyAuthenticationData)
		pkt.Properties.AuthenticationData = []byte{2}
		return pkt, nil
	}

	sendPacket(t, nc, auth1.Packet)
	_ = receivePacket(t, nc, auth2.Packet)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	for ctx.Err() == nil {
		if client.State() == ClientReAuthenticating {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if client.State() != ClientReAuthenticating {
		t.Fatalf("Unexpected state\nwant: %v\ngot:  %v", ClientReAuthenticating, client.State())
	}
}

func TestServerAuthPacketInvalidState(t *testing.T) {
	auth := readPacketFixture(t, "auth.json", "Re-authenticate With Auth Method and Data")
	e := &mockEnhancedAuth{}

	s := newServer(t, WithEnhancedAuths([]EnhancedAuth{e}))
	defer s.Close()
	startServer(t, s)

	nc, conn := newConnection(t)
	defer func() { _ = nc.Close() }()

	serveConnection(t, s, conn)
	sendPacket(t, nc, auth.Packet)

	_, err := nc.Read(nil)
	if !errors.Is(err, io.EOF) {
		t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
	}

	if e.calls() != 0 {
		t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 0, e.calls())
	}
}

func TestServerConnectUnavailableWhenGetSessionReturnsError(t *testing.T) {
	testCases := []struct {
		connect string
		connack string
	}{
		{"V3.1", "V3 Server Unavailable"},
		{"V3.1.1", "V3 Server Unavailable"},
		{"V5.0", "V5.0 Server Unavailable"},
	}

	for _, tc := range testCases {
		t.Run(tc.connect, func(t *testing.T) {
			connect := readPacketFixture(t, "connect.json", tc.connect)
			connack := readPacketFixture(t, "connack.json", tc.connack)

			ss := &mockSessionStore{}
			ss.getCB = func(_ []byte, s *Session) error { return errSessionStoreFailed }

			connectErr := make(chan error)
			h := &mockOnConnectFailedHook{cb: func(_ *Client, err error) { connectErr <- err }}

			s := newServer(t, WithSessionStore(ss), WithHooks([]Hook{h}))
			defer s.Close()
			startServer(t, s)

			nc, conn := newConnection(t)
			defer func() { _ = nc.Close() }()

			serveConnection(t, s, conn)
			sendPacket(t, nc, connect.Packet)
			_ = receivePacket(t, nc, connack.Packet)

			err := <-connectErr
			if !errors.Is(err, packet.ErrServerUnavailable) {
				t.Errorf("Unexpected error\nwant: %v\ngot:  %v", packet.ErrServerUnavailable, err)
			}

			_, err = nc.Read(nil)
			if !errors.Is(err, io.EOF) {
				t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
			}
		})
	}
}

func TestServerConnectUnavailableWhenDeleteSessionReturnsError(t *testing.T) {
	testCases := []struct {
		connect string
		connack string
	}{
		{"V3.1 Clean Session", "V3 Server Unavailable"},
		{"V3.1.1 Clean Session", "V3 Server Unavailable"},
		{"V5.0 Clean Start", "V5.0 Server Unavailable"},
	}

	for _, tc := range testCases {
		t.Run(tc.connect, func(t *testing.T) {
			connect := readPacketFixture(t, "connect.json", tc.connect)
			connack := readPacketFixture(t, "connack.json", tc.connack)

			ss := &mockSessionStore{}
			ss.deleteCB = func(_ []byte) error { return errSessionStoreFailed }

			connectErr := make(chan error)
			h := &mockOnConnectFailedHook{cb: func(_ *Client, err error) { connectErr <- err }}

			s := newServer(t, WithSessionStore(ss), WithHooks([]Hook{h}))
			defer s.Close()
			startServer(t, s)

			nc, conn := newConnection(t)
			defer func() { _ = nc.Close() }()

			serveConnection(t, s, conn)
			sendPacket(t, nc, connect.Packet)
			_ = receivePacket(t, nc, connack.Packet)

			if ss.deleteCalls() != 1 {
				t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, ss.deleteCalls())
			}

			err := <-connectErr
			if !errors.Is(err, packet.ErrServerUnavailable) {
				t.Errorf("Unexpected error\nwant: %v\ngot:  %v", packet.ErrServerUnavailable, err)
			}

			_, err = nc.Read(nil)
			if !errors.Is(err, io.EOF) {
				t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
			}
		})
	}
}

func TestServerConnectUnavailableWhenSaveSessionReturnsError(t *testing.T) {
	testCases := []struct {
		connect string
		connack string
	}{
		{"V3.1", "V3 Server Unavailable"},
		{"V3.1.1", "V3 Server Unavailable"},
		{"V5.0 Session Expiry Interval", "V5.0 Server Unavailable"},
	}

	for _, tc := range testCases {
		t.Run(tc.connect, func(t *testing.T) {
			connect := readPacketFixture(t, "connect.json", tc.connect)
			connack := readPacketFixture(t, "connack.json", tc.connack)

			ss := &mockSessionStore{}
			ss.saveCB = func(_ []byte, s *Session) error { return errSessionStoreFailed }

			connectErr := make(chan error)
			h := &mockOnConnectFailedHook{cb: func(_ *Client, err error) { connectErr <- err }}

			s := newServer(t, WithSessionStore(ss), WithHooks([]Hook{h}))
			defer s.Close()
			startServer(t, s)

			nc, conn := newConnection(t)
			defer func() { _ = nc.Close() }()

			serveConnection(t, s, conn)
			sendPacket(t, nc, connect.Packet)
			_ = receivePacket(t, nc, connack.Packet)

			err := <-connectErr
			if !errors.Is(err, packet.ErrServerUnavailable) {
				t.Errorf("Unexpected error\nwant: %v\ngot:  %v", packet.ErrServerUnavailable, err)
			}

			_, err = nc.Read(nil)
			if !errors.Is(err, io.EOF) {
				t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
			}
		})
	}
}

func TestServerConnectProtocolErrorOnDuplicatePacket(t *testing.T) {
	testCases := []struct {
		connect string
		connack []string
	}{
		{"V3.1", []string{"V3 Accepted"}},
		{"V3.1.1", []string{"V3 Accepted"}},
		{"V5.0", []string{"V5.0 Success", "V5.0 Protocol Error"}},
	}

	for _, tc := range testCases {
		t.Run(tc.connect, func(t *testing.T) {
			connect := readPacketFixture(t, "connect.json", tc.connect)
			connack := make([]testdata.PacketFixture, 0, len(tc.connack))
			for _, p := range tc.connack {
				connack = append(connack, readPacketFixture(t, "connack.json", p))
			}

			connClosedErr := make(chan error, 1)
			h := &mockOnConnectionClosedHook{cb: func(_ *Connection, err error) {
				connClosedErr <- err
			}}

			s := newServer(t, WithHooks([]Hook{h}))
			defer s.Close()
			startServer(t, s)

			nc, conn := newConnection(t)
			defer func() { _ = nc.Close() }()

			serveConnection(t, s, conn)
			sendPacket(t, nc, connect.Packet)
			_ = receivePacket(t, nc, connack[0].Packet)

			sendPacket(t, nc, connect.Packet)
			if len(connack) > 1 {
				_ = receivePacket(t, nc, connack[1].Packet)
			}

			_, err := nc.Read(nil)
			if !errors.Is(err, io.EOF) {
				t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
			}

			err = <-connClosedErr
			if !errors.Is(err, packet.ErrProtocolError) {
				t.Errorf("Unexpected error\nwant: %v\ngot:  %v", packet.ErrProtocolError, err)
			}
		})
	}
}

func TestServerConnectWithOnConnectReturningValidPacketError(t *testing.T) {
	testCases := []struct {
		connect string
		connack string
	}{
		{"V3.1", "V3 Server Unavailable"},
		{"V3.1.1", "V3 Server Unavailable"},
		{"V5.0", "V5.0 Server Unavailable"},
	}

	for _, tc := range testCases {
		t.Run(tc.connect, func(t *testing.T) {
			connect := readPacketFixture(t, "connect.json", tc.connect)
			connack := readPacketFixture(t, "connack.json", tc.connack)
			h := &mockOnConnectHook{cb: func(_ *Client, _ *packet.Connect) error { return packet.ErrServerUnavailable }}

			s := newServer(t, WithHooks([]Hook{h}))
			defer s.Close()
			startServer(t, s)

			nc, conn := newConnection(t)
			defer func() { _ = nc.Close() }()

			serveConnection(t, s, conn)
			sendPacket(t, nc, connect.Packet)
			_ = receivePacket(t, nc, connack.Packet)

			_, err := nc.Read(nil)
			if !errors.Is(err, io.EOF) {
				t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
			}
		})
	}
}

func TestServerConnectWithOnConnectReturningUnknownError(t *testing.T) {
	testCases := []string{"V3.1", "V3.1.1", "V5.0"}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			connect := readPacketFixture(t, "connect.json", tc)
			h := &mockOnConnectHook{cb: func(_ *Client, _ *packet.Connect) error { return errHookFailed }}

			s := newServer(t, WithHooks([]Hook{h}))
			defer s.Close()
			startServer(t, s)

			nc, conn := newConnection(t)
			defer func() { _ = nc.Close() }()

			serveConnection(t, s, conn)
			sendPacket(t, nc, connect.Packet)

			_, err := nc.Read(nil)
			if !errors.Is(err, io.EOF) {
				t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
			}
		})
	}
}

func TestServerConnectIgnoreReasonCodesFromOnConnect(t *testing.T) {
	testCases := []struct {
		connect string
		code    packet.ReasonCode
	}{
		{"V3.1", packet.ReasonCodeSuccess},
		{"V3.1.1", packet.ReasonCodeSuccess},
		{"V5.0", packet.ReasonCodeSuccess},
		{"V3.1", packet.ReasonCodeMalformedPacket},
		{"V3.1.1", packet.ReasonCodeMalformedPacket},
		{"V5.0", packet.ReasonCodeMalformedPacket},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s-%v", tc.connect, tc.code), func(t *testing.T) {
			connect := readPacketFixture(t, "connect.json", tc.connect)
			h := &mockOnConnectHook{cb: func(_ *Client, _ *packet.Connect) error { return packet.Error{Code: tc.code} }}

			s := newServer(t, WithHooks([]Hook{h}))
			defer s.Close()
			startServer(t, s)

			nc, conn := newConnection(t)
			defer func() { _ = nc.Close() }()

			serveConnection(t, s, conn)
			sendPacket(t, nc, connect.Packet)

			_, err := nc.Read(nil)
			if !errors.Is(err, io.EOF) {
				t.Errorf("Unexpected error\nwant: %v\ngot:  %v", io.EOF, err)
			}
		})
	}
}

func TestServerUpdateSessionWhenClosingConnectedClient(t *testing.T) {
	testCases := []struct {
		connect string
		connack string
	}{
		{"V3.1", "V3 Accepted"},
		{"V3.1.1", "V3 Accepted"},
		{"V5.0", "V5.0 Success"},
	}

	for _, tc := range testCases {
		t.Run(tc.connect, func(t *testing.T) {
			connect := readPacketFixture(t, "connect.json", tc.connect)
			connack := readPacketFixture(t, "connack.json", tc.connack)
			ss := &mockSessionStore{}
			onConnected := &mockOnConnectedHook{}
			onConnClosed := &mockOnConnectionClosedHook{}

			s := newServer(t, WithSessionStore(ss), WithHooks([]Hook{onConnected, onConnClosed}))
			defer s.Close()
			startServer(t, s)

			nc, conn := newConnection(t)
			defer func() { _ = nc.Close() }()

			connected := make(chan struct{})
			onConnected.cb = func(_ *Client) { close(connected) }

			serveConnection(t, s, conn)
			sendPacket(t, nc, connect.Packet)
			_ = receivePacket(t, nc, connack.Packet)
			<-connected

			sessionStream := make(chan *Session, 1)
			ss.saveCB = func(_ []byte, session *Session) error {
				sessionStream <- session
				return nil
			}

			ss._saveCalls.Store(0)

			closed := make(chan struct{})
			onConnClosed.cb = func(_ *Connection, _ error) { close(closed) }

			_ = nc.Close()
			<-closed

			session := <-sessionStream
			if session == nil {
				t.Fatal("A session was expected")
			}
			if session.DisconnectedAt == 0 {
				t.Error("Disconnected timestamp should be set")
			}
			if ss.saveCalls() != 1 {
				t.Errorf("Unexpected calls\nwant: %v\ngot:  %v", 1, ss.saveCalls())
			}
		})
	}
}

func BenchmarkHandleConnect(b *testing.B) {
	testCases := []struct {
		connect string
		connack string
	}{
		{"V3.1", "V3 Accepted"},
		{"V3.1.1", "V3 Accepted"},
		{"V5.0", "V5.0 Success"},
	}

	for _, tc := range testCases {
		b.Run(tc.connect, func(b *testing.B) {
			connect := readPacketFixture(b, "connect.json", tc.connect)
			connack := readPacketFixture(b, "connack.json", tc.connack)

			s := newServer(b)
			defer s.Close()
			startServer(b, s)

			for i := 0; i < b.N; i++ {
				nc, conn := newConnection(b)

				err := s.Serve(conn)
				if err != nil {
					b.Errorf("Unexpected error: %s", err)
				}

				sendPacket(b, nc, connect.Packet)
				_ = receivePacket(b, nc, connack.Packet)

				err = nc.Close()
				if err != nil {
					b.Errorf("Unexpected error: %s", err)
				}
			}
		})
	}
}

func readPacketFixture(tb testing.TB, path, name string) testdata.PacketFixture {
	tb.Helper()
	fixture, err := testdata.ReadPacketFixture(path, name)
	if err != nil {
		tb.Fatalf("Unexpected error\n%v", err)
	}
	return fixture
}

func newServer(tb testing.TB, opts ...OptionsFunc) *Server {
	tb.Helper()
	s, err := NewServer(opts...)
	if err != nil {
		tb.Fatalf("Unexpected error\n%v", err)
	}
	if s == nil {
		tb.Fatal("A server was expected")
	}
	return s
}

func startServer(tb testing.TB, s *Server) {
	tb.Helper()
	err := s.Start()
	if err != nil {
		tb.Fatalf("Unexpected error\n%v", err)
	}
}

func stopServer(tb testing.TB, s *Server) {
	tb.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := s.Stop(ctx)
	if err != nil {
		tb.Fatalf("Unexpected error\n%v", err)
	}
}

func serveConnection(tb testing.TB, s *Server, c *Connection) {
	tb.Helper()
	err := s.Serve(c)
	if err != nil {
		tb.Fatalf("Unexpected error\n%v", err)
	}
}

func sendPacket(tb testing.TB, nc net.Conn, p []byte) {
	tb.Helper()
	_, err := nc.Write(p)
	if err != nil {
		tb.Fatalf("Unexpected error\n%v", err)
	}
}

func receivePacket(tb testing.TB, nc net.Conn, pkt []byte) []byte {
	tb.Helper()
	p := make([]byte, len(pkt))
	n, err := nc.Read(p)
	if err != nil {
		tb.Fatalf("Unexpected error\n%v", err)
	}
	if n != len(pkt) {
		tb.Errorf("Unexpected number of bytes read\nwant: %v\ngot:  %v", len(pkt), n)
	}
	if !bytes.Equal(pkt, p) {
		tb.Errorf("Unexpected packet\nwant: %v\ngot:  %v", pkt, p)
	}
	return p
}
