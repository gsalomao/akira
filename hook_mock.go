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
	"github.com/stretchr/testify/mock"
)

type hookMock struct {
	mock.Mock
}

func newHookMock() *hookMock {
	h := hookMock{}
	return &h
}

func (h *hookMock) Name() string {
	return "hook-mock"
}

func (h *hookMock) OnStart(s *Server) error {
	args := h.Called(s)
	if len(args) > 0 {
		return args.Error(0)
	}
	return nil
}

func (h *hookMock) OnStop(s *Server) {
	h.Called(s)
}

func (h *hookMock) OnServerStart(s *Server) error {
	args := h.Called(s)
	if len(args) > 0 {
		return args.Error(0)
	}
	return nil
}

func (h *hookMock) OnServerStartFailed(s *Server, err error) {
	h.Called(s, err)
}

func (h *hookMock) OnServerStarted(s *Server) {
	h.Called(s)
}

func (h *hookMock) OnServerStop(s *Server) {
	h.Called(s)
}

func (h *hookMock) OnServerStopped(s *Server) {
	h.Called(s)
}

func (h *hookMock) OnConnectionOpen(s *Server, l Listener) error {
	args := h.Called(s, l)
	if len(args) > 0 {
		return args.Error(0)
	}
	return nil
}

func (h *hookMock) OnConnectionOpened(s *Server, l Listener) {
	h.Called(s, l)
}

func (h *hookMock) OnConnectionClose(s *Server, l Listener, err error) {
	h.Called(s, l, err)
}

func (h *hookMock) OnConnectionClosed(s *Server, l Listener, err error) {
	h.Called(s, l, err)
}

func (h *hookMock) OnPacketReceive(s *Server, c *Client) error {
	args := h.Called(s, c)
	if len(args) > 0 {
		return args.Error(0)
	}
	return nil
}

func (h *hookMock) OnPacketReceiveError(s *Server, c *Client, err error) error {
	args := h.Called(s, c, err)
	if len(args) > 0 {
		return args.Error(0)
	}
	return nil
}

func (h *hookMock) OnPacketReceived(s *Server, c *Client, p Packet) error {
	args := h.Called(s, c, p)
	if len(args) > 0 {
		return args.Error(0)
	}
	return nil
}

type hookSpy struct{}

func (h *hookSpy) Name() string {
	return "hook-spy"
}

func (h *hookSpy) OnConnectionOpen(_ *Server, _ Listener) error {
	return nil
}

func (h *hookSpy) OnConnectionOpened(_ *Server, _ Listener) {
}

func (h *hookSpy) OnConnectionClose(_ *Server, _ Listener, _ error) {
}

func (h *hookSpy) OnConnectionClosed(_ *Server, _ Listener, _ error) {
}

func (h *hookSpy) OnPacketReceive(_ *Server, _ *Client) error {
	return nil
}

func (h *hookSpy) OnPacketReceived(_ *Server, _ *Client, _ Packet) error {
	return nil
}
