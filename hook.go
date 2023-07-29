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
	onConnectHook
	onConnectFailedHook
	onConnectedHook
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
	OnStart() error
}

// OnStopHook is the hook interface that wraps the OnStop method. The OnStop method is called by the server to stop
// the hook.
type OnStopHook interface {
	OnStop()
}

// OnServerStartHook is the hook interface that wraps the OnServerStart method. The OnServerStart method is called by
// the server when it is starting. When this method is called, the server is in ServerStarting state. If this method
// returns any error, the start process fails.
type OnServerStartHook interface {
	OnServerStart() error
}

// OnServerStartFailedHook is the hook interface that wraps the OnServerStartFailed method. The OnServerStartFailed
// method is called by the server when it has failed to start.
type OnServerStartFailedHook interface {
	OnServerStartFailed(err error)
}

// OnServerStartedHook is the hook interface that wraps the OnServerStarted method. The OnServerStarted method is
// called by the server when it has started with success.
type OnServerStartedHook interface {
	OnServerStarted()
}

// OnServerStopHook is the hook interface that wraps the OnServerStop method. The OnServerStop method is called by the
// server when it is stopping.
type OnServerStopHook interface {
	OnServerStop()
}

// OnServerStoppedHook is the hook interface that wraps the OnServerStopped method. The OnServerStopped method is
// called by the server when it has stopped.
type OnServerStoppedHook interface {
	OnServerStopped()
}

// OnConnectionOpenHook is the hook interface that wraps the OnConnectionOpen method. The OnConnectionOpen method is
// called by the server when a new network connection was opened. If this method returns any error, the connection is
// closed.
type OnConnectionOpenHook interface {
	OnConnectionOpen(c *Connection) error
}

// OnClientOpenedHook is the hook interface that wraps the OnClientOpened method. The OnClientOpened method is called
// by the server when the client opened the connection, and after the network connection has been opened with success.
type OnClientOpenedHook interface {
	OnClientOpened(c *Client)
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
	OnReceivePacket(c *Client) error
}

// OnPacketReceiveHook is the hook interface that wraps the OnPacketReceive method. The OnPacketReceive method is
// called by the server when it has received the fixed header of the packet from the client, but it did not receive
// yet the remaining bytes of the packet. If this method returns an  error, the server stops to receive the packet
// and closes the client.
type OnPacketReceiveHook interface {
	OnPacketReceive(c *Client, h packet.FixedHeader) error
}

// OnPacketReceiveFailedHook is the hook interface that wraps the OnPacketReceiveFailed method. The
// OnPacketReceiveFailed method is called by the server when it fails to receive a packet from the client for any
// reason other than the network connection closed. After the server calls this method, it closes the client.
type OnPacketReceiveFailedHook interface {
	OnPacketReceiveFailed(c *Client, err error)
}

// OnPacketReceivedHook is the hook interface that wraps the OnPacketReceived method. The OnPacketReceived method is
// called by the server when a new packet is received. If this method returns an error, the server discards the
// packet and closes the client.
type OnPacketReceivedHook interface {
	OnPacketReceived(c *Client, p Packet) error
}

// OnPacketSendHook is the hook interface that wraps the OnPacketSend method. The OnPacketSend method is called by the
// server before the packet has been sent to the client. If this method returns an error, the server discards the
// packet and closes the client.
type OnPacketSendHook interface {
	OnPacketSend(c *Client, p Packet) error
}

// OnPacketSendFailedHook is the hook interface that wraps the OnPacketSendFailed method. The OnPacketSendFailed
// method is called by the server when it fails to send the packet to the client for any reason other than the
// network connection closed. After the server calls this method, it closes the client.
type OnPacketSendFailedHook interface {
	OnPacketSendFailed(c *Client, p Packet, err error)
}

// OnPacketSentHook is the hook interface that wraps the OnPacketSent method. The OnPacketSent method is called by the
// server after the packet has been sent to the client.
type OnPacketSentHook interface {
	OnPacketSent(c *Client, p Packet)
}

// OnConnectHook is the hook interface that wraps the OnConnect method. The OnConnect method is called by the server
// when it receives a CONNECT Packet. If this method returns an error, the server stops the connection process and
// closes the client. If the returned error is a packet.Error with a reason code other than packet.ReasonCodeSuccess
// and packet.ReasonCodeMalformedPacket, the server sends a CONNACK Packet with the reason code before it closes the
// client.
type OnConnectHook interface {
	OnConnect(c *Client, p *packet.Connect) error
}

// OnConnectFailedHook is the hook interface that wraps the OnConnectFailed method. The OnConnectFailed method is
// called by the server when it fails to connect the client. The connection process can fail for different reasons,
// and the error indicating the reason is passed as parameter. After the server calls this method, it closes the
// client.
type OnConnectFailedHook interface {
	OnConnectFailed(c *Client, p *packet.Connect, err error)
}

