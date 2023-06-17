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
	"fmt"
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
	onClientOpenHook
	onClientOpenedHook
	onClientCloseHook
	onClientClosedHook
	numHookTypes
)

// ErrHookAlreadyExists indicates that the Hook already exists based on its name.
var ErrHookAlreadyExists = errors.New("hook already exists")

// Hook represents the base interface which any hook must implement. All others hook interfaces are optional.
// A hook is a way to extend the Server based on different events which happens in the Server. For each event, the
// Server calls the respective hook method.
type Hook interface {
	// Name returns the name of the Hook. The Server supports only one Hook for each name. If a Hook is added into the
	// Server with the same name of another hook, the ErrHookAlreadyExists is returned.
	Name() string
}

// OnStartHook represents a hook which receives the event to start it.
type OnStartHook interface {
	// OnStart is called by the Server to start the Hook. If this method returns any error, the Server considers that
	// the Hook failed to start.
	OnStart() error
}

// OnStopHook represents a hook which receives the event to stop it.
type OnStopHook interface {
	// OnStop is called by the Server to stop the Hook.
	OnStop()
}

// OnServerStartHook represents a hook which receives the event indicating the Server started the start process.
type OnServerStartHook interface {
	// OnServerStart is called by the Server when it starts the start process. When this method is called, the Server
	// is in ServerStarting state. If this method returns any error, the start process stops, and the ServerState
	// changes to the ServerFailed state.
	OnServerStart(*Server) error
}

// OnServerStartFailedHook represents a hook which receives the event indicating the Server failed to start.
type OnServerStartFailedHook interface {
	// OnServerStartFailed is called by the Server during the start process when the Server failed to start and
	// changed to the ServerFailed state.
	OnServerStartFailed(*Server, error)
}

// OnServerStartedHook represents a hook which receives the event indicating the Server completed the start process
// with success.
type OnServerStartedHook interface {
	// OnServerStarted is called by the Server when it completed the start process and changed to the ServerRunning
	// state.
	OnServerStarted(*Server)
}

// OnServerStopHook represents a hook which receives the event indicating the Server started the stop process.
type OnServerStopHook interface {
	// OnServerStop is called by the Server when it starts the stop process and changed to the ServerStopping state.
	OnServerStop(*Server)
}

// OnServerStoppedHook represents a hook which receives the event indicating the Server completed the stop process.
type OnServerStoppedHook interface {
	// OnServerStopped is called by the Server when it completed the stop process and changed to the ServerStopped
	// State.
	OnServerStopped(*Server)
}

// OnClientOpenHook represents a hook which receives the event indicating that a new Client is being opened.
type OnClientOpenHook interface {
	// OnClientOpen is called by the Server when a new Client is being opened. If this method returns any error,
	// the Client is closed.
	OnClientOpen(*Server, Listener, *Client) error
}

// OnClientOpenedHook represents a hook which receives the event indicating that a new Client was opened.
type OnClientOpenedHook interface {
	// OnClientOpened is called by the Server when a new Client was opened.
	OnClientOpened(*Server, Listener, *Client)
}

// OnClientCloseHook represents a hook which receives the event indicating that the Client is being closed.
type OnClientCloseHook interface {
	// OnClientClose is called by the Server when the Client is being closed.
	OnClientClose(*Client)
}

// OnClientClosedHook represents a hook which receives the event indicating that the Client was closed.
type OnClientClosedHook interface {
	// OnClientClosed is called by the Server when the Client was closed.
	OnClientClosed(*Client)
}

var hooksRegistries = map[hookType]func(*hooks, Hook, hookType){
	onStartHook:             registerHook[OnStartHook],
	onStopHook:              registerHook[OnStopHook],
	onServerStartHook:       registerHook[OnServerStartHook],
	onServerStartFailedHook: registerHook[OnServerStartFailedHook],
	onServerStartedHook:     registerHook[OnServerStartedHook],
	onServerStopHook:        registerHook[OnServerStopHook],
	onServerStoppedHook:     registerHook[OnServerStoppedHook],
	onClientOpenHook:        registerHook[OnClientOpenHook],
	onClientOpenedHook:      registerHook[OnClientOpenedHook],
	onClientCloseHook:       registerHook[OnClientCloseHook],
	onClientClosedHook:      registerHook[OnClientClosedHook],
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
			return fmt.Errorf("%s.OnStart: %w", hook.Name(), err)
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
			return fmt.Errorf("%s.OnServerStart: %w", hook.Name(), err)
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

func (h *hooks) onClientOpen(s *Server, l Listener, c *Client) error {
	if !h.hasHook(onClientOpenHook) {
		return nil
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.hooks[onClientOpenHook] {
		hk := hook.(OnClientOpenHook)

		err := hk.OnClientOpen(s, l, c)
		if err != nil {
			return fmt.Errorf("%s.OnClientOpen: %w", hook.Name(), err)
		}
	}

	return nil
}

func (h *hooks) onClientOpened(s *Server, l Listener, c *Client) {
	if !h.hasHook(onClientOpenedHook) {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.hooks[onClientOpenedHook] {
		hk := hook.(OnClientOpenedHook)
		hk.OnClientOpened(s, l, c)
	}
}

func (h *hooks) onClientClose(c *Client) {
	if !h.hasHook(onClientCloseHook) {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.hooks[onClientCloseHook] {
		hk := hook.(OnClientCloseHook)
		hk.OnClientClose(c)
	}
}

func (h *hooks) onClientClosed(c *Client) {
	if !h.hasHook(onClientClosedHook) {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, hook := range h.hooks[onClientClosedHook] {
		hk := hook.(OnClientClosedHook)
		hk.OnClientClosed(c)
	}
}
