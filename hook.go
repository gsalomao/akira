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
	"sync"
	"sync/atomic"

	"github.com/gsalomao/akira/packet"
)

const (
	onStartHook hookType = iota
	onStopHook
	onServerStartHook
	onServerStartFailedHook
	onServerStartedHook
	onServerStopHook
	onServerStoppedHook
	onConnectionOpenHook
	onClientOpenedHook
	onClientCloseHook
	onConnectionClosedHook
	onReceivePacketHook
	onPacketReceiveHook
	onPacketReceiveFailedHook
	onPacketReceivedHook
	onPacketSendHook
	onPacketSendFailedHook
	onPacketSentHook
	onConnectPacketHook
	onConnectFailedHook
	onConnectedHook
	onAuthPacketHook
	maxHooks
)

// ErrHookAlreadyExists indicates that the Hook already exists based on its name.
var ErrHookAlreadyExists = errors.New("hook already exists")

type hookType int

// Hook is the minimal interface which any hook must implement. All others hook interfaces are optional.
// During the lifecycle of the Server, several events are generated which can be handled by the hook
// based on each method it implements.
type Hook interface {
	// Name returns the name of the hook. The Server supports only one hook for each name. If a hook is added into the
	// server with the same name of another hook, the ErrHookAlreadyExists is returned.
	Name() string
}

// OnStartHook is the hook interface that wraps the OnStart method. The OnStart method is called by the server to
// start the hook, and it's called after the server has started. If this method returns any error, the server
// considers that the hook failed to start.
type OnStartHook interface {
	OnStart(ctx context.Context) error
}

// OnStopHook is the hook interface that wraps the OnStop method. The OnStop method is called by the server to stop
// the hook.
type OnStopHook interface {
	OnStop(ctx context.Context)
}

// OnServerStartHook is the hook interface that wraps the OnServerStart method. The OnServerStart method is called by
// the server when it is starting. When this method is called, the server is in ServerStarting state. If this method
// returns any error, the start process fails.
type OnServerStartHook interface {
	OnServerStart(ctx context.Context) error
}

// OnServerStartFailedHook is the hook interface that wraps the OnServerStartFailed method. The OnServerStartFailed
// method is called by the server when it has failed to start.
type OnServerStartFailedHook interface {
	OnServerStartFailed(ctx context.Context, err error)
}

// OnServerStartedHook is the hook interface that wraps the OnServerStarted method. The OnServerStarted method is
// called by the server when it has started with success.
type OnServerStartedHook interface {
	OnServerStarted(ctx context.Context)
}

// OnServerStopHook is the hook interface that wraps the OnServerStop method. The OnServerStop method is called by the
// server when it is stopping.
type OnServerStopHook interface {
	OnServerStop(ctx context.Context)
}

// OnServerStoppedHook is the hook interface that wraps the OnServerStopped method. The OnServerStopped method is
// called by the server when it has stopped.
type OnServerStoppedHook interface {
	OnServerStopped(ctx context.Context)
}

// OnConnectionOpenHook is the hook interface that wraps the OnConnectionOpen method. The OnConnectionOpen method is
// called by the server when a new network connection was opened. If this method returns any error, the connection is
// closed.
type OnConnectionOpenHook interface {
	OnConnectionOpen(ctx context.Context, c *Connection) error
}

// OnClientOpenedHook is the hook interface that wraps the OnClientOpened method. The OnClientOpened method is called
// by the server when the client opened the connection, and after the network connection has been opened with success.
type OnClientOpenedHook interface {
	OnClientOpened(ctx context.Context, c *Client)
}

// OnClientCloseHook is the hook interface that wraps the OnClientClose method. The OnClientClose method is called by
// the server when the client is being closed, but before the network connection is closed. If the client is being
// closed due to an error, this error is passed as parameter.
type OnClientCloseHook interface {
	OnClientClose(c *Client, err error)
}

// OnConnectionClosedHook is the hook interface that wraps the OnConnectionClosed method. The OnConnectionClosed
// method is called by the server after the network connection was closed. If the connection was closed due to ab
// error, this error is passed as parameter.
type OnConnectionClosedHook interface {
	OnConnectionClosed(c *Connection, err error)
}