// OnConnectedHook is the hook interface that wraps the OnConnected method. The OnConnected method is called by the
// server when it completes the connection process to connect the client with success.
type OnConnectedHook interface {
	OnConnected(c *Client)
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
	onConnectHook:             registerHook[OnConnectHook],
	onConnectFailedHook:       registerHook[OnConnectFailedHook],
	onConnectedHook:           registerHook[OnConnectedHook],
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

	return nil
}

func (h *hooks) onStart() error {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onStartHook] {
		hk := hook.(OnStartHook)
		return hk.OnStart()
	}
	return nil
}

func (h *hooks) onStop() {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onStopHook] {
		hk := hook.(OnStopHook)
		hk.OnStop()
	}
}

func (h *hooks) onServerStart() error {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onServerStartHook] {
		hk := hook.(OnServerStartHook)
		return hk.OnServerStart()
	}
	return nil
}

func (h *hooks) onServerStartFailed(err error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onServerStartFailedHook] {
		hk := hook.(OnServerStartFailedHook)
		hk.OnServerStartFailed(err)
	}
}

func (h *hooks) onServerStarted() {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onServerStartedHook] {
		hk := hook.(OnServerStartedHook)
		hk.OnServerStarted()
	}
}

func (h *hooks) onServerStop() {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onServerStopHook] {
		hk := hook.(OnServerStopHook)
		hk.OnServerStop()
	}
}

func (h *hooks) onServerStopped() {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onServerStoppedHook] {
		hk := hook.(OnServerStoppedHook)
		hk.OnServerStopped()
	}
}

func (h *hooks) onConnectionOpen(c *Connection) error {
	if !h.hasHook(onConnectionOpenHook) {
		return nil
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onConnectionOpenHook] {
		hk := hook.(OnConnectionOpenHook)
		return hk.OnConnectionOpen(c)
	}
	return nil
}

func (h *hooks) onClientOpened(c *Client) {
	if !h.hasHook(onClientOpenedHook) {
		return
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onClientOpenedHook] {
		hk := hook.(OnClientOpenedHook)
		hk.OnClientOpened(c)
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

func (h *hooks) onReceivePacket(c *Client) error {
	if !h.hasHook(onReceivePacketHook) {
		return nil
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onReceivePacketHook] {
		hk := hook.(OnReceivePacketHook)
		return hk.OnReceivePacket(c)
	}
	return nil
}

func (h *hooks) onPacketReceive(c *Client, header packet.FixedHeader) error {
	if !h.hasHook(onPacketReceiveHook) {
		return nil
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onPacketReceiveHook] {
		hk := hook.(OnPacketReceiveHook)

		err := hk.OnPacketReceive(c, header)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *hooks) onPacketReceiveFailed(c *Client, err error) {
	if !h.hasHook(onPacketReceiveFailedHook) {
		return
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onPacketReceiveFailedHook] {
		hk := hook.(OnPacketReceiveFailedHook)
		hk.OnPacketReceiveFailed(c, err)
	}
}

func (h *hooks) onPacketReceived(c *Client, p Packet) error {
	if !h.hasHook(onPacketReceivedHook) {
		return nil
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onPacketReceivedHook] {
		hk := hook.(OnPacketReceivedHook)
		return hk.OnPacketReceived(c, p)
	}
	return nil
}

func (h *hooks) onPacketSend(c *Client, p Packet) error {
	if !h.hasHook(onPacketSendHook) {
		return nil
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onPacketSendHook] {
		hk := hook.(OnPacketSendHook)
		return hk.OnPacketSend(c, p)
	}
	return nil
}

func (h *hooks) onPacketSendFailed(c *Client, p Packet, err error) {
	if !h.hasHook(onPacketSendFailedHook) {
		return
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onPacketSendFailedHook] {
		hk := hook.(OnPacketSendFailedHook)
		hk.OnPacketSendFailed(c, p, err)
	}
}

func (h *hooks) onPacketSent(c *Client, p Packet) {
	if !h.hasHook(onPacketSentHook) {
		return
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onPacketSentHook] {
		hk := hook.(OnPacketSentHook)
		hk.OnPacketSent(c, p)
	}
}

func (h *hooks) onConnect(c *Client, p *packet.Connect) error {
	if !h.hasHook(onConnectHook) {
		return nil
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onConnectHook] {
		hk := hook.(OnConnectHook)
		return hk.OnConnect(c, p)
	}
	return nil
}

func (h *hooks) onConnectFailed(c *Client, p *packet.Connect, err error) {
	if !h.hasHook(onConnectFailedHook) {
		return
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onConnectFailedHook] {
		hk := hook.(OnConnectFailedHook)
		hk.OnConnectFailed(c, p, err)
	}
}

func (h *hooks) onConnected(c *Client) {
	if !h.hasHook(onConnectedHook) {
		return
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onConnectedHook] {
		hk := hook.(OnConnectedHook)
		hk.OnConnected(c)
	}
}
