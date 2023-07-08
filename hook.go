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
)

const (
	hookOnStart hookType = iota
	hookOnStop
	hookOnServerStart
	hookOnServerStartFailed
	hookOnServerStarted
	hookOnServerStop
	hookOnServerStopped
	hookOnConnectionOpen
	hookOnConnectionOpened
	hookOnConnectionClose
	hookOnConnectionClosed
	hookOnPacketReceive
	hookOnPacketReceiveError
	hookOnPacketReceived
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

// HookOnStart is the hook interface that wraps the OnStart method. The OnStart method is called by the server to
// start the hook. If this method returns any error, the server considers that the hook failed to start.
type HookOnStart interface {
	OnStart(s *Server) error
}

// HookOnStop is the hook interface that wraps the OnStop method. The OnStop method is called by the server to stop
// the hook.
type HookOnStop interface {
	OnStop(s *Server)
}

// HookOnServerStart is the hook interface that wraps the OnServerStart method. The OnServerStart method is called by
// the server when it is starting. When this method is called, the server is in ServerStarting state. If this method
// returns any error, the start process fails.
type HookOnServerStart interface {
	OnServerStart(s *Server) error
}

// HookOnServerStartFailed is the hook interface that wraps the OnServerStartFailed method. The OnServerStartFailed
// method is called by the server when it has failed to start.
type HookOnServerStartFailed interface {
	OnServerStartFailed(s *Server, err error)
}

// HookOnServerStarted is the hook interface that wraps the OnServerStarted method. The OnServerStarted method is
// called by the server when it has started with success.
type HookOnServerStarted interface {
	OnServerStarted(s *Server)
}

// HookOnServerStop is the hook interface that wraps the OnServerStop method. The OnServerStop method is called by the
// server when it is stopping.
type HookOnServerStop interface {
	OnServerStop(s *Server)
}

// HookOnServerStopped is the hook interface that wraps the OnServerStopped method. The OnServerStopped method is
// called by the server when it has stopped.
type HookOnServerStopped interface {
	OnServerStopped(s *Server)
}

// HookOnConnectionOpen is the hook interface that wraps the OnConnectionOpen method. The OnConnectionOpen method is
// called by the server when a new connection is being opened. If this method returns any error, the connection is
// closed.
type HookOnConnectionOpen interface {
	OnConnectionOpen(s *Server, l Listener) error
}

// HookOnConnectionOpened is the hook interface that wraps the OnConnectionOpened method. The OnConnectionOpened
// method is called by the server when a new connection has opened. This method is called after the OnConnectionOpen
// method.
type HookOnConnectionOpened interface {
	OnConnectionOpened(s *Server, l Listener)
}

// HookOnClientClose is the hook interface that wraps the OnConnectionClose method. The OnConnectionClose method is
// called by the server when the connection is being closed. If the connection is being closed due to any error, the
// error is passed as parameter.
type HookOnClientClose interface {
	OnConnectionClose(s *Server, l Listener, err error)
}

// HookOnConnectionClosed is the hook interface that wraps the OnConnectionClosed method. The OnConnectionClosed
// method is called by the server when the connection has closed. This method is called after the OnConnectionClose
// method. If the connection was closed due to any error, the error is passed as parameter.
type HookOnConnectionClosed interface {
	OnConnectionClosed(s *Server, l Listener, err error)
}

// HookOnPacketReceive is the hook interface that wraps the OnPacketReceive method. The OnPacketReceive method is
// called by the server before try to receive any packet. If this method returns an error, the server doesn't try
// to receive a new packet and the client is closed.
type HookOnPacketReceive interface {
	OnPacketReceive(c *Client) error
}

// HookOnPacketReceiveError is the hook interface that wraps the OnPacketReceiveError method. The OnPacketReceiveError
// method is called by the server, with the error as parameter, when it fails to receive a new packet. If this method
// returns an error, the server closes the client. Otherwise, it tries to receive a new packet again.
type HookOnPacketReceiveError interface {
	OnPacketReceiveError(c *Client, err error) error
}

// HookOnPacketReceived is the hook interface that wraps the OnPacketReceived method. The OnPacketReceived method is
// called by the server when a new packet is received. If this method returns an error, the server discards the
// received packet.
type HookOnPacketReceived interface {
	OnPacketReceived(c *Client, p Packet) error
}

var hooksRegistries = map[hookType]func(*hooks, Hook, hookType){
	hookOnStart:              registerHook[HookOnStart],
	hookOnStop:               registerHook[HookOnStop],
	hookOnServerStart:        registerHook[HookOnServerStart],
	hookOnServerStartFailed:  registerHook[HookOnServerStartFailed],
	hookOnServerStarted:      registerHook[HookOnServerStarted],
	hookOnServerStop:         registerHook[HookOnServerStop],
	hookOnServerStopped:      registerHook[HookOnServerStopped],
	hookOnConnectionOpen:     registerHook[HookOnConnectionOpen],
	hookOnConnectionOpened:   registerHook[HookOnConnectionOpened],
	hookOnConnectionClose:    registerHook[HookOnClientClose],
	hookOnConnectionClosed:   registerHook[HookOnConnectionClosed],
	hookOnPacketReceive:      registerHook[HookOnPacketReceive],
	hookOnPacketReceiveError: registerHook[HookOnPacketReceiveError],
	hookOnPacketReceived:     registerHook[HookOnPacketReceived],
}

func registerHook[T any](h *hooks, hook Hook, t hookType) {
	if _, ok := hook.(T); ok {
		h.hookNames[t][hook.Name()] = struct{}{}
		h.hooks[t] = append(h.hooks[t], hook)
		h.hookPresent[t].Store(true)
	}
}

type hooks struct {
	mu          sync.RWMutex
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
	h.mu.Lock()
	defer h.mu.Unlock()

	for t := hookType(0); t < numOfHookTypes; t++ {
		_, ok := h.hookNames[t][hook.Name()]
		if ok {
			return ErrHookAlreadyExists
		}
	}

	for t := hookType(0); t < numOfHookTypes; t++ {
		hooksRegistries[t](h, hook, t)
	}

	return nil
}

func (h *hooks) onStart(s *Server) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.hooks[hookOnStart] {
		hk := hook.(HookOnStart)

		err := hk.OnStart(s)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *hooks) onStop(s *Server) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.hooks[hookOnStop] {
		hk := hook.(HookOnStop)
		hk.OnStop(s)
	}
}