// OnReceivePacketHook is the hook interface that wraps the OnReceivePacket method. The OnReceivePacket method is
// called by the server before try to receive any packet from the client. If this method returns an error, the server
// doesn't try to receive a new packet and the client is closed.
type OnReceivePacketHook interface {
	OnReceivePacket(ctx context.Context, c *Client) error
}

// OnPacketReceiveHook is the hook interface that wraps the OnPacketReceive method. The OnPacketReceive method is
// called by the server when it has received the fixed header of the packet from the client, but it did not receive
// yet the remaining bytes of the packet. If this method returns an  error, the server stops to receive the packet
// and closes the client.
type OnPacketReceiveHook interface {
	OnPacketReceive(ctx context.Context, c *Client, h packet.FixedHeader) error
}

// OnPacketReceiveFailedHook is the hook interface that wraps the OnPacketReceiveFailed method. The
// OnPacketReceiveFailed method is called by the server when it fails to receive a packet from the client for any
// reason other than the network connection closed. After the server calls this method, it closes the client.
type OnPacketReceiveFailedHook interface {
	OnPacketReceiveFailed(ctx context.Context, c *Client, err error)
}

// OnPacketReceivedHook is the hook interface that wraps the OnPacketReceived method. The OnPacketReceived method is
// called by the server when a new packet is received. If this method returns an error, the server discards the
// packet and closes the client.
type OnPacketReceivedHook interface {
	OnPacketReceived(ctx context.Context, c *Client, p Packet) error
}

// OnPacketSendHook is the hook interface that wraps the OnPacketSend method. The OnPacketSend method is called by the
// server before the packet has been sent to the client. If this method returns an error, the server discards the
// packet and closes the client.
type OnPacketSendHook interface {
	OnPacketSend(ctx context.Context, c *Client, p Packet) error
}

// OnPacketSendFailedHook is the hook interface that wraps the OnPacketSendFailed method. The OnPacketSendFailed
// method is called by the server when it fails to send the packet to the client for any reason other than the
// network connection closed. After the server calls this method, it closes the client.
type OnPacketSendFailedHook interface {
	OnPacketSendFailed(ctx context.Context, c *Client, p Packet, err error)
}

// OnPacketSentHook is the hook interface that wraps the OnPacketSent method. The OnPacketSent method is called by the
// server after the packet has been sent to the client.
type OnPacketSentHook interface {
	OnPacketSent(ctx context.Context, c *Client, p Packet)
}

// OnConnectPacketHook is the hook interface that wraps the OnConnectPacket method. The OnConnectPacket method is
// called by the server when it receives a CONNECT Packet. If this method returns an error, the server stops the
// connection process and closes the client. If the returned error is a packet.Error with a reason code other than
// packet.ReasonCodeSuccess and packet.ReasonCodeMalformedPacket, the server sends a CONNACK Packet with the reason
// code before it closes the client.
type OnConnectPacketHook interface {
	OnConnectPacket(ctx context.Context, c *Client, p *packet.Connect) error
}

// OnConnectFailedHook is the hook interface that wraps the OnConnectFailed method. The OnConnectFailed method is
// called by the server when it fails to connect the client. The connection process can fail for different reasons,
// and the error indicating the reason is passed as parameter. After the server calls this method, it closes the
// client.
type OnConnectFailedHook interface {
	OnConnectFailed(ctx context.Context, c *Client, err error)
}

// OnConnectedHook is the hook interface that wraps the OnConnected method. The OnConnected method is called by the
// server when it completes the connection process to connect the client with success.
type OnConnectedHook interface {
	OnConnected(ctx context.Context, c *Client)
}

// OnAuthPacketHook is the hook interface that wraps the OnAuthPacket method. The OnAuthPacket method is called by the
// server when it receives an AUTH Packet. If this method returns an error, the server closes the connection.
// When this method returns an error, no packet is sent to the client before closing the connection, regardless if
// error is a packet.Error or not.
type OnAuthPacketHook interface {
	OnAuthPacket(ctx context.Context, c *Client, p *packet.Auth) error
}

