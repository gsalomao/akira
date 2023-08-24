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
	"context"
	"sync/atomic"
)

type mockEnhancedAuth struct {
	name   string
	cb     func(*Client, Packet) (PacketEncodable, error)
	_calls atomic.Int32
}

func (m *mockEnhancedAuth) Name() string {
	if m.name == "" {
		return "mock"
	}
	return m.name
}

func (m *mockEnhancedAuth) Authenticate(_ context.Context, c *Client, p Packet) (PacketEncodable, error) {
	m._calls.Add(1)
	if m.cb == nil {
		return nil, nil
	}
	return m.cb(c, p)
}

func (m *mockEnhancedAuth) calls() int {
	return int(m._calls.Load())
}
