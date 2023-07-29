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
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gsalomao/akira/packet"
)

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

// ErrInvalidServerState indicates that the Server is not in a valid state for the operation.
var ErrInvalidServerState = errors.New("invalid server state")

// ErrServerStopped indicates that the Server has stopped.
var ErrServerStopped = errors.New("server stopped")

// ServerState represents the current state of the Server.
type ServerState uint32

// Server is a MQTT broker responsible for implementing the MQTT 3.1, 3.1.1, and 5.0 specifications.
// To create a Server instance, use the NewServer factory method.
type Server struct {
	config      Config
	store       store
	listeners   *listeners
	hooks       *hooks
	readBufPool sync.Pool
	ctx         context.Context
	cancelCtx   context.CancelCauseFunc
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
		store:     store{sessionStore: opts.SessionStore},
		listeners: newListeners(),
		hooks:     newHooks(),
		readBufPool: sync.Pool{
			New: func() any { return bufio.NewReaderSize(nil, s.config.ReadBufferSize) },
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
		err := l.Listen(s.handleConnection)
		if err != nil {
			return err
		}
	}
	s.listeners.add(l)
	return nil
}

// AddHook adds the provided Hook into the list of hooks managed by the Server.
// If the Server is already running at the time this function is called, the Server calls immediately the OnStart hook
// method, if this method is implemented by the Hook.
// If the OnStart hook returns an error, the Hook is not added into the Server and the error is returned.
func (s *Server) AddHook(h Hook) error {
	if s.State() == ServerRunning {
		if hook, ok := h.(OnStartHook); ok {
			err := hook.OnStart()
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
func (s *Server) Start() error {
	state := s.State()
	if state != ServerNotStarted && state != ServerStopped {
		return ErrInvalidServerState
	}

	err := s.setState(ServerStarting, nil)
	if err != nil {
		_ = s.setState(ServerFailed, err)
		return err
	}

	s.stopOnce = sync.Once{}
	s.ctx, s.cancelCtx = context.WithCancelCause(context.Background())

	err = s.listeners.listenAll(s.handleConnection)
	if err != nil {
		_ = s.setState(ServerFailed, err)
		return err
	}

	s.wg.Add(1)
	readyCh := make(chan struct{})
	s.stopOnce = sync.Once{}

	go func() {
		defer s.wg.Done()
		close(readyCh)

		s.backgroundLoop(s.ctx)
	}()

	<-readyCh
	_ = s.setState(ServerRunning, nil)
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
		// TODO: Make sure that the server doesn't end up in the stopped state when this goroutine sets the state to
		// stopped after the server has been closed.
		_ = s.setState(ServerStopped, nil)
		close(stoppedCh)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-stoppedCh:
		return nil
	}
}

// Close closes the Server.
// If the Server is in ServerRunning state, it stops the server first before close it.
// When the Server is closed, it cannot be started again.
func (s *Server) Close() {
	switch s.State() {
	case ServerNotStarted, ServerStopped, ServerFailed:
		_ = s.setState(ServerClosed, nil)
		return
	case ServerRunning:
		s.stop()
		_ = s.setState(ServerStopped, nil)
	case ServerClosed:
		return
	default:
	}
	_ = s.setState(ServerClosed, nil)
}

// State returns the current ServerState.
func (s *Server) State() ServerState {
	return ServerState(s.state.Load())
}

func (s *Server) setState(state ServerState, err error) error {
	s.state.Store(uint32(state))
	switch state {
	case ServerStarting:
		err = s.hooks.onServerStart()
		if err == nil {
			err = s.hooks.onStart()
		}
	case ServerStopping:
		s.hooks.onServerStop()
		s.hooks.onStop()
	case ServerStopped:
		s.hooks.onServerStopped()
	case ServerRunning:
		s.hooks.onServerStarted()
	case ServerFailed:
		s.hooks.onServerStartFailed(err)
	default:
		err = ErrInvalidServerState
	}
	return err
}

func (s *Server) stop() {
	s.stopOnce.Do(func() {
		_ = s.setState(ServerStopping, nil)
		s.listeners.closeAll()
		s.cancelCtx(ErrServerStopped)
	})
}

func (s *Server) handleConnection(l Listener, nc net.Conn) {
	c := &Client{
		Connection: Connection{
			netConn:   nc,
			listener:  l,
			KeepAlive: s.config.ConnectTimeout,
		},
	}

	if err := s.hooks.onClientOpen(s, c); err != nil {
		_ = c.Connection.netConn.Close()
		s.hooks.onClientClosed(s, c, err)
		return
	}

	ctx, cancelCtx := context.WithCancelCause(s.ctx)
	s.wg.Add(2)

	// The inbound goroutine is responsible for receive and handle the received packets from client.
	go func() {
		defer s.wg.Done()

		err := s.inboundLoop(c)
		if err != nil {
			cancelCtx(err)
		}
	}()

	// The outbound goroutine is responsible to send outbound packets to client.
	go func() {
		defer s.wg.Done()

		err := s.outboundLoop(ctx)
		if err != nil {
			_ = c.Connection.netConn.Close()
			<-ctx.Done()
			s.hooks.onClientClosed(s, c, err)
		}
	}()
}

func (s *Server) inboundLoop(c *Client) error {
	for {
		p, err := s.receivePacket(c)
		if err != nil {
			return err
		}
		if p == nil {
			continue
		}

		err = s.handlePacket(c, p)
		if err != nil {
			return err
		}
	}
}

func (s *Server) outboundLoop(ctx context.Context) error {
	<-ctx.Done()
	return context.Cause(ctx)
}

func (s *Server) backgroundLoop(ctx context.Context) {
	<-ctx.Done()
}

func (s *Server) receivePacket(c *Client) (p Packet, err error) {
	err = s.hooks.onPacketReceive(c)
	if err != nil {
		return nil, err
	}

	buf := s.readBufPool.Get().(*bufio.Reader)
	defer s.readBufPool.Put(buf)

	p, _, err = c.Connection.receivePacket(r)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, io.EOF
		}
		if errors.Is(err, os.ErrDeadlineExceeded) {
			return nil, err
		}
		err = fmt.Errorf("failed to read packet: %w", err)
		s.hooks.onPacketReceiveFailed(c, err)
		return nil, err
	}

	err = s.hooks.onPacketReceived(c, p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (s *Server) handlePacket(c *Client, p Packet) error {
	switch pkt := p.(type) {
	case *packet.Connect:
		return s.handlePacketConnect(c, pkt)
	default:
		return errors.New("unsupported packet")
	}
}

func (s *Server) handlePacketConnect(c *Client, connect *packet.Connect) error {
	// The first step must be to set the version of the MQTT connection as this information is required for further
	// processing.
	c.Connection.Version = connect.Version

	err := s.hooks.onConnect(c, connect)
	if err != nil {
		var pktErr packet.Error
		if errors.As(err, &pktErr) {
			if pktErr.Code != packet.ReasonCodeSuccess && pktErr.Code != packet.ReasonCodeMalformedPacket {
				_ = s.sendConnAck(c, pktErr.Code, false, nil)
			}
		}
		return err
	}

	err = s.connectClient(c, connect)
	if err != nil {
		s.hooks.onConnectFailed(c, connect, err)
		return err
	}

	s.hooks.onConnected(c)
	return nil
}

func (s *Server) connectClient(c *Client, connect *packet.Connect) error {
	var (
		sessionPresent bool
		err            error
	)

	if !isClientIDValid(connect.Version, len(connect.ClientID), &s.config) {
		_ = s.sendConnAck(c, packet.ReasonCodeClientIDNotValid, false, nil)
		return packet.ErrClientIDNotValid
	}

	if !isKeepAliveValid(connect.Version, connect.KeepAlive, s.config.MaxKeepAliveSec) {
		// For MQTT v3.1 and v3.1.1, there is no mechanism to tell the clients what Keep Alive value they should use.
		// If an MQTT v3.1 or v3.1.1 client specifies a Keep Alive time greater than maximum keep alive, the CONNACK
		// packet shall be sent with the reason code "identifier rejected".
		_ = s.sendConnAck(c, packet.ReasonCodeClientIDNotValid, false, nil)
		return packet.ErrClientIDNotValid
	}

	// If the client requested a clean start, the server must delete any existing session for the given client
	// identifier. Otherwise, resume any existing session for the given client identifier.
	if connect.Flags.CleanStart() {
		err = s.store.deleteSession(connect.ClientID)
	} else {
		err = s.store.readSession(connect.ClientID, &c.Session)
		if err == nil {
			sessionPresent = true
		}
	}
	if err != nil && !errors.Is(err, ErrSessionNotFound) {
		// If the session store fails to get the session, or to delete the session, the server replies to client
		// indicating that the service is unavailable.
		pErr := packet.ErrServerUnavailable
		_ = s.sendConnAck(c, pErr.Code, false, nil)
		return fmt.Errorf("%w: %s", pErr, err.Error())
	}

	c.ID = connect.ClientID
	c.Session.Connected = true
	c.Session.ConnectedAt = time.Now().UnixMilli()
	c.Session.Properties = sessionProperties(connect.Properties, &s.config)
	c.Session.LastWill = lastWill(connect)
	c.Connection.KeepAliveMs = uint32(sessionKeepAlive(connect.KeepAlive, s.config.MaxKeepAliveSec)) * 1000

	if (c.Connection.Version != packet.MQTT50 && !connect.Flags.CleanStart()) || sessionExpiryInterval(&c.Session) > 0 {
		err = s.store.saveSession(c.ID, &c.Session)
		if err != nil {
			// If the session store fails to save the session, the server replies to client indicating that the
			// service is unavailable.
			pErr := packet.ErrServerUnavailable
			_ = s.sendConnAck(c, pErr.Code, false, nil)
			return fmt.Errorf("%w: %s", pErr, err.Error())
		}
	}

	err = s.sendConnAck(c, packet.ReasonCodeSuccess, sessionPresent, connect)
	if err != nil {
		return err
	}

	c.connected.Store(true)
	return nil
}

func (s *Server) sendConnAck(c *Client, code packet.ReasonCode, present bool, connect *packet.Connect) error {
	var props *packet.ConnAckProperties

	if code == packet.ReasonCodeClientIDNotValid && c.Connection.Version != packet.MQTT50 {
		code = packet.ReasonCodeV3IdentifierRejected
	}
	if code == packet.ReasonCodeServerUnavailable && c.Connection.Version != packet.MQTT50 {
		code = packet.ReasonCodeV3ServerUnavailable
	}
	if code == packet.ReasonCodeSuccess && c.Connection.Version == packet.MQTT50 {
		props = connAckProperties(c, &s.config, connect)
	}
	connack := packet.ConnAck{Version: c.Connection.Version, Code: code, SessionPresent: present, Properties: props}
	return s.sendPacket(c, &connack)
}

func (s *Server) sendPacket(c *Client, p PacketEncodable) error {
	err := s.hooks.onPacketSend(c, p)
	if err != nil {
		return err
	}

	_, err = c.Connection.sendPacket(p)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			s.hooks.onPacketSendFailed(c, p, err)
		}
		return err
	}

	s.hooks.onPacketSent(c, p)
	return nil
}
