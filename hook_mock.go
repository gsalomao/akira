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
	"github.com/stretchr/testify/mock"
)

type mockHook struct {
	mock.Mock
}

func newMockHook() *mockHook {
	h := mockHook{}
	return &h
}

func (h *mockHook) Name() string {
	return "mock"
}

func (h *mockHook) OnStart() error {
	args := h.Called()
	if len(args) > 0 {
		return args.Error(0)
	}
	return nil
}

func (h *mockHook) OnStop() {
	h.Called()
}

func (h *mockHook) OnServerStart(s *Server) error {
	args := h.Called(s)
	if len(args) > 0 {
		return args.Error(0)
	}
	return nil
}

func (h *mockHook) OnServerStartFailed(s *Server, err error) {
	h.Called(s, err)
}

func (h *mockHook) OnServerStarted(s *Server) {
	h.Called(s)
}

func (h *mockHook) OnServerStop(s *Server) {
	h.Called(s)
}

func (h *mockHook) OnServerStopped(s *Server) {
	h.Called(s)
}

func (h *mockHook) OnClientOpen(s *Server, l Listener, c *Client) error {
	args := h.Called(s, l, c)
	if len(args) > 0 {
		return args.Error(0)
	}
	return nil
}

func (h *mockHook) OnClientOpened(s *Server, l Listener, c *Client) {
	h.Called(s, l, c)
}

func (h *mockHook) OnClientClose(c *Client) {
	h.Called(c)
}

func (h *mockHook) OnClientClosed(c *Client) {
	h.Called(c)
}