var hooksRegistryFunc = map[hookType]func(*hooks, Hook, hookType){
	onStartHook:               registerHook[OnStartHook],
	onStopHook:                registerHook[OnStopHook],
	onServerStartHook:         registerHook[OnServerStartHook],
	onServerStartFailedHook:   registerHook[OnServerStartFailedHook],
	onServerStartedHook:       registerHook[OnServerStartedHook],
	onServerStopHook:          registerHook[OnServerStopHook],
	onServerStoppedHook:       registerHook[OnServerStoppedHook],
	onConnectionOpenHook:      registerHook[OnConnectionOpenHook],
	onClientOpenedHook:        registerHook[OnClientOpenedHook],
	onClientCloseHook:         registerHook[OnClientCloseHook],
	onConnectionClosedHook:    registerHook[OnConnectionClosedHook],
	onReceivePacketHook:       registerHook[OnReceivePacketHook],
	onPacketReceiveHook:       registerHook[OnPacketReceiveHook],
	onPacketReceiveFailedHook: registerHook[OnPacketReceiveFailedHook],
	onPacketReceivedHook:      registerHook[OnPacketReceivedHook],
	onPacketSendHook:          registerHook[OnPacketSendHook],
	onPacketSendFailedHook:    registerHook[OnPacketSendFailedHook],
	onPacketSentHook:          registerHook[OnPacketSentHook],
	onConnectPacketHook:       registerHook[OnConnectPacketHook],
	onConnectFailedHook:       registerHook[OnConnectFailedHook],
	onConnectedHook:           registerHook[OnConnectedHook],
	onAuthPacketHook:          registerHook[OnAuthPacketHook],
}

func registerHook[T any](h *hooks, hook Hook, t hookType) {
	if _, ok := hook.(T); ok {
		h.hookNames[t][hook.Name()] = struct{}{}
		h.hooks[t] = append(h.hooks[t], hook)
		h.hookPresent[t].Store(true)
	}
}

type hooks struct {
	mutex       sync.RWMutex
	cnt         int
	hookNames   [maxHooks]map[string]struct{}
	hooks       [maxHooks][]Hook
	hookPresent [maxHooks]atomic.Bool
}

func newHooks() *hooks {
	var names [maxHooks]map[string]struct{}
	var hks [maxHooks][]Hook

	for t := hookType(0); t < maxHooks; t++ {
		names[t] = make(map[string]struct{})
		hks[t] = make([]Hook, 0)
	}

	return &hooks{
		hookNames: names,
		hooks:     hks,
	}
}

func (h *hooks) hasHook(t hookType) bool {
	return h.hookPresent[t].Load()
}

func (h *hooks) add(hook Hook) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for t := hookType(0); t < maxHooks; t++ {
		_, ok := h.hookNames[t][hook.Name()]
		if ok {
			return ErrHookAlreadyExists
		}
	}

	for t := hookType(0); t < maxHooks; t++ {
		hooksRegistryFunc[t](h, hook, t)
	}

	h.cnt++
	return nil
}

func (h *hooks) len() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return h.cnt
}

func (h *hooks) onStart(ctx context.Context) error {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onStartHook] {
		hk := hook.(OnStartHook)

		err := hk.OnStart(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *hooks) onStop(ctx context.Context) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onStopHook] {
		hk := hook.(OnStopHook)
		hk.OnStop(ctx)
	}
}

func (h *hooks) onServerStart(ctx context.Context) error {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onServerStartHook] {
		hk := hook.(OnServerStartHook)

		err := hk.OnServerStart(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *hooks) onServerStartFailed(ctx context.Context, err error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onServerStartFailedHook] {
		hk := hook.(OnServerStartFailedHook)
		hk.OnServerStartFailed(ctx, err)
	}
}

func (h *hooks) onServerStarted(ctx context.Context) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onServerStartedHook] {
		hk := hook.(OnServerStartedHook)
		hk.OnServerStarted(ctx)
	}
}

func (h *hooks) onServerStop(ctx context.Context) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onServerStopHook] {
		hk := hook.(OnServerStopHook)
		hk.OnServerStop(ctx)
	}
}

func (h *hooks) onServerStopped(ctx context.Context) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onServerStoppedHook] {
		hk := hook.(OnServerStoppedHook)
		hk.OnServerStopped(ctx)
	}
}