func (h *hooks) onServerStart(s *Server) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.hooks[hookOnServerStart] {
		hk := hook.(HookOnServerStart)

		err := hk.OnServerStart(s)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *hooks) onServerStartFailed(s *Server, err error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.hooks[hookOnServerStartFailed] {
		hk := hook.(HookOnServerStartFailed)
		hk.OnServerStartFailed(s, err)
	}
}

func (h *hooks) onServerStarted(s *Server) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.hooks[hookOnServerStarted] {
		hk := hook.(HookOnServerStarted)
		hk.OnServerStarted(s)
	}
}

func (h *hooks) onServerStop(s *Server) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.hooks[hookOnServerStop] {
		hk := hook.(HookOnServerStop)
		hk.OnServerStop(s)
	}
}

func (h *hooks) onServerStopped(s *Server) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.hooks[hookOnServerStopped] {
		hk := hook.(HookOnServerStopped)
		hk.OnServerStopped(s)
	}
}

func (h *hooks) onConnectionOpen(s *Server, l Listener) error {
	if !h.hasHook(hookOnConnectionOpen) {
		return nil
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.hooks[hookOnConnectionOpen] {
		hk := hook.(HookOnConnectionOpen)

		err := hk.OnConnectionOpen(s, l)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *hooks) onConnectionOpened(s *Server, l Listener) {
	if !h.hasHook(hookOnConnectionOpened) {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.hooks[hookOnConnectionOpened] {
		hk := hook.(HookOnConnectionOpened)
		hk.OnConnectionOpened(s, l)
	}
}

func (h *hooks) onConnectionClose(s *Server, l Listener, err error) {
	if !h.hasHook(hookOnConnectionClose) {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.hooks[hookOnConnectionClose] {
		hk := hook.(HookOnClientClose)
		hk.OnConnectionClose(s, l, err)
	}
}

func (h *hooks) onConnectionClosed(s *Server, l Listener, err error) {
	if !h.hasHook(hookOnConnectionClosed) {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.hooks[hookOnConnectionClosed] {
		hk := hook.(HookOnConnectionClosed)
		hk.OnConnectionClosed(s, l, err)
	}
}

func (h *hooks) onPacketReceive(c *Client) error {
	if !h.hasHook(hookOnPacketReceive) {
		return nil
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.hooks[hookOnPacketReceive] {
		hk := hook.(HookOnPacketReceive)

		err := hk.OnPacketReceive(c)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *hooks) onPacketReceiveError(c *Client, err error) error {
	if !h.hasHook(hookOnPacketReceiveError) {
		return nil
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.hooks[hookOnPacketReceiveError] {
		hk := hook.(HookOnPacketReceiveError)

		newErr := hk.OnPacketReceiveError(c, err)
		if newErr != nil {
			return newErr
		}
	}

	return nil
}

func (h *hooks) onPacketReceived(c *Client, p Packet) error {
	if !h.hasHook(hookOnPacketReceived) {
		return nil
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.hooks[hookOnPacketReceived] {
		hk := hook.(HookOnPacketReceived)

		err := hk.OnPacketReceived(c, p)
		if err != nil {
			return err
		}
	}

	return nil
}
