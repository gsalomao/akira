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
	"errors"
	"reflect"
	"sync/atomic"

	"github.com/gsalomao/akira/packet"
)

var errHookFailed = errors.New("hook failed")

type mockHook struct {
	_calls atomic.Int32
}

func (m *mockHook) called() {
	m._calls.Add(1)
}

func (m *mockHook) calls() int {
	return int(m._calls.Load())
}

type mockOnStartHook struct {
	mockHook
	cb func() error
}

func (m *mockOnStartHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnStartHook) OnStart(_ context.Context) error {
	m.called()
	if m.cb == nil {
		return nil
	}
	return m.cb()
}

type mockOnStopHook struct {
	mockHook
	cb func()
}

func (m *mockOnStopHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnStopHook) OnStop(_ context.Context) {
	m.called()
	if m.cb == nil {
		return
	}
	m.cb()
}

type mockOnStartStopHook struct {
	mockHook
	startCB func() error
	stopCB  func()
}

func (m *mockOnStartStopHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnStartStopHook) OnStart(_ context.Context) error {
	m.called()
	if m.startCB == nil {
		return nil
	}
	return m.startCB()
}

func (m *mockOnStartStopHook) OnStop(_ context.Context) {
	m.called()
	if m.stopCB == nil {
		return
	}
	m.stopCB()
}

type mockOnServerStartHook struct {
	mockHook
	cb func() error
}

func (m *mockOnServerStartHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnServerStartHook) OnServerStart(_ context.Context) error {
	m.called()
	if m.cb == nil {
		return nil
	}
	return m.cb()
}

type mockOnServerStartFailedHook struct {
	mockHook
	cb func(err error)
}

func (m *mockOnServerStartFailedHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnServerStartFailedHook) OnServerStartFailed(_ context.Context, err error) {
	m.called()
	if m.cb == nil {
		return
	}
	m.cb(err)
}

type mockOnServerStartedHook struct {
	mockHook
	cb func()
}

func (m *mockOnServerStartedHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnServerStartedHook) OnServerStarted(_ context.Context) {
	m.called()
	if m.cb == nil {
		return
	}
	m.cb()
}

type mockOnServerStopHook struct {
	mockHook
	cb func()
}

func (m *mockOnServerStopHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnServerStopHook) OnServerStop(_ context.Context) {
	m.called()
	if m.cb == nil {
		return
	}
	m.cb()
}

type mockOnServerStoppedHook struct {
	mockHook
	cb func()
}

func (m *mockOnServerStoppedHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnServerStoppedHook) OnServerStopped(_ context.Context) {
	m.called()
	if m.cb == nil {
		return
	}
	m.cb()
}

type mockOnConnectionOpenHook struct {
	mockHook
	cb func(c *Connection) error
}

func (m *mockOnConnectionOpenHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnConnectionOpenHook) OnConnectionOpen(_ context.Context, c *Connection) error {
	m.called()
	if m.cb == nil {
		return nil
	}
	return m.cb(c)
}

type mockOnClientOpenedHook struct {
	mockHook
	cb func(c *Client)
}

func (m *mockOnClientOpenedHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnClientOpenedHook) OnClientOpened(_ context.Context, c *Client) {
	m.called()
	if m.cb == nil {
		return
	}
	m.cb(c)
}

type mockOnClientCloseHook struct {
	mockHook
	cb func(c *Client, err error)
}

func (m *mockOnClientCloseHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnClientCloseHook) OnClientClose(c *Client, err error) {
	m.called()
	if m.cb == nil {
		return
	}
	m.cb(c, err)
}

type mockOnConnectionClosedHook struct {
	mockHook
	cb func(c *Connection, err error)
}

func (m *mockOnConnectionClosedHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnConnectionClosedHook) OnConnectionClosed(c *Connection, err error) {
	m.called()
	if m.cb == nil {
		return
	}
	m.cb(c, err)
}

type mockOnReceivePacketHook struct {
	mockHook
	cb func(c *Client) error
}

func (m *mockOnReceivePacketHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnReceivePacketHook) OnReceivePacket(_ context.Context, c *Client) error {
	m.called()
	if m.cb == nil {
		return nil
	}
	return m.cb(c)
}

type mockOnPacketReceiveHook struct {
	mockHook
	cb func(c *Client, h packet.FixedHeader) error
}

func (m *mockOnPacketReceiveHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnPacketReceiveHook) OnPacketReceive(_ context.Context, c *Client, h packet.FixedHeader) error {
	m.called()
	if m.cb == nil {
		return nil
	}
	return m.cb(c, h)
}

type mockOnPacketReceiveFailedHook struct {
	mockHook
	cb func(c *Client, err error)
}

func (m *mockOnPacketReceiveFailedHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnPacketReceiveFailedHook) OnPacketReceiveFailed(_ context.Context, c *Client, err error) {
	m.called()
	if m.cb == nil {
		return
	}
	m.cb(c, err)
}

type mockOnPacketReceivedHook struct {
	mockHook
	cb func(c *Client, p Packet) error
}

func (m *mockOnPacketReceivedHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnPacketReceivedHook) OnPacketReceived(_ context.Context, c *Client, p Packet) error {
	m.called()
	if m.cb == nil {
		return nil
	}
	return m.cb(c, p)
}

type mockOnPacketSendHook struct {
	mockHook
	cb func(c *Client, p Packet) error
}

func (m *mockOnPacketSendHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnPacketSendHook) OnPacketSend(_ context.Context, c *Client, p Packet) error {
	m.called()
	if m.cb == nil {
		return nil
	}
	return m.cb(c, p)
}

type mockOnPacketSendFailedHook struct {
	mockHook
	cb func(c *Client, p Packet, err error)
}

func (m *mockOnPacketSendFailedHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnPacketSendFailedHook) OnPacketSendFailed(_ context.Context, c *Client, p Packet, err error) {
	m.called()
	if m.cb == nil {
		return
	}
	m.cb(c, p, err)
}

type mockOnPacketSentHook struct {
	mockHook
	cb func(c *Client, p Packet)
}

func (m *mockOnPacketSentHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnPacketSentHook) OnPacketSent(_ context.Context, c *Client, p Packet) {
	m.called()
	if m.cb == nil {
		return
	}
	m.cb(c, p)
}

type mockOnConnectHook struct {
	mockHook
	cb func(c *Client, p *packet.Connect) error
}

func (m *mockOnConnectHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnConnectHook) OnConnectPacket(_ context.Context, c *Client, p *packet.Connect) error {
	m.called()
	if m.cb == nil {
		return nil
	}
	return m.cb(c, p)
}

type mockOnConnectFailedHook struct {
	mockHook
	cb func(c *Client, err error)
}

func (m *mockOnConnectFailedHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnConnectFailedHook) OnConnectFailed(_ context.Context, c *Client, err error) {
	m.called()
	if m.cb == nil {
		return
	}
	m.cb(c, err)
}

type mockOnConnectedHook struct {
	mockHook
	cb func(c *Client)
}

func (m *mockOnConnectedHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnConnectedHook) OnConnected(_ context.Context, c *Client) {
	m.called()
	if m.cb == nil {
		return
	}
	m.cb(c)
}

type mockOnAuthPacketHook struct {
	mockHook
	cb func(c *Client, p *packet.Auth) error
}

func (m *mockOnAuthPacketHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnAuthPacketHook) OnAuthPacket(_ context.Context, c *Client, p *packet.Auth) error {
	m.called()
	if m.cb == nil {
		return nil
	}
	return m.cb(c, p)
}
