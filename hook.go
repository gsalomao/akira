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

package melitte

import (
	"errors"
	"sync"
	"sync/atomic"
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
	onConnectionOpenedHook
	onConnectionCloseHook
	onConnectionClosedHook
	onPacketReceiveHook
	numHookTypes
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
	OnStart(s *Server) error
}

// OnStopHook is the hook interface that wraps the OnStop method. The OnStop method is called by the server to stop
// the hook.
type OnStopHook interface {
	OnStop(s *Server)
}

// OnServerStartHook is the hook interface that wraps the OnServerStart method. The OnServerStart method is called by
// the server when it is starting. When this method is called, the server is in ServerStarting state. If this method
// returns any error, the start process fails.
type OnServerStartHook interface {
	OnServerStart(s *Server) error
}

// OnServerStartFailedHook is the hook interface that wraps the OnServerStartFailed method. The OnServerStartFailed
// method is called by the server when it has failed to start.
type OnServerStartFailedHook interface {
	OnServerStartFailed(s *Server, err error)
}

// OnServerStartedHook is the hook interface that wraps the OnServerStarted method. The OnServerStarted method is
// called by the server when it has started with success.
type OnServerStartedHook interface {
	OnServerStarted(s *Server)
}

// OnServerStopHook is the hook interface that wraps the OnServerStop method. The OnServerStop method is called by the
// server when it is stopping.
type OnServerStopHook interface {
	OnServerStop(s *Server)
}

// OnServerStoppedHook is the hook interface that wraps the OnServerStopped method. The OnServerStopped method is
// called by the server when it has stopped.
type OnServerStoppedHook interface {
	OnServerStopped(s *Server)
}

// OnConnectionOpenHook is the hook interface that wraps the OnConnectionOpen method. The OnConnectionOpen method is
// called by the server when a new connection is being opened. If this method returns any error, the connection is
// closed.
type OnConnectionOpenHook interface {
	OnConnectionOpen(s *Server, l Listener) error
}

// OnConnectionOpenedHook is the hook interface that wraps the OnConnectionOpened method. The OnConnectionOpened
// method is called by the server when a new connection has opened. This method is called after the OnConnectionOpen
// method.
type OnConnectionOpenedHook interface {
	OnConnectionOpened(s *Server, l Listener)
}

// OnClientCloseHook is the hook interface that wraps the OnConnectionClose method. The OnConnectionClose method is
// called by the server when the connection is being closed.
type OnClientCloseHook interface {
	OnConnectionClose(s *Server, l Listener)
}

// OnConnectionClosedHook is the hook interface that wraps the OnConnectionClosed method. The OnConnectionClosed
// method is called by the server when the connection has closed. This method is called after the OnConnectionClose
// method.
type OnConnectionClosedHook interface {
	OnConnectionClosed(s *Server, l Listener)
}

// OnPacketReceiveHook is the hook interface that wraps the OnPacketReceive method. The OnPacketReceive method is
// called by the server before try to receive any packet. If this method returns an error, the server doesn't try
// to receive a new packet and the client is closed.
type OnPacketReceiveHook interface {
	OnPacketReceive(s *Server, c *Client) error
}

var hooksRegistries = map[hookType]func(*hooks, Hook, hookType){
	onStartHook:             registerHook[OnStartHook],
	onStopHook:              registerHook[OnStopHook],
	onServerStartHook:       registerHook[OnServerStartHook],
	onServerStartFailedHook: registerHook[OnServerStartFailedHook],
	onServerStartedHook:     registerHook[OnServerStartedHook],
	onServerStopHook:        registerHook[OnServerStopHook],
	onServerStoppedHook:     registerHook[OnServerStoppedHook],
	onConnectionOpenHook:    registerHook[OnConnectionOpenHook],
	onConnectionOpenedHook:  registerHook[OnConnectionOpenedHook],
	onConnectionCloseHook:   registerHook[OnClientCloseHook],
	onConnectionClosedHook:  registerHook[OnConnectionClosedHook],
	onPacketReceiveHook:     registerHook[OnPacketReceiveHook],
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
	hookNames   [numHookTypes]map[string]struct{}
	hooks       [numHookTypes][]Hook
	hookPresent [numHookTypes]atomic.Bool
}

func newHooks() *hooks {
	var names [numHookTypes]map[string]struct{}
	var hks [numHookTypes][]Hook

	for t := hookType(0); t < numHookTypes; t++ {
		names[t] = make(map[string]struct{})
		hks[t] = make([]Hook, 0)
	}

	h := hooks{
		hookNames: names,
		hooks:     hks,
	}
	return &h
}

func (h *hooks) hasHook(t hookType) bool {
	return h.hookPresent[t].Load()
}

func (h *hooks) add(hook Hook) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	for t := hookType(0); t < numHookTypes; t++ {
		_, ok := h.hookNames[t][hook.Name()]
		if ok {
			return ErrHookAlreadyExists
		}
	}

	for t := hookType(0); t < numHookTypes; t++ {
		hooksRegistries[t](h, hook, t)
	}

	return nil
}

func (h *hooks) onStart(s *Server) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

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
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.hooks[onStopHook] {
		hk := hook.(OnStopHook)
		hk.OnStop(s)
	}
}

func (h *hooks) onServerStart(s *Server) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

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
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.hooks[onServerStartFailedHook] {
		hk := hook.(OnServerStartFailedHook)
		hk.OnServerStartFailed(s, err)
	}
}

func (h *hooks) onServerStarted(s *Server) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.hooks[onServerStartedHook] {
		hk := hook.(OnServerStartedHook)
		hk.OnServerStarted(s)
	}
}

func (h *hooks) onServerStop(s *Server) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.hooks[onServerStopHook] {
		hk := hook.(OnServerStopHook)
		hk.OnServerStop(s)
	}
}

func (h *hooks) onServerStopped(s *Server) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.hooks[onServerStoppedHook] {
		hk := hook.(OnServerStoppedHook)
		hk.OnServerStopped(s)
	}
}

func (h *hooks) onConnectionOpen(s *Server, l Listener) error {
	if !h.hasHook(onConnectionOpenHook) {
		return nil
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.hooks[onConnectionOpenHook] {
		hk := hook.(OnConnectionOpenHook)

		err := hk.OnConnectionOpen(s, l)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *hooks) onConnectionOpened(s *Server, l Listener) {
	if !h.hasHook(onConnectionOpenedHook) {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.hooks[onConnectionOpenedHook] {
		hk := hook.(OnConnectionOpenedHook)
		hk.OnConnectionOpened(s, l)
	}
}

func (h *hooks) onConnectionClose(s *Server, l Listener) {
	if !h.hasHook(onConnectionCloseHook) {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.hooks[onConnectionCloseHook] {
		hk := hook.(OnClientCloseHook)
		hk.OnConnectionClose(s, l)
	}
}

func (h *hooks) onConnectionClosed(s *Server, l Listener) {
	if !h.hasHook(onConnectionClosedHook) {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.hooks[onConnectionClosedHook] {
		hk := hook.(OnConnectionClosedHook)
		hk.OnConnectionClosed(s, l)
	}
}

func (h *hooks) onPacketReceive(s *Server, c *Client) error {
	if !h.hasHook(onPacketReceiveHook) {
		return nil
	}

	for _, hook := range h.hooks[onPacketReceiveHook] {
		hk := hook.(OnPacketReceiveHook)

		err := hk.OnPacketReceive(s, c)
		if err != nil {
			return err
		}
	}

	return nil
}
