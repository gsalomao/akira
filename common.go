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
	"time"

	"github.com/gsalomao/akira/packet"
)

func newIfNotExists[T any](v *T) *T {
	if v != nil {
		return v
	}
	return new(T)
}

func maxIfExceeded[T ~int | ~int8 | ~uint8 | ~int16 | ~uint16 | ~int32 | ~uint32 | ~int64 | ~uint64](val, max T) T {
	if max > 0 && val > max {
		return max
	}
	return val
}

func isClientIDValid(version packet.Version, clientIDSize int, config *Config) bool {
	if clientIDSize == 0 {
		return false
	}
	if version == packet.MQTT31 && clientIDSize > 23 {
		return false
	}
	if clientIDSize > config.MaxClientIDSize {
		return false
	}
	return true
}

func isKeepAliveValid(version packet.Version, keepAlive, maxKeepAlive uint16) bool {
	// For V5.0 clients, the Keep Alive sent in the CONNECT Packet can be overwritten in the CONNACK Packet.
	if version == packet.MQTT50 || maxKeepAlive == 0 {
		return true
	}
	if keepAlive > 0 && keepAlive <= maxKeepAlive {
		return true
	}
	return false
}

func sessionKeepAlive(keepAlive, maxKeepAlive uint16) uint16 {
	if maxKeepAlive > 0 && (keepAlive == 0 || keepAlive > maxKeepAlive) {
		return maxKeepAlive
	}
	return keepAlive
}

func isPersistentSession(c *Client, cleanStart bool) bool {
	if (c.Connection.Version != packet.MQTT50 && !cleanStart) || sessionExpiryInterval(&c.Session) > 0 {
		return true
	}
	return false
}

func sessionExpiryInterval(s *Session) time.Duration {
	if !s.Properties.Has(packet.PropertySessionExpiryInterval) {
		return 0
	}
	return s.Properties.SessionExpiryInterval
}

func newLastWill(connect *packet.Connect) *LastWill {
	if !connect.Flags.WillFlag() {
		return nil
	}

	will := &LastWill{
		Topic:   connect.WillTopic,
		Payload: connect.WillPayload,
		QoS:     connect.Flags.WillQoS(),
		Retain:  connect.Flags.WillRetain(),
	}

	props := connect.WillProperties
	will.Properties = &packet.WillProperties{Flags: props.Flags}

	if props.Has(packet.PropertyWillDelayInterval) {
		will.Properties.WillDelayInterval = props.WillDelayInterval
		will.Properties.Set(packet.PropertyWillDelayInterval)
	}
	if props.Has(packet.PropertyPayloadFormatIndicator) {
		will.Properties.PayloadFormatIndicator = props.PayloadFormatIndicator
		will.Properties.Set(packet.PropertyPayloadFormatIndicator)
	}
	if props.Has(packet.PropertyMessageExpiryInterval) {
		will.Properties.MessageExpiryInterval = props.MessageExpiryInterval
		will.Properties.Set(packet.PropertyMessageExpiryInterval)
	}
	if props.Has(packet.PropertyContentType) {
		will.Properties.ContentType = props.ContentType
		will.Properties.Set(packet.PropertyContentType)
	}
	if props.Has(packet.PropertyResponseTopic) {
		will.Properties.ResponseTopic = props.ResponseTopic
		will.Properties.Set(packet.PropertyResponseTopic)
	}
	if props.Has(packet.PropertyCorrelationData) {
		will.Properties.CorrelationData = props.CorrelationData
		will.Properties.Set(packet.PropertyCorrelationData)
	}
	if props.Has(packet.PropertyUserProperty) {
		will.Properties.UserProperties = props.UserProperties
		will.Properties.Set(packet.PropertyUserProperty)
	}

	return will
}

