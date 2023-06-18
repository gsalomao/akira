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

type hookType int

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
	numHookTypes
)

// ErrHookAlreadyExists indicates that the Hook already exists based on its name.
var ErrHookAlreadyExists = errors.New("hook already exists")

// Hook is the minimal interface which any hook must implement. All others hook interfaces are optional.
// During the lifecycle of the Server, several events are generated which can be handled by the hook
// based on each handler it implements.
type Hook interface {
	// Name returns the name of the hook. The Server supports only one hook for each name. If a hook is added into the
	// server with the same name of another hook, the ErrHookAlreadyExists is returned.
	Name() string
}

// OnStartHook is a hook which receives the event to start it.
type OnStartHook interface {
	// OnStart is called by the server to start the hook. If this method returns any error, the server considers that
	// the hook failed to start.
	OnStart() error
}

// OnStopHook is a hook which receives the event to stop it.
type OnStopHook interface {
	// OnStop is called by the server to stop the hook.
	OnStop()
}

// OnServerStartHook is a hook which receives the event indicating the server is starting.
type OnServerStartHook interface {
	// OnServerStart is called by the server when it is starting. When this method is called, the server
	// is in ServerStarting state. If this method returns any error, the start process fails.
	OnServerStart(s *Server) error
}

// OnServerStartFailedHook is a hook which receives the event indicating the server has failed to start.
type OnServerStartFailedHook interface {
	// OnServerStartFailed is called by the server when it has failed to start.
	OnServerStartFailed(s *Server, err error)
}

// OnServerStartedHook is a hook which receives the event indicating the server has started.
type OnServerStartedHook interface {
	// OnServerStarted is called by the server when it has started with success.
	OnServerStarted(s *Server)
}

// OnServerStopHook is a hook which receives the event indicating the server is stopping.
type OnServerStopHook interface {
	// OnServerStop is called by the server when it is stopping.
	OnServerStop(s *Server)
}

// OnServerStoppedHook is a hook which receives the event indicating the server has stopped.
type OnServerStoppedHook interface {
	// OnServerStopped is called by the server when it has stopped.
	OnServerStopped(s *Server)
}

// OnConnectionOpenHook is a hook which receives the event indicating that a new connection is being opened.
type OnConnectionOpenHook interface {
	// OnConnectionOpen is called by the server when a new connection is being opened. If this method returns any
	// error, the connection is closed.
	OnConnectionOpen(s *Server, l Listener) error
}

// OnConnectionOpenedHook is a hook which receives the event indicating that a new connection was opened.
type OnConnectionOpenedHook interface {
	// OnConnectionOpened is called by the server when a new connection was opened.
	OnConnectionOpened(s *Server, l Listener)
}

// OnClientCloseHook is a hook which receives the event indicating that the connection is being closed.
type OnClientCloseHook interface {
	// OnConnectionClose is called by the server when the connection is being closed.
	OnConnectionClose(s *Server, l Listener)
}

// OnConnectionClosedHook is a hook which receives the event indicating that the connection was closed.
type OnConnectionClosedHook interface {
	// OnConnectionClosed is called by the server when the connection was closed.
	OnConnectionClosed(s *Server, l Listener)
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

func (h *hooks) onStart() error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.hooks[onStartHook] {
		hk := hook.(OnStartHook)

		err := hk.OnStart()
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *hooks) onStop() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.hooks[onStopHook] {
		hk := hook.(OnStopHook)
		hk.OnStop()
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
