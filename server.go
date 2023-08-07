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

// String returns a human-friendly name for the ServerState.
func (s ServerState) String() string {
	switch s {
	case ServerNotStarted:
		return "Not Started"
	case ServerStarting:
		return "Starting"
	case ServerRunning:
		return "Running"
	case ServerStopping:
		return "Stopping"
	case ServerStopped:
		return "Stopped"
	case ServerFailed:
		return "Failed"
	case ServerClosed:
		return "Closed"
	default:
		return "Invalid"
	}
}

// Server is a MQTT broker responsible for implementing the MQTT 3.1, 3.1.1, and 5.0 specifications. To create a Server
// instance, use the NewServer or NewServerWithOptions factory functions.
type Server struct {
	// Metrics provides the server metrics.
	Metrics Metrics

	config       Config
	logger       Logger
	sessionStore SessionStore
	listeners    *listeners
	hooks        *hooks
	readersPool  sync.Pool
	ctx          context.Context
	cancelCtx    context.CancelCauseFunc
	state        atomic.Uint32
	wg           sync.WaitGroup
	stopOnce     sync.Once
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
	if opts.Logger == nil {
		opts.Logger = &noOpLogger{}
	}
	if opts.SessionStore == nil {
		opts.SessionStore = newInMemorySessionStore()
	}

	s = &Server{
		config:       *opts.Config,
		sessionStore: opts.SessionStore,
		logger:       opts.Logger,
		listeners:    newListeners(),
		hooks:        newHooks(),
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
	st := s.State()
	s.logger.Log("Adding listener", "listeners", s.listeners.len(), "state", st.String())

	if st == ServerRunning {
		err := l.Listen(s.Serve)
		if err != nil {
			s.logger.Log("Failed to start listener", "error", err, "listeners", s.listeners.len())
			return err
		}
	}
	s.listeners.add(l)
	s.logger.Log("Listener added with success", "listeners", s.listeners.len(), "state", st.String())
	return nil
}

// AddHook adds the provided Hook into the list of hooks managed by the Server.
//
// If the server is already running at the time this method is called, the server calls immediately the OnStart
// hook, if this hook implements it. If the OnStart hook returns an error, the hook is not added into the server
// and the error is returned.
func (s *Server) AddHook(h Hook) error {
	st := s.State()
	s.logger.Log("Adding hook", "hook", h.Name(), "hooks", s.hooks.len(), "state", st.String())

	if s.State() == ServerRunning {
		if hook, ok := h.(OnStartHook); ok {
			err := hook.OnStart()
			if err != nil {
				s.logger.Log("Failed to start hook", "error", err, "hook", h.Name(), "hooks", s.hooks.len())
				return err
			}
		}
	}

	err := s.hooks.add(h)
	if err != nil {
		s.logger.Log("Failed to add hook", "error", err, "hook", h.Name(), "hooks", s.hooks.len())
		return err
	}

	s.logger.Log("Hook added with success", "hook", h.Name(), "hooks", s.hooks.len(), "state", st.String())
	return nil
}

// Start starts the Server and returns immediately.
//
// During the start of the server, the hooks OnServerStart and OnStart are called. If any of these hooks return
// an error, the start process fails and the OnServerStartFailed hook is called.
//
// If the server is not in the ServerNotStarted or ServerStopped state, it returns ErrInvalidServerState.
func (s *Server) Start() error {
	st := s.State()
	s.logger.Log("Starting server", "state", st.String())

	if st != ServerNotStarted && st != ServerStopped {
		s.logger.Log("Failed start server due to invalid start", "state", st.String())
		return ErrInvalidServerState
	}

	err := s.setState(ServerStarting, nil)
	if err != nil {
		s.logger.Log("Failed to set server to starting state", "error", err, "state", st.String())

		// The setState method never returns error when setting to ServerFailed.
		_ = s.setState(ServerFailed, err)
		return err
	}

	s.stopOnce = sync.Once{}
	s.ctx, s.cancelCtx = context.WithCancelCause(context.Background())

	err = s.listeners.listenAll(s.Serve)
	if err != nil {
		s.logger.Log("Failed to start listeners", "error", err, "state", st.String())

		// The setState method never returns error when setting to ServerFailed.
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

	// The setState method never returns error when setting to ServerRunning.
	_ = s.setState(ServerRunning, nil)
	s.logger.Log("Server started with success", "state", s.State().String())

	return nil
}

// Stop stops the Server gracefully.
//
// This method blocks until the server has been stopped or the context has been cancelled. When the server
// is stopped, it can be started again.
//
// If the server is not in the ServerRunning state, this function has no side effect.
func (s *Server) Stop(ctx context.Context) error {
	st := s.State()
	s.logger.Log("Stopping server", "state", st.String())

	if st != ServerRunning {
		s.logger.Log("Server not running", "state", st.String())
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
		s.logger.Log("Stop server cancelled", "state", s.State().String())
		return ctx.Err()
	case <-stoppedCh:
		s.logger.Log("Server stopped with success", "state", s.State().String())
		return nil
	}
}

// Close closes the Server without waiting for a graceful stop.
//
// When the Server is closed, it cannot be started again. If the server is in ServerRunning state, it stops
// the server first before close it.
func (s *Server) Close() {
	st := s.State()
	s.logger.Log("Closing server", "state", st.String())

	if st == ServerClosed {
		s.logger.Log("Server already closed", "state", st.String())
		return
	}
	if st == ServerRunning {
		s.stop()
		// The setState method never returns error when setting to ServerStopped.
		_ = s.setState(ServerStopped, nil)
	}

	// The setState method never returns error when setting to ServerClosed.
	_ = s.setState(ServerClosed, nil)
	s.logger.Log("Server closed with success", "state", s.State().String())
}

// Serve serves the provided Connection.
//
// This method does not block the caller and returns an error only if the server is not able to serve the
// connection.
//
// When this method is called, the server creates a Client, associate the client with the connection, and
// returns immediately. If this method is called while the server is not running, it returns ErrServerNotRunning.
func (s *Server) Serve(c *Connection) error {
	st := s.State()
	if c == nil || c.Listener == nil || c.netConn == nil {
		s.logger.Log("Cannot serve connection as it is invalid", "connection", c, "state", st.String())
		return ErrInvalidConnection
	}
	if st != ServerRunning {
		s.logger.Log("Cannot serve connection as server not running",
			"address", c.Address,
			"state", st.String(),
		)
		return ErrServerNotRunning
	}

	c.Address = c.netConn.RemoteAddr().String()
	c.KeepAliveMs = s.config.ConnectTimeoutMs
	c.sendTimeoutMs = s.config.SendPacketTimeoutMs

	clientCtx, cancelClientCtx := context.WithCancelCause(s.ctx)

	if err := s.hooks.onConnectionOpen(clientCtx, c); err != nil {
		_ = c.netConn.Close()
		s.hooks.onConnectionClosed(c, err)
		s.logger.Log("Cannot serve connection due to an error from OnConnectionOpen",
			"address", c.Address,
			"error", err,
		)
		return err
	}

	client := &Client{Connection: c}
	s.wg.Add(2)

	// The inbound context must be detached from the server context to allow the outbound goroutine to wait for the
	// inbound goroutine to finish before it can finish.
	inboundCtx, cancelInboundCtx := context.WithCancel(context.Background())

	// The inbound goroutine is responsible for receive and handle the received packets from client.
	go func() {
		defer s.wg.Done()

		s.logger.Log("Running inbound loop", "address", c.Address)
		err := s.inboundLoop(clientCtx, client)

		cancelClientCtx(err)
		cancelInboundCtx()

		s.logger.Log("Stopping inbound loop",
			"address", c.Address,
			"id", string(client.ID),
			"state", s.State().String(),
			"version", c.Version.String(),
		)
	}()

	// The outbound goroutine is responsible to send outbound packets to client.
	go func() {
		defer s.wg.Done()

		s.logger.Log("Running outbound loop", "address", c.Address)
		err := s.outboundLoop(clientCtx)
		s.hooks.onClientClose(client, err)

		_ = client.Connection.netConn.Close()
		<-inboundCtx.Done()

		if errors.Is(err, io.EOF) || errors.Is(err, ErrServerStopped) {
			s.logger.Log("Client closed",
				"address", c.Address,
				"id", string(client.ID),
				"state", s.State().String(),
				"version", c.Version.String(),
			)
		} else {
			s.logger.Log("Client closed due to an error",
				"address", c.Address,
				"error", err,
				"id", string(client.ID),
				"state", s.State().String(),
				"version", c.Version.String(),
			)
		}

		conn := client.Connection
		client.Connection = nil
		s.hooks.onConnectionClosed(conn, err)

		s.logger.Log("Stopping outbound loop",
			"address", c.Address,
			"id", string(client.ID),
			"state", s.State().String(),
			"version", c.Version.String(),
		)
	}()

	s.hooks.onClientOpened(clientCtx, client)
	s.logger.Log("Serving connection", "address", c.Address)
	return nil
}

// State returns the current ServerState.
func (s *Server) State() ServerState {
	return ServerState(s.state.Load())
}

func (s *Server) setState(st ServerState, err error) error {
	s.state.Store(uint32(st))
	switch st {
	case ServerStarting:
		err = s.hooks.onServerStart()
		if err == nil {
			err = s.hooks.onStart()
		}
	case ServerRunning:
		s.hooks.onServerStarted()
	case ServerStopping:
		s.hooks.onServerStop()
		s.hooks.onStop()
	case ServerStopped:
		s.hooks.onServerStopped()
	case ServerFailed:
		s.hooks.onServerStartFailed(err)
	case ServerClosed:
		// The closed state doesn't have any hook.
	default:
		panic(fmt.Sprintf("invalid server state: %v", st))
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

func (s *Server) inboundLoop(ctx context.Context, c *Client) error {
	for {
		var p Packet

		err := s.hooks.onReceivePacket(ctx, c)
		if err != nil {
			return err
		}

		p, err = s.receivePacket(ctx, c)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, os.ErrDeadlineExceeded) {
				return err
			}
			s.hooks.onPacketReceiveFailed(ctx, c, err)
			return err
		}

		err = s.hooks.onPacketReceived(ctx, c, p)
		if err != nil {
			return err
		}

		err = s.handlePacket(ctx, c, p)
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

func (s *Server) receivePacket(ctx context.Context, c *Client) (Packet, error) {
	var h packet.FixedHeader

	r := s.readersPool.Get().(*bufio.Reader)
	defer s.readersPool.Put(r)

	headerSize, err := c.Connection.readFixedHeader(r, &h)
	if err != nil {
		return nil, fmt.Errorf("failed to read fixed header: %w", err)
	}

	pSize := headerSize + h.RemainingLength
	if s.config.MaxPacketSize > 0 && pSize > int(s.config.MaxPacketSize) {
		return nil, ErrPacketSizeExceeded
	}

	err = s.hooks.onPacketReceive(ctx, c, h)
	if err != nil {
		return nil, err
	}

	var (
		p Packet
		n int
	)

	p, n, err = c.Connection.receivePacket(r, h)
	if err != nil {
		return nil, fmt.Errorf("failed to read packet: %w", err)
	}

	pSize = headerSize + n
	s.Metrics.PacketReceived.value.Add(1)
	s.Metrics.BytesReceived.value.Add(uint64(pSize))

	s.logger.Log("Received packet from client",
		"address", c.Connection.Address,
		"id", string(c.ID),
		"packet", p.Type().String(),
		"size", pSize,
		"version", c.Connection.Version.String(),
	)
	return p, nil
}

func (s *Server) handlePacket(ctx context.Context, c *Client, p Packet) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	switch pkt := p.(type) {
	case *packet.Connect:
		return s.handlePacketConnect(ctx, c, pkt)
	default:
		return errors.New("unsupported packet")
	}
}

func (s *Server) handlePacketConnect(ctx context.Context, c *Client, connect *packet.Connect) error {
	if c.Connected() {
		s.logger.Log("Duplicate CONNECT packet",
			"address", c.Connection.Address,
			"id", string(c.ID),
			"version", connect.Version.String(),
		)

		if connect.Version == packet.MQTT50 {
			// As the client is going to be closed, there's nothing else to do with the error returned from the
			// sendConnAck, as it was already logged.
			_ = s.sendConnAck(ctx, c, packet.ReasonCodeProtocolError, false, nil)
		}
		return fmt.Errorf("%w: duplicate CONNECT packet", packet.ErrProtocolError)
	}

	// The first step must be to set the version of the MQTT connection as this information is required for further
	// processing.
	c.Connection.Version = connect.Version

	err := s.hooks.onConnect(ctx, c, connect)
	if err != nil {
		s.logger.Log("Stopping to connect client due to an error from OnConnect",
			"address", c.Connection.Address,
			"error", err,
			"version", connect.Version.String(),
		)

		var pktErr packet.Error
		if errors.As(err, &pktErr) {
			if pktErr.Code != packet.ReasonCodeSuccess && pktErr.Code != packet.ReasonCodeMalformedPacket {
				// As the client is going to be closed, there's nothing else to do with the error returned from the
				// sendConnAck, as it was already logged.
				_ = s.sendConnAck(ctx, c, pktErr.Code, false, nil)
			}
		}
		return err
	}

	err = s.connectClient(ctx, c, connect)
	if err != nil {
		s.hooks.onConnectFailed(ctx, c, connect, err)
		return err
	}

	s.hooks.onConnected(ctx, c)
	return nil
}

func (s *Server) connectClient(ctx context.Context, c *Client, connect *packet.Connect) error {
	var (
		sessionPresent bool
		err            error
	)

	if !isClientIDValid(connect.Version, len(connect.ClientID), &s.config) {
		s.logger.Log("Failed to connect client due to invalid client ID",
			"address", c.Connection.Address,
			"id", string(connect.ClientID),
			"version", connect.Version.String(),
		)

		_ = s.sendConnAck(ctx, c, packet.ReasonCodeClientIDNotValid, false, nil)
		return packet.ErrClientIDNotValid
	}

	if !isKeepAliveValid(connect.Version, connect.KeepAlive, s.config.MaxKeepAliveSec) {
		s.logger.Log("Failed to connect client due to invalid keep alive",
			"address", c.Connection.Address,
			"id", string(connect.ClientID),
			"keep_alive", connect.KeepAlive,
			"version", connect.Version.String(),
		)

		// For MQTT v3.1 and v3.1.1, there is no mechanism to tell the clients what Keep Alive value they should use.
		// If an MQTT v3.1 or v3.1.1 client specifies a Keep Alive time greater than maximum keep alive, the CONNACK
		// packet shall be sent with the reason code "identifier rejected".
		_ = s.sendConnAck(ctx, c, packet.ReasonCodeClientIDNotValid, false, nil)
		return packet.ErrClientIDNotValid
	}

	// If the client requested a clean start, the server must delete any existing session for the given client
	// identifier. Otherwise, resume any existing session for the given client identifier.
	if connect.Flags.CleanStart() {
		err = s.sessionStore.DeleteSession(ctx, connect.ClientID)
	} else {
		err = s.sessionStore.GetSession(ctx, connect.ClientID, &c.Session)
		if err == nil {
			if c.Session.Expired() {
				err = s.sessionStore.DeleteSession(ctx, connect.ClientID)
			} else {
				sessionPresent = true
			}
		}
	}
	if err != nil && !errors.Is(err, ErrSessionNotFound) {
		s.logger.Log("Failed to get/delete session",
			"address", c.Connection.Address,
			"clean_start", connect.Flags.CleanStart(),
			"id", string(connect.ClientID),
			"error", err,
			"version", connect.Version.String(),
		)

		// If the session store fails to get the session, or to delete the session, the server replies to client
		// indicating that the service is unavailable.
		pErr := packet.ErrServerUnavailable

		_ = s.sendConnAck(ctx, c, pErr.Code, false, nil)
		return fmt.Errorf("%w: %s", pErr, err.Error())
	}

	c.ID = connect.ClientID
	c.Session.ConnectedAt = time.Now().UnixMilli()
	c.Session.Properties = sessionProperties(connect.Properties, &s.config)
	c.Session.LastWill = lastWill(connect)
	c.Connection.KeepAliveMs = uint32(sessionKeepAlive(connect.KeepAlive, s.config.MaxKeepAliveSec)) * 1000

	persistentSession := isPersistentSession(c.Connection.Version, connect.Flags.CleanStart(),
		sessionExpiryInterval(&c.Session))

	if persistentSession {
		err = s.sessionStore.SaveSession(ctx, c.ID, &c.Session)
		if err != nil {
			s.logger.Log("Failed save session",
				"address", c.Connection.Address,
				"error", err,
				"id", string(c.ID),
				"session_present", sessionPresent,
				"version", connect.Version.String(),
			)

			// If the session store fails to save the session, the server replies to client indicating that the
			// service is unavailable.
			pErr := packet.ErrServerUnavailable

			_ = s.sendConnAck(ctx, c, pErr.Code, false, nil)
			return fmt.Errorf("%w: %s", pErr, err.Error())
		}
	}

	err = s.sendConnAck(ctx, c, packet.ReasonCodeSuccess, sessionPresent, connect)
	if err != nil {
		return err
	}

	c.connected.Store(true)
	s.Metrics.ClientsConnected.value.Add(1)

	if persistentSession {
		s.Metrics.PersistentSessions.value.Add(1)
	}

	s.logger.Log("Client connected with success",
		"address", c.Connection.Address,
		"clean_start", connect.Flags.CleanStart(),
		"id", string(c.ID),
		"keep_alive", c.Connection.KeepAliveMs/1000,
		"persistent_session", persistentSession,
		"session_present", sessionPresent,
		"version", c.Connection.Version.String(),
	)
	return nil
}

func (s *Server) sendConnAck(
	ctx context.Context, c *Client, code packet.ReasonCode, present bool, connect *packet.Connect,
) error {
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
	return s.sendPacket(ctx, c, &connack)
}

func (s *Server) sendPacket(ctx context.Context, c *Client, p PacketEncodable) error {
	err := s.hooks.onPacketSend(ctx, c, p)
	if err != nil {
		s.logger.Log("Stopping to send packet to client due to an error from OnPacketSend",
			"address", c.Connection.Address,
			"error", err,
			"id", string(c.ID),
			"packet", p.Type().String(),
		)
		return err
	}

	var pSize int

	pSize, err = c.Connection.sendPacket(p)
	if err != nil {
		s.logger.Log("Failed to send packet to client",
			"address", c.Connection.Address,
			"error", err,
			"id", string(c.ID),
			"packet", p.Type().String(),
			"version", c.Connection.Version.String(),
		)

		if !errors.Is(err, io.EOF) {
			s.hooks.onPacketSendFailed(ctx, c, p, err)
		}
		return err
	}

	s.hooks.onPacketSent(ctx, c, p)

	s.Metrics.PacketSent.value.Add(1)
	s.Metrics.BytesSent.value.Add(uint64(pSize))

	s.logger.Log("Packet sent to client",
		"address", c.Connection.Address,
		"id", string(c.ID),
		"packet", p.Type().String(),
		"size", p.Size(),
		"version", c.Connection.Version.String(),
	)
	return nil
}
