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

func BenchmarkHooksOnConnectionOpen(b *testing.B) {
	b.Run("No Hook", func(b *testing.B) {
		h := newHooks()

		for i := 0; i < b.N; i++ {
			_ = h.onConnectionOpen(nil)
		}
	})

	b.Run("With Hook", func(b *testing.B) {
		h := newHooks()
		_ = h.add(&mockOnConnectionOpenHook{})

		for i := 0; i < b.N; i++ {
			_ = h.onConnectionOpen(nil)
		}
	})
}

func BenchmarkHooksOnClientOpened(b *testing.B) {
	b.Run("No Hook", func(b *testing.B) {
		h := newHooks()

		for i := 0; i < b.N; i++ {
			h.onClientOpened(nil)
		}
	})

	b.Run("With Hook", func(b *testing.B) {
		h := newHooks()
		_ = h.add(&mockOnClientOpenedHook{})

		for i := 0; i < b.N; i++ {
			h.onClientOpened(nil)
		}
	})
}

func BenchmarkHooksOnClientClose(b *testing.B) {
	b.Run("No Hook", func(b *testing.B) {
		h := newHooks()

		for i := 0; i < b.N; i++ {
			h.onClientClose(nil, nil)
		}
	})

	b.Run("With Hook", func(b *testing.B) {
		h := newHooks()
		_ = h.add(&mockOnClientCloseHook{})

		for i := 0; i < b.N; i++ {
			h.onClientClose(nil, nil)
		}
	})
}

func BenchmarkHooksOnConnectionClosed(b *testing.B) {
	b.Run("No Hook", func(b *testing.B) {
		h := newHooks()

		for i := 0; i < b.N; i++ {
			h.onConnectionClosed(nil, nil)
		}
	})

	b.Run("With Hook", func(b *testing.B) {
		h := newHooks()
		_ = h.add(&mockOnConnectionClosedHook{})

		for i := 0; i < b.N; i++ {
			h.onConnectionClosed(nil, nil)
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
		_ = h.add(&mockOnPacketReceiveHook{})

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
		_ = h.add(&mockOnPacketReceivedHook{})

		for i := 0; i < b.N; i++ {
			_ = h.onPacketReceived(nil, nil)
		}
	})
}

func BenchmarkHooksOnPacketSend(b *testing.B) {
	b.Run("No Hook", func(b *testing.B) {
		h := newHooks()

		for i := 0; i < b.N; i++ {
			_ = h.onPacketSend(nil, nil)
		}
	})

	b.Run("With Hook", func(b *testing.B) {
		h := newHooks()
		_ = h.add(&mockOnPacketSendHook{})

		for i := 0; i < b.N; i++ {
			_ = h.onPacketSend(nil, nil)
		}
	})
}

func BenchmarkHooksOnPacketSent(b *testing.B) {
	b.Run("No Hook", func(b *testing.B) {
		h := newHooks()

		for i := 0; i < b.N; i++ {
			h.onPacketSent(nil, nil)
		}
	})

	b.Run("With Hook", func(b *testing.B) {
		h := newHooks()
		_ = h.add(&mockOnPacketSentHook{})

		for i := 0; i < b.N; i++ {
			h.onPacketSent(nil, nil)
		}
	})
}
