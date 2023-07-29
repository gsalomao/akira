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

func (m *mockOnStartHook) OnStart() error {
	m.called()
	if m.cb != nil {
		return m.cb()
	}
	return nil
}

type mockOnStopHook struct {
	mockHook
	cb func()
}

func (m *mockOnStopHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnStopHook) OnStop() {
	m.called()
	if m.cb != nil {
		m.cb()
	}
}

type mockOnServerStartHook struct {
	mockHook
	cb func() error
}

func (m *mockOnServerStartHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnServerStartHook) OnServerStart() error {
	m.called()
	if m.cb != nil {
		return m.cb()
	}
	return nil
}

type mockOnServerStartFailedHook struct {
	mockHook
	cb func(err error)
}

func (m *mockOnServerStartFailedHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnServerStartFailedHook) OnServerStartFailed(err error) {
	m.called()
	if m.cb != nil {
		m.cb(err)
	}
}

type mockOnServerStartedHook struct {
	mockHook
	cb func()
}

func (m *mockOnServerStartedHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnServerStartedHook) OnServerStarted() {
	m.called()
	if m.cb != nil {
		m.cb()
	}
}

type mockOnServerStopHook struct {
	mockHook
	cb func()
}

func (m *mockOnServerStopHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnServerStopHook) OnServerStop() {
	m.called()
	if m.cb != nil {
		m.cb()
	}
}

type mockOnServerStoppedHook struct {
	mockHook
	cb func()
}

func (m *mockOnServerStoppedHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnServerStoppedHook) OnServerStopped() {
	m.called()
	if m.cb != nil {
		m.cb()
	}
}

type mockOnConnectionOpenHook struct {
	mockHook
	cb func(c *Connection) error
}

func (m *mockOnConnectionOpenHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnConnectionOpenHook) OnConnectionOpen(c *Connection) error {
	m.called()
	if m.cb != nil {
		return m.cb(c)
	}
	return nil
}

type mockOnClientOpenedHook struct {
	mockHook
	cb func(c *Client)
}

func (m *mockOnClientOpenedHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnClientOpenedHook) OnClientOpened(c *Client) {
	m.called()
	if m.cb != nil {
		m.cb(c)
	}
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
	if m.cb != nil {
		m.cb(c, err)
	}
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
	if m.cb != nil {
		m.cb(c, err)
	}
}

type mockOnPacketReceiveHook struct {
	mockHook
	cb func(c *Client) error
}

func (m *mockOnPacketReceiveHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnPacketReceiveHook) OnPacketReceive(c *Client) error {
	m.called()
	if m.cb != nil {
		return m.cb(c)
	}
	return nil
}

type mockOnPacketReceiveFailedHook struct {
	mockHook
	cb func(c *Client, err error)
}

func (m *mockOnPacketReceiveFailedHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnPacketReceiveFailedHook) OnPacketReceiveFailed(c *Client, err error) {
	m.called()
	if m.cb != nil {
		m.cb(c, err)
	}
}

type mockOnPacketReceivedHook struct {
	mockHook
	cb func(c *Client, p Packet) error
}

func (m *mockOnPacketReceivedHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnPacketReceivedHook) OnPacketReceived(c *Client, p Packet) error {
	m.called()
	if m.cb != nil {
		return m.cb(c, p)
	}
	return nil
}

type mockOnPacketSendHook struct {
	mockHook
	cb func(c *Client, p Packet) error
}

func (m *mockOnPacketSendHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnPacketSendHook) OnPacketSend(c *Client, p Packet) error {
	m.called()
	if m.cb != nil {
		return m.cb(c, p)
	}
	return nil
}

type mockOnPacketSendFailedHook struct {
	mockHook
	cb func(c *Client, p Packet, err error)
}

func (m *mockOnPacketSendFailedHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnPacketSendFailedHook) OnPacketSendFailed(c *Client, p Packet, err error) {
	m.called()
	if m.cb != nil {
		m.cb(c, p, err)
	}
}

type mockOnPacketSentHook struct {
	mockHook
	cb func(c *Client, p Packet)
}

func (m *mockOnPacketSentHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnPacketSentHook) OnPacketSent(c *Client, p Packet) {
	m.called()
	if m.cb != nil {
		m.cb(c, p)
	}
}

type mockOnConnectHook struct {
	mockHook
	cb func(c *Client, p *packet.Connect) error
}

func (m *mockOnConnectHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnConnectHook) OnConnect(c *Client, p *packet.Connect) error {
	m.called()
	if m.cb != nil {
		return m.cb(c, p)
	}
	return nil
}

type mockOnConnectFailedHook struct {
	mockHook
	cb func(c *Client, p *packet.Connect, err error)
}

func (m *mockOnConnectFailedHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnConnectFailedHook) OnConnectFailed(c *Client, p *packet.Connect, err error) {
	m.called()
	if m.cb != nil {
		m.cb(c, p, err)
	}
}

type mockOnConnectedHook struct {
	mockHook
	cb func(c *Client)
}

func (m *mockOnConnectedHook) Name() string {
	return reflect.TypeOf(m).String()
}

func (m *mockOnConnectedHook) OnConnected(c *Client) {
	m.called()
	if m.cb != nil {
		m.cb(c)
	}
}
