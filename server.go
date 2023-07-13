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
	"bufio"
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
)

// ErrInvalidServerState indicates that the Server is not in a valid state for the operation.
var ErrInvalidServerState = errors.New("invalid server state")

const (
	// ServerNotStarted indicates that the Server has not started yet.
	ServerNotStarted ServerState = iota

	// ServerStarting indicates that the Server is starting.
	ServerStarting

	// ServerRunning indicates that the Server has started and is running.
	ServerRunning

	// ServerStopping indicates that the Server is stopping.
	ServerStopping

	// ServerStopped indicates that the Server has stopped.
	ServerStopped

	// ServerFailed indicates that the Server has failed to start.
	ServerFailed

	// ServerClosed indicates that the Server has closed.
	ServerClosed
)

// ServerState represents the current state of the Server.
type ServerState uint32

// Server is a MQTT broker responsible for implementing the MQTT 3.1, 3.1.1, and 5.0 specifications.
// To create a Server instance, use the NewServer factory method.
type Server struct {
	config      Config
	listeners   *listeners
	hooks       *hooks
	clients     *clients
	readBufPool sync.Pool
	cancelCtx   context.CancelFunc
	state       atomic.Uint32
	wg          sync.WaitGroup
	stopOnce    sync.Once
}

// NewServer creates a new Server.
// If the Options parameter is not provided, it uses the default options. If the provided Options does not contain the
// Config, it uses the default configuration.
func NewServer(opts *Options) (s *Server, err error) {
	if opts == nil {
		opts = NewDefaultOptions()
	}

	if opts.Config == nil {
		opts.Config = NewDefaultConfig()
	}

	s = &Server{
		config:    *opts.Config,
		listeners: newListeners(),
		hooks:     newHooks(),
		clients:   newClients(),
		readBufPool: sync.Pool{
			New: func() interface{} { return bufio.NewReaderSize(nil, int(s.config.ReadBufferSize)) },
		},
	}

	for _, l := range opts.Listeners {
		err = s.AddListener(l)
		if err != nil {
			return nil, err
		}
	}

	for _, h := range opts.Hooks {
		err = s.AddHook(h)
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

// AddListener adds the provided Listener into the list of listeners managed by the Server.
// If the Server is running at the time this function is called, the Server calls immediately the Listener's Listen
// method.
// If there's a Listener with the same name already managed by the Server, the ErrListenerAlreadyExists is returned.
func (s *Server) AddListener(l Listener) error {
	if s.State() == ServerRunning {
		err := s.listeners.listen(l, s.handleConnection)
		if err != nil {
			return err
		}
	}

	s.listeners.add(l)
	return nil
}

// AddHook adds the provided Hook into the list of hooks managed by the Server.
// If the Server is already running at the time this function is called, the Server calls immediately the HookOnStart
// method, if this method is implemented by the Hook.
// If the HookOnStart returns an error, the Hook is not added into the Server and the error is returned.
func (s *Server) AddHook(h Hook) error {
	if s.State() == ServerRunning {
		if hook, ok := h.(HookOnStart); ok {
			err := hook.OnStart(s)
			if err != nil {
				return err
			}
		}
	}

	return s.hooks.add(h)
}

// Start starts the Server and returns immediately.
// If the OnServerStart or OnStart hooks return an error, the start process fails and the OnServerStartFailed hook is
// called.
// If the Server is not in the ServerNotStarted or ServerStopped state, it returns ErrInvalidServerState.
func (s *Server) Start(ctx context.Context) error {
	state := s.State()

	if state != ServerNotStarted && state != ServerStopped {
		return ErrInvalidServerState
	}

	err := s.setState(ServerStarting)
	if err != nil {
		return err
	}

	newCtx, cancel := context.WithCancel(ctx)
	s.cancelCtx = cancel

	err = s.listeners.listenAll(s.handleConnection)
	if err != nil {
		_ = s.setState(ServerFailed)
		return err
	}

	s.stopOnce = sync.Once{}
	s.startDaemon(newCtx)

	_ = s.setState(ServerRunning)
	return nil
}

// Stop stops the Server gracefully.
// This function blocks until the Server has been stopped or the context has been cancelled.
// If the Server is not in the ServerRunning state, this function has no side effect.
// When the Server is stopped, it can be started again.
func (s *Server) Stop(ctx context.Context) error {
	if s.State() != ServerRunning {
		return nil
	}

	s.stop()

	stoppedCh := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(stoppedCh)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-stoppedCh:
		_ = s.setState(ServerStopped)
		return nil
	}
}

// Close closes the Server.
// If the Server is in ServerRunning or ServerStopping state, it blocks until the Server has stopped.
// When the Server is closed, it cannot be started again.
func (s *Server) Close() {
	state := s.State()
	if state != ServerRunning && state != ServerStopping {
		_ = s.setState(ServerClosed)
		return
	}

	s.stop()
	s.wg.Wait()

	if state == ServerRunning {
		_ = s.setState(ServerStopped)
	}

	_ = s.setState(ServerClosed)
}

// State returns the current ServerState.
func (s *Server) State() ServerState {
	return ServerState(s.state.Load())
}

func (s *Server) stop() {
	s.stopOnce.Do(func() {
		_ = s.setState(ServerStopping)
		s.listeners.stopAll()
		s.stopDaemon()
		s.clients.closeAll()
	})
}

func (s *Server) setState(state ServerState) error {
	var err error

	s.state.Store(uint32(state))

	if state == ServerStarting {
		err = s.hooks.onServerStart(s)
		if err == nil {
			err = s.hooks.onStart(s)
		}
	}

	if state == ServerStopping {
		s.hooks.onServerStop(s)
		s.hooks.onStop(s)
	}

	if state == ServerStopped {
		s.hooks.onServerStopped(s)
	}

	if state == ServerRunning {
		s.hooks.onServerStarted(s)
	}

	if state == ServerFailed || err != nil {
		s.state.Store(uint32(ServerFailed))
		s.hooks.onServerStartFailed(s, err)
	}

	return err
}

func (s *Server) startDaemon(ctx context.Context) {
	readyCh := make(chan struct{})

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		close(readyCh)
		<-ctx.Done()
	}()

	<-readyCh
}

func (s *Server) stopDaemon() {
	s.cancelCtx()
}

func (s *Server) handleConnection(l Listener, nc net.Conn) {
	client := newClient(nc, s, l)

	if err := s.hooks.onConnectionOpen(s, l); err != nil {
		client.Close(err)
		return
	}

	s.clients.add(client)
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()
		defer s.clients.remove(client)

		var err error
		defer func() { client.Close(err) }()

		for {
			var pkt Packet

			pkt, err = s.receivePacket(client)
			if err != nil {
				return
			}
			if pkt == nil {
				continue
			}
		}
	}()

	s.hooks.onConnectionOpened(s, client.Connection.listener)
}

func (s *Server) receivePacket(c *Client) (p Packet, err error) {
	buf := s.readBufPool.Get().(*bufio.Reader)
	defer s.readBufPool.Put(buf)

	buf.Reset(c.Connection.netConn)
	c.refreshDeadline()

	return c.receivePacket(buf)
}