func newSessionProperties(props *packet.ConnectProperties, conf *Config) *SessionProperties {
	if props == nil {
		return nil
	}

	p := &SessionProperties{}

	if props.Has(packet.PropertySessionExpiryInterval) {
		interval := maxIfExceeded(props.SessionExpiryInterval, conf.MaxSessionExpiryIntervalSec)
		p.SessionExpiryInterval = time.Duration(interval) * time.Second
		p.Set(packet.PropertySessionExpiryInterval)
	}
	if props.Has(packet.PropertyMaximumPacketSize) {
		size := props.MaximumPacketSize
		p.MaximumPacketSize = maxIfExceeded(size, conf.MaxPacketSize)
		p.Set(packet.PropertyMaximumPacketSize)
	}
	if props.Has(packet.PropertyReceiveMaximum) {
		maximum := props.ReceiveMaximum
		p.ReceiveMaximum = maxIfExceeded(maximum, conf.MaxInflightMessages)
		p.Set(packet.PropertyReceiveMaximum)
	}
	if props.Has(packet.PropertyTopicAliasMaximum) {
		maximum := props.TopicAliasMaximum
		p.TopicAliasMaximum = maxIfExceeded(maximum, conf.TopicAliasMax)
		p.Set(packet.PropertyTopicAliasMaximum)
	}
	if props.Has(packet.PropertyRequestResponseInfo) {
		p.RequestResponseInfo = props.RequestResponseInfo
		p.Set(packet.PropertyRequestResponseInfo)
	}
	if props.Has(packet.PropertyRequestProblemInfo) {
		p.RequestProblemInfo = props.RequestProblemInfo
		p.Set(packet.PropertyRequestProblemInfo)
	}
	if props.Has(packet.PropertyUserProperty) {
		p.UserProperties = props.UserProperties
		p.Set(packet.PropertyUserProperty)
	}

	return p
}

func newConnAckProperties(c *Client, conf *Config) *packet.ConnAckProperties {
	var props *packet.ConnAckProperties
	sProps := c.Session.Properties

	if sProps.Has(packet.PropertySessionExpiryInterval) {
		if c.sessionExpiryInterval != sProps.SessionExpiryInterval {
			props = newIfNotExists(props)
			props.Set(packet.PropertySessionExpiryInterval)
			props.SessionExpiryInterval = uint32(c.Session.Properties.SessionExpiryInterval.Seconds())
		}
	}
	if c.keepAlive != c.Connection.KeepAlive {
		props = newIfNotExists(props)
		props.Set(packet.PropertyServerKeepAlive)
		props.ServerKeepAlive = uint16(c.Connection.KeepAlive.Seconds())
	}
	if conf.MaxInflightMessages > 0 {
		props = newIfNotExists(props)
		props.Set(packet.PropertyReceiveMaximum)
		props.ReceiveMaximum = conf.MaxInflightMessages
	}
	if conf.MaxPacketSize > 0 {
		props = newIfNotExists(props)
		props.Set(packet.PropertyMaximumPacketSize)
		props.MaximumPacketSize = conf.MaxPacketSize
	}
	if conf.MaxQoS < byte(packet.QoS2) {
		props = newIfNotExists(props)
		props.Set(packet.PropertyMaximumQoS)
		props.MaximumQoS = conf.MaxQoS
	}
	if conf.TopicAliasMax > 0 {
		props = newIfNotExists(props)
		props.Set(packet.PropertyTopicAliasMaximum)
		props.TopicAliasMaximum = conf.TopicAliasMax
	}
	if !conf.RetainAvailable {
		props = newIfNotExists(props)
		props.Set(packet.PropertyRetainAvailable)
		props.RetainAvailable = false
	}
	if !conf.WildcardSubscriptionAvailable {
		props = newIfNotExists(props)
		props.Set(packet.PropertyWildcardSubscriptionAvailable)
		props.WildcardSubscriptionAvailable = false
	}
	if !conf.SubscriptionIDAvailable {
		props = newIfNotExists(props)
		props.Set(packet.PropertySubscriptionIDAvailable)
		props.SubscriptionIDAvailable = false
	}
	if !conf.SharedSubscriptionAvailable {
		props = newIfNotExists(props)
		props.Set(packet.PropertySharedSubscriptionAvailable)
		props.SharedSubscriptionAvailable = false
	}

	return props
}

func setConnAckPropertiesWithAuth(props *packet.ConnAckProperties, connack *packet.ConnAck) *packet.ConnAckProperties {
	if props == nil || connack == nil || connack.Properties == nil {
		return props
	}

	if connack.Properties.Has(packet.PropertyAuthenticationData) {
		props = newIfNotExists(props)
		props.Set(packet.PropertyAuthenticationData)
		props.AuthenticationData = connack.Properties.AuthenticationData
	}
	if connack.Properties.Has(packet.PropertyReasonString) {
		props = newIfNotExists(props)
		props.Set(packet.PropertyReasonString)
		props.ReasonString = connack.Properties.ReasonString
	}

	return props
}
