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
	onClientOpenHook
	onClientClosedHook
	onPacketReceiveHook
	onPacketReceiveErrorHook
	onPacketReceivedHook
	onPacketSendHook
	onPacketSendErrorHook
	onPacketSentHook
	onConnectHook
	onConnectErrorHook
	onConnectedHook
	numOfHookTypes
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
// start the hook. If this method returns any error, the server considers that the hook failed to start.
type OnStartHook interface {
	Hook
	OnStart(s *Server) error
}

// OnStopHook is the hook interface that wraps the OnStop method. The OnStop method is called by the server to stop
// the hook.
type OnStopHook interface {
	Hook
	OnStop(s *Server)
}

// OnServerStartHook is the hook interface that wraps the OnServerStart method. The OnServerStart method is called by
// the server when it is starting. When this method is called, the server is in ServerStarting state. If this method
// returns any error, the start process fails.
type OnServerStartHook interface {
	Hook
	OnServerStart(s *Server) error
}

// OnServerStartFailedHook is the hook interface that wraps the OnServerStartFailed method. The OnServerStartFailed
// method is called by the server when it has failed to start.
type OnServerStartFailedHook interface {
	Hook
	OnServerStartFailed(s *Server, err error)
}

// OnServerStartedHook is the hook interface that wraps the OnServerStarted method. The OnServerStarted method is
// called by the server when it has started with success.
type OnServerStartedHook interface {
	Hook
	OnServerStarted(s *Server)
}

// OnServerStopHook is the hook interface that wraps the OnServerStop method. The OnServerStop method is called by the
// server when it is stopping.
type OnServerStopHook interface {
	Hook
	OnServerStop(s *Server)
}

// OnServerStoppedHook is the hook interface that wraps the OnServerStopped method. The OnServerStopped method is
// called by the server when it has stopped.
type OnServerStoppedHook interface {
	Hook
	OnServerStopped(s *Server)
}

// OnClientOpenHook is the hook interface that wraps the OnClientOpen method. The OnClientOpen method is called by the
// server when a client opens a new connection. If this method returns any error, the client and its connection are
// closed.
type OnClientOpenHook interface {
	Hook
	OnClientOpen(s *Server, c *Client) error
}

// OnClientClosedHook is the hook interface that wraps the OnClientClosed method. The OnClientClosed method is called
// by the server when the client and its connection have closed. If the client was closed due to any error, the error
// is passed as parameter.
type OnClientClosedHook interface {
	Hook
	OnClientClosed(s *Server, c *Client, err error)
}

// OnPacketReceiveHook is the hook interface that wraps the OnPacketReceive method. The OnPacketReceive method is
// called by the server before try to receive any packet. If this method returns an error, the server doesn't try
// to receive a new packet and the client is closed.
type OnPacketReceiveHook interface {
	Hook
	OnPacketReceive(c *Client) error
}

// OnPacketReceiveErrorHook is the hook interface that wraps the OnPacketReceiveError method. The OnPacketReceiveError
// method is called by the server when it fails to receive a packet from the client for any reason other than the
// network connection closed. If this method returns an error, the server closes the client. Otherwise, it tries to
// receive a new packet again.
type OnPacketReceiveErrorHook interface {
	Hook
	OnPacketReceiveError(c *Client, err error) error
}

// OnPacketReceivedHook is the hook interface that wraps the OnPacketReceived method. The OnPacketReceived method is
// called by the server when a new packet is received. If this method returns an error, the server discards the
// received packet.
type OnPacketReceivedHook interface {
	Hook
	OnPacketReceived(c *Client, p Packet) error
}

// OnPacketSendHook is the hook interface that wraps the OnPacketSend method. The OnPacketSend method is called by the
// server before the packet has been sent to the client. If this method returns an error, the server discards the
// packet and closes the client.
type OnPacketSendHook interface {
	Hook
	OnPacketSend(c *Client, p Packet) error
}

// OnPacketSendErrorHook is the hook interface that wraps the OnPacketSendError method. The OnPacketSendError method
// is called by the server when it fails to send the packet to the client for any reason other than the network
// connection closed.
type OnPacketSendErrorHook interface {
	Hook
	OnPacketSendError(c *Client, p Packet, err error)
}

// OnPacketSentHook is the hook interface that wraps the OnPacketSent method. The OnPacketSent method is called by the
// server after the packet has been sent to the client.
type OnPacketSentHook interface {
	Hook
	OnPacketSent(c *Client, p Packet)
}

// OnConnectHook is the hook interface that wraps the OnConnect method. The OnConnect method is called by the server
// when it receives a CONNECT Packet. If this method returns an error, the server stops the connection process and
// closes the client. If the returned error is a packet.Error with a reason code other than packet.ReasonCodeSuccess
// and packet.ReasonCodeMalformedPacket, the server sends a CONNACK Packet with the reason code before it closes the
// client.
type OnConnectHook interface {
	Hook
	OnConnect(c *Client, p *packet.Connect) error
}

// OnConnectErrorHook is the hook interface that wraps the OnConnectError method. The OnConnectError method is called
// by the server when it fails to connect the client. The connection process can fail for different reasons, and the
// error indicating the reason is passed as parameter.
type OnConnectErrorHook interface {
	Hook
	OnConnectError(c *Client, p *packet.Connect, err error)
}

// OnConnectedHook is the hook interface that wraps the OnConnected method. The OnConnected method is called by the
// server when it completes the connection process to connect the client with success.
type OnConnectedHook interface {
	Hook
	OnConnected(c *Client)
}

