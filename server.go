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

// ErrServerNotRunning indicates that the Server is not running.
var ErrServerNotRunning = errors.New("server not running")

// ServerState represents the current state of the Server.
type ServerState uint32

// Server is a MQTT broker responsible for implementing the MQTT 3.1, 3.1.1, and 5.0 specifications. To create a Server
// instance, use the NewServer or NewServerWithOptions factory functions.
type Server struct {
	config      Config
	store       store
	listeners   *listeners
	hooks       *hooks
	readersPool sync.Pool
	ctx         context.Context
	cancelCtx   context.CancelCauseFunc
	state       atomic.Uint32
	wg          sync.WaitGroup
	stopOnce    sync.Once
}

// NewServer creates a new Server. If no options functions are provided, the server is created with the default
// options.
func NewServer(f ...OptionsFunc) (s *Server, err error) {
	opts := NewDefaultOptions()
	for _, opt := range f {
		opt(opts)
	}
	return NewServerWithOptions(opts)
}

// NewServerWithOptions creates a new Server. If the Options parameter is not provided, it uses the default options.
// If the provided Options does not contain the Config, it uses the default configuration.
func NewServerWithOptions(opts *Options) (s *Server, err error) {
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
		readersPool: sync.Pool{
			New: func() any { return bufio.NewReaderSize(nil, s.config.ReadBufferSize) },
		},
	}

	for _, l := range opts.Listeners {
		s.listeners.add(l)
	}

	for _, h := range opts.Hooks {
		err = s.hooks.add(h)
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

// AddListener adds the provided Listener to the list of listeners managed by the Server.
//
// Once a listener is added into the server, the server manages the listener by calling the Listen method on
// start, and close method on stop. If the server is running at the time this method is called, the server
// calls the Listen method immediately.
//
// For listeners which should not be managed by the server, don't add them into the server and call the Serve
// method for each Connection to be served.
func (s *Server) AddListener(l Listener) error {
	if s.State() == ServerRunning {
		err := l.Listen(s.Serve)
		if err != nil {
			return err
		}
	}
	s.listeners.add(l)
	return nil
}

// AddHook adds the provided Hook into the list of hooks managed by the Server.
//
// If the server is already running at the time this method is called, the server calls immediately the OnStart
// hook, if this hook implements it. If the OnStart hook returns an error, the hook is not added into the server
// and the error is returned.
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
//
// During the start of the server, the hooks OnServerStart and OnStart are called. If any of these hooks return
// an error, the start process fails and the OnServerStartFailed hook is called.
//
// If the server is not in the ServerNotStarted or ServerStopped state, it returns ErrInvalidServerState.
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

	err = s.listeners.listenAll(s.Serve)
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
//
// This method blocks until the server has been stopped or the context has been cancelled. When the server
// is stopped, it can be started again.
//
// If the server is not in the ServerRunning state, this function has no side effect.
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

// Close closes the Server without waiting for a graceful stop.
//
// When the Server is closed, it cannot be started again. If the server is in ServerRunning state, it stops
// the server first before close it.
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

// Serve serves the provided Connection.
//
// This method does not block the caller and returns an error only if the server is not able to serve the
// connection.
//
// When this method is called, the server creates a Client, associate the client with the connection, and
// returns immediately. If this method is called while the server is not running, it returns ErrServerNotRunning.
func (s *Server) Serve(c *Connection) error {
	if s.State() != ServerRunning {
		return ErrServerNotRunning
	}
	if c == nil || c.Listener == nil || c.netConn == nil {
		return ErrInvalidConnection
	}

	c.KeepAliveMs = s.config.ConnectTimeoutMs

	if err := s.hooks.onConnectionOpen(c); err != nil {
		_ = c.netConn.Close()
		s.hooks.onConnectionClosed(c, err)
		return err
	}

	client := &Client{Connection: c}
	s.wg.Add(2)

	inboundCtx, cancelInboundCtx := context.WithCancel(context.Background())
	outboundCtx, cancelOutboundCtx := context.WithCancelCause(s.ctx)

	// The inbound goroutine is responsible for receive and handle the received packets from client.
	go func() {
		defer s.wg.Done()
		defer cancelInboundCtx()

		err := s.inboundLoop(client)
		if err != nil {
			cancelOutboundCtx(err)
		}
	}()

	// The outbound goroutine is responsible to send outbound packets to client.
	go func() {
		defer s.wg.Done()

		err := s.outboundLoop(outboundCtx)
		if err != nil {
			s.hooks.onClientClose(client, err)
			_ = client.Connection.netConn.Close()
			<-inboundCtx.Done()

			conn := client.Connection
			client.Connection = nil
			s.hooks.onConnectionClosed(conn, err)
		}
	}()

	s.hooks.onClientOpened(client)
	return nil
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

func (s *Server) inboundLoop(c *Client) error {
	for {
		var p Packet

		err := s.hooks.onReceivePacket(c)
		if err != nil {
			return err
		}

		p, err = s.receivePacket(c)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, os.ErrDeadlineExceeded) {
				return err
			}
			s.hooks.onPacketReceiveFailed(c, err)
			return err
		}

		err = s.hooks.onPacketReceived(c, p)
		if err != nil {
			return err
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
	r := s.readersPool.Get().(*bufio.Reader)
	defer s.readersPool.Put(r)

	var h packet.FixedHeader

	_, err = c.Connection.readFixedHeader(r, &h)
	if err != nil {
		return nil, fmt.Errorf("failed to read fixed header: %w", err)
	}

	err = s.hooks.onPacketReceive(c, h)
	if err != nil {
		return nil, err
	}

	p, _, err = c.Connection.receivePacket(r, h)
	if err != nil {
		return nil, fmt.Errorf("failed to read packet: %w", err)
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
		err = s.store.getSession(connect.ClientID, &c.Session)
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
