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
	"testing"
)

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

func (h *hookSpy) OnPacketReceive(_ *Client) error {
	return nil
}

func (h *hookSpy) OnPacketReceived(_ *Client, _ Packet) error {
	return nil
}

func BenchmarkHooksOnConnectionOpen(b *testing.B) {
	b.Run("No Hook", func(b *testing.B) {
		h := newHooks()

		for i := 0; i < b.N; i++ {
			_ = h.onConnectionOpen(nil, nil)
		}
	})

	b.Run("With Hook", func(b *testing.B) {
		h := newHooks()
		_ = h.add(&hookSpy{})

		for i := 0; i < b.N; i++ {
			_ = h.onConnectionOpen(nil, nil)
		}
	})
}

func BenchmarkHooksOnConnectionOpened(b *testing.B) {
	b.Run("No Hook", func(b *testing.B) {
		h := newHooks()

		for i := 0; i < b.N; i++ {
			h.onConnectionOpened(nil, nil)
		}
	})

	b.Run("With Hook", func(b *testing.B) {
		h := newHooks()
		_ = h.add(&hookSpy{})

		for i := 0; i < b.N; i++ {
			h.onConnectionOpened(nil, nil)
		}
	})
}

func BenchmarkHooksOnConnectionClose(b *testing.B) {
	b.Run("No Hook", func(b *testing.B) {
		h := newHooks()

		for i := 0; i < b.N; i++ {
			h.onConnectionClose(nil, nil, nil)
		}
	})

	b.Run("With Hook", func(b *testing.B) {
		h := newHooks()
		_ = h.add(&hookSpy{})

		for i := 0; i < b.N; i++ {
			h.onConnectionClose(nil, nil, nil)
		}
	})
}

func BenchmarkHooksOnConnectionClosed(b *testing.B) {
	b.Run("No Hook", func(b *testing.B) {
		h := newHooks()

		for i := 0; i < b.N; i++ {
			h.onConnectionClosed(nil, nil, nil)
		}
	})

	b.Run("With Hook", func(b *testing.B) {
		h := newHooks()
		_ = h.add(&hookSpy{})

		for i := 0; i < b.N; i++ {
			h.onConnectionClosed(nil, nil, nil)
		}
	})
}

func BenchmarkHooksOnPacketReceive(b *testing.B) {
	b.Run("No Hook", func(b *testing.B) {
		h := newHooks()

		for i := 0; i < b.N; i++ {
			_ = h.onPacketReceive(nil)
		}
	})

	b.Run("With Hook", func(b *testing.B) {
		h := newHooks()
		_ = h.add(&hookSpy{})

		for i := 0; i < b.N; i++ {
			_ = h.onPacketReceive(nil)
		}
	})
}

func BenchmarkHooksOnPacketReceived(b *testing.B) {
	b.Run("No Hook", func(b *testing.B) {
		h := newHooks()

		for i := 0; i < b.N; i++ {
			_ = h.onPacketReceived(nil, nil)
		}
	})

	b.Run("With Hook", func(b *testing.B) {
		h := newHooks()
		_ = h.add(&hookSpy{})

		for i := 0; i < b.N; i++ {
			_ = h.onPacketReceived(nil, nil)
		}
	})
}