var hooksRegistryFunc = map[hookType]func(*hooks, Hook, hookType){
	onStartHook:              registerHook[OnStartHook],
	onStopHook:               registerHook[OnStopHook],
	onServerStartHook:        registerHook[OnServerStartHook],
	onServerStartFailedHook:  registerHook[OnServerStartFailedHook],
	onServerStartedHook:      registerHook[OnServerStartedHook],
	onServerStopHook:         registerHook[OnServerStopHook],
	onServerStoppedHook:      registerHook[OnServerStoppedHook],
	onClientOpenHook:         registerHook[OnClientOpenHook],
	onClientClosedHook:       registerHook[OnClientClosedHook],
	onPacketReceiveHook:      registerHook[OnPacketReceiveHook],
	onPacketReceiveErrorHook: registerHook[OnPacketReceiveErrorHook],
	onPacketReceivedHook:     registerHook[OnPacketReceivedHook],
	onPacketSendHook:         registerHook[OnPacketSendHook],
	onPacketSendErrorHook:    registerHook[OnPacketSendErrorHook],
	onPacketSentHook:         registerHook[OnPacketSentHook],
	onConnectHook:            registerHook[OnConnectHook],
	onConnectErrorHook:       registerHook[OnConnectErrorHook],
	onConnectedHook:          registerHook[OnConnectedHook],
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
	hookNames   [numOfHookTypes]map[string]struct{}
	hooks       [numOfHookTypes][]Hook
	hookPresent [numOfHookTypes]atomic.Bool
}

func newHooks() *hooks {
	var names [numOfHookTypes]map[string]struct{}
	var hks [numOfHookTypes][]Hook

	for t := hookType(0); t < numOfHookTypes; t++ {
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

	for t := hookType(0); t < numOfHookTypes; t++ {
		_, ok := h.hookNames[t][hook.Name()]
		if ok {
			return ErrHookAlreadyExists
		}
	}

	for t := hookType(0); t < numOfHookTypes; t++ {
		hooksRegistryFunc[t](h, hook, t)
	}

	return nil
}

func (h *hooks) onStart(s *Server) error {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onStartHook] {
		hk := hook.(OnStartHook)

		err := hk.OnStart(s)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *hooks) onStop(s *Server) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onStopHook] {
		hk := hook.(OnStopHook)
		hk.OnStop(s)
	}
}

func (h *hooks) onServerStart(s *Server) error {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onServerStartHook] {
		hk := hook.(OnServerStartHook)

		err := hk.OnServerStart(s)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *hooks) onServerStartFailed(s *Server, err error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onServerStartFailedHook] {
		hk := hook.(OnServerStartFailedHook)
		hk.OnServerStartFailed(s, err)
	}
}

func (h *hooks) onServerStarted(s *Server) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onServerStartedHook] {
		hk := hook.(OnServerStartedHook)
		hk.OnServerStarted(s)
	}
}

func (h *hooks) onServerStop(s *Server) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onServerStopHook] {
		hk := hook.(OnServerStopHook)
		hk.OnServerStop(s)
	}
}

func (h *hooks) onServerStopped(s *Server) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onServerStoppedHook] {
		hk := hook.(OnServerStoppedHook)
		hk.OnServerStopped(s)
	}
}

func (h *hooks) onClientOpen(s *Server, c *Client) error {
	if !h.hasHook(onClientOpenHook) {
		return nil
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onClientOpenHook] {
		hk := hook.(OnClientOpenHook)

		err := hk.OnClientOpen(s, c)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *hooks) onClientClosed(s *Server, c *Client, err error) {
	if !h.hasHook(onClientClosedHook) {
		return
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onClientClosedHook] {
		hk := hook.(OnClientClosedHook)
		hk.OnClientClosed(s, c, err)
	}
}

func (h *hooks) onPacketReceive(c *Client) error {
	if !h.hasHook(onPacketReceiveHook) {
		return nil
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onPacketReceiveHook] {
		hk := hook.(OnPacketReceiveHook)

		err := hk.OnPacketReceive(c)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *hooks) onPacketReceiveError(c *Client, err error) error {
	if !h.hasHook(onPacketReceiveErrorHook) {
		return err
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onPacketReceiveErrorHook] {
		hk := hook.(OnPacketReceiveErrorHook)

		newErr := hk.OnPacketReceiveError(c, err)
		if newErr != nil {
			return newErr
		}
	}

	return nil
}

func (h *hooks) onPacketReceived(c *Client, p Packet) error {
	if !h.hasHook(onPacketReceivedHook) {
		return nil
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onPacketReceivedHook] {
		hk := hook.(OnPacketReceivedHook)

		err := hk.OnPacketReceived(c, p)
		if err != nil {
			return err
		}
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

		err := hk.OnPacketSend(c, p)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *hooks) onPacketSendError(c *Client, p Packet, err error) {
	if !h.hasHook(onPacketSendErrorHook) {
		return
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onPacketSendErrorHook] {
		hk := hook.(OnPacketSendErrorHook)
		hk.OnPacketSendError(c, p, err)
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

		err := hk.OnConnect(c, p)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *hooks) onConnectError(c *Client, p *packet.Connect, err error) {
	if !h.hasHook(onConnectErrorHook) {
		return
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, hook := range h.hooks[onConnectErrorHook] {
		hk := hook.(OnConnectErrorHook)
		hk.OnConnectError(c, p, err)
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