func (h *hooks) onConnectionOpen(ctx context.Context, c *Connection) error {
	if !h.hasHook(onConnectionOpenHook) {
		return nil
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onConnectionOpenHook] {
		hk := hook.(OnConnectionOpenHook)

		err := hk.OnConnectionOpen(ctx, c)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *hooks) onClientOpened(ctx context.Context, c *Client) {
	if !h.hasHook(onClientOpenedHook) {
		return
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onClientOpenedHook] {
		hk := hook.(OnClientOpenedHook)
		hk.OnClientOpened(ctx, c)
	}
}

func (h *hooks) onClientClose(c *Client, err error) {
	if !h.hasHook(onClientCloseHook) {
		return
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onClientCloseHook] {
		hk := hook.(OnClientCloseHook)
		hk.OnClientClose(c, err)
	}
}

func (h *hooks) onConnectionClosed(c *Connection, err error) {
	if !h.hasHook(onConnectionClosedHook) {
		return
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onConnectionClosedHook] {
		hk := hook.(OnConnectionClosedHook)
		hk.OnConnectionClosed(c, err)
	}
}

func (h *hooks) onReceivePacket(ctx context.Context, c *Client) error {
	if !h.hasHook(onReceivePacketHook) {
		return nil
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onReceivePacketHook] {
		hk := hook.(OnReceivePacketHook)

		err := hk.OnReceivePacket(ctx, c)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *hooks) onPacketReceive(ctx context.Context, c *Client, header packet.FixedHeader) error {
	if !h.hasHook(onPacketReceiveHook) {
		return nil
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onPacketReceiveHook] {
		hk := hook.(OnPacketReceiveHook)

		err := hk.OnPacketReceive(ctx, c, header)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *hooks) onPacketReceiveFailed(ctx context.Context, c *Client, err error) {
	if !h.hasHook(onPacketReceiveFailedHook) {
		return
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onPacketReceiveFailedHook] {
		hk := hook.(OnPacketReceiveFailedHook)
		hk.OnPacketReceiveFailed(ctx, c, err)
	}
}

func (h *hooks) onPacketReceived(ctx context.Context, c *Client, p Packet) error {
	if !h.hasHook(onPacketReceivedHook) {
		return nil
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onPacketReceivedHook] {
		hk := hook.(OnPacketReceivedHook)

		err := hk.OnPacketReceived(ctx, c, p)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *hooks) onPacketSend(ctx context.Context, c *Client, p Packet) error {
	if !h.hasHook(onPacketSendHook) {
		return nil
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onPacketSendHook] {
		hk := hook.(OnPacketSendHook)

		err := hk.OnPacketSend(ctx, c, p)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *hooks) onPacketSendFailed(ctx context.Context, c *Client, p Packet, err error) {
	if !h.hasHook(onPacketSendFailedHook) {
		return
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onPacketSendFailedHook] {
		hk := hook.(OnPacketSendFailedHook)
		hk.OnPacketSendFailed(ctx, c, p, err)
	}
}

func (h *hooks) onPacketSent(ctx context.Context, c *Client, p Packet) {
	if !h.hasHook(onPacketSentHook) {
		return
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onPacketSentHook] {
		hk := hook.(OnPacketSentHook)
		hk.OnPacketSent(ctx, c, p)
	}
}

func (h *hooks) onConnectPacket(ctx context.Context, c *Client, p *packet.Connect) error {
	if !h.hasHook(onConnectPacketHook) {
		return nil
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onConnectPacketHook] {
		hk := hook.(OnConnectPacketHook)

		err := hk.OnConnectPacket(ctx, c, p)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *hooks) onConnectFailed(ctx context.Context, c *Client, err error) {
	if !h.hasHook(onConnectFailedHook) {
		return
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onConnectFailedHook] {
		hk := hook.(OnConnectFailedHook)
		hk.OnConnectFailed(ctx, c, err)
	}
}

func (h *hooks) onConnected(ctx context.Context, c *Client) {
	if !h.hasHook(onConnectedHook) {
		return
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onConnectedHook] {
		hk := hook.(OnConnectedHook)
		hk.OnConnected(ctx, c)
	}
}

func (h *hooks) onAuthPacket(ctx context.Context, c *Client, p *packet.Auth) error {
	if !h.hasHook(onAuthPacketHook) {
		return nil
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onAuthPacketHook] {
		hk := hook.(OnAuthPacketHook)

		err := hk.OnAuthPacket(ctx, c, p)
		if err != nil {
			return err
		}
	}

	return nil
}
